package reconcile

import (
	"context"
	"fmt"
	"github.com/ardianferdianto/reconciliation-service/internal/domain"
	"github.com/ardianferdianto/reconciliation-service/internal/repository"
	"github.com/google/uuid"
	"strings"
	"time"
)

const (
	MinMatchScore = 2
)

type IUseCase interface {
	ProcessReconciliation(ctx context.Context, startDate time.Time, endDate time.Time) (domain.ReconciliationResult, error)
	GetReconciliationSummary(ctx context.Context, jobID string) (domain.ReconciliationSummary, error)
}

type useCase struct {
	recRepo  repository.ReconciliationRepository
	dataRepo repository.DataRepository
}

func NewReconciliationUseCase(
	recRepo repository.ReconciliationRepository,
	dataRepo repository.DataRepository,
) IUseCase {
	return &useCase{
		recRepo:  recRepo,
		dataRepo: dataRepo,
	}
}

func (s *useCase) GetReconciliationSummary(ctx context.Context, jobID string) (domain.ReconciliationSummary, error) {
	result, err := s.recRepo.GetReconciliationResult(ctx, jobID)
	if err != nil {
		return domain.ReconciliationSummary{}, fmt.Errorf("failed to get reconciliation result: %w", err)
	}

	unmatchedSystemTx, err := s.recRepo.GetUnmatchedSystemTx(ctx, jobID)
	if err != nil {
		return domain.ReconciliationSummary{}, fmt.Errorf("failed to get unmatched system transactions: %w", err)
	}

	unmatchedByBank, err := s.recRepo.GetUnmatchedBankTxGroupedByBank(ctx, jobID)
	if err != nil {
		return domain.ReconciliationSummary{}, fmt.Errorf("failed to get unmatched bank transactions grouped by bank: %w", err)
	}

	return domain.ReconciliationSummary{
		TotalTransactionsProcessed: result.TotalSystemTxCount,
		TotalMatchedTransactions:   result.MatchedCount,
		TotalUnmatchedTransactions: result.UnmatchedSystemCount,
		UnmatchedSystemTx:          unmatchedSystemTx,
		UnmatchedBankTxByBank:      unmatchedByBank,
		TotalDiscrepancies:         result.TotalDiscrepancies,
	}, nil
}

func (s *useCase) ProcessReconciliation(ctx context.Context, startDate, endDate time.Time) (domain.ReconciliationResult, error) {
	jobID := uuid.New().String()
	job := domain.ReconciliationJob{
		JobID:     jobID,
		StartDate: startDate,
		EndDate:   endDate,
	}
	if err := s.recRepo.CreateJob(ctx, job); err != nil {
		return domain.ReconciliationResult{}, err
	}

	systemTx, err := s.dataRepo.FindSystemTxByDateRange(ctx, startDate, endDate)
	if err != nil {
		return domain.ReconciliationResult{}, err
	}
	bankStmts, err := s.dataRepo.FindBankStmtsByDateRange(ctx, startDate, endDate)
	if err != nil {
		return domain.ReconciliationResult{}, err
	}

	matchedRecords, unmatchedSystemTx, unmatchedBankStmts, totalDiscrepancies := s.matchRecords(jobID, systemTx, bankStmts)

	for _, matched := range matchedRecords {
		if _, err := s.recRepo.StoreMatchedRecord(ctx, matched); err != nil {
			return domain.ReconciliationResult{}, err
		}
	}

	if err := s.recRepo.StoreUnmatchedSystemTx(ctx, unmatchedSystemTx); err != nil {
		return domain.ReconciliationResult{}, err
	}
	if err := s.recRepo.StoreUnmatchedBankTx(ctx, unmatchedBankStmts); err != nil {
		return domain.ReconciliationResult{}, err
	}

	result := domain.ReconciliationResult{
		JobID:                jobID,
		TotalSystemTxCount:   len(systemTx),
		TotalBankTxCount:     len(bankStmts),
		MatchedCount:         len(matchedRecords),
		UnmatchedSystemCount: len(unmatchedSystemTx),
		UnmatchedBankCount:   len(unmatchedBankStmts),
		TotalDiscrepancies:   totalDiscrepancies,
	}
	_, err = s.recRepo.StoreResult(ctx, result)
	if err != nil {
		return domain.ReconciliationResult{}, err
	}

	return result, nil
}

