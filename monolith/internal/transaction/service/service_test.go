package service

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	ledgerRepository "github.com/gootibi/golang-wallet-microservice/monolith/internal/ledger/repository"
	transactionModel "github.com/gootibi/golang-wallet-microservice/monolith/internal/transaction/model"
	transactionRepository "github.com/gootibi/golang-wallet-microservice/monolith/internal/transaction/repository"
	userModel "github.com/gootibi/golang-wallet-microservice/monolith/internal/user/model"
	userRepository "github.com/gootibi/golang-wallet-microservice/monolith/internal/user/repository"
	walletModel "github.com/gootibi/golang-wallet-microservice/monolith/internal/wallet/model"
	walletRepository "github.com/gootibi/golang-wallet-microservice/monolith/internal/wallet/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestTransfer_Success(t *testing.T) {
	// Sql mock
	db, dbMock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	// Initiate mock reposirories
	mockTxRepo := new(transactionRepository.MockTransactionRepository)
	mockUserRepo := new(userRepository.MockUserRepository)
	mockWalletRepo := new(walletRepository.MockWalletRepository)
	mockLedgerRepo := new(ledgerRepository.MockLedgerRepository)

	// We no need real sql.DB connection, so we sent sql mock
	svc := NewTransactionService(db, mockTxRepo, mockUserRepo, mockWalletRepo, mockLedgerRepo)

	ctx := context.TODO()
	senderUserID := "sender-uuid"
	req := transactionModel.TransferRequest{
		ReceiverEmail:  "reciver@example.com",
		Amount:         200.0,
		Description:    "Test",
		IdempotencyKey: "ip-unique-key",
	}
	receiverUser := &userModel.User{
		ID:    "receiver-uuid",
		Email: req.ReceiverEmail,
	}
	senderWallet := &walletModel.Wallet{
		ID:      "wallet-sender-uuid",
		UserID:  senderUserID,
		Balance: 1000.0,
		Version: 1,
	}
	receiverWallet := &walletModel.Wallet{
		ID:      "wallet-receiver-uuid",
		UserID:  "receiver-uuid",
		Balance: 500.0,
		Version: 2,
	}

	// Check idempoteny key
	mockTxRepo.On("GetByIdempotencyKey", ctx, req.IdempotencyKey).Return(nil, errors.New("Not found"))

	// Look receiver
	mockUserRepo.On("GetByEmail", ctx, req.ReceiverEmail).Return(receiverUser, nil)

	// Expect sql transaction
	dbMock.ExpectBegin()
	dbMock.ExpectCommit()

	// Look for receiver and sender wallet
	mockWalletRepo.On("GetByUserID", ctx, senderUserID).Return(senderWallet, nil)
	mockWalletRepo.On("GetByUserID", ctx, "receiver-uuid").Return(receiverWallet, nil)

	// Update balances
	mockWalletRepo.On("UpdateBalanceTx", ctx, mock.Anything, senderWallet.ID, 800.0, senderWallet.Version).Return(nil)
	mockWalletRepo.On("UpdateBalanceTx", ctx, mock.Anything, receiverWallet.ID, 700.0, receiverWallet.Version).Return(nil)

	// Create transaction
	mockTxRepo.On("CreateTX", ctx, mock.Anything, mock.Anything).Return(nil)

	// Create two ledger rows
	mockLedgerRepo.On("CreateTx", ctx, mock.Anything, mock.Anything).Return(nil).Twice()

	// Run the function
	txRes, err := svc.Transfer(ctx, senderUserID, req)

	// Verify result
	assert.NoError(t, err)
	assert.NotNil(t, txRes)
	assert.Equal(t, req.Amount, txRes.Amount)
	assert.Equal(t, transactionModel.SUCCESS, txRes.Status)

	// Make sure all mock if called and expected
	mockTxRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
	mockWalletRepo.AssertExpectations(t)
	mockLedgerRepo.AssertExpectations(t)
	assert.NoError(t, dbMock.ExpectationsWereMet())
}

