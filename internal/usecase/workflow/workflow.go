package workflow

import (
	"context"
	"fmt"
	"github.com/ardianferdianto/reconciliation-service/internal/domain"
	enum_parser "github.com/ardianferdianto/reconciliation-service/internal/domain/enum/parser"
	enum_status "github.com/ardianferdianto/reconciliation-service/internal/domain/enum/status"
	"github.com/ardianferdianto/reconciliation-service/internal/repository"
	"github.com/ardianferdianto/reconciliation-service/internal/usecase/ingestion"
	"github.com/ardianferdianto/reconciliation-service/internal/usecase/reconcile"
	"github.com/google/uuid"
	"sync"
	"time"
)

type IUseCase interface {
	StartWorkflow(ctx context.Context, sysFile string, bankFiles []string, startDate, endDate time.Time) (string, error)
	OnSystemIngestionComplete(ctx context.Context, workflowID, jobID string, success bool) error
	OnBankIngestionComplete(ctx context.Context, workflowID, jobID string, success bool) error
	OnReconciliationComplete(ctx context.Context, workflowID, jobID string, success bool) error
	GetWorkflowSummary(ctx context.Context, workflowID string) (*domain.Workflow, error)
}

type workflowUseCase struct {
	workflowRepo repository.WorkflowRepository
	ingestionUC  ingestion.IUseCase
	reconcileUC  reconcile.IUseCase
	mu           sync.Mutex // Mutex to handle concurrent updates
	wg           sync.WaitGroup
}

func NewWorkflowUseCase(
	wfRepo repository.WorkflowRepository,
	ingestionUC ingestion.IUseCase,
	reconcileUC reconcile.IUseCase,
) IUseCase {
	return &workflowUseCase{
		workflowRepo: wfRepo,
		ingestionUC:  ingestionUC,
		reconcileUC:  reconcileUC,
	}
}

func (uc *workflowUseCase) GetWorkflowSummary(ctx context.Context, workflowID string) (*domain.Workflow, error) {
	wf, err := uc.workflowRepo.GetWorkflow(ctx, workflowID)
	if err != nil {
		return nil, fmt.Errorf("error retrieving workflow: %w", err)
	}
	return &wf, nil
}

func (uc *workflowUseCase) StartWorkflow(
	ctx context.Context,
	sysFile string,
	bankFiles []string,
	startDate, endDate time.Time,
) (string, error) {

	workflowID := uuid.New().String()

	wf := domain.Workflow{
		WorkflowID: workflowID,
		Status:     enum_status.IN_PROGRESS.String(),
		StartDate:  startDate,
		EndDate:    endDate,
	}

	if err := uc.workflowRepo.CreateWorkflow(ctx, wf); err != nil {
		return "", fmt.Errorf("failed to create workflow: %w", err)
	}

	sysObjInfo, err := uc.ingestionUC.FetchFileMetadata(ctx, sysFile)
	if err != nil {
		wf.Status = enum_status.FAILED.String()
		uc.workflowRepo.UpdateWorkflow(ctx, wf)
		return "", fmt.Errorf("failed to fetch system transaction file metadata: %w", err)
	}

	sysJob := &domain.IngestionJob{
		JobID:    uuid.New().String(),
		FileType: enum_parser.SYSTEM_TRX,
		FileName: sysObjInfo.Key,
		Status:   enum_status.IN_PROGRESS.String(),
	}
	if err := uc.ingestionUC.CreateIngestionJob(ctx, sysJob); err != nil {
		wf.Status = enum_status.FAILED.String()
		uc.workflowRepo.UpdateWorkflow(ctx, wf)
		return "", fmt.Errorf("failed to create system ingestion job: %w", err)
	}

	uc.wg.Add(1)
	go func() {
		defer uc.wg.Done()
		err := uc.ingestionUC.ProcessIngestionJob(ctx, sysJob)
		success := err == nil
		uc.OnSystemIngestionComplete(ctx, workflowID, sysJob.JobID, success)
	}()

	for _, bankFile := range bankFiles {
		bankObjInfo, err := uc.ingestionUC.FetchFileMetadata(ctx, bankFile)
		if err != nil {
			wf.Status = enum_status.FAILED.String()
			uc.workflowRepo.UpdateWorkflow(ctx, wf)
			return "", fmt.Errorf("failed to fetch bank statement file metadata: %w", err)
		}

		bankJob := &domain.IngestionJob{
			JobID:    uuid.New().String(),
			FileType: enum_parser.BANK_STATEMENT,
			FileName: bankObjInfo.Key,
			Status:   enum_status.IN_PROGRESS.String(),
		}
		if err := uc.ingestionUC.CreateIngestionJob(ctx, bankJob); err != nil {
			wf.Status = enum_status.FAILED.String()
			uc.workflowRepo.UpdateWorkflow(ctx, wf)
			return "", fmt.Errorf("failed to create bank ingestion job: %w", err)
		}
		uc.wg.Add(1)
		go func(job *domain.IngestionJob) {
			defer uc.wg.Done()
			err := uc.ingestionUC.ProcessIngestionJob(ctx, job)
			success := err == nil
			uc.OnBankIngestionComplete(ctx, workflowID, job.JobID, success)
		}(bankJob)
	}
	uc.wg.Wait()
	return workflowID, nil
}

