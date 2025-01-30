package contract

import (
	"github.com/ardianferdianto/reconciliation-service/internal/domain"
	"time"
)

type StartWorkflowRequest struct {
	SystemTransactionFilePath string    `json:"system_transaction_file_path"`
	BankStatementFilePaths    []string  `json:"bank_statement_file_paths"`
	StartDate                 time.Time `json:"start_date"`
	EndDate                   time.Time `json:"end_date"`
}

type WorkflowSummaryResponse struct {
	WorkflowID       string                        `json:"workflow_id"`
	Status           string                        `json:"status"`
	StartDate        time.Time                     `json:"start_date"`
	EndDate          time.Time                     `json:"end_date"`
	ReconcileSummary *domain.ReconciliationSummary `json:"reconciliation_summary"`
}
