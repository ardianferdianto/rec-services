package parser

import (
	"fmt"
	"github.com/ardianferdianto/reconciliation-service/internal/domain"
	"strconv"
	"strings"
	"time"
)

type BankStatementParser struct{}

func (b *BankStatementParser) ParseLine(fields []string) (interface{}, error) {
	if len(fields) < 4 {
		return nil, fmt.Errorf("not enough columns for BCA bank statement")
	}
	amt, err := strconv.ParseFloat(fields[1], 64)
	if err != nil {
		return nil, fmt.Errorf("parse amount error: %w", err)
	}
	dt, err := time.Parse("2006-01-02", strings.TrimSpace(fields[2]))
	if err != nil {
		return nil, fmt.Errorf("parse date error: %w", err)
	}
	return domain.BankStatement{
		UniqueID:      fields[0],
		Amount:        amt,
		StatementTime: dt,
		BankCode:      fields[3],
	}, nil
}
