package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

type TransactionType string

const (
	Debit  = "DEBIT"
	Credit = "CREDIT"
)

// Transaction represents the system transaction data.
type Transaction struct {
	ID              int
	TrxID           string
	Amount          float64
	Type            string
	TransactionTime time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// BankStatement represents the external bank statement.
type BankStatement struct {
	ID            int
	UniqueID      string
	Amount        float64 // Negative for debits, positive for credits
	StatementTime time.Time
	BankCode      string
	CreatedAt     time.Time
	UpdatedAt     time.Time
	HashCode      string
}

func (b *BankStatement) GenerateHashCode() string {
	dateStr := b.StatementTime.Format("2006-01-02") // Extract YYYY-MM-DD
	hashInput := fmt.Sprintf("%s|%s|%.2f|%s", b.UniqueID, b.BankCode, b.Amount, dateStr)

	hash := sha256.Sum256([]byte(hashInput))
	return hex.EncodeToString(hash[:])
}

// ReconciliationJob for auditing
type ReconciliationJob struct {
	JobID     string
	StartDate time.Time
	EndDate   time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}

// ReconciliationResult holds summary stats for a job
type ReconciliationResult struct {
	ID                   int
	JobID                string
	TotalSystemTxCount   int
	TotalBankTxCount     int
	MatchedCount         int
	UnmatchedSystemCount int
	UnmatchedBankCount   int
	TotalDiscrepancies   float64
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

// MatchedRecord links 1 systemTx to 1 bankStatement for a job
type MatchedRecord struct {
	ID              int
	JobID           string
	SystemTxID      int
	BankStatementID int
	Discrepancy     float64
	MatchedAt       time.Time
}

// UnmatchedSystemTx and UnmatchedBankTx store unmatched items
type UnmatchedSystemTx struct {
	ID              int
	JobID           string
	TrxID           string
	Amount          float64
	Type            string
	TransactionTime time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type UnmatchedBankTx struct {
	ID            int
	JobID         string
	UniqueID      string
	Amount        float64
	StatementDate time.Time
	BankCode      string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type ReconciliationSummary struct {
	TotalTransactionsProcessed int                          `json:"total_transactions_processed"`
	TotalMatchedTransactions   int                          `json:"total_matched_transactions"`
	TotalUnmatchedTransactions int                          `json:"total_unmatched_transactions"`
	UnmatchedSystemTx          []UnmatchedSystemTx          `json:"unmatched_system_transactions"`
	UnmatchedBankTxByBank      map[string][]UnmatchedBankTx `json:"unmatched_bank_transactions_by_bank"`
	TotalDiscrepancies         float64                      `json:"total_discrepancies"`
}
