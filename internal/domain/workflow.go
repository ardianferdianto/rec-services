package domain

import "time"

type Workflow struct {
	WorkflowID           string
	SystemIngestionJobID *string
	BankIngestionJobID   *string
	ReconciliationJobID  *string
	Status               string // e.g. "IN_PROGRESS", "COMPLETED", "FAILED"
	StartDate            time.Time
	EndDate              time.Time
	CreatedAt            time.Time
	UpdatedAt            time.Time
}
