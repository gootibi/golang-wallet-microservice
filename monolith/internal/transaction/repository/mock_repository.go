package repository

import (
	"context"
	"database/sql"

	transactionModel "github.com/gootibi/golang-wallet-microservice/monolith/internal/transaction/model"
	"github.com/stretchr/testify/mock"
)

type MockTransactionRepository struct {
	mock.Mock
}

func (m *MockTransactionRepository) CreateTX(ctx context.Context, tx *sql.Tx, t *transactionModel.Transaction) error {
	args := m.Called(ctx, tx, t)
	return args.Error(0)
}

func (m *MockTransactionRepository) GetByIdempotencyKey(ctx context.Context, key string) (*transactionModel.Transaction, error) {
	args := m.Called(ctx, key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*transactionModel.Transaction), args.Error(1)
}

func (m *MockTransactionRepository) GetHistory(ctx context.Context, walletID string, params transactionModel.PaginationParams) ([]transactionModel.Transaction, int64, error) {
	args := m.Called(ctx, walletID, params)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]transactionModel.Transaction), args.Get(1).(int64), args.Error(2)
}
