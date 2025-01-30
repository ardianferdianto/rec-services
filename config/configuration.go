package config

import (
	"context"
	"fmt"
	"github.com/ardianferdianto/reconciliation-service/pkg/config"
	"github.com/ardianferdianto/reconciliation-service/pkg/db"
	"github.com/ardianferdianto/reconciliation-service/pkg/logger"
	"log/slog"
)

type Configuration struct {
	App       AppConfiguration      `mapstructure:"app"`
	Server    ServerConfiguration   `mapstructure:"server"`
	Worker    WorkerConfiguration   `mapstructure:"worker"`
	Database  DatabaseConfiguration `mapstructure:"database"`
	Log       LogConfig             `mapstructure:"log"`
	BasicAuth []BasicAuthConfig     `mapstructure:"basic_auth"`
	Storage   StorageConfiguration  `mapstructure:"storage"`
}

type AppConfiguration struct {
	ENV                   string `mapstructure:"env"`
	ApiPrefix             string `mapstructure:"api_prefix"`
	Version               string `mapstructure:"version"`
	IsSingleDeployment    bool   `mapstructure:"is_single_deployment"`
	AccountInquiryEnabled bool   `mapstructure:"account_inquiry_enabled"`
	EnableDatadog         bool   `mapstructure:"enable_datadog"`
	Name                  string `mapstructure:"name"`
}

type ServerConfiguration struct {
	Port int `mapstructure:"port"`
}

type WorkerConfiguration struct {
	MaxWorkers string `mapstructure:"max_workers"`
}

type DatabaseConfiguration struct {
	Master db.Config `mapstructure:"master"`
}

type LogConfig struct {
	Level string `mapstructure:"level"`
}

type BasicAuthConfig struct {
	ClientID     string `mapstructure:"client_id"`
	ClientSecret string `mapstructure:"client_secret"`
}

type StorageConfiguration struct {
	Endpoint     string `mapstructure:"endpoint"`
	ClientID     string `mapstructure:"client_id"`
	ClientSecret string `mapstructure:"client_secret"`
	Bucket       string `mapstructure:"bucket"`
}

var (
	configuration *Configuration
)

func InitWithPath(path string) (*Configuration, error) {
	err := config.New(config.WithPath(path)).Read(&configuration)
	if err != nil {
		slog.ErrorContext(context.Background(), "failed to initiate config", logger.ErrAttr(err))
		return nil, err
	}
	slog.DebugContext(context.Background(), "loaded configuration", fmt.Sprintf("%+v", configuration))
	return configuration, nil
}

func Init() (*Configuration, error) {
	return InitWithPath(".")
}

func Get() *Configuration {
	return configuration
}

func GetCredentials() map[string]string {
	creds := make(map[string]string)
	for _, v := range configuration.BasicAuth {
		creds[v.ClientID] = v.ClientSecret
	}
	return creds
}
