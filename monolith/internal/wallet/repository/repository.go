package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/gootibi/golang-wallet-microservice/monolith/internal/wallet/model"
)

type WalletRepository interface {
	CreateTx(ctx context.Context, tx *sql.Tx, w *model.Wallet) error
	GetByUserID(ctx context.Context, userID string) (*model.Wallet, error)
	UpdateBalanceTx(ctx context.Context, tx *sql.Tx, walletID string, newBalance float64, currentWersion int) error
}

func NewMySQLWalletRepository(db *sql.DB) WalletRepository {
	return &mysqlWalletRepository{
		db: db,
	}
}

type mysqlWalletRepository struct {
	db *sql.DB
}

// CreateTx implements [WalletRepository].
func (r *mysqlWalletRepository) CreateTx(ctx context.Context, tx *sql.Tx, w *model.Wallet) error {
	query := `INSERT INTO wallets ( id, user_id, balance, currency, status ) VALUES (?, ?, ?, ?, ?);`
	_, err := tx.ExecContext(ctx, query, w.ID, w.UserID, w.Balance, w.Currency, w.Status)
	return err
}

// GetByUserID implements [WalletRepository].
func (r *mysqlWalletRepository) GetByUserID(ctx context.Context, userID string) (*model.Wallet, error) {
	query := `SELECT id, user_id, balance, currency, status, version, created_at, updated_at  FROM wallets WHERE id = ? AND deleted_at IS NULL;`
	w := &model.Wallet{}

	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&w.ID,
		&w.UserID,
		&w.Balance,
		&w.Currency,
		&w.Status,
		&w.Version,
		&w.CreatedAt,
		&w.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("Wallet not found")
		}
		return nil, err
	}

	return w, nil
}

// UpdateWallet implements [WalletRepository].
func (r *mysqlWalletRepository) UpdateBalanceTx(ctx context.Context, tx *sql.Tx, walletID string, newBalance float64, currentVersion int) error {
	query := `UPDATE wallets SET balance = ?, version = version + 1 WHERE id = ? AND version = ?;`

	result, err := tx.ExecContext(ctx, query, newBalance, walletID, currentVersion)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	// If 0 rows affected, it means database version has changhed - concurreny conflict
	if rowsAffected == 0 {
		return errors.New("Concurrent update detected: version mismatch")
	}

	return nil
}
