package model

import "time"

type WalletStatus string

const (
	ACTIVE WalletStatus = "active"
	FROZEN WalletStatus = "frozen"
)

type Wallet struct {
	ID        string       `json:"id"`
	UserID    string       `json:"user_id"`
	Balance   float64      `json:"balance"`
	Currency  string       `json:"currency"`
	Status    WalletStatus `json:"status"` // active, frozen
	Version   int          `json:"-"`      // used for optimistic control
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`
}
