package config_test

import (
	"github.com/ardianferdianto/reconciliation-service/pkg/config"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type testConfig struct {
	Name     string
	Port     int
	Upstream struct {
		Url     string
		Timeout time.Duration
		ApiKey  string `mapstructure:"api_key"`
	}
	Db struct {
		Host     string
		Port     int
		Username string
		Password string
	}
}

func TestLoadConfigFromFile(t *testing.T) {
	var cfg testConfig
	err := config.New(config.WithPath("./testdata")).Read(&cfg)
	assert.NoError(t, err)

	assert.Equal(t, "fooApp", cfg.Name)
	assert.Equal(t, 9090, cfg.Port)
	assert.Equal(t, "http://localhost:8181", cfg.Upstream.Url)
	assert.Equal(t, 2*time.Second, cfg.Upstream.Timeout)
	assert.Equal(t, "dummy", cfg.Upstream.ApiKey)
	assert.Equal(t, "localhost", cfg.Db.Host)
	assert.Equal(t, 5432, cfg.Db.Port)
	assert.Equal(t, "foo", cfg.Db.Username)
	assert.Equal(t, "bar", cfg.Db.Password)
}
