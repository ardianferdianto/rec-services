package logger_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/ardianferdianto/reconciliation-service/pkg/contextprop"
	"github.com/ardianferdianto/reconciliation-service/pkg/logger"
	"github.com/stretchr/testify/assert"
	"time"

	"log/slog"
	"testing"
)

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		input         string
		expectedLevel slog.Level
	}{
		{"DEBUG", slog.LevelDebug},
		{"INFO", slog.LevelInfo},
		{"WARN", slog.LevelWarn},
		{"ERROR", slog.LevelError},
		{"ANY", slog.LevelWarn},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			level := logger.ParseLevel(tt.input)
			assert.Equal(t, tt.expectedLevel, level)
		})
	}
}

type log struct {
	Time          string `json:"time"`
	Level         string `json:"level"`
	Msg           string `json:"msg"`
	Error         string `json:"error"`
	Source        string `json:"source"`
	CorrelationID string `json:"correlation_id"`
	Pin           string `json:"pin"`
	User          string `json:"user"`
}

func TestLogger(t *testing.T) {
	buffer := bytes.NewBuffer(nil)
	defer buffer.Reset()
	logger.InitializeLogger(
		logger.WithWriter(buffer),
		logger.WithLevel("INFO"),
	)

	tests := []struct {
		name        string
		ex          func()
		expectedLog log
	}{
		{"InfoLog", func() { slog.Info("foo") }, log{Msg: "foo", Level: "INFO"}},
		{"ErrorLogShouldHaveSource", func() { slog.Error("bar") }, log{Msg: "bar", Level: "ERROR", Source: "github.com/ardianferdianto/reconciliation-service/pkg/logger_test.TestLogger.func2:61"}},
		{"ErrorLogWithErrorAttribute", func() {
			slog.Error("bar", logger.ErrAttr(fmt.Errorf("invalid")))
		}, log{Msg: "bar", Level: "ERROR", Error: "invalid", Source: "github.com/ardianferdianto/reconciliation-service/pkg/logger_test.TestLogger.func3:63"}},
		{"InfoLogWithCorrelationID", func() {
			slog.InfoContext(context.WithValue(context.Background(), contextprop.CorrelationIDKey, "bar"), "foo")
		}, log{Msg: "foo", Level: "INFO", CorrelationID: "bar"}},
		{"InfoLogWithSensitiveKey", func() {
			slog.InfoContext(context.WithValue(context.Background(), contextprop.CorrelationIDKey, "foo"), "bar", slog.String("pin", "123456"))
		}, log{Msg: "bar", Level: "INFO", CorrelationID: "foo", Pin: "*"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer buffer.Reset()
			tt.ex()

			log := &log{}
			err := json.Unmarshal(buffer.Bytes(), log)
			assert.NoError(t, err)

			assert.Equal(t, tt.expectedLog.Msg, log.Msg)
			assert.Equal(t, tt.expectedLog.Level, log.Level)
			assert.NotEmpty(t, log.Time)
			_, err = time.Parse("2006-01-02T15:04:05.000", log.Time)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedLog.Error, log.Error)
			assert.Equal(t, tt.expectedLog.CorrelationID, log.CorrelationID)
			assert.Equal(t, tt.expectedLog.Source, log.Source)
			assert.Equal(t, tt.expectedLog.Pin, log.Pin)
		})
	}
}

func TestWithLoggerShouldNotOverrideAttributeFromHandler(t *testing.T) {
	buffer := bytes.NewBuffer(nil)
	logger.InitializeLogger(
		logger.WithWriter(buffer),
	)

	ctx := context.WithValue(context.Background(), contextprop.CorrelationIDKey, "corr-id")
	slog.With(slog.String("pin", "123456")).ErrorContext(ctx, "foo message", slog.String("user", "CUSTOMER"))

	log := &log{}
	err := json.Unmarshal(buffer.Bytes(), log)
	assert.NoError(t, err)

	assert.Equal(t, "foo message", log.Msg)
	assert.Equal(t, slog.LevelError.String(), log.Level)
	assert.NotEmpty(t, log.Time)
	_, err = time.Parse("2006-01-02T15:04:05.000", log.Time)
	assert.NoError(t, err)
	assert.Equal(t, "corr-id", log.CorrelationID)
	assert.Equal(t, "CUSTOMER", log.User)
	assert.Equal(t, "*", log.Pin)
	assert.NotEmpty(t, log.Source)
}
