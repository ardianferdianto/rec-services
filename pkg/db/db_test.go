package db

import (
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
)

func TestConfig_URL(t *testing.T) {
	cfg := Config{
		Host:              "localhost",
		Port:              "5432",
		User:              "baskara",
		Password:          "foo",
		Name:              "dbname",
		MaxOpen:           10,
		MinIdle:           20,
		MaxLifetime:       time.Duration(30) * time.Minute,
		MaxIdleTime:       time.Duration(5) * time.Minute,
		ConnectionTimeout: 11,
		StatementTimeout:  12,
	}

	assert.Equal(t,
		"postgres://baskara:foo@localhost:5432/dbname?connect_timeout=11&pool_min_conns=20&pool_max_conns=10&pool_max_conn_lifetime=30m0s&pool_max_conn_idle_time=5m0s",
		cfg.URL())
}

func TestConfig_GetPgxPoolConfig(t *testing.T) {
	cfg := Config{
		Host:              "localhost",
		Port:              "5432",
		User:              "baskara",
		Password:          "foo",
		Name:              "dbname",
		MaxOpen:           10,
		MinIdle:           20,
		MaxLifetime:       time.Duration(30) * time.Minute,
		MaxIdleTime:       time.Duration(5) * time.Minute,
		ConnectionTimeout: 11,
	}

	poolCfg, err := cfg.GetPgxPoolConfig()
	assert.Nil(t, err)

	want := &pgxpool.Config{
		MaxConns:        int32(cfg.MaxOpen),
		MinConns:        int32(cfg.MinIdle),
		MaxConnLifetime: cfg.MaxLifetime,
		MaxConnIdleTime: cfg.MaxIdleTime,
		ConnConfig: &pgx.ConnConfig{
			Config: pgconn.Config{
				ConnectTimeout: time.Duration(cfg.ConnectionTimeout) * time.Second,
			},
		},
	}
	assert.Equal(t, want.MinConns, poolCfg.MinConns)
	assert.Equal(t, want.MaxConns, poolCfg.MaxConns)
	assert.Equal(t, want.MaxConnLifetime, poolCfg.MaxConnLifetime)
	assert.Equal(t, want.MaxConnIdleTime, poolCfg.MaxConnIdleTime)
	assert.Equal(t, want.ConnConfig.ConnectTimeout, poolCfg.ConnConfig.ConnectTimeout)
}

func TestConfig_GetPgxPoolConfig_WithDefaultValue(t *testing.T) {
	cfg := Config{}

	poolCfg, err := cfg.GetPgxPoolConfig()
	assert.Nil(t, err)

	want := &pgxpool.Config{
		MaxConns:        int32(MaxOpen),
		MinConns:        int32(MinIdle),
		MaxConnLifetime: MaxLifetime,
		MaxConnIdleTime: MaxIdleTime,
		ConnConfig: &pgx.ConnConfig{
			Config: pgconn.Config{
				ConnectTimeout: time.Duration(ConnectionTimeout) * time.Second,
			},
		},
	}
	assert.Equal(t, want.MinConns, poolCfg.MinConns)
	assert.Equal(t, want.MaxConns, poolCfg.MaxConns)
	assert.Equal(t, want.MaxConnLifetime, poolCfg.MaxConnLifetime)
	assert.Equal(t, want.MaxConnIdleTime, poolCfg.MaxConnIdleTime)
	assert.Equal(t, want.ConnConfig.ConnectTimeout, poolCfg.ConnConfig.ConnectTimeout)
}
