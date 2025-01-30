package domain

import "time"

// IngestionJob tracks a single CSV ingestion into DB (system or bank).
type IngestionJob struct {
	JobID               string
	FileType            string // "SYSTEM_TX" or "BANK_STMT"
	FileName            string
	TotalLinesProcessed int64
	Status              string // "IN_PROGRESS", "COMPLETED", "FAILED"
	CreatedAt           time.Time
	UpdatedAt           time.Time
}
