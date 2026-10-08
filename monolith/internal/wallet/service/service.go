package service

import (
	"context"
	"net/http"

	customError "github.com/gootibi/golang-wallet-microservice/monolith/internal/errors"
	"github.com/gootibi/golang-wallet-microservice/monolith/internal/wallet/model"
	"github.com/gootibi/golang-wallet-microservice/monolith/internal/wallet/repository"
)

type WalletService interface {
	GetWalletByUserID(ctx context.Context, userID string) (*model.Wallet, error)
}

func NewWalletService(repo repository.WalletRepository) WalletService {
	return &walletService{
		repo: repo,
	}
}

type walletService struct {
	repo repository.WalletRepository
}

// GetWalletByUserID implements [WalletService].
func (s *walletService) GetWalletByUserID(ctx context.Context, userID string) (*model.Wallet, error) {
	w, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, customError.NewAppError(http.StatusNotFound, "WALLET_NOT_FOUND", "Wallet not found.")
	}

	return w, nil
}
