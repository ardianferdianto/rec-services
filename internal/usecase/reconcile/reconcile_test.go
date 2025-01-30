package reconcile

import (
	"context"
	"github.com/ardianferdianto/reconciliation-service/internal/domain"
	mock_repository "github.com/ardianferdianto/reconciliation-service/internal/repository/_mock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

type ReconcileUseCaseSuite struct {
	suite.Suite
	mockDataRepo *mock_repository.MockDataRepository
	mockRecRepo  *mock_repository.MockReconciliationRepository
	uc           IUseCase
	controller   *gomock.Controller
}

func (suite *ReconcileUseCaseSuite) SetupTest() {
	suite.controller = gomock.NewController(suite.T())
	suite.mockDataRepo = mock_repository.NewMockDataRepository(suite.controller)
	suite.mockRecRepo = mock_repository.NewMockReconciliationRepository(suite.controller)
	suite.uc = NewReconciliationUseCase(suite.mockRecRepo, suite.mockDataRepo)
}

func (suite *ReconcileUseCaseSuite) TearDownTest() {
	suite.controller.Finish()
}

func (suite *ReconcileUseCaseSuite) TestProcessReconciliation() {
	ctx := context.Background()
	startDate, endDate := time.Date(2021, 01, 01, 0, 0, 0, 0, time.UTC), time.Date(2021, 01, 31, 0, 0, 0, 0, time.UTC)

	testCases := []struct {
		name          string
		setupMocks    func()
		expectedError bool
		verifyResult  func(result domain.ReconciliationResult)
	}{
		{
			name: "Perfect Match",
			setupMocks: func() {
				transactions := []domain.Transaction{{ID: 1, TrxID: "TX1001", Amount: 100.0, TransactionTime: startDate}}
				statements := []domain.BankStatement{{ID: 1, UniqueID: "TX1001", Amount: 100.0, StatementTime: startDate}}

				suite.mockDataRepo.EXPECT().FindSystemTxByDateRange(ctx, startDate, endDate).Return(transactions, nil)
				suite.mockDataRepo.EXPECT().FindBankStmtsByDateRange(ctx, startDate, endDate).Return(statements, nil)
				suite.mockRecRepo.EXPECT().CreateJob(ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, job domain.ReconciliationJob) error {
					return nil
				})
				suite.mockRecRepo.EXPECT().StoreMatchedRecord(ctx, gomock.Any()).Return(1, nil)
				suite.mockRecRepo.EXPECT().StoreUnmatchedSystemTx(ctx, gomock.Any()).Return(nil)
				suite.mockRecRepo.EXPECT().StoreUnmatchedBankTx(ctx, gomock.Any()).Return(nil)
				suite.mockRecRepo.EXPECT().StoreResult(ctx, gomock.Any()).Return(1, nil)
			},
			expectedError: false,
			verifyResult: func(result domain.ReconciliationResult) {
				suite.Regexp(`\b[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}\b`, result.JobID)
				suite.Equal(1, result.MatchedCount)
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			tc.setupMocks()
			result, err := suite.uc.ProcessReconciliation(ctx, startDate, endDate)
			if tc.expectedError {
				suite.NotNil(err)
			} else {
				suite.NoError(err)
				tc.verifyResult(result)
			}
		})
	}
}

func TestReconcileUseCaseSuite(t *testing.T) {
	suite.Run(t, new(ReconcileUseCaseSuite))
}