func (s *useCase) matchRecords(jobID string, systemTx []domain.Transaction, bankStmts []domain.BankStatement) ([]domain.MatchedRecord, []domain.UnmatchedSystemTx, []domain.UnmatchedBankTx, float64) {
	var matched []domain.MatchedRecord
	var unmatchedSystem []domain.UnmatchedSystemTx
	var unmatchedBank []domain.UnmatchedBankTx
	var totalDiscrepancies float64

	bankMap := make(map[string][]domain.BankStatement) // Keyed by date and amount for initial coarse matching
	for _, stmt := range bankStmts {
		key := stmt.StatementTime.Format("20060102") + "_" + formatAmount(stmt.Amount)
		bankMap[key] = append(bankMap[key], stmt)
	}

	// Attempt to match transactions
	for _, tx := range systemTx {
		expectedAmount := tx.Amount
		if tx.Type == domain.Debit {
			expectedAmount = -tx.Amount
		}
		txKey := fmt.Sprintf("%s_%.2f", tx.TransactionTime.Format("20060102"), expectedAmount)
		if stmts, exists := bankMap[txKey]; exists {
			for i, stmt := range stmts {
				score := calculateMatchScore(tx, stmt)
				if score >= MinMatchScore {
					discrepancy := calculateDiscrepancy(tx.Amount, stmt.Amount)
					totalDiscrepancies += discrepancy
					matchedRecord := domain.MatchedRecord{
						JobID:           jobID,
						SystemTxID:      tx.ID,
						BankStatementID: stmt.ID,
						Discrepancy:     discrepancy,
					}
					matched = append(matched, matchedRecord)
					// Remove the matched statement
					stmts = append(stmts[:i], stmts[i+1:]...)
					bankMap[txKey] = stmts
					break
				}
			}
			if len(stmts) == 0 {
				delete(bankMap, txKey)
			}
		} else {
			unmatchedSystem = append(unmatchedSystem, domain.UnmatchedSystemTx{
				JobID:           jobID,
				TrxID:           tx.TrxID,
				Amount:          tx.Amount,
				Type:            tx.Type,
				TransactionTime: tx.TransactionTime,
			})
		}
	}

	// Collect any remaining unmatched bank statements
	for _, stmts := range bankMap {
		for _, stmt := range stmts {
			unmatchedBank = append(unmatchedBank, domain.UnmatchedBankTx{
				JobID:         jobID,
				UniqueID:      stmt.UniqueID,
				Amount:        stmt.Amount,
				StatementDate: stmt.StatementTime,
				BankCode:      stmt.BankCode,
			})
		}
	}

	return matched, unmatchedSystem, unmatchedBank, totalDiscrepancies
}

func calculateMatchScore(tx domain.Transaction, stmt domain.BankStatement) int {
	score := 0
	if tx.TransactionTime.Format("20060102") == stmt.StatementTime.Format("20060102") {
		score++
	}
	if formatAmount(tx.Amount) == formatAmount(stmt.Amount) {
		score++
	}
	if strings.Contains(stmt.UniqueID, tx.TrxID) || strings.Contains(tx.TrxID, stmt.UniqueID) {
		score++
	}
	return score
}

func calculateDiscrepancy(txAmount, stmtAmount float64) float64 {
	discrepancy := txAmount - stmtAmount
	if discrepancy < 0 {
		return -discrepancy
	}
	return discrepancy
}

func formatAmount(amount float64) string {
	return fmt.Sprintf("%.2f", amount)
}
