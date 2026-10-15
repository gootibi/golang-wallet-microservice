package repository

import (
	"context"
	"database/sql"

	"github.com/gootibi/golang-wallet-microservice/monolith/internal/transaction/model"
)

type TransactionRepository interface {
	CreateTX(ctx context.Context, tx *sql.Tx, t *model.Transaction) error
	GetByIdempotencyKey(ctx context.Context, key string) (*model.Transaction, error)
}

func NewMySQLTransactionRepositor(db *sql.DB) TransactionRepository {
	return &mysqlTransactionRepository{
		db: db,
	}
}

type mysqlTransactionRepository struct {
	db *sql.DB
}

// CreateTX implements [TransactionRepository].
func (r *mysqlTransactionRepository) CreateTX(ctx context.Context, tx *sql.Tx, t *model.Transaction) error {
	query := `
		INSERT INTO transactions ( id, sender_wallet_id, receiver_wallet_id, amount, description, idempotency_key, status ) VALUES(
     		?, ?, ?, ?, ?, ?, ?
		);
	`
	_, err := tx.ExecContext(ctx, query,
		t.ID,
		t.SenderWalletID,
		t.ReceiverWalletID,
		t.Amount,
		t.Description,
		t.IdempotencyKey,
		t.Status,
	)

	return err
}

// GetByIdempotencyKey implements [TransactionRepository].
func (r *mysqlTransactionRepository) GetByIdempotencyKey(ctx context.Context, key string) (*model.Transaction, error) {
	query := `SELECT id, sender_wallet_id, receiver_wallet_id, amount, description, idempotency_key, status, created_at FROM transactions WHERE idempotency_key = ?;`
	t := &model.Transaction{}
	var sender sql.NullString

	err := r.db.QueryRowContext(ctx, query, key).Scan(
		&t.ID,
		&sender,
		&t.ReceiverWalletID,
		&t.Amount,
		&t.Description,
		&t.IdempotencyKey,
		&t.Status,
		&t.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	if sender.Valid {
		t.SenderWalletID = &sender.String
	}

	return t, nil
}
