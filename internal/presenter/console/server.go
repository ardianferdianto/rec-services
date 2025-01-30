package console

import (
	"context"
	"errors"
	"fmt"
	"github.com/ardianferdianto/reconciliation-service/config"
	"github.com/ardianferdianto/reconciliation-service/internal/infrastructure"
	"github.com/ardianferdianto/reconciliation-service/internal/presenter/rest"
	"github.com/ardianferdianto/reconciliation-service/internal/repository"
	"github.com/ardianferdianto/reconciliation-service/internal/usecase/ingestion"
	"github.com/ardianferdianto/reconciliation-service/internal/usecase/reconcile"
	"github.com/ardianferdianto/reconciliation-service/internal/usecase/workflow"
	"github.com/ardianferdianto/reconciliation-service/pkg/logger"
	"github.com/ardianferdianto/reconciliation-service/pkg/middleware"
	"gopkg.in/ukautz/clif.v1"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/gorilla/mux"
)

type ServerConsole struct{}

func (c *ServerConsole) StartServer() *clif.Command {
	return clif.NewCommand("start", "starting http server.", func(o *clif.Command, in clif.Input, out clif.Output) error {
		ctx := context.Background()
		slog.InfoContext(ctx, "Runtime go version "+runtime.Version())
		conf := config.Get()
		infra, err := infrastructure.NewInfra(ctx, *conf)
		if err != nil {
			return err
		}

		routes := SetupRoute(infra)
		server := http.Server{
			Handler: routes,
			Addr:    fmt.Sprintf(":%d", conf.Server.Port),
		}

		gracefulShutdown := make(chan os.Signal, 1)
		signal.Notify(gracefulShutdown, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

		go func() {
			slog.InfoContext(ctx, "server starting on port "+fmt.Sprintf("%d", conf.Server.Port))
			if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				slog.ErrorContext(ctx, "failed to start server", logger.ErrAttr(err))
			}
		}()

		<-gracefulShutdown
		slog.InfoContext(ctx, "shutting down")
		time.Sleep(5 * time.Second)
		err = server.Shutdown(ctx)
		if err != nil {
			slog.ErrorContext(ctx, "error shutting down server", logger.ErrAttr(err))
		}

		return nil
	})
}

func SetupRoute(infra infrastructure.Infrastructure) *mux.Router {
	conf := config.Get()
	wfRepo := repository.NewWorkflowRepo(infra.SQLStore())
	dtRepo := repository.NewDataRepo(infra.SQLStore())
	ingRepo := repository.NewIngestionRepo(infra.SQLStore())
	recRepo := repository.NewReconciliationRepo(infra.SQLStore())

	ingestionUC := ingestion.NewIngestionUseCase(ingRepo, dtRepo, infra.Minio())
	reconcileUC := reconcile.NewReconciliationUseCase(recRepo, dtRepo)
	workflowUC := workflow.NewWorkflowUseCase(wfRepo, ingestionUC, reconcileUC)

	baseRouter := mux.NewRouter()
	baseRouter.NotFoundHandler = http.HandlerFunc(rest.NotFoundHandler)
	baseRouter.HandleFunc("/ping", rest.PingHandler).Methods(http.MethodGet)

	apiRouter := baseRouter.PathPrefix(fmt.Sprintf("/%s/v1", conf.App.ApiPrefix)).Subrouter()
	apiRouter.Use(middleware.BasicAuthMiddleware(config.GetCredentials()))
	apiRouter.Use(middleware.RecoveryHandler())

	workflowHandler := rest.NewWorkflowHandler(workflowUC, reconcileUC)
	apiRouter.HandleFunc("/workflow", workflowHandler.StartWorkflowHandler).Methods(http.MethodPost)
	apiRouter.HandleFunc("/workflow/{workflowID}", workflowHandler.GetWorkflowSummary).Methods(http.MethodGet)

	return baseRouter
}
