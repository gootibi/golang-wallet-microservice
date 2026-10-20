package service

import (
	"context"
	"errors"
	"testing"

	ledgerModel "github.com/gootibi/golang-wallet-microservice/monolith/internal/ledger/model"
	ledgerRepository "github.com/gootibi/golang-wallet-microservice/monolith/internal/ledger/repository"
	walletModel "github.com/gootibi/golang-wallet-microservice/monolith/internal/wallet/model"
	walletRepository "github.com/gootibi/golang-wallet-microservice/monolith/internal/wallet/repository"
	"github.com/stretchr/testify/assert"
)

func TestReconcileWalletBalance_Consistent(t *testing.T) {
	// Initiate mock reposirories
	mockLedgerRepo := new(ledgerRepository.MockLedgerRepository)
	mockWalletRepo := new(walletRepository.MockWalletRepository)

	// Initiate new wallet service
	svc := NewLedgerService(mockLedgerRepo, mockWalletRepo)

	ctx := context.TODO()
	userID := "user-uuid"
	wallet := &walletModel.Wallet{
		ID:      "wallet-uuid",
		Balance: 600.0,
	}

	// Expect mock
	mockWalletRepo.On("GetByUserID", ctx, userID).Return(wallet, nil)
	mockLedgerRepo.On("GetBalanceByWalletID", ctx, wallet.ID).Return(600.0, nil)

	// Run the function
	isConsistent, walletBalance, calculatedBalance, err := svc.ReconcileWalletBalance(ctx, userID)

	// Verify result
	assert.NoError(t, err)
	assert.True(t, isConsistent)
	assert.Equal(t, 600.0, walletBalance)
	assert.Equal(t, 600.0, calculatedBalance)

	// Make sure all mock if called and expected
	mockLedgerRepo.AssertExpectations(t)
	mockWalletRepo.AssertExpectations(t)
}

func TestReconcileWalletBalance_Discrepancy(t *testing.T) {
	// Initiate mock reposirories
	mockLedgerRepo := new(ledgerRepository.MockLedgerRepository)
	mockWalletRepo := new(walletRepository.MockWalletRepository)

	// Initiate new wallet service
	svc := NewLedgerService(mockLedgerRepo, mockWalletRepo)

	ctx := context.TODO()
	userID := "user-uuid"
	wallet := &walletModel.Wallet{
		ID:      "wallet-uuid",
		Balance: 600.0,
	}

	// Expect mock
	mockWalletRepo.On("GetByUserID", ctx, userID).Return(wallet, nil)
	mockLedgerRepo.On("GetBalanceByWalletID", ctx, wallet.ID).Return(500.0, nil)

	// Run the function
	isConsistent, walletBalance, calculatedBalance, err := svc.ReconcileWalletBalance(ctx, userID)

	// Verify result
	assert.NoError(t, err)
	assert.False(t, isConsistent)
	assert.Equal(t, 600.0, walletBalance)
	assert.Equal(t, 500.0, calculatedBalance)

	// Make sure all mock if called and expected
	mockLedgerRepo.AssertExpectations(t)
	mockWalletRepo.AssertExpectations(t)
}

func TestReconcileWalletBalance_WalletError_NotFound(t *testing.T) {
	// Initiate mock reposirories
	mockLedgerRepo := new(ledgerRepository.MockLedgerRepository)
	mockWalletRepo := new(walletRepository.MockWalletRepository)

	// Initiate new wallet service
	svc := NewLedgerService(mockLedgerRepo, mockWalletRepo)

	ctx := context.TODO()
	userID := "user-uuid"

	// Expect mock
	mockWalletRepo.On("GetByUserID", ctx, userID).Return(nil, errors.New("db error"))

	// Run the function
	isConsistent, walletBalance, calculatedBalance, err := svc.ReconcileWalletBalance(ctx, userID)

	// Verify result
	assert.Error(t, err)
	assert.False(t, isConsistent)
	assert.Equal(t, 0.0, walletBalance)
	assert.Equal(t, 0.0, calculatedBalance)

	// Make sure all mock if called and expected
	mockLedgerRepo.AssertExpectations(t)
	mockWalletRepo.AssertExpectations(t)
}

func TestReconcileWalletBalance_LedgerError_NotFound(t *testing.T) {
	// Initiate mock reposirories
	mockLedgerRepo := new(ledgerRepository.MockLedgerRepository)
	mockWalletRepo := new(walletRepository.MockWalletRepository)

	// Initiate new wallet service
	svc := NewLedgerService(mockLedgerRepo, mockWalletRepo)

	ctx := context.TODO()
	userID := "suer-uuid"
	wallet := &walletModel.Wallet{
		ID:      "wallet-uuid",
		Balance: 600.0,
	}

	// Expect mock
	mockWalletRepo.On("GetByUserID", ctx, userID).Return(wallet, nil)
	mockLedgerRepo.On("GetBalanceByWalletID", ctx, wallet.ID).Return(0.0, errors.New("ledger error"))

	// Run the function
	isConsistent, walletBalance, calculatedBalance, err := svc.ReconcileWalletBalance(ctx, userID)

	// Verify result
	assert.Error(t, err)
	assert.False(t, isConsistent)
	assert.Equal(t, 0.0, walletBalance)
	assert.Equal(t, 0.0, calculatedBalance)

	// Make sure all mock if called and expected
	mockLedgerRepo.AssertExpectations(t)
	mockWalletRepo.AssertExpectations(t)
}

func TestGetMutationHistory_Success(t *testing.T) {
	// Initiate mock reposirories
	mockLedgerRepo := new(ledgerRepository.MockLedgerRepository)
	mockWalletRepo := new(walletRepository.MockWalletRepository)

	// Initiate new wallet service
	svc := NewLedgerService(mockLedgerRepo, mockWalletRepo)

	ctx := context.TODO()
	userID := "user-uuid"
	wallet := &walletModel.Wallet{
		ID: "wallet-uuid",
	}
	expectedEtries := []ledgerModel.LedgerEntry{
		{
			ID:        "entry-uuid",
			WalletID:  "wallet-uuid",
			EntryType: "credit",
			Amount:    200,
		},
	}

	// Expect mock
	mockWalletRepo.On("GetByUserID", ctx, userID).Return(wallet, nil)
	mockLedgerRepo.On("GetEntriesByWalletID", ctx, wallet.ID).Return(expectedEtries, nil)

	// Run the function
	entries, err := svc.GetMutationHistory(ctx, userID)

	// Verify result
	assert.NoError(t, err)
	assert.Equal(t, expectedEtries, entries)

	// Make sure all mock if called and expected
	mockLedgerRepo.AssertExpectations(t)
	mockWalletRepo.AssertExpectations(t)
}

func TestGetMutationHistory_WalletError_NotFound(t *testing.T) {
	// Initiate mock reposirories
	mockLedgerRepo := new(ledgerRepository.MockLedgerRepository)
	mockWalletRepo := new(walletRepository.MockWalletRepository)

	// Initiate new wallet service
	svc := NewLedgerService(mockLedgerRepo, mockWalletRepo)

	ctx := context.TODO()
	userID := "user-uuid"

	// Expect mock
	mockWalletRepo.On("GetByUserID", ctx, userID).Return(nil, errors.New("db error"))

	// Run the function
	entries, err := svc.GetMutationHistory(ctx, userID)

	// Verify result
	assert.Error(t, err)
	assert.Nil(t, entries)

	// Make sure all mock if called and expected
	mockLedgerRepo.AssertExpectations(t)
	mockWalletRepo.AssertExpectations(t)
}

