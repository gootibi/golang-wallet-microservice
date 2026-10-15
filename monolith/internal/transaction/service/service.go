package service

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/google/uuid"
	customError "github.com/gootibi/golang-wallet-microservice/monolith/internal/errors"
	ledgerModel "github.com/gootibi/golang-wallet-microservice/monolith/internal/ledger/model"
	ledgerRepo "github.com/gootibi/golang-wallet-microservice/monolith/internal/ledger/repository"
	transactionModel "github.com/gootibi/golang-wallet-microservice/monolith/internal/transaction/model"
	transactionRepo "github.com/gootibi/golang-wallet-microservice/monolith/internal/transaction/repository"
	userRepo "github.com/gootibi/golang-wallet-microservice/monolith/internal/user/repository"
	walletRepo "github.com/gootibi/golang-wallet-microservice/monolith/internal/wallet/repository"
)

type TransactionService interface {
	Transfer(ctx context.Context, senderUserID string, req transactionModel.TransferRequest) (*transactionModel.Transaction, error)
}

func NewTransactionService(
	db *sql.DB,
	tRepo transactionRepo.TransactionRepository,
	uRepo userRepo.UserRepository,
	wRepo walletRepo.WalletRepository,
	lRepo ledgerRepo.LedgerRepository,
) TransactionService {
	return &transactionService{
		db:         db,
		txRepo:     tRepo,
		userRepo:   uRepo,
		walletRepo: wRepo,
		ledgerRepo: lRepo,
	}
}

type transactionService struct {
	db         *sql.DB
	txRepo     transactionRepo.TransactionRepository
	userRepo   userRepo.UserRepository
	walletRepo walletRepo.WalletRepository
	ledgerRepo ledgerRepo.LedgerRepository
}

// Transfer implements [TransactionService].
func (s *transactionService) Transfer(ctx context.Context, senderUserID string, req transactionModel.TransferRequest) (*transactionModel.Transaction, error) {
	// Check idempoteny key - this is for checking to not process the same request
	existing, _ := s.txRepo.GetByIdempotencyKey(ctx, req.IdempotencyKey)
	if existing != nil {
		return existing, nil
	}

	// Look receiver by email
	receiverUser, err := s.userRepo.GetByEmail(ctx, req.ReceiverEmail)
	if err != nil {
		return nil, customError.NewAppError(http.StatusNotFound, "RECEIVER_NOT_FOUND", "Receiver not found.")
	}

	// Start DB transaction
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, customError.ErrInternalServer
	}
	defer tx.Rollback()

	// Look for receiver and sender wallet
	senderWallet, err := s.walletRepo.GetByUserID(ctx, senderUserID)
	if err != nil {
		return nil, customError.NewAppError(http.StatusNotFound, "SENDER_WALLET_NOT_FOUND", "Sender wallet not found.")
	}

	receiverWallet, err := s.walletRepo.GetByUserID(ctx, receiverUser.ID)
	if err != nil {
		return nil, customError.NewAppError(http.StatusNotFound, "RECEIVER_WALLET_NOT_FOUND", "Receiver wallet not found.")
	}

	// Sender and receiver cannot be same
	if senderWallet.ID == receiverWallet.ID {
		return nil, customError.NewAppError(http.StatusBadRequest, "INVALID_TRANSFER", "Cannot transfer to self account.")
	}

	// Validate sender wallet balance is enough or not
	if senderWallet.Balance < req.Amount {
		return nil, customError.NewAppError(http.StatusBadRequest, "INSUFFICIENT_BALANCE", "Insufficient balance.")
	}

	// OPTIMISTIC LOCKING PROCESS
	// -------------------------------------------------------------------------
	// Reduce sender wallet and add receiver wallet with tx and checking version
	newSenderBalance := senderWallet.Balance - req.Amount
	err = s.walletRepo.UpdateBalanceTx(ctx, tx, senderWallet.ID, newSenderBalance, senderWallet.Version)
	if err != nil {
		// If failed because version mismatch, return special error so client can retry
		return nil, customError.NewAppError(http.StatusConflict, "CONCURRENCY_CONFLICT", "Transaction is busy, please retry in the following minutes")
	}

	newReceiverBalance := receiverWallet.Balance + req.Amount
	err = s.walletRepo.UpdateBalanceTx(ctx, tx, receiverWallet.ID, newReceiverBalance, receiverWallet.Version)
	if err != nil {
		// If failed because version mismatch, return special error so client can retry
		return nil, customError.NewAppError(http.StatusConflict, "CONCURRENCY_CONFLICT", "Transaction is busy, please retry in the following minutes")
	}

	// Create transaction record
	transactionID := uuid.New().String()
	transaction := &transactionModel.Transaction{
		ID:               transactionID,
		SenderWalletID:   &senderWallet.ID,
		ReceiverWalletID: receiverWallet.ID,
		Amount:           req.Amount,
		Description:      req.Description,
		IdempotencyKey:   req.IdempotencyKey,
		Status:           transactionModel.SUCCESS, // pending, processing, success ect.
	}

	if err = s.txRepo.CreateTX(ctx, tx, transaction); err != nil {
		return nil, customError.ErrInternalServer
	}

	// Create two ledger rows (debit for sender and credit for receiver)
	debitEntry := &ledgerModel.LedgerEntry{
		ID:            uuid.New().String(),
		WalletID:      senderWallet.ID,
		TransactionID: transactionID,
		EntryType:     ledgerModel.DEBIT, // "debit"
		Amount:        req.Amount,
	}
	if err = s.ledgerRepo.CreateTx(ctx, tx, debitEntry); err != nil {
		return nil, customError.ErrInternalServer
	}

	creditEntry := &ledgerModel.LedgerEntry{
		ID:            uuid.New().String(),
		WalletID:      receiverWallet.ID,
		TransactionID: transactionID,
		EntryType:     ledgerModel.CREDIT, // "credit"
		Amount:        req.Amount,
	}
	if err = s.ledgerRepo.CreateTx(ctx, tx, creditEntry); err != nil {
		return nil, customError.ErrInternalServer
	}

	// Commit the db transaction
	if err = tx.Commit(); err != nil {
		return nil, customError.ErrInternalServer
	}

	return transaction, nil
}
