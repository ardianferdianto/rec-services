package sqlstore

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"testing"
)

func testContext(t *testing.T) context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	return ctx
}

func TestSQLStore_GetConn_WithDBTrx(t *testing.T) {
	cfg := config.All(`../../../`)
	ctx := testContext(t)
	dbConfig := cfg.Database.Master
	store, _ := sqlstore.NewSQLStore(ctx, dbConfig)
	db := store.GetDB()

	ctx = db.BeginTx(ctx)
	assert.Equal(t, int32(1), db.Master.Stat().AcquiredConns())

	conn, deferFunc, err := db.GetConn(ctx)
	assert.NoError(t, err)
	_, _ = conn.Exec(ctx, `CREATE TABLE samples (column1 text);`)
	defer func() {
		_, _ = conn.Exec(context.Background(), `DROP TABLE samples;`)
		db.Master.Close()
	}()
	deferFunc()

	err = db.CommitTx(ctx)
	assert.NoError(t, err)
	assert.Equal(t, int32(0), db.Master.Stat().AcquiredConns())
}

func TestSQLStore_GetConn_WithoutDBTrx(t *testing.T) {
	cfg := config.All(`../../../`)
	ctx := testContext(t)
	dbConfig := cfg.Database.Master
	store, _ := sqlstore.NewSQLStore(ctx, dbConfig)
	db := store.GetDB()

	assert.Equal(t, int32(0), db.Master.Stat().AcquiredConns())

	conn, deferFunc, err := db.GetConn(ctx)
	assert.NoError(t, err)
	_, _ = conn.Exec(ctx, `CREATE TABLE samples (column1 text);`)
	defer func() {
		_, _ = conn.Exec(ctx, `DROP TABLE samples;`)
		db.Master.Close()
	}()
	deferFunc()

	assert.Equal(t, int32(0), db.Master.Stat().AcquiredConns())
}

func TestSQLStore_BeginTx(t *testing.T) {
	cfg := config.All(`../../../`)
	ctx := testContext(t)
	dbConfig := cfg.Database.Master
	store, _ := sqlstore.NewSQLStore(ctx, dbConfig)

	ctx = store.BeginTx(ctx)
	defer func() {
		_ = store.RollbackTx(ctx)
		store.GetDB().Master.Close()
	}()

	tx, ok := ctx.Value(sqlstore.Tx).(pgx.Tx)
	assert.NotEmpty(t, tx)
	assert.True(t, ok)
}

func TestSQLStore_CommitTx(t *testing.T) {
	cfg := config.All(`../../../`)
	ctx := testContext(t)
	dbConfig := cfg.Database.Master
	store, _ := sqlstore.NewSQLStore(ctx, dbConfig)
	defer store.GetDB().Master.Close()

	ctx = store.BeginTx(ctx)

	testcases := []struct {
		name    string
		ctx     context.Context
		isError bool
	}{
		{
			name:    "begin transaction is initiated",
			ctx:     ctx,
			isError: false,
		},
		{
			name:    "begin transaction has not been initiated",
			ctx:     context.Background(),
			isError: true,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.isError, store.CommitTx(tc.ctx) != nil)
		})
	}
}

func TestSQLStore_RollbackTx(t *testing.T) {
	cfg := config.All(`../../../`)
	ctx := testContext(t)
	dbConfig := cfg.Database.Master
	store, _ := sqlstore.NewSQLStore(ctx, dbConfig)
	defer store.GetDB().Master.Close()

	ctx = store.BeginTx(ctx)

	testcases := []struct {
		name    string
		ctx     context.Context
		isError bool
	}{
		{
			name:    "begin transaction is initiated",
			ctx:     ctx,
			isError: false,
		},
		{
			name:    "begin transaction has not been initiated",
			ctx:     context.Background(),
			isError: true,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.isError, store.RollbackTx(tc.ctx) != nil)
		})
	}
}
