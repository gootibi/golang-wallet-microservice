package repository

import (
	"context"
	"database/sql"

	"github.com/gootibi/golang-wallet-microservice/monolith/internal/wallet/model"
	"github.com/stretchr/testify/mock"
)

type MockWalletRepository struct {
	mock.Mock
}

func (m *MockWalletRepository) CreateTx(ctx context.Context, tx *sql.Tx, w *model.Wallet) error {
	args := m.Called(ctx, tx, w)
	return args.Error(0)
}

func (m *MockWalletRepository) GetByUserID(ctx context.Context, userID string) (*model.Wallet, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*model.Wallet), args.Error(1)
}

func (m *MockWalletRepository) UpdateBalanceTx(ctx context.Context, tx *sql.Tx, walletID string, newBalance float64, currentWersion int) error {
	args := m.Called(ctx, tx, walletID, newBalance, currentWersion)
	return args.Error(0)
}
