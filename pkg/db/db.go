package db

import (
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

const (
	ConnectionTimeout = 5
	MinIdle           = 1
	MaxOpen           = 5
	MaxLifetime       = 1 * time.Minute
	MaxIdleTime       = 1 * time.Minute
)

type Config struct {
	Host              string        `mapstructure:"host"`
	Port              string        `mapstructure:"port"`
	User              string        `mapstructure:"user"`
	Password          string        `mapstructure:"password"`
	Name              string        `mapstructure:"name"`
	MaxOpen           int           `mapstructure:"max_open"`      // maximum open connection
	MinIdle           int           `mapstructure:"min_idle"`      // minimum idle connection
	MaxLifetime       time.Duration `mapstructure:"max_lifetime"`  // in duration, connections are closed and replaced every this duration
	MaxIdleTime       time.Duration `mapstructure:"max_idle_time"` // in duration, idle connections are closed after this duration
	ParseTime         bool
	Driver            string `mapstructure:"driver"`
	ConnectionTimeout int    `mapstructure:"connection_timeout"`
	StatementTimeout  int    `mapstructure:"statement_timeout"` // in seconds, any value below 2 will be interpreted as 2, ref: https://www.postgresql.org/docs/current/libpq-connect.html#LIBPQ-CONNECT-CONNECT-TIMEOUT
}

func (c *Config) initDefault() {
	if c.ConnectionTimeout == 0 {
		c.ConnectionTimeout = ConnectionTimeout
	}
	if c.MinIdle == 0 {
		c.MinIdle = MinIdle
	}
	if c.MaxOpen == 0 {
		c.MaxOpen = MaxOpen
	}
	if c.MaxLifetime == 0 {
		c.MaxLifetime = MaxLifetime
	}
	if c.MaxIdleTime == 0 {
		c.MaxIdleTime = MaxIdleTime
	}
}

func (c *Config) URL() string {
	c.initDefault()
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?connect_timeout=%d&pool_min_conns=%d&pool_max_conns=%d&pool_max_conn_lifetime=%v&pool_max_conn_idle_time=%v&sslmode=disable",
		c.User,
		c.Password,
		c.Host,
		c.Port,
		c.Name,
		c.ConnectionTimeout,
		c.MinIdle,
		c.MaxOpen,
		c.MaxLifetime,
		c.MaxIdleTime,
	)
}

func (c *Config) GetPgxPoolConfig() (*pgxpool.Config, error) {
	c.initDefault()
	return pgxpool.ParseConfig(c.URL())
}