func (uc *workflowUseCase) OnSystemIngestionComplete(ctx context.Context, workflowID, jobID string, success bool) error {
	uc.mu.Lock()
	defer uc.mu.Unlock()

	wf, err := uc.workflowRepo.GetWorkflow(ctx, workflowID)
	if err != nil {
		return fmt.Errorf("failed to get workflow: %w", err)
	}

	if success {
		jobIDCopy := jobID
		wf.SystemIngestionJobID = &jobIDCopy
	} else {
		wf.Status = enum_status.FAILED.String()
		uc.workflowRepo.UpdateWorkflow(ctx, wf)
		return fmt.Errorf("system ingestion failed")
	}

	if wf.SystemIngestionJobID != nil && wf.BankIngestionJobID != nil {
		return uc.startReconciliation(ctx, workflowID, wf.StartDate, wf.EndDate)
	}

	uc.workflowRepo.UpdateWorkflow(ctx, wf)
	return nil
}

func (uc *workflowUseCase) OnBankIngestionComplete(ctx context.Context, workflowID, jobID string, success bool) error {
	uc.mu.Lock()
	defer uc.mu.Unlock()

	wf, err := uc.workflowRepo.GetWorkflow(ctx, workflowID)
	if err != nil {
		return fmt.Errorf("failed to get workflow: %w", err)
	}

	if success {
		jobIDCopy := jobID
		wf.BankIngestionJobID = &jobIDCopy
	} else {
		wf.Status = enum_status.FAILED.String()
		uc.workflowRepo.UpdateWorkflow(ctx, wf)
		return fmt.Errorf("bank ingestion failed")
	}

	if wf.SystemIngestionJobID != nil && wf.BankIngestionJobID != nil {
		return uc.startReconciliation(ctx, workflowID, wf.StartDate, wf.EndDate)
	}

	uc.workflowRepo.UpdateWorkflow(ctx, wf)
	return nil
}

func (uc *workflowUseCase) OnReconciliationComplete(ctx context.Context, workflowID, jobID string, success bool) error {
	uc.mu.Lock()
	defer uc.mu.Unlock()

	wf, err := uc.workflowRepo.GetWorkflow(ctx, workflowID)
	if err != nil {
		return fmt.Errorf("failed to get workflow: %w", err)
	}

	if success {
		wf.ReconciliationJobID = &jobID
		wf.Status = enum_status.COMPLETED.String()
	} else {
		wf.Status = enum_status.FAILED.String()
	}

	return uc.workflowRepo.UpdateWorkflow(ctx, wf)
}

func (uc *workflowUseCase) startReconciliation(ctx context.Context, workflowID string, startDate, endDate time.Time) error {
	result, err := uc.reconcileUC.ProcessReconciliation(ctx, startDate, endDate)
	if err != nil {
		return err
	}

	go func() {
		uc.OnReconciliationComplete(ctx, workflowID, result.JobID, true)
	}()

	return nil
}
