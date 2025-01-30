// Code generated by MockGen. DO NOT EDIT.
// Source: reconciliation_repository.go

// Package mock_repository is a generated GoMock package.
package mock_repository

import (
	context "context"
	reflect "reflect"

	domain "github.com/ardianferdianto/reconciliation-service/internal/domain"
	gomock "github.com/golang/mock/gomock"
)

// MockReconciliationRepository is a mock of ReconciliationRepository interface.
type MockReconciliationRepository struct {
	ctrl     *gomock.Controller
	recorder *MockReconciliationRepositoryMockRecorder
}

// MockReconciliationRepositoryMockRecorder is the mock recorder for MockReconciliationRepository.
type MockReconciliationRepositoryMockRecorder struct {
	mock *MockReconciliationRepository
}

// NewMockReconciliationRepository creates a new mock instance.
func NewMockReconciliationRepository(ctrl *gomock.Controller) *MockReconciliationRepository {
	mock := &MockReconciliationRepository{ctrl: ctrl}
	mock.recorder = &MockReconciliationRepositoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockReconciliationRepository) EXPECT() *MockReconciliationRepositoryMockRecorder {
	return m.recorder
}

// CreateJob mocks base method.
func (m *MockReconciliationRepository) CreateJob(ctx context.Context, job domain.ReconciliationJob) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateJob", ctx, job)
	ret0, _ := ret[0].(error)
	return ret0
}

// CreateJob indicates an expected call of CreateJob.
func (mr *MockReconciliationRepositoryMockRecorder) CreateJob(ctx, job interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateJob", reflect.TypeOf((*MockReconciliationRepository)(nil).CreateJob), ctx, job)
}

// StoreMatchedRecord mocks base method.
func (m *MockReconciliationRepository) StoreMatchedRecord(ctx context.Context, rec domain.MatchedRecord) (int, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "StoreMatchedRecord", ctx, rec)
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// StoreMatchedRecord indicates an expected call of StoreMatchedRecord.
func (mr *MockReconciliationRepositoryMockRecorder) StoreMatchedRecord(ctx, rec interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "StoreMatchedRecord", reflect.TypeOf((*MockReconciliationRepository)(nil).StoreMatchedRecord), ctx, rec)
}

// StoreResult mocks base method.
func (m *MockReconciliationRepository) StoreResult(ctx context.Context, result domain.ReconciliationResult) (int, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "StoreResult", ctx, result)
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// StoreResult indicates an expected call of StoreResult.
func (mr *MockReconciliationRepositoryMockRecorder) StoreResult(ctx, result interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "StoreResult", reflect.TypeOf((*MockReconciliationRepository)(nil).StoreResult), ctx, result)
}

// StoreUnmatchedBankTx mocks base method.
func (m *MockReconciliationRepository) StoreUnmatchedBankTx(ctx context.Context, txList []domain.UnmatchedBankTx) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "StoreUnmatchedBankTx", ctx, txList)
	ret0, _ := ret[0].(error)
	return ret0
}

// StoreUnmatchedBankTx indicates an expected call of StoreUnmatchedBankTx.
func (mr *MockReconciliationRepositoryMockRecorder) StoreUnmatchedBankTx(ctx, txList interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "StoreUnmatchedBankTx", reflect.TypeOf((*MockReconciliationRepository)(nil).StoreUnmatchedBankTx), ctx, txList)
}

// StoreUnmatchedSystemTx mocks base method.
func (m *MockReconciliationRepository) StoreUnmatchedSystemTx(ctx context.Context, txList []domain.UnmatchedSystemTx) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "StoreUnmatchedSystemTx", ctx, txList)
	ret0, _ := ret[0].(error)
	return ret0
}

// StoreUnmatchedSystemTx indicates an expected call of StoreUnmatchedSystemTx.
func (mr *MockReconciliationRepositoryMockRecorder) StoreUnmatchedSystemTx(ctx, txList interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "StoreUnmatchedSystemTx", reflect.TypeOf((*MockReconciliationRepository)(nil).StoreUnmatchedSystemTx), ctx, txList)
}
