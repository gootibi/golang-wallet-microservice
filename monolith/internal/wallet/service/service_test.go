package service

import (
	"context"
	"errors"
	"testing"

	walletModel "github.com/gootibi/golang-wallet-microservice/monolith/internal/wallet/model"
	walletRepository "github.com/gootibi/golang-wallet-microservice/monolith/internal/wallet/repository"
	"github.com/stretchr/testify/assert"
)

func TestGetWalletByUserID_Success(t *testing.T) {
	// Initiate mock reposirories
	mockWalletRepo := new(walletRepository.MockWalletRepository)

	// Initiate new wallet service
	svc := NewWalletService(mockWalletRepo)

	ctx := context.TODO()
	userID := "some-uuid"
	existingWallet := &walletModel.Wallet{
		ID:       "wallet-uuid",
		UserID:   userID,
		Balance:  0.0,
		Currency: "IDR",
		Status:   "active",
	}

	// Expect mock
	mockWalletRepo.On("GetByUserID", ctx, userID).Return(existingWallet, nil)

	// Run the function
	wallet, err := svc.GetWalletByUserID(ctx, userID)

	// Verify result
	assert.NoError(t, err)
	assert.Equal(t, existingWallet, wallet)

	// Make sure all mock if called and expected
	mockWalletRepo.AssertExpectations(t)
}

func TestGetWalletByUserID_NotFound(t *testing.T) {
	// Initiate mock reposirories
	mockWalletRepo := new(walletRepository.MockWalletRepository)

	// Initiate new wallet service
	svc := NewWalletService(mockWalletRepo)

	ctx := context.TODO()
	userID := "some-uuid"

	// Expect mock
	mockWalletRepo.On("GetByUserID", ctx, userID).Return(nil, errors.New("Wallet not found."))

	// Run the function
	wallet, err := svc.GetWalletByUserID(ctx, userID)

	// Verify result
	assert.Error(t, err)
	assert.Nil(t, wallet)

	// Make sure all mock if called and expected
	mockWalletRepo.AssertExpectations(t)
}
