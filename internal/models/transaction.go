package models

import "time"

type Transaction struct {
	ID           string     `json:"id"`
	UserID       string     `json:"user_id"`
	CategoryID   *string    `json:"category_id,omitempty"`
	CategoryName *string    `json:"category_name,omitempty"`
	Amount       float64    `json:"amount"`
	Description  *string    `json:"description,omitempty"`
	OccurredAt   time.Time  `json:"occurred_at"`
	CreatedAt    time.Time  `json:"created_at"`
}

type CreateTransactionRequest struct {
	UserID      string  `json:"user_id"`
	CategoryID  *string `json:"category_id,omitempty"`
	Amount      float64 `json:"amount"`
	Description *string `json:"description,omitempty"`
	OccurredAt  *time.Time `json:"occurred_at,omitempty"`
}
