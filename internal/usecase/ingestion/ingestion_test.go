package ingestion

import (
	"context"
	"github.com/ardianferdianto/reconciliation-service/internal/domain"
	mock_infrastructure "github.com/ardianferdianto/reconciliation-service/internal/infrastructure/_mock"
	mock_repository "github.com/ardianferdianto/reconciliation-service/internal/repository/_mock"
	"github.com/ardianferdianto/reconciliation-service/internal/usecase/parser/_mock"
	"github.com/golang/mock/gomock"
	"github.com/minio/minio-go/v7"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestIngestCSVJob(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockJobRepo := mock_repository.NewMockIngestionJobRepository(ctrl)
	mockDataRepo := mock_repository.NewMockDataRepository(ctrl)
	mockMinioClient := mock_infrastructure.NewMockIMinioClient(ctrl)
	mockParser := mock_parser.NewMockCSVParser(ctrl)

	ctx := context.Background()
	bucket := "test-bucket"
	job := &domain.IngestionJob{
		JobID:    "job123",
		FileName: "test.csv",
		FileType: "csv",
	}

	// Mock the parser response
	mockParser.EXPECT().ParseLine(gomock.Any()).DoAndReturn(func(record []string) (interface{}, error) {
		// Simulate parsing
		return domain.Transaction{ID: 123, TrxID: "trx123"}, nil
	}).AnyTimes()

	// Mock the Minio client response
	//csvContent := "header1,header2\nvalue1,value2\n"
	mockMinioClient.EXPECT().GetObject(ctx, job.FileName, minio.GetObjectOptions{}).Return(gomock.Any(), nil)

	// Mock the repository methods
	mockJobRepo.EXPECT().UpdateJobProgress(ctx, gomock.Any(), gomock.Any(), "IN_PROGRESS").AnyTimes()
	mockJobRepo.EXPECT().UpdateJobProgress(ctx, gomock.Any(), gomock.Any(), "COMPLETED").AnyTimes()
	mockDataRepo.EXPECT().BatchInsertSystemTx(ctx, gomock.Any()).Return(nil).AnyTimes()
	mockDataRepo.EXPECT().BatchInsertBankStmts(ctx, gomock.Any()).Return(nil).AnyTimes()

	// Execute the function
	err := ingestCSVJob(ctx, mockJobRepo, mockDataRepo, mockMinioClient, bucket, job)

	// Assert no error
	assert.NoError(t, err)
}
