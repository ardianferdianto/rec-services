package repository

import (
	"context"
	"fmt"
	"github.com/ardianferdianto/reconciliation-service/internal/domain"
	"github.com/ardianferdianto/reconciliation-service/internal/infrastructure/sqlstore"
)

//go:generate mockgen -source=reconciliation_repository.go -destination=_mock/reconciliation_repository.go
type ReconciliationRepository interface {
	CreateJob(ctx context.Context, job domain.ReconciliationJob) error
	StoreResult(ctx context.Context, result domain.ReconciliationResult) (int, error)
	StoreMatchedRecord(ctx context.Context, rec domain.MatchedRecord) (int, error)
	StoreUnmatchedSystemTx(ctx context.Context, txList []domain.UnmatchedSystemTx) error
	StoreUnmatchedBankTx(ctx context.Context, txList []domain.UnmatchedBankTx) error
	GetReconciliationResult(ctx context.Context, jobID string) (*domain.ReconciliationResult, error)
	GetUnmatchedSystemTx(ctx context.Context, jobID string) ([]domain.UnmatchedSystemTx, error)
	GetUnmatchedBankTxGroupedByBank(ctx context.Context, jobID string) (map[string][]domain.UnmatchedBankTx, error)
}

type reconciliationRepo struct {
	db sqlstore.Store
}

// CreateJob creates a new reconciliation job record in the database
func (r *reconciliationRepo) CreateJob(ctx context.Context, job domain.ReconciliationJob) error {
	const query = `
        INSERT INTO reconciliation_jobs (job_id, start_date, end_date, created_at, updated_at)
        VALUES ($1, $2, $3, NOW(), NOW())
    `
	conn, deferFunc, err := r.db.GetConn(ctx)
	if err != nil {
		return fmt.Errorf("GetConn error: %w", err)
	}
	defer deferFunc()

	_, err = conn.Exec(ctx, query, job.JobID, job.StartDate, job.EndDate)
	return err
}

// StoreResult stores the final reconciliation results
func (r *reconciliationRepo) StoreResult(ctx context.Context, result domain.ReconciliationResult) (int, error) {
	const query = `
        INSERT INTO reconciliation_results (
            job_id, total_system_tx_count, total_bank_tx_count, matched_count,
            unmatched_system_count, unmatched_bank_count, total_discrepancies, created_at, updated_at
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
        RETURNING id
    `
	conn, deferFunc, err := r.db.GetConn(ctx)
	if err != nil {
		return 0, fmt.Errorf("GetConn error: %w", err)
	}
	defer deferFunc()

	var id int
	err = conn.QueryRow(ctx, query, result.JobID, result.TotalSystemTxCount, result.TotalBankTxCount, result.MatchedCount,
		result.UnmatchedSystemCount, result.UnmatchedBankCount, result.TotalDiscrepancies).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("query row scan error: %w", err)
	}
	return id, nil
}

// StoreMatchedRecord stores a record of a match between a system transaction and a bank statement
func (r *reconciliationRepo) StoreMatchedRecord(ctx context.Context, rec domain.MatchedRecord) (int, error) {
	const query = `
        INSERT INTO reconciliation_matched_records (
            job_id, system_tx_id, bank_statement_id, discrepancy, matched_at
        ) VALUES ($1, $2, $3, $4, NOW())
        RETURNING id
    `
	conn, deferFunc, err := r.db.GetConn(ctx)
	if err != nil {
		return 0, fmt.Errorf("GetConn error: %w", err)
	}
	defer deferFunc()

	var id int
	err = conn.QueryRow(ctx, query, rec.JobID, rec.SystemTxID, rec.BankStatementID, rec.Discrepancy).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("query row scan error: %w", err)
	}
	return id, nil
}

// StoreUnmatchedSystemTx stores details of unmatched system transactions
func (r *reconciliationRepo) StoreUnmatchedSystemTx(ctx context.Context, txList []domain.UnmatchedSystemTx) error {
	const query = `
        INSERT INTO reconciliation_unmatched_system_tx (
            job_id, trx_id, amount, trx_type, transaction_time, created_at, updated_at
        ) VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
    `
	conn, deferFunc, err := r.db.GetConn(ctx)
	if err != nil {
		return fmt.Errorf("GetConn error: %w", err)
	}
	defer deferFunc()

	for _, tx := range txList {
		_, err := conn.Exec(ctx, query, tx.JobID, tx.TrxID, tx.Amount, tx.Type, tx.TransactionTime)
		if err != nil {
			return fmt.Errorf("execute insert error: %w", err)
		}
	}
	return nil
}