func TestTransfer_ReceiverNotFound(t *testing.T) {
	// Sql mock
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	// Initiate mock reposirories
	mockTxRepo := new(transactionRepository.MockTransactionRepository)
	mockUserRepo := new(userRepository.MockUserRepository)
	mockWalletRepo := new(walletRepository.MockWalletRepository)
	mockLedgerRepo := new(ledgerRepository.MockLedgerRepository)

	// We no need real sql.DB connection, so we sent sql mock
	svc := NewTransactionService(db, mockTxRepo, mockUserRepo, mockWalletRepo, mockLedgerRepo)

	ctx := context.TODO()
	senderUserID := "sender-uuid"
	req := transactionModel.TransferRequest{
		ReceiverEmail:  "reciver@example.com",
		Amount:         200.0,
		Description:    "Test",
		IdempotencyKey: "ip-unique-key",
	}

	// Expect mock
	// Check idempoteny key
	mockTxRepo.On("GetByIdempotencyKey", ctx, req.IdempotencyKey).Return(nil, errors.New("Not found"))
	// Look receiver
	mockUserRepo.On("GetByEmail", ctx, req.ReceiverEmail).Return(nil, errors.New("Receiver not found."))

	// Run the function
	txRes, err := svc.Transfer(ctx, senderUserID, req)

	// Verify result
	assert.Error(t, err)
	assert.Nil(t, txRes)

	// Make sure all mock if called and expected
	mockTxRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
	mockWalletRepo.AssertExpectations(t)
	mockLedgerRepo.AssertExpectations(t)
}

func TestTransfer_SelfTransferNotAllowed(t *testing.T) {
	// Sql mock
	db, dbMock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	// Initiate mock reposirories
	mockTxRepo := new(transactionRepository.MockTransactionRepository)
	mockUserRepo := new(userRepository.MockUserRepository)
	mockWalletRepo := new(walletRepository.MockWalletRepository)
	mockLedgerRepo := new(ledgerRepository.MockLedgerRepository)

	// We no need real sql.DB connection, so we sent sql mock
	svc := NewTransactionService(db, mockTxRepo, mockUserRepo, mockWalletRepo, mockLedgerRepo)

	ctx := context.TODO()
	senderUserID := "sender-uuid"
	req := transactionModel.TransferRequest{
		ReceiverEmail:  "reciver@example.com",
		Amount:         200.0,
		Description:    "Test",
		IdempotencyKey: "ip-unique-key",
	}
	receiverUser := &userModel.User{
		ID:    senderUserID,
		Email: req.ReceiverEmail,
	}
	senderWallet := &walletModel.Wallet{
		ID:      "wallet-sender-uuid",
		UserID:  senderUserID,
		Balance: 1000.0,
		Version: 1,
	}
	receiverWallet := &walletModel.Wallet{
		ID:      "wallet-sender-uuid",
		UserID:  senderUserID,
		Balance: 1000.0,
		Version: 1,
	}

	// Check idempoteny key
	mockTxRepo.On("GetByIdempotencyKey", ctx, req.IdempotencyKey).Return(nil, errors.New("Not found"))

	// Look receiver
	mockUserRepo.On("GetByEmail", ctx, req.ReceiverEmail).Return(receiverUser, nil)

	// Expect sql transaction
	dbMock.ExpectBegin()
	dbMock.ExpectCommit()

	// Look for receiver and sender wallet
	// Both return same wallet ID
	mockWalletRepo.On("GetByUserID", ctx, senderUserID).Return(senderWallet, nil)
	mockWalletRepo.On("GetByUserID", ctx, receiverUser.ID).Return(receiverWallet, nil)

	// Run the function
	txRes, err := svc.Transfer(ctx, senderUserID, req)

	// Verify result
	assert.Error(t, err)
	assert.Nil(t, txRes)

	// Make sure all mock if called and expected
	mockTxRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
	mockWalletRepo.AssertExpectations(t)
	mockLedgerRepo.AssertExpectations(t)
	assert.Error(t, dbMock.ExpectationsWereMet())
}

