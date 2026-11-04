package service

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/go-redis/redismock/v9"
	"github.com/gootibi/golang-wallet-microservice/monolith/internal/logger"
	walletModel "github.com/gootibi/golang-wallet-microservice/monolith/internal/wallet/model"
	walletRepository "github.com/gootibi/golang-wallet-microservice/monolith/internal/wallet/repository"
	"github.com/stretchr/testify/assert"
)

func init() {
	logger.InnitLogger()
}

func TestGetWalletByUserID_Success(t *testing.T) {
	// Initiate mock reposirories
	mockWalletRepo := new(walletRepository.MockWalletRepository)

	// Initiate new Redis client (redis mock)
	rdb, mockRedis := redismock.NewClientMock()

	// Initiate new wallet service
	svc := NewWalletService(mockWalletRepo, rdb)

	ctx := context.TODO()
	userID := "user-uuid"
	expectedWallet := &walletModel.Wallet{
		ID:       "wallet-uuid",
		UserID:   userID,
		Balance:  0.0,
		Currency: "IDR",
		Status:   "active",
	}
	expectedWalletBytes, _ := json.Marshal(expectedWallet)

	// Mock Redis cache miss
	mockRedis.ExpectGet("wallet:user:user-uuid").RedisNil()

	// Mock DB success
	mockWalletRepo.On("GetByUserID", ctx, userID).Return(expectedWallet, nil)

	// Mock Redis cache save
	mockRedis.ExpectSet("wallet:user:user-uuid", expectedWalletBytes, 5*time.Minute).SetVal("OK")

	// Run the function
	wallet, err := svc.GetWalletByUserID(ctx, userID)

	// Verify result
	assert.NoError(t, err)
	assert.NoError(t, mockRedis.ExpectationsWereMet())
	assert.Equal(t, expectedWallet, wallet)

	// Make sure all mock if called and expected
	mockWalletRepo.AssertExpectations(t)
}

func TestGetWalletByUserID_NotFound(t *testing.T) {
	// Initiate mock reposirories
	mockWalletRepo := new(walletRepository.MockWalletRepository)

	// Initiate new Redis client (redis mock)
	rdb, mockRedis := redismock.NewClientMock()

	// Initiate new wallet service
	svc := NewWalletService(mockWalletRepo, rdb)

	ctx := context.TODO()
	userID := "user-uuid"

	// Mock Redis cache miss
	mockRedis.ExpectGet("wallet:user:user-uuid").RedisNil()

	// Mock DB failure
	mockWalletRepo.On("GetByUserID", ctx, userID).Return(nil, errors.New("Wallet not found."))

	// Run the function
	wallet, err := svc.GetWalletByUserID(ctx, userID)

	// Verify result
	assert.Error(t, err)
	assert.NoError(t, mockRedis.ExpectationsWereMet())
	assert.Nil(t, wallet)

	// Make sure all mock if called and expected
	mockWalletRepo.AssertExpectations(t)
}

func TestGetWalletByUserID_CacheHit(t *testing.T) {
	// Initiate mock reposirories
	mockRepo := new(walletRepository.MockWalletRepository)

	// Initiate new Redis client (redis mock)
	rdb, mockRedis := redismock.NewClientMock()

	// Initiate new wallet service
	svc := NewWalletService(mockRepo, rdb)

	ctx := context.TODO()
	userID := "user-uuid"
	expectedWallet := &walletModel.Wallet{
		ID:       "wallet-uuid",
		UserID:   userID,
		Balance:  1000.0,
		Currency: "IDR",
		Status:   "active",
	}

	expectedWalletBytes, _ := json.Marshal(expectedWallet)

	// Mock Redis cache hit
	mockRedis.ExpectGet("wallet:user:user-uuid").SetVal(string(expectedWalletBytes))

	// DB should NOT be called (since it's a cache hit)
	// Run the function
	wallet, err := svc.GetWalletByUserID(ctx, userID)

	// Verify result
	assert.NoError(t, err)
	assert.NoError(t, mockRedis.ExpectationsWereMet())
	assert.Equal(t, expectedWallet, wallet)

	// Make sure all mock if called and expected
	mockRepo.AssertExpectations(t)
}
