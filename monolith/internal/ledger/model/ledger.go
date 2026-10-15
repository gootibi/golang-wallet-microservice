package model

import "time"

type LedgerEntryType string

const (
	DEBIT  LedgerEntryType = "debit"
	CREDIT LedgerEntryType = "credit"
)

type LedgerEntry struct {
	ID            string          `json:"id"`
	WalletID      string          `json:"wallet_id"`
	TransactionID string          `json:"transaction_id"`
	EntryType     LedgerEntryType `json:"entry_type"` // credit (+) or debit (-)
	Amount        float64         `json:"amount"`
	CreatedAt     time.Time       `json:"created_at"`
}