func TestTransfer_InsufficientBalance(t *testing.T) {
	// Sql mock
	db, dbMock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	// Initiate mock reposirories
	mockTxRepo := new(transactionRepository.MockTransactionRepository)
	mockUserRepo := new(userRepository.MockUserRepository)
	mockWalletRepo := new(walletRepository.MockWalletRepository)
	mockLedgerRepo := new(ledgerRepository.MockLedgerRepository)

	// We no need real sql.DB connection, so we sent sql mock
	svc := NewTransactionService(db, mockTxRepo, mockUserRepo, mockWalletRepo, mockLedgerRepo)

	ctx := context.TODO()
	senderUserID := "sender-uuid"
	req := transactionModel.TransferRequest{
		ReceiverEmail:  "reciver@example.com",
		Amount:         1100.0, // more than sender user balance
		Description:    "Test",
		IdempotencyKey: "ip-unique-key",
	}
	receiverUser := &userModel.User{
		ID:    "receiver-uuid",
		Email: req.ReceiverEmail,
	}
	senderWallet := &walletModel.Wallet{
		ID:      "wallet-sender-uuid",
		UserID:  senderUserID,
		Balance: 1000.0,
		Version: 1,
	}
	receiverWallet := &walletModel.Wallet{
		ID:      "wallet-receiver-uuid",
		UserID:  "receiver-uuid",
		Balance: 500.0,
		Version: 2,
	}

	// Check idempoteny key
	mockTxRepo.On("GetByIdempotencyKey", ctx, req.IdempotencyKey).Return(nil, errors.New("Not found"))

	// Look receiver
	mockUserRepo.On("GetByEmail", ctx, req.ReceiverEmail).Return(receiverUser, nil)

	// Expect sql transaction
	dbMock.ExpectBegin()
	dbMock.ExpectCommit()

	// Look for receiver and sender wallet
	mockWalletRepo.On("GetByUserID", ctx, senderUserID).Return(senderWallet, nil)
	mockWalletRepo.On("GetByUserID", ctx, "receiver-uuid").Return(receiverWallet, nil)

	// Run the function
	txRes, err := svc.Transfer(ctx, senderUserID, req)

	// Verify result
	assert.Error(t, err)
	assert.Nil(t, txRes)

	// Make sure all mock if called and expected
	mockTxRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
	mockWalletRepo.AssertExpectations(t)
	mockLedgerRepo.AssertExpectations(t)
	assert.Error(t, dbMock.ExpectationsWereMet())
}

func TestTransfer_SenderWalletNotFound(t *testing.T) {
	// Sql mock
	db, dbMock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	// Initiate mock reposirories
	mockTxRepo := new(transactionRepository.MockTransactionRepository)
	mockUserRepo := new(userRepository.MockUserRepository)
	mockWalletRepo := new(walletRepository.MockWalletRepository)
	mockLedgerRepo := new(ledgerRepository.MockLedgerRepository)

	// We no need real sql.DB connection, so we sent sql mock
	svc := NewTransactionService(db, mockTxRepo, mockUserRepo, mockWalletRepo, mockLedgerRepo)

	ctx := context.TODO()
	senderUserID := "sender-uuid"
	req := transactionModel.TransferRequest{
		ReceiverEmail:  "reciver@example.com",
		Amount:         200.0,
		Description:    "Test",
		IdempotencyKey: "ip-unique-key",
	}
	receiverUser := &userModel.User{
		ID:    "receiver-uuid",
		Email: req.ReceiverEmail,
	}

	receiverWallet := &walletModel.Wallet{
		ID:      "wallet-receiver-uuid",
		UserID:  "receiver-uuid",
		Balance: 500.0,
		Version: 2,
	}

	// Check idempoteny key
	mockTxRepo.On("GetByIdempotencyKey", ctx, req.IdempotencyKey).Return(nil, errors.New("Not found"))

	// Look receiver
	mockUserRepo.On("GetByEmail", ctx, req.ReceiverEmail).Return(receiverUser, nil)

	// Expect sql transaction
	dbMock.ExpectBegin()
	dbMock.ExpectCommit()

	// Look for receiver and sender wallet
	mockWalletRepo.On("GetByUserID", ctx, senderUserID).Return(nil, errors.New("Not found"))
	mockWalletRepo.On("GetByUserID", ctx, "receiver-uuid").Return(receiverWallet, nil)

	// Run the function
	txRes, err := svc.Transfer(ctx, senderUserID, req)

	// Verify result
	assert.Error(t, err)
	assert.Nil(t, txRes)

	// Make sure all mock if called and expected
	mockTxRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
	mockLedgerRepo.AssertExpectations(t)
	assert.Error(t, dbMock.ExpectationsWereMet())
}

