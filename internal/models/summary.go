package models

import "time"

type CategorySummary struct {
	CategoryID   *string `json:"category_id"`
	CategoryName string  `json:"category_name"`
	Total        float64 `json:"total"`
}

type Summary struct {
	UserID    string            `json:"user_id"`
	From      time.Time         `json:"from"`
	To        time.Time         `json:"to"`
	Categories []CategorySummary `json:"categories"`
}
