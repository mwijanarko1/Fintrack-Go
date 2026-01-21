package testutil

import (
	"time"

	"github.com/google/uuid"
	"fintrack-go/internal/models"
)

type UserFixture struct {
	ID    string
	Email  string
	Now    time.Time
}

func NewUserFixture() *UserFixture {
	now := time.Now()
	return &UserFixture{
		ID:   uuid.New().String(),
		Email: "test-" + uuid.New().String() + "@example.com",
		Now:  now,
	}
}

type CategoryFixture struct {
	ID     string
	UserID string
	Name   string
	Now    time.Time
}

func NewCategoryFixture(userID string) *CategoryFixture {
	now := time.Now()
	return &CategoryFixture{
		ID:     uuid.New().String(),
		UserID: userID,
		Name:   "Test Category",
		Now:    now,
	}
}

type TransactionFixture struct {
	ID          string
	UserID      string
	CategoryID  *string
	Amount      float64
	Description *string
	OccurredAt  time.Time
	CreatedAt   time.Time
}

func NewTransactionFixture(userID string, categoryID *string) *TransactionFixture {
	now := time.Now()
	desc := "Test transaction"
	return &TransactionFixture{
		ID:          uuid.New().String(),
		UserID:      userID,
		CategoryID:  categoryID,
		Amount:      10.50,
		Description:  &desc,
		OccurredAt:  now.Add(-time.Hour),
		CreatedAt:   now,
	}
}

func NewUserModelFromFixture(f *UserFixture) *models.User {
	return &models.User{
		ID:        f.ID,
		Email:     f.Email,
		CreatedAt: f.Now,
	}
}

func NewCategoryModelFromFixture(f *CategoryFixture) *models.Category {
	return &models.Category{
		ID:        f.ID,
		UserID:    f.UserID,
		Name:      f.Name,
		CreatedAt: f.Now,
	}
}

func NewTransactionModelFromFixture(f *TransactionFixture) *models.Transaction {
	return &models.Transaction{
		ID:          f.ID,
		UserID:      f.UserID,
		CategoryID:  f.CategoryID,
		Amount:      f.Amount,
		Description:  f.Description,
		OccurredAt:  f.OccurredAt,
		CreatedAt:   f.CreatedAt,
	}
}

type BulkTransactionOptions struct {
	UserID      string
	CategoryID  *string
	Count       int
	Amount      float64
	StartDate   *time.Time
}

func CreateBulkTransactionFixtures(opts BulkTransactionOptions) []*TransactionFixture {
	if opts.Count <= 0 {
		opts.Count = 10
	}
	if opts.Amount <= 0 {
		opts.Amount = 10.0
	}

	fixtures := make([]*TransactionFixture, opts.Count)
	startTime := time.Now()
	if opts.StartDate != nil {
		startTime = *opts.StartDate
	}

	for i := 0; i < opts.Count; i++ {
		occurredAt := startTime.Add(time.Duration(i) * time.Hour)
		desc := "Bulk transaction " + string(rune('0'+i))
		
		fixtures[i] = &TransactionFixture{
			ID:          uuid.New().String(),
			UserID:      opts.UserID,
			CategoryID:  opts.CategoryID,
			Amount:      opts.Amount,
			Description:  &desc,
			OccurredAt:  occurredAt,
			CreatedAt:   time.Now(),
		}
	}

	return fixtures
}
