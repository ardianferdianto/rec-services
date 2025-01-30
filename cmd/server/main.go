package main

import (
	"context"
	"github.com/ardianferdianto/reconciliation-service/config"
	enum_parser "github.com/ardianferdianto/reconciliation-service/internal/domain/enum/parser"
	"github.com/ardianferdianto/reconciliation-service/internal/presenter/console"
	parser2 "github.com/ardianferdianto/reconciliation-service/internal/usecase/parser"
	"github.com/ardianferdianto/reconciliation-service/pkg/logger"
	"gopkg.in/ukautz/clif.v1"
	"os"
)

func main() {
	ctx := context.Background()
	conf, err := config.Init()
	if err != nil {
		os.Exit(1)
	}
	logger.InitializeLogger(logger.WithLevel(conf.Log.Level))

	// No need to run CLI if there is no argument
	if len(os.Args) == 1 {
		return
	}

	cli := clif.New("reconciliation-service", "1.0.0", "")
	cmdServer := console.ServerConsole{}
	cmdMigrate := console.NewMigrateConsole(conf.Database.Master)
	//cmdWorker := console.WorkerConsole{}
	cli.Add(cmdServer.StartServer())
	cli.Add(cmdMigrate.MigrateCreate())
	cli.Add(cmdMigrate.MigrateRun(ctx))
	cli.Add(cmdMigrate.MigrateRollback())
	//cli.Add(cmdWorker.StartWorker())
	cli.Run()
}

func init() {
	parser2.RegisterParser(enum_parser.BANK_STATEMENT, &parser2.BankStatementParser{})
	parser2.RegisterParser(enum_parser.SYSTEM_TRX, &parser2.SystemTxParser{})
}
