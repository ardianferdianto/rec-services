package repository

import (
	"context"
	"fmt"
	"github.com/ardianferdianto/reconciliation-service/internal/domain"
	"github.com/ardianferdianto/reconciliation-service/internal/infrastructure/sqlstore"
	"time"
)

//go:generate mockgen -source=data_repository.go -destination=_mock/data_repository.go
type DataRepository interface {
	BatchInsertSystemTx(ctx context.Context, txList []domain.Transaction) error
	BatchInsertBankStmts(ctx context.Context, stmts []domain.BankStatement) error
	FindSystemTxByDateRange(ctx context.Context, startDate, endDate time.Time) ([]domain.Transaction, error)
	FindBankStmtsByDateRange(ctx context.Context, startDate, endDate time.Time) ([]domain.BankStatement, error)
}

type dataRepo struct {
	db sqlstore.Store
}

func NewDataRepo(db sqlstore.Store) DataRepository {
	return &dataRepo{db: db}
}

func (r *dataRepo) BatchInsertSystemTx(ctx context.Context, txList []domain.Transaction) error {
	ctx = r.db.BeginTx(ctx)
	defer r.db.RollbackTx(ctx)

	const query = `
        INSERT INTO system_transactions (trx_id, amount, trx_type, transaction_time, created_at, updated_at)
        VALUES ($1, $2, $3, $4, NOW(), NOW())
		ON CONFLICT (trx_id) DO NOTHING
    `

	conn, deferFunc, err := r.db.GetConn(ctx)
	if err != nil {
		return fmt.Errorf("GetConn error: %w", err)
	}
	defer deferFunc()

	_, err = conn.Prepare(ctx, "insertSystemTx", query)
	if err != nil {
		return fmt.Errorf("prepare statement error: %w", err)
	}
	defer conn.Deallocate(ctx, "insertSystemTx")

	for _, tx := range txList {
		_, err := conn.Exec(ctx, "insertSystemTx", tx.TrxID, tx.Amount, tx.Type, tx.TransactionTime)
		if err != nil {
			return fmt.Errorf("execute insert error: %w", err)
		}
	}

	if err := r.db.CommitTx(ctx); err != nil {
		return fmt.Errorf("commit tx error: %w", err)
	}
	return nil
}

func (r *dataRepo) BatchInsertBankStmts(ctx context.Context, stmtList []domain.BankStatement) error {
	ctx = r.db.BeginTx(ctx)
	defer r.db.RollbackTx(ctx)

	const query = `
        INSERT INTO bank_statements (unique_id, amount, statement_time, bank_code, hash_code, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
        ON CONFLICT (hash_code) DO NOTHING
    `

	conn, deferFunc, err := r.db.GetConn(ctx)
	if err != nil {
		return fmt.Errorf("GetConn error: %w", err)
	}
	defer deferFunc()

	_, err = conn.Prepare(ctx, "insertBankStmt", query)
	if err != nil {
		return fmt.Errorf("prepare statement error: %w", err)
	}
	defer conn.Deallocate(ctx, "insertBankStmt")

	for _, stmt := range stmtList {
		stmt.HashCode = stmt.GenerateHashCode()
		_, err := conn.Exec(ctx, "insertBankStmt", stmt.UniqueID, stmt.Amount, stmt.StatementTime, stmt.BankCode, stmt.HashCode)
		if err != nil {
			return fmt.Errorf("execute insert error: %w", err)
		}
	}

	if err := r.db.CommitTx(ctx); err != nil {
		return fmt.Errorf("commit tx error: %w", err)
	}
	return nil
}

func (r *dataRepo) FindSystemTxByDateRange(ctx context.Context, startDate, endDate time.Time) ([]domain.Transaction, error) {
	const query = `
        SELECT id, trx_id, amount, trx_type, transaction_time, created_at, updated_at
        FROM system_transactions
        WHERE transaction_time BETWEEN $1 AND $2
        ORDER BY transaction_time ASC
    `

	conn, deferFunc, err := r.db.GetConn(ctx)
	if err != nil {
		return nil, fmt.Errorf("GetConn error: %w", err)
	}
	defer deferFunc()

	rows, err := conn.Query(ctx, query, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("query error: %w", err)
	}
	defer rows.Close()

	var transactions []domain.Transaction
	for rows.Next() {
		var t domain.Transaction
		if err := rows.Scan(&t.ID, &t.TrxID, &t.Amount, &t.Type, &t.TransactionTime, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, fmt.Errorf("row scan error: %w", err)
		}
		transactions = append(transactions, t)
	}

	return transactions, nil
}

// FindBankStmtsByDateRange retrieves bank statements within a specified date range
func (r *dataRepo) FindBankStmtsByDateRange(ctx context.Context, startDate, endDate time.Time) ([]domain.BankStatement, error) {
	const query = `
        SELECT id, unique_id, amount, statement_time, bank_code, created_at, updated_at
        FROM bank_statements
        WHERE statement_time BETWEEN $1 AND $2
        ORDER BY statement_time ASC
    `

	conn, deferFunc, err := r.db.GetConn(ctx)
	if err != nil {
		return nil, fmt.Errorf("GetConn error: %w", err)
	}
	defer deferFunc()

	rows, err := conn.Query(ctx, query, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("query error: %w", err)
	}
	defer rows.Close()

	var statements []domain.BankStatement
	for rows.Next() {
		var b domain.BankStatement
		if err := rows.Scan(&b.ID, &b.UniqueID, &b.Amount, &b.StatementTime, &b.BankCode, &b.CreatedAt, &b.UpdatedAt); err != nil {
			return nil, fmt.Errorf("row scan error: %w", err)
		}
		statements = append(statements, b)
	}

	return statements, nil
}
