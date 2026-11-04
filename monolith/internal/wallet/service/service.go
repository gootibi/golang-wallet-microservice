package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	customError "github.com/gootibi/golang-wallet-microservice/monolith/internal/errors"
	"github.com/gootibi/golang-wallet-microservice/monolith/internal/logger"
	"github.com/gootibi/golang-wallet-microservice/monolith/internal/wallet/model"
	"github.com/gootibi/golang-wallet-microservice/monolith/internal/wallet/repository"
	"github.com/redis/go-redis/v9"
)

type WalletService interface {
	GetWalletByUserID(ctx context.Context, userID string) (*model.Wallet, error)
}

func NewWalletService(repo repository.WalletRepository, rdb *redis.Client) WalletService {
	return &walletService{
		repo: repo,
		rdb:  rdb,
	}
}

type walletService struct {
	repo repository.WalletRepository
	rdb  *redis.Client
}

// GetWalletByUserID implements [WalletService].
func (s *walletService) GetWalletByUserID(ctx context.Context, userID string) (*model.Wallet, error) {
	cacheKey := fmt.Sprintf("wallet:user:%s", userID)

	// Check if data exist in redis
	cachedVal, err := s.rdb.Get(ctx, cacheKey).Result()
	if err != nil {
		if err != redis.Nil {
			// Redis is down or has an issues, don't fail the request
			logger.Warn(
				ctx,
				"Redis error during cache lookup, falling back to MySQL",
				"error", err.Error(), "user_id", userID,
			)
		}
		// Cache miss or redis down
		logger.Info(
			ctx,
			"Cache miss for wallet, fetching from MySQL...",
			"user_id", userID,
		)
	} else {
		// Cache hit, Deserialize JSON string to model.Wallet struct
		wallet := &model.Wallet{}
		if err := json.Unmarshal([]byte(cachedVal), wallet); err == nil {
			logger.Info(
				ctx,
				"Cache hit for wallet, returning from Redis",
				"user_id", userID,
			)

			return wallet, nil
		}

		logger.Warn(
			ctx,
			"Failed to deserialize cached wallet data, falling back to MySQL",
			"error", err.Error(), "user_id", userID,
		)
	}

	wallet, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, customError.NewAppError(http.StatusNotFound, "WALLET_NOT_FOUND", "Wallet not found.")
	}

	// Save to Redis for 5 minutes TTL
	walletBytes, err := json.Marshal(wallet)
	if err == nil {
		s.rdb.Set(ctx, cacheKey, walletBytes, 5*time.Minute)
	}

	return wallet, nil
}
