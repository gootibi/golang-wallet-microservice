package service

import (
	"context"

	"github.com/gootibi/golang-wallet-microservice/monolith/internal/ledger/model"
	"github.com/gootibi/golang-wallet-microservice/monolith/internal/ledger/repository"
	walletRepo "github.com/gootibi/golang-wallet-microservice/monolith/internal/wallet/repository"
)

type LedgerService interface {
	ReconcileWalletBalance(ctx context.Context, userID string) (bool, float64, float64, error)
	GetMutationHistory(ctx context.Context, userID string) ([]model.LedgerEntry, error)
}

func NewLedgerService(lRepo repository.LedgerRepository, wRepo walletRepo.WalletRepository) LedgerService {
	return &ledgerService{
		ledgerRepo: lRepo,
		walletRepo: wRepo,
	}
}

type ledgerService struct {
	ledgerRepo repository.LedgerRepository
	walletRepo walletRepo.WalletRepository
}

// ReconcileWalletBalance implements [LedgerService].
func (s *ledgerService) ReconcileWalletBalance(ctx context.Context, userID string) (bool, float64, float64, error) {
	// Get the user wallet data
	wallet, err := s.walletRepo.GetByUserID(ctx, userID)
	if err != nil {
		return false, 0, 0, err
	}

	// Get the sum of all entries
	calculatedBalanc, err := s.ledgerRepo.GetBalanceByWalletID(ctx, wallet.ID)
	if err != nil {
		return false, 0, 0, err
	}

	// Check if there is a discrepancy
	isConsistent := wallet.Balance == calculatedBalanc

	return isConsistent, wallet.Balance, calculatedBalanc, nil
}

// GetMutationHistory implements [LedgerService].
func (s *ledgerService) GetMutationHistory(ctx context.Context, userID string) ([]model.LedgerEntry, error) {
	wallet, err := s.walletRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return s.ledgerRepo.GetEntriesByWalletID(ctx, wallet.ID)
}