func TestTransfer_ReceiverWalletNotFound(t *testing.T) {
	// Sql mock
	db, dbMock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	// Initiate mock reposirories
	mockTxRepo := new(transactionRepository.MockTransactionRepository)
	mockUserRepo := new(userRepository.MockUserRepository)
	mockWalletRepo := new(walletRepository.MockWalletRepository)
	mockLedgerRepo := new(ledgerRepository.MockLedgerRepository)

	// We no need real sql.DB connection, so we sent sql mock
	svc := NewTransactionService(db, mockTxRepo, mockUserRepo, mockWalletRepo, mockLedgerRepo)

	ctx := context.TODO()
	senderUserID := "sender-uuid"
	req := transactionModel.TransferRequest{
		ReceiverEmail:  "reciver@example.com",
		Amount:         200.0,
		Description:    "Test",
		IdempotencyKey: "ip-unique-key",
	}
	receiverUser := &userModel.User{
		ID:    "receiver-uuid",
		Email: req.ReceiverEmail,
	}
	senderWallet := &walletModel.Wallet{
		ID:      "wallet-sender-uuid",
		UserID:  senderUserID,
		Balance: 1000.0,
		Version: 1,
	}

	// Check idempoteny key
	mockTxRepo.On("GetByIdempotencyKey", ctx, req.IdempotencyKey).Return(nil, errors.New("Not found"))

	// Look receiver
	mockUserRepo.On("GetByEmail", ctx, req.ReceiverEmail).Return(receiverUser, nil)

	// Expect sql transaction
	dbMock.ExpectBegin()
	dbMock.ExpectCommit()

	// Look for receiver and sender wallet
	mockWalletRepo.On("GetByUserID", ctx, senderUserID).Return(senderWallet, nil)
	mockWalletRepo.On("GetByUserID", ctx, "receiver-uuid").Return(nil, errors.New("Not found"))

	// Run the function
	txRes, err := svc.Transfer(ctx, senderUserID, req)

	// Verify result
	assert.Error(t, err)
	assert.Nil(t, txRes)

	// Make sure all mock if called and expected
	mockTxRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
	mockLedgerRepo.AssertExpectations(t)
	assert.Error(t, dbMock.ExpectationsWereMet())
}

func TestTransfer_UpdateBalanceFailed(t *testing.T) {
	// Sql mock
	db, dbMock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	// Initiate mock reposirories
	mockTxRepo := new(transactionRepository.MockTransactionRepository)
	mockUserRepo := new(userRepository.MockUserRepository)
	mockWalletRepo := new(walletRepository.MockWalletRepository)
	mockLedgerRepo := new(ledgerRepository.MockLedgerRepository)

	// We no need real sql.DB connection, so we sent sql mock
	svc := NewTransactionService(db, mockTxRepo, mockUserRepo, mockWalletRepo, mockLedgerRepo)

	ctx := context.TODO()
	senderUserID := "sender-uuid"
	req := transactionModel.TransferRequest{
		ReceiverEmail:  "reciver@example.com",
		Amount:         200.0,
		Description:    "Test",
		IdempotencyKey: "ip-unique-key",
	}
	receiverUser := &userModel.User{
		ID:    "receiver-uuid",
		Email: req.ReceiverEmail,
	}
	senderWallet := &walletModel.Wallet{
		ID:      "wallet-sender-uuid",
		UserID:  senderUserID,
		Balance: 1000.0,
		Version: 1,
	}
	receiverWallet := &walletModel.Wallet{
		ID:      "wallet-receiver-uuid",
		UserID:  "receiver-uuid",
		Balance: 500.0,
		Version: 2,
	}

	// Check idempoteny key
	mockTxRepo.On("GetByIdempotencyKey", ctx, req.IdempotencyKey).Return(nil, errors.New("Not found"))

	// Look receiver
	mockUserRepo.On("GetByEmail", ctx, req.ReceiverEmail).Return(receiverUser, nil)

	// Expect sql transaction
	dbMock.ExpectBegin()
	dbMock.ExpectCommit()

	// Look for receiver and sender wallet
	mockWalletRepo.On("GetByUserID", ctx, senderUserID).Return(senderWallet, nil)
	mockWalletRepo.On("GetByUserID", ctx, "receiver-uuid").Return(receiverWallet, nil)

	// Update balances
	mockWalletRepo.On("UpdateBalanceTx", ctx, mock.Anything, senderWallet.ID, 800.0, senderWallet.Version).Return(errors.New("db error"))

	// Run the function
	txRes, err := svc.Transfer(ctx, senderUserID, req)

	// Verify result
	assert.Error(t, err)
	assert.Nil(t, txRes)

	// Make sure all mock if called and expected
	mockTxRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
	mockWalletRepo.AssertExpectations(t)
	mockLedgerRepo.AssertExpectations(t)
	assert.Error(t, dbMock.ExpectationsWereMet())
}

