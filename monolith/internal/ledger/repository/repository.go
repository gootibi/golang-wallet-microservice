package repository

import (
	"context"
	"database/sql"

	"github.com/gootibi/golang-wallet-microservice/monolith/internal/ledger/model"
)

type LedgerRepository interface {
	CreateTx(ctx context.Context, tx *sql.Tx, entry *model.LedgerEntry) error
	GetBalanceByWalletID(ctx context.Context, walletID string) (float64, error)
	GetEntriesByWalletID(ctx context.Context, walletID string) ([]model.LedgerEntry, error)
}

func NewMySQLLedgerRepository(db *sql.DB) LedgerRepository {
	return &mysqlLedgerRepository{
		db: db,
	}
}

type mysqlLedgerRepository struct {
	db *sql.DB
}

// CreateTx implements [LedgerRepository].
func (r *mysqlLedgerRepository) CreateTx(ctx context.Context, tx *sql.Tx, entry *model.LedgerEntry) error {
	query := `INSERT INTO ledger_entries (id, wallet_id, transaction_id, entry_type, amount) VALUES (?, ?, ?, ?, ?);`
	_, err := tx.ExecContext(ctx, query,
		entry.ID,
		entry.WalletID,
		entry.TransactionID,
		entry.EntryType,
		entry.Amount,
	)
	return err
}

// GetBalanceByWalletID implements [LedgerRepository].
func (r *mysqlLedgerRepository) GetBalanceByWalletID(ctx context.Context, walletID string) (float64, error) {
	// balance = summ(credit ) - sum(debit)
	query := `
		SELECT
		    COALESCE(SUM(CASE WHEN entry_type = 'credit' THEN amount ELSE 0 END),0) -
		    COALESCE(SUM(CASE WHEN entry_type = 'debit' THEN amount ELSE 0 END),0)
		FROM
    		ledger_entries
		WHERE
    		wallet_id = ?;
	`
	var balance float64

	err := r.db.QueryRowContext(ctx, query, walletID).Scan(&balance)
	if err != nil {
		return 0, err
	}

	return balance, nil
}

// GetEntriesByWalletID implements [LedgerRepository].
func (r *mysqlLedgerRepository) GetEntriesByWalletID(ctx context.Context, walletID string) ([]model.LedgerEntry, error) {
	query := `SELECT id, wallet_id, transaction_id, entry_type, amount, created_at FROM ledger_entries WHERE wallet_id = ? ORDER BY created_at DESC;`

	rows, err := r.db.QueryContext(ctx, query, walletID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []model.LedgerEntry
	for rows.Next() {
		var entry model.LedgerEntry
		if err := rows.Scan(
			&entry.ID,
			&entry.WalletID,
			&entry.TransactionID,
			&entry.EntryType,
			&entry.Amount,
			&entry.CreatedAt,
		); err != nil {
			return nil, err
		}

		entries = append(entries, entry)
	}

	return entries, nil
}
