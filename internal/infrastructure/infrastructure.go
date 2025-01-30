package infrastructure

import (
	"context"
	"github.com/ardianferdianto/reconciliation-service/config"
	"github.com/ardianferdianto/reconciliation-service/internal/infrastructure/sqlstore"
	"github.com/ardianferdianto/reconciliation-service/pkg/logger"
	"log"
	"log/slog"
)

type Infrastructure interface {
	SQLStore() sqlstore.Store
	Minio() *MinioClient
}

type Infra struct {
	sqlStore    sqlstore.Store
	minioClient *MinioClient
}

func NewInfra(ctx context.Context, config config.Configuration) (Infrastructure, error) {
	sqlStore, err := sqlstore.NewSQLStore(ctx, config.Database.Master)
	if err != nil {
		slog.ErrorContext(ctx, "error when initializing sql store", logger.ErrAttr(err))
		return nil, err
	}

	storageConf := config.Storage
	minioCl, err := NewMinioClient(storageConf.Endpoint, storageConf.ClientID, storageConf.ClientSecret, false, storageConf.Bucket)
	if err != nil {
		log.Fatalf("NewMinioClient error: %v", err)
	}

	return &Infra{
		sqlStore:    sqlStore,
		minioClient: minioCl,
	}, nil
}

func (i *Infra) SQLStore() sqlstore.Store {
	return i.sqlStore.GetDB()
}

func (i *Infra) Minio() *MinioClient {
	return i.minioClient
}