func TestTransfer_CreateTransactionFailed(t *testing.T) {
	// Sql mock
	db, dbMock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	// Initiate mock reposirories
	mockTxRepo := new(transactionRepository.MockTransactionRepository)
	mockUserRepo := new(userRepository.MockUserRepository)
	mockWalletRepo := new(walletRepository.MockWalletRepository)
	mockLedgerRepo := new(ledgerRepository.MockLedgerRepository)

	// We no need real sql.DB connection, so we sent sql mock
	svc := NewTransactionService(db, mockTxRepo, mockUserRepo, mockWalletRepo, mockLedgerRepo)

	ctx := context.TODO()
	senderUserID := "sender-uuid"
	req := transactionModel.TransferRequest{
		ReceiverEmail:  "reciver@example.com",
		Amount:         200.0,
		Description:    "Test",
		IdempotencyKey: "ip-unique-key",
	}
	receiverUser := &userModel.User{
		ID:    "receiver-uuid",
		Email: req.ReceiverEmail,
	}
	senderWallet := &walletModel.Wallet{
		ID:      "wallet-sender-uuid",
		UserID:  senderUserID,
		Balance: 1000.0,
		Version: 1,
	}
	receiverWallet := &walletModel.Wallet{
		ID:      "wallet-receiver-uuid",
		UserID:  "receiver-uuid",
		Balance: 500.0,
		Version: 2,
	}

	// Check idempoteny key
	mockTxRepo.On("GetByIdempotencyKey", ctx, req.IdempotencyKey).Return(nil, errors.New("Not found"))

	// Look receiver
	mockUserRepo.On("GetByEmail", ctx, req.ReceiverEmail).Return(receiverUser, nil)

	// Expect sql transaction
	dbMock.ExpectBegin()
	dbMock.ExpectCommit()

	// Look for receiver and sender wallet
	mockWalletRepo.On("GetByUserID", ctx, senderUserID).Return(senderWallet, nil)
	mockWalletRepo.On("GetByUserID", ctx, "receiver-uuid").Return(receiverWallet, nil)

	// Update balances
	mockWalletRepo.On("UpdateBalanceTx", ctx, mock.Anything, senderWallet.ID, 800.0, senderWallet.Version).Return(nil)
	mockWalletRepo.On("UpdateBalanceTx", ctx, mock.Anything, receiverWallet.ID, 700.0, receiverWallet.Version).Return(nil)

	// Create transaction
	mockTxRepo.On("CreateTX", ctx, mock.Anything, mock.Anything).Return(errors.New("db error"))

	// Run the function
	txRes, err := svc.Transfer(ctx, senderUserID, req)

	// Verify result
	assert.Error(t, err)
	assert.Nil(t, txRes)

	// Make sure all mock if called and expected
	mockTxRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
	mockWalletRepo.AssertExpectations(t)
	mockLedgerRepo.AssertExpectations(t)
	assert.Error(t, dbMock.ExpectationsWereMet())
}

