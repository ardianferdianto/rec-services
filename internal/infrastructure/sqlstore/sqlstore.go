package sqlstore

import (
	"context"
	"errors"
	"fmt"
	"github.com/ardianferdianto/reconciliation-service/pkg/db"
	"github.com/ardianferdianto/reconciliation-service/pkg/logger"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"
)

const (
	Tx = "tx"
)

// Store is the wrapper for dto.
type Store interface {
	BeginTx(ctx context.Context) context.Context
	CommitTx(ctx context.Context) error
	RollbackTx(ctx context.Context) error

	GetConn(ctx context.Context) (conn *pgx.Conn, deferFunc func(), err error)
	GetDB() *SQLStore
	Close()
}

type SQLStore struct {
	Master *pgxpool.Pool
}

func NewSQLStore(ctx context.Context, dbConfigMaster db.Config) (Store, error) {
	slog.InfoContext(ctx,
		fmt.Sprintf("db configs master, max open %d, min idle %d, max lifetime %v",
			dbConfigMaster.MaxOpen, dbConfigMaster.MinIdle, dbConfigMaster.MaxLifetime))

	configMaster := db.Config{
		Host:        dbConfigMaster.Host,
		Port:        dbConfigMaster.Port,
		User:        dbConfigMaster.User,
		Password:    dbConfigMaster.Password,
		Name:        dbConfigMaster.Name,
		MaxOpen:     dbConfigMaster.MaxOpen,
		MinIdle:     dbConfigMaster.MinIdle,
		MaxLifetime: dbConfigMaster.MaxLifetime,
		MaxIdleTime: dbConfigMaster.MaxIdleTime,
		ParseTime:   true,
		Driver:      dbConfigMaster.Driver,
	}

	poolCfg, err := configMaster.GetPgxPoolConfig()
	if err != nil {
		slog.ErrorContext(ctx, "failed to initialize db config when initiating sql store, error", logger.ErrAttr(err))
		return nil, err
	}
	dbMaster, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		slog.ErrorContext(ctx, "failed to connect to master db", logger.ErrAttr(err))
		return nil, err
	}

	return &SQLStore{
		Master: dbMaster,
	}, nil
}

func (s *SQLStore) BeginTx(ctx context.Context) context.Context {
	tx, err := s.Master.Begin(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "failed to begin transaction", logger.ErrAttr(err))
	}
	ctx = context.WithValue(ctx, Tx, tx)
	return ctx
}

func (s *SQLStore) CommitTx(ctx context.Context) error {
	tx, ok := ctx.Value(Tx).(pgx.Tx)
	if !ok {
		return errors.New("failed to commit on non transaction mode")
	}

	return tx.Commit(ctx)
}

func (s *SQLStore) RollbackTx(ctx context.Context) error {
	tx, ok := ctx.Value(Tx).(pgx.Tx)
	if !ok {
		return errors.New("failed to rollback on non transaction mode")
	}
	_ = tx.Rollback(ctx)
	return nil
}

func (s *SQLStore) getDBConn(ctx context.Context) (*pgxpool.Conn, *pgx.Conn, error) {
	var pgxConn *pgx.Conn
	var poolConn *pgxpool.Conn

	tx, ok := ctx.Value(Tx).(pgx.Tx)
	if ok {
		pgxConn = tx.Conn()
	} else {
		dbm := s.GetDB().Master
		conn, err := dbm.Acquire(ctx)
		if err != nil {
			slog.ErrorContext(ctx, "error acquiring db connection", logger.ErrAttr(err))
			return nil, nil, errors.New("error acquiring db connection, max pool exhausted")
		}
		poolConn = conn
	}

	return poolConn, pgxConn, nil
}

func (s *SQLStore) GetConn(ctx context.Context) (conn *pgx.Conn, deferFunc func(), err error) {
	deferFunc = func() {}
	poolConn, pgxConn, err := s.getDBConn(ctx)
	if err != nil {
		return nil, deferFunc, err
	}

	if poolConn != nil {
		deferFunc = func() {
			slog.InfoContext(ctx, "connection released") // will be removed later, for testing purpose
			poolConn.Release()
		}
		pgxConn = poolConn.Conn()
	}

	return pgxConn, deferFunc, nil
}

func (s *SQLStore) GetDB() *SQLStore {
	return s
}

func (s *SQLStore) Close() {
	s.Master.Close()
}
