package console

import (
	"context"
	"github.com/ardianferdianto/reconciliation-service/config"
	"github.com/ardianferdianto/reconciliation-service/internal/infrastructure"
	"github.com/ardianferdianto/reconciliation-service/internal/repository"
	"github.com/ardianferdianto/reconciliation-service/internal/usecase/ingestion"
	"gopkg.in/ukautz/clif.v1"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"syscall"
	"time"
)

type WorkerConsole struct{}

func (c *WorkerConsole) StartWorker() *clif.Command {
	return clif.NewCommand("start", "starting workers.", func(o *clif.Command, in clif.Input, out clif.Output) error {
		ctx := context.Background()
		slog.InfoContext(ctx, "Runtime go version "+runtime.Version())
		conf := config.Get()
		infra, err := infrastructure.NewInfra(ctx, *conf)
		if err != nil {
			return err
		}

		workerConcurrency := 1
		if val := conf.Worker.MaxWorkers; val != "" {
			if v, err := strconv.Atoi(val); err == nil && v > 0 {
				workerConcurrency = v
			}
		}
		jobRepo := repository.NewIngestionRepo(infra.SQLStore())
		dataRepo := repository.NewDataRepo(infra.SQLStore())

		ingestionUC := ingestion.NewIngestionUseCase(jobRepo, dataRepo, infra.Minio())

		log.Printf("Starting worker with concurrency = %d\n", workerConcurrency)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		go ingestion.WorkerIngestionLoop(ctx, ingestionUC, jobRepo, workerConcurrency)

		// wait for signal
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		log.Println("Shutting down worker gracefully...")
		slog.InfoContext(ctx, "shutting down")
		time.Sleep(5 * time.Second)

		return nil
	})
}

func init() {

}