func TestTransfer_CreateLedgerFailed(t *testing.T) {
	// Sql mock
	db, dbMock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	// Initiate mock reposirories
	mockTxRepo := new(transactionRepository.MockTransactionRepository)
	mockUserRepo := new(userRepository.MockUserRepository)
	mockWalletRepo := new(walletRepository.MockWalletRepository)
	mockLedgerRepo := new(ledgerRepository.MockLedgerRepository)

	// We no need real sql.DB connection, so we sent sql mock
	svc := NewTransactionService(db, mockTxRepo, mockUserRepo, mockWalletRepo, mockLedgerRepo)

	ctx := context.TODO()
	senderUserID := "sender-uuid"
	req := transactionModel.TransferRequest{
		ReceiverEmail:  "reciver@example.com",
		Amount:         200.0,
		Description:    "Test",
		IdempotencyKey: "ip-unique-key",
	}
	receiverUser := &userModel.User{
		ID:    "receiver-uuid",
		Email: req.ReceiverEmail,
	}
	senderWallet := &walletModel.Wallet{
		ID:      "wallet-sender-uuid",
		UserID:  senderUserID,
		Balance: 1000.0,
		Version: 1,
	}
	receiverWallet := &walletModel.Wallet{
		ID:      "wallet-receiver-uuid",
		UserID:  "receiver-uuid",
		Balance: 500.0,
		Version: 2,
	}

	// Check idempoteny key
	mockTxRepo.On("GetByIdempotencyKey", ctx, req.IdempotencyKey).Return(nil, errors.New("Not found"))

	// Look receiver
	mockUserRepo.On("GetByEmail", ctx, req.ReceiverEmail).Return(receiverUser, nil)

	// Expect sql transaction
	dbMock.ExpectBegin()
	dbMock.ExpectCommit()

	// Look for receiver and sender wallet
	mockWalletRepo.On("GetByUserID", ctx, senderUserID).Return(senderWallet, nil)
	mockWalletRepo.On("GetByUserID", ctx, "receiver-uuid").Return(receiverWallet, nil)

	// Update balances
	mockWalletRepo.On("UpdateBalanceTx", ctx, mock.Anything, senderWallet.ID, 800.0, senderWallet.Version).Return(nil)
	mockWalletRepo.On("UpdateBalanceTx", ctx, mock.Anything, receiverWallet.ID, 700.0, receiverWallet.Version).Return(nil)

	// Create transaction
	mockTxRepo.On("CreateTX", ctx, mock.Anything, mock.Anything).Return(nil)

	// Create two ledger rows
	mockLedgerRepo.On("CreateTx", ctx, mock.Anything, mock.Anything).Return(errors.New("db error")).Twice()

	// Run the function
	txRes, err := svc.Transfer(ctx, senderUserID, req)

	// Verify result
	assert.Error(t, err)
	assert.Nil(t, txRes)

	// Make sure all mock if called and expected
	mockTxRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
	mockWalletRepo.AssertExpectations(t)
	assert.Error(t, dbMock.ExpectationsWereMet())
}

