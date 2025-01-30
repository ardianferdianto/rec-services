package ingestion

import (
	"context"
	"fmt"
	"github.com/ardianferdianto/reconciliation-service/internal/domain"
	"github.com/ardianferdianto/reconciliation-service/internal/repository"
	"time"
)

// WorkerIngestionLoop runs a manager goroutine plus a pool of worker goroutines.
func WorkerIngestionLoop(
	ctx context.Context,
	uc IUseCase,
	repo repository.IngestionJobRepository,
	concurrency int,
) {
	// 1) create a channel of jobs
	jobCh := make(chan domain.IngestionJob, concurrency*2)

	// 2) spawn worker goroutines
	for i := 0; i < concurrency; i++ {
		go func(workerID int) {
			for {
				select {
				case <-ctx.Done():
					return
				case job, ok := <-jobCh:
					if !ok {
						return
					}
					// process the job
					if err := uc.ProcessIngestionJob(ctx, &job); err != nil {
						fmt.Printf("[worker %d] error ingesting job=%s: %v\n", workerID, job.JobID, err)
					}
				}
			}
		}(i)
	}

	// 3) manager loop that fetches pending jobs and feeds them to jobCh
	for {
		select {
		case <-ctx.Done():
			close(jobCh) // signal workers to exit
			return
		default:
			pendingJobs, err := repo.ListPendingJobs(ctx, 10)
			if err != nil {
				fmt.Printf("ListPendingJobs error: %v\n", err)
				time.Sleep(5 * time.Second)
				continue
			}
			if len(pendingJobs) == 0 {
				// no jobs => sleep a bit
				time.Sleep(5 * time.Second)
				continue
			}

			for _, job := range pendingJobs {
				// Mark the job IN_PROGRESS
				if err := repo.MarkJobInProgress(ctx, job.JobID); err != nil {
					fmt.Printf("MarkJobInProgress error for job=%s: %v\n", job.JobID, err)
					continue
				}
				// Send job to a worker
				select {
				case <-ctx.Done():
					close(jobCh)
					return
				case jobCh <- job:
					// enqueued successfully
				}
			}
		}
	}
}
