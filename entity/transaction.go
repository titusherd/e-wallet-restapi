package entity

import "time"

type Transaction struct {
	ID              int       `json:"id"`
	FromWalletID    *int      `json:"from_wallet_id,omitempty"`
	ToWalletID      int       `json:"to_wallet_id"`
	Amount          float64   `json:"amount"`
	Description     string    `json:"description"`
	SourceOfFundID  int       `json:"source_of_fund_id"`
	TransactionType string    `json:"transaction_type"`
	CreatedAt       time.Time `json:"created_at"`
	// Additional fields for response
	FromWalletNumber string `json:"from_wallet_number,omitempty"`
	ToWalletNumber   string `json:"to_wallet_number"`
	RecipientName    string `json:"recipient_name"`
}
