package ingestion

import (
	"bufio"
	"context"
	"encoding/csv"
	"fmt"
	"github.com/ardianferdianto/reconciliation-service/internal/domain"
	"github.com/ardianferdianto/reconciliation-service/internal/infrastructure"
	"github.com/ardianferdianto/reconciliation-service/internal/repository"
	"github.com/ardianferdianto/reconciliation-service/internal/usecase/parser"
	"github.com/minio/minio-go/v7"
	"io"
	"log/slog"
)

const defaultBatchSize = 1000

type IUseCase interface {
	CreateIngestionJob(ctx context.Context, job *domain.IngestionJob) error
	ProcessIngestionJob(ctx context.Context, job *domain.IngestionJob) error
	FetchFileMetadata(ctx context.Context, objectName string) (*minio.ObjectInfo, error)
}

type useCase struct {
	jobRepo     repository.IngestionJobRepository
	dataRepo    repository.DataRepository
	minioClient infrastructure.IMinioClient
}

func NewIngestionUseCase(
	jobRepo repository.IngestionJobRepository,
	dataRepo repository.DataRepository,
	minioClient infrastructure.IMinioClient,
) IUseCase {
	return &useCase{
		jobRepo:     jobRepo,
		dataRepo:    dataRepo,
		minioClient: minioClient,
	}
}

func (u *useCase) CreateIngestionJob(ctx context.Context, job *domain.IngestionJob) error {
	return u.jobRepo.CreateJob(ctx, job)
}

func (u *useCase) ProcessIngestionJob(ctx context.Context, job *domain.IngestionJob) error {
	return u.ingestCSVJob(ctx, job)
}

func (u *useCase) ingestCSVJob(ctx context.Context, job *domain.IngestionJob) error {
	prsr := parser.GetParser(job.FileType)
	if prsr == nil {
		u.jobRepo.UpdateJobProgress(ctx, job.JobID, job.TotalLinesProcessed, "FAILED")
		return fmt.Errorf("no parser registered for fileType=%s", job.FileType)
	}

	obj, err := u.minioClient.GetObject(ctx, job.FileName, minio.GetObjectOptions{})
	if err != nil {
		u.jobRepo.UpdateJobProgress(ctx, job.JobID, job.TotalLinesProcessed, "FAILED")
		return fmt.Errorf("GetObject error: %w", err)
	}
	defer obj.Close()

	cReader := csv.NewReader(bufio.NewReader(obj))

	// Skip the header row
	if _, err := cReader.Read(); err != nil {
		u.jobRepo.UpdateJobProgress(ctx, job.JobID, job.TotalLinesProcessed, "FAILED")
		return fmt.Errorf("failed to read header: %w", err)
	}

	var sysBatch []domain.Transaction
	var bankBatch []domain.BankStatement
	linesProcessed := int64(0)

	for {
		record, err := cReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			slog.ErrorContext(ctx, fmt.Sprintf("CSV parse error: %s", err.Error()))
			continue
		}
		linesProcessed++

		objVal, parseErr := prsr.ParseLine(record)
		if parseErr != nil {
			slog.ErrorContext(ctx, fmt.Sprintf("CSV parse line: %d error: %s", linesProcessed, err.Error()))
			continue
		}

		switch val := objVal.(type) {
		case domain.Transaction:
			sysBatch = append(sysBatch, val)
			if len(sysBatch) >= defaultBatchSize {
				if err := u.dataRepo.BatchInsertSystemTx(ctx, sysBatch); err != nil {
					u.jobRepo.UpdateJobProgress(ctx, job.JobID, linesProcessed, "FAILED")
					return err
				}
				sysBatch = sysBatch[:0]
				u.jobRepo.UpdateJobProgress(ctx, job.JobID, linesProcessed, "IN_PROGRESS")
			}
		case domain.BankStatement:
			bankBatch = append(bankBatch, val)
			if len(bankBatch) >= defaultBatchSize {
				if err := u.dataRepo.BatchInsertBankStmts(ctx, bankBatch); err != nil {
					u.jobRepo.UpdateJobProgress(ctx, job.JobID, linesProcessed, "FAILED")
					return err
				}
				bankBatch = bankBatch[:0]
				u.jobRepo.UpdateJobProgress(ctx, job.JobID, linesProcessed, "IN_PROGRESS")
			}
		default:
			slog.ErrorContext(ctx, fmt.Sprintf("CSV error: Unknown object type from parser"))
			continue
		}
	}

	if len(sysBatch) > 0 {
		if err := u.dataRepo.BatchInsertSystemTx(ctx, sysBatch); err != nil {
			u.jobRepo.UpdateJobProgress(ctx, job.JobID, linesProcessed, "FAILED")
			return err
		}
	}
	if len(bankBatch) > 0 {
		if err := u.dataRepo.BatchInsertBankStmts(ctx, bankBatch); err != nil {
			u.jobRepo.UpdateJobProgress(ctx, job.JobID, linesProcessed, "FAILED")
			return err
		}
	}

	u.jobRepo.UpdateJobProgress(ctx, job.JobID, linesProcessed, "COMPLETED")
	return nil
}

func (u *useCase) FetchFileMetadata(ctx context.Context, objectName string) (*minio.ObjectInfo, error) {
	objInfo, err := u.minioClient.StatObject(ctx, objectName)
	if err != nil {
		return nil, fmt.Errorf("file not found in storage: %w", err)
	}
	return objInfo, nil
}
