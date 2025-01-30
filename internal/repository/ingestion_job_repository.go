package repository

import (
	"context"
	"fmt"
	"github.com/ardianferdianto/reconciliation-service/internal/domain"
	"github.com/ardianferdianto/reconciliation-service/internal/infrastructure/sqlstore"
	"github.com/google/uuid"
)

//go:generate mockgen -source=ingestion_job_repository.go -destination=_mock/ingestion_job_repository.go
type IngestionJobRepository interface {
	CreateJob(ctx context.Context, job *domain.IngestionJob) error
	UpdateJobProgress(ctx context.Context, jobID string, linesProcessed int64, status string) error
	ListPendingJobs(ctx context.Context, limit int) ([]domain.IngestionJob, error)
	MarkJobInProgress(ctx context.Context, jobID string) error
}

type ingestionRepo struct {
	db sqlstore.Store
}

func NewIngestionRepo(db sqlstore.Store) IngestionJobRepository {
	return &ingestionRepo{db: db}
}

func (r *ingestionRepo) CreateJob(ctx context.Context, job *domain.IngestionJob) error {
	conn, deferFunc, err := r.db.GetConn(ctx)
	if err != nil {
		return err
	}
	if job.JobID == "" {
		job.JobID = uuid.New().String()
	}
	if job.Status == "" {
		job.Status = "PENDING"
	}
	const q = `
	INSERT INTO ingestion_jobs (job_id, file_type, file_name, total_lines_processed, status, created_at, updated_at)
	VALUES ($1, $2, $3, 0, $4, NOW(), NOW())
	`
	_, err = conn.Exec(ctx, q, job.JobID, job.FileType, job.FileName, job.Status)
	defer deferFunc()
	return err
}

func (r *ingestionRepo) UpdateJobProgress(ctx context.Context, jobID string, linesProcessed int64, status string) error {
	conn, deferFunc, err := r.db.GetConn(ctx)
	if err != nil {
		return err
	}
	const q = `
	UPDATE ingestion_jobs
	SET total_lines_processed = $1,
	    status = $2,
	    updated_at = NOW()
	WHERE job_id = $3
	`
	_, err = conn.Exec(ctx, q, linesProcessed, status, jobID)
	defer deferFunc()
	return err
}

func (r *ingestionRepo) ListPendingJobs(ctx context.Context, limit int) ([]domain.IngestionJob, error) {
	conn, deferFunc, err := r.db.GetConn(ctx)
	if err != nil {
		return nil, fmt.Errorf("GetConn error: %w", err)
	}
	defer deferFunc() // ensure connection is released

	const q = `
        SELECT job_id, file_type, file_name, total_lines_processed, status, created_at, updated_at
        FROM ingestion_jobs
        WHERE status IN ('PENDING')
        ORDER BY created_at ASC
        LIMIT $1
    `

	// Use Query(...) instead of Exec(...) to get a row set
	rows, err := conn.Query(ctx, q, limit)
	if err != nil {
		return nil, fmt.Errorf("query error: %w", err)
	}
	defer rows.Close()

	var jobs []domain.IngestionJob
	for rows.Next() {
		var job domain.IngestionJob
		if scanErr := rows.Scan(
			&job.JobID,
			&job.FileType,
			&job.FileName,
			&job.TotalLinesProcessed,
			&job.Status,
			&job.CreatedAt,
			&job.UpdatedAt,
		); scanErr != nil {
			return nil, fmt.Errorf("rows.Scan error: %w", scanErr)
		}
		jobs = append(jobs, job)
	}

	// Handle any iteration error
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return jobs, nil
}

func (r *ingestionRepo) MarkJobInProgress(ctx context.Context, jobID string) error {
	conn, deferFunc, err := r.db.GetConn(ctx)
	if err != nil {
		return fmt.Errorf("GetConn error: %w", err)
	}
	defer deferFunc() // ensure connection is released

	const q = `
	UPDATE ingestion_jobs
	SET status = 'IN_PROGRESS',
	    updated_at = NOW()
	WHERE job_id = $1
	`
	_, err = conn.Exec(ctx, q, jobID)
	return err
}