// StoreUnmatchedBankTx stores details of unmatched bank transactions
func (r *reconciliationRepo) StoreUnmatchedBankTx(ctx context.Context, txList []domain.UnmatchedBankTx) error {
	const query = `
        INSERT INTO reconciliation_unmatched_bank_tx (
            job_id, unique_id, amount, statement_time, bank_code, created_at, updated_at
        ) VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
    `
	conn, deferFunc, err := r.db.GetConn(ctx)
	if err != nil {
		return fmt.Errorf("GetConn error: %w", err)
	}
	defer deferFunc()

	for _, b := range txList {
		_, err := conn.Exec(ctx, query, b.JobID, b.UniqueID, b.Amount, b.StatementDate, b.BankCode)
		if err != nil {
			return fmt.Errorf("execute insert error: %w", err)
		}
	}
	return nil
}

func (r *reconciliationRepo) GetReconciliationResult(ctx context.Context, jobID string) (*domain.ReconciliationResult, error) {
	const query = `
        SELECT job_id, total_system_tx_count, total_bank_tx_count, matched_count, unmatched_system_count, unmatched_bank_count, total_discrepancies, created_at, updated_at
        FROM reconciliation_results
        WHERE job_id = $1
    `
	conn, deferFunc, err := r.db.GetConn(ctx)
	if err != nil {
		return nil, fmt.Errorf("GetConn error: %w", err)
	}
	defer deferFunc()

	var wf domain.ReconciliationResult
	err = conn.QueryRow(ctx, query, jobID).Scan(
		&wf.JobID,
		&wf.TotalSystemTxCount,
		&wf.TotalBankTxCount,
		&wf.MatchedCount,
		&wf.UnmatchedSystemCount,
		&wf.UnmatchedBankCount,
		&wf.TotalDiscrepancies,
		&wf.CreatedAt,
		&wf.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("query error: %w", err)
	}

	return &wf, nil
}

func (r *reconciliationRepo) GetUnmatchedBankTxGroupedByBank(ctx context.Context, jobID string) (map[string][]domain.UnmatchedBankTx, error) {
	const query = `
        SELECT bank_code, unique_id, amount, statement_time
        FROM reconciliation_unmatched_bank_tx
        WHERE job_id = $1
        ORDER BY bank_code, statement_time
    `

	conn, deferFunc, err := r.db.GetConn(ctx)
	if err != nil {
		return nil, fmt.Errorf("GetConn error: %w", err)
	}
	defer deferFunc()

	rows, err := conn.Query(ctx, query, jobID)
	if err != nil {
		return nil, fmt.Errorf("query error: %w", err)
	}
	defer rows.Close()
	groupedByBank := make(map[string][]domain.UnmatchedBankTx)

	for rows.Next() {
		var bankTx domain.UnmatchedBankTx
		if err := rows.Scan(&bankTx.BankCode, &bankTx.UniqueID, &bankTx.Amount, &bankTx.StatementDate); err != nil {
			return nil, fmt.Errorf("scan unmatched bank statement: %w", err)
		}

		groupedByBank[bankTx.BankCode] = append(groupedByBank[bankTx.BankCode], bankTx)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating unmatched bank transactions: %w", err)
	}

	return groupedByBank, nil
}

func (r *reconciliationRepo) GetUnmatchedSystemTx(ctx context.Context, jobID string) ([]domain.UnmatchedSystemTx, error) {
	const query = `
        SELECT trx_id, amount, trx_type, transaction_time
        FROM reconciliation_unmatched_system_tx
        WHERE job_id = $1
        ORDER BY transaction_time
    `

	conn, deferFunc, err := r.db.GetConn(ctx)
	if err != nil {
		return nil, fmt.Errorf("GetConn error: %w", err)
	}
	defer deferFunc()

	rows, err := conn.Query(ctx, query, jobID)
	if err != nil {
		return nil, fmt.Errorf("query error: %w", err)
	}
	defer rows.Close()

	var unmatchedSystemTx []domain.UnmatchedSystemTx

	for rows.Next() {
		var tx domain.UnmatchedSystemTx
		if err := rows.Scan(&tx.TrxID, &tx.Amount, &tx.Type, &tx.TransactionTime); err != nil {
			return nil, fmt.Errorf("scan unmatched system transactions: %w", err)
		}

		tx.JobID = jobID
		unmatchedSystemTx = append(unmatchedSystemTx, tx)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating unmatched bank transactions: %w", err)
	}

	return unmatchedSystemTx, nil
}

func NewReconciliationRepo(db sqlstore.Store) ReconciliationRepository {
	return &reconciliationRepo{db: db}
}
