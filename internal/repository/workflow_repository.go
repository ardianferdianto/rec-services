package repository

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/ardianferdianto/reconciliation-service/internal/domain"
	"github.com/ardianferdianto/reconciliation-service/internal/infrastructure/sqlstore"
)

type WorkflowRepository interface {
	CreateWorkflow(ctx context.Context, wf domain.Workflow) error
	GetWorkflow(ctx context.Context, workflowID string) (domain.Workflow, error)
	UpdateWorkflow(ctx context.Context, wf domain.Workflow) error
}

// workflowRepo works with the sqlstore.Store to manage ReconciliationWorkflow records
type workflowRepo struct {
	db sqlstore.Store
}

func NewWorkflowRepo(db sqlstore.Store) WorkflowRepository {
	return &workflowRepo{db: db}
}

// CreateWorkflow inserts a new workflow record into the reconciliation_workflows table
func (r *workflowRepo) CreateWorkflow(ctx context.Context, wf domain.Workflow) error {
	const query = `
        INSERT INTO reconciliation_workflows (
            workflow_id,
            system_ingestion_job_id,
            bank_ingestion_job_id,
            reconciliation_job_id,
            status,
            start_date,
            end_date
        ) VALUES ($1, $2, $3, $4, $5, $6, $7)
    `

	// Begin a new transaction
	ctx = r.db.BeginTx(ctx)
	defer r.db.RollbackTx(ctx) // Ensure rollback in case of failure

	conn, deferFunc, err := r.db.GetConn(ctx)
	if err != nil {
		return fmt.Errorf("GetConn error: %w", err)
	}
	defer deferFunc()

	// Execute the query to insert the workflow record
	_, err = conn.Exec(ctx, query,
		wf.WorkflowID,
		wf.SystemIngestionJobID,
		wf.BankIngestionJobID,
		wf.ReconciliationJobID,
		wf.Status,
		wf.StartDate,
		wf.EndDate,
	)
	if err != nil {
		return fmt.Errorf("insert workflow error: %w", err)
	}

	// Commit the transaction
	if err := r.db.CommitTx(ctx); err != nil {
		return fmt.Errorf("commit tx error: %w", err)
	}

	return nil
}

// GetWorkflow retrieves a workflow by its ID from the reconciliation_workflows table
func (r *workflowRepo) GetWorkflow(ctx context.Context, id string) (domain.Workflow, error) {
	const query = `
        SELECT
          workflow_id,
          system_ingestion_job_id,
          bank_ingestion_job_id,
          reconciliation_job_id,
          status,
          start_date,
          end_date,
          created_at,
          updated_at
        FROM reconciliation_workflows
        WHERE workflow_id = $1
    `

	conn, deferFunc, err := r.db.GetConn(ctx)
	if err != nil {
		return domain.Workflow{}, fmt.Errorf("GetConn error: %w", err)
	}
	defer deferFunc()

	var wf domain.Workflow
	err = conn.QueryRow(ctx, query, id).Scan(
		&wf.WorkflowID,
		&wf.SystemIngestionJobID,
		&wf.BankIngestionJobID,
		&wf.ReconciliationJobID,
		&wf.Status,
		&wf.StartDate,
		&wf.EndDate,
		&wf.CreatedAt,
		&wf.UpdatedAt,
	)
	if err != nil {
		return domain.Workflow{}, fmt.Errorf("query error: %w", err)
	}

	return wf, nil
}

// UpdateWorkflow updates an existing workflow record in the reconciliation_workflows table
func (r *workflowRepo) UpdateWorkflow(ctx context.Context, wf domain.Workflow) error {
	const query = `
        UPDATE reconciliation_workflows
        SET system_ingestion_job_id = $1, 
            bank_ingestion_job_id = $2, 
            reconciliation_job_id = $3,
            status = $4,
            updated_at = NOW()
        WHERE workflow_id = $5
    `

	// Begin a new transaction
	ctx = r.db.BeginTx(ctx)
	defer r.db.RollbackTx(ctx) // Ensure rollback in case of failure

	conn, deferFunc, err := r.db.GetConn(ctx)
	if err != nil {
		return fmt.Errorf("GetConn error: %w", err)
	}
	defer deferFunc()

	// Convert pointers to values or NULL
	systemIngestionJobID := sql.NullString{}
	if wf.SystemIngestionJobID != nil {
		systemIngestionJobID.String = *wf.SystemIngestionJobID
		systemIngestionJobID.Valid = true
	}

	bankIngestionJobID := sql.NullString{}
	if wf.BankIngestionJobID != nil {
		bankIngestionJobID.String = *wf.BankIngestionJobID
		bankIngestionJobID.Valid = true
	}

	reconciliationJobID := sql.NullString{}
	if wf.ReconciliationJobID != nil {
		reconciliationJobID.String = *wf.ReconciliationJobID
		reconciliationJobID.Valid = true
	}

	// Execute the query to update the workflow record
	_, err = conn.Exec(ctx, query,
		systemIngestionJobID,
		bankIngestionJobID,
		reconciliationJobID,
		wf.Status,
		wf.WorkflowID,
	)
	if err != nil {
		return fmt.Errorf("update workflow error: %w", err)
	}

	// Commit the transaction
	if err := r.db.CommitTx(ctx); err != nil {
		return fmt.Errorf("commit tx error: %w", err)
	}

	return nil
}