func TestTransfer_TxCommitError(t *testing.T) {
	// Sql mock
	db, dbMock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	// Initiate mock reposirories
	mockTxRepo := new(transactionRepository.MockTransactionRepository)
	mockUserRepo := new(userRepository.MockUserRepository)
	mockWalletRepo := new(walletRepository.MockWalletRepository)
	mockLedgerRepo := new(ledgerRepository.MockLedgerRepository)

	// We no need real sql.DB connection, so we sent sql mock
	svc := NewTransactionService(db, mockTxRepo, mockUserRepo, mockWalletRepo, mockLedgerRepo)

	ctx := context.TODO()
	senderUserID := "sender-uuid"
	req := transactionModel.TransferRequest{
		ReceiverEmail:  "reciver@example.com",
		Amount:         200.0,
		Description:    "Test",
		IdempotencyKey: "ip-unique-key",
	}
	receiverUser := &userModel.User{
		ID:    "receiver-uuid",
		Email: req.ReceiverEmail,
	}
	senderWallet := &walletModel.Wallet{
		ID:      "wallet-sender-uuid",
		UserID:  senderUserID,
		Balance: 1000.0,
		Version: 1,
	}
	receiverWallet := &walletModel.Wallet{
		ID:      "wallet-receiver-uuid",
		UserID:  "receiver-uuid",
		Balance: 500.0,
		Version: 2,
	}

	// Check idempoteny key
	mockTxRepo.On("GetByIdempotencyKey", ctx, req.IdempotencyKey).Return(nil, errors.New("Not found"))

	// Look receiver
	mockUserRepo.On("GetByEmail", ctx, req.ReceiverEmail).Return(receiverUser, nil)

	// Expect sql transaction
	dbMock.ExpectBegin()

	// Look for receiver and sender wallet
	mockWalletRepo.On("GetByUserID", ctx, senderUserID).Return(senderWallet, nil)
	mockWalletRepo.On("GetByUserID", ctx, "receiver-uuid").Return(receiverWallet, nil)

	// Update balances
	mockWalletRepo.On("UpdateBalanceTx", ctx, mock.Anything, senderWallet.ID, 800.0, senderWallet.Version).Return(nil)
	mockWalletRepo.On("UpdateBalanceTx", ctx, mock.Anything, receiverWallet.ID, 700.0, receiverWallet.Version).Return(nil)

	// Create transaction
	mockTxRepo.On("CreateTX", ctx, mock.Anything, mock.Anything).Return(nil)

	// Create two ledger rows
	mockLedgerRepo.On("CreateTx", ctx, mock.Anything, mock.Anything).Return(nil).Twice()

	// Run the function
	txRes, err := svc.Transfer(ctx, senderUserID, req)

	// Commit the db transaction
	dbMock.ExpectCommit().WillReturnError(errors.New("Internal server error"))

	// Verify result
	assert.Error(t, err)
	assert.Nil(t, txRes)

	// Make sure all mock if called and expected
	mockTxRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
	mockWalletRepo.AssertExpectations(t)
	mockLedgerRepo.AssertExpectations(t)
	assert.Error(t, dbMock.ExpectationsWereMet())
}

func TestTransfer_TxBeginError(t *testing.T) {
	// Sql mock
	db, dbMock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	// Initiate mock reposirories
	mockTxRepo := new(transactionRepository.MockTransactionRepository)
	mockUserRepo := new(userRepository.MockUserRepository)
	mockWalletRepo := new(walletRepository.MockWalletRepository)
	mockLedgerRepo := new(ledgerRepository.MockLedgerRepository)

	// We no need real sql.DB connection, so we sent sql mock
	svc := NewTransactionService(db, mockTxRepo, mockUserRepo, mockWalletRepo, mockLedgerRepo)

	ctx := context.TODO()
	senderUserID := "sender-uuid"
	req := transactionModel.TransferRequest{
		ReceiverEmail:  "reciver@example.com",
		Amount:         200.0,
		Description:    "Test",
		IdempotencyKey: "ip-unique-key",
	}
	receiverUser := &userModel.User{
		ID:    "receiver-uuid",
		Email: req.ReceiverEmail,
	}

	// Check idempoteny key
	mockTxRepo.On("GetByIdempotencyKey", ctx, req.IdempotencyKey).Return(nil, errors.New("Not found"))

	// Look receiver
	mockUserRepo.On("GetByEmail", ctx, req.ReceiverEmail).Return(receiverUser, nil)

	// Expect sql transaction
	dbMock.ExpectBegin().WillReturnError(errors.New("Internal server error"))

	// Run the function
	txRes, err := svc.Transfer(ctx, senderUserID, req)

	// Verify result
	assert.Error(t, err)
	assert.Nil(t, txRes)

	// Make sure all mock if called and expected
	mockTxRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
	mockWalletRepo.AssertExpectations(t)
	mockLedgerRepo.AssertExpectations(t)
}

