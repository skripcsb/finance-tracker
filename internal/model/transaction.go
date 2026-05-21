package model

import "time"

type TransactionType string

const (
	Income  TransactionType = "income"
	Expense TransactionType = "expense"
)

type Transaction struct {
	ID         int             `json:"id"`
	UserID     int             `json:"user_id"`
	Type       TransactionType `json:"type"`
	Amount     float64         `json:"amount"`
	CategoryID int             `json:"category_id"`
	Category   string          `json:"category,omitempty"`
	Note       string          `json:"note,omitempty"`
	Date       time.Time       `json:"date"`
	CreatedAt  time.Time       `json:"created_at"`
}

type CreateTransactionRequest struct {
	Type       TransactionType `json:"type"`
	Amount     float64         `json:"amount"`
	CategoryID int             `json:"category_id"`
	Note       string          `json:"note"`
	Date       string          `json:"date"` // формат YYYY-MM-DD
}

type TransactionFilter struct {
	Type       TransactionType
	CategoryID int
	From       string
	To         string
	Page       int
	PerPage    int
}
