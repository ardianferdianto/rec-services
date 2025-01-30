package parser

import (
	"fmt"
	"github.com/ardianferdianto/reconciliation-service/internal/domain"
	"strconv"
	"strings"
	"time"
)

type SystemTxParser struct{}

func (p *SystemTxParser) ParseLine(fields []string) (interface{}, error) {
	if len(fields) < 4 {
		return nil, fmt.Errorf("not enough columns for system tx")
	}
	amt, err := strconv.ParseFloat(fields[1], 64)
	if err != nil {
		return nil, fmt.Errorf("parse amount error: %w", err)
	}
	dt, err := time.Parse("2006-01-02 15:04:05", strings.TrimSpace(fields[3]))
	if err != nil {
		return nil, fmt.Errorf("parse datetime error: %w", err)
	}
	return domain.Transaction{
		TrxID:           fields[0],
		Amount:          amt,
		Type:            fields[2],
		TransactionTime: dt,
	}, nil
}