func TestGetHistory_Success(t *testing.T) {
	// Sql mock
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	// Initiate mock reposirories
	mockTxRepo := new(transactionRepository.MockTransactionRepository)
	mockUserRepo := new(userRepository.MockUserRepository)
	mockWalletRepo := new(walletRepository.MockWalletRepository)
	mockLedgerRepo := new(ledgerRepository.MockLedgerRepository)

	// We no need real sql.DB connection, so we sent sql mock
	svc := NewTransactionService(db, mockTxRepo, mockUserRepo, mockWalletRepo, mockLedgerRepo)

	ctx := context.TODO()
	userID := "user-uuid"
	wallet := &walletModel.Wallet{ID: "wallet-uuid"}
	params := transactionModel.PaginationParams{Page: 1, Limit: 10}

	expectedHistory := []transactionModel.Transaction{
		{ID: "tx-uuid", ReceiverWalletID: "wallet-uuid", Amount: 500.0},
	}

	mockWalletRepo.On("GetByUserID", ctx, userID).Return(wallet, nil)
	mockTxRepo.On("GetHistory", ctx, wallet.ID, params).Return(expectedHistory, int64(1), nil)

	// Run the function
	txs, meta, err := svc.GetHistory(ctx, userID, params)

	// Verify result
	assert.NoError(t, err)
	assert.NotNil(t, txs)
	assert.Equal(t, expectedHistory, txs)
	assert.Equal(t, int64(1), meta.Total)
	assert.Equal(t, 1, meta.TotalPage)

	// Make sure all mock if called and expected
	mockTxRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
	mockWalletRepo.AssertExpectations(t)
	mockLedgerRepo.AssertExpectations(t)
}

func TestGetHistory_WalletNotFound(t *testing.T) {
	// Sql mock
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	// Initiate mock reposirories
	mockTxRepo := new(transactionRepository.MockTransactionRepository)
	mockUserRepo := new(userRepository.MockUserRepository)
	mockWalletRepo := new(walletRepository.MockWalletRepository)
	mockLedgerRepo := new(ledgerRepository.MockLedgerRepository)

	// We no need real sql.DB connection, so we sent sql mock
	svc := NewTransactionService(db, mockTxRepo, mockUserRepo, mockWalletRepo, mockLedgerRepo)

	ctx := context.TODO()
	userID := "user-uuid"
	params := transactionModel.PaginationParams{Page: 1, Limit: 10}

	mockWalletRepo.On("GetByUserID", ctx, userID).Return(nil, errors.New("Not found"))

	// Run the function
	txs, meta, err := svc.GetHistory(ctx, userID, params)

	// Verify result
	assert.Error(t, err)
	assert.Nil(t, txs)
	assert.Nil(t, meta)

	// Make sure all mock if called and expected
	mockTxRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
	mockWalletRepo.AssertExpectations(t)
	mockLedgerRepo.AssertExpectations(t)
}

func TestGetHistory_TxGetHistoryFailed(t *testing.T) {
	// Sql mock
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	// Initiate mock reposirories
	mockTxRepo := new(transactionRepository.MockTransactionRepository)
	mockUserRepo := new(userRepository.MockUserRepository)
	mockWalletRepo := new(walletRepository.MockWalletRepository)
	mockLedgerRepo := new(ledgerRepository.MockLedgerRepository)

	// We no need real sql.DB connection, so we sent sql mock
	svc := NewTransactionService(db, mockTxRepo, mockUserRepo, mockWalletRepo, mockLedgerRepo)

	ctx := context.TODO()
	userID := "user-uuid"
	wallet := &walletModel.Wallet{ID: "wallet-uuid"}
	params := transactionModel.PaginationParams{Page: 1, Limit: 10}

	mockWalletRepo.On("GetByUserID", ctx, userID).Return(wallet, nil)
	mockTxRepo.On("GetHistory", ctx, wallet.ID, params).Return(nil, int64(0), errors.New("db error"))

	// Run the function
	txs, meta, err := svc.GetHistory(ctx, userID, params)

	// Verify result
	assert.Error(t, err)
	assert.Nil(t, txs)
	assert.Nil(t, meta)

	// Make sure all mock if called and expected
	mockTxRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
	mockWalletRepo.AssertExpectations(t)
	mockLedgerRepo.AssertExpectations(t)
}
