package db

import (
	"context"
	"time"

	"fintrack-go/internal/models"
)

type Database interface {
	Ping(ctx context.Context) error
	CreateUser(ctx context.Context, email string) (*models.User, error)
	GetUserByID(ctx context.Context, id string) (*models.User, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	CreateCategory(ctx context.Context, userID, name string) (*models.Category, error)
	ListCategories(ctx context.Context, userID string) ([]models.Category, error)
	GetCategoryByID(ctx context.Context, id string) (*models.Category, error)
	CreateTransaction(ctx context.Context, userID string, categoryID *string, amount float64, description *string, occurredAt time.Time) (*models.Transaction, error)
	ListTransactions(ctx context.Context, userID string, from, to *time.Time) ([]models.Transaction, error)
	GetTransactionByID(ctx context.Context, id string) (*models.Transaction, error)
	ValidateCategoryOwnership(ctx context.Context, categoryID, userID string) error
	GetSummary(ctx context.Context, userID string, from, to *time.Time) (*models.Summary, error)
}

var _ Database = (*DB)(nil)
