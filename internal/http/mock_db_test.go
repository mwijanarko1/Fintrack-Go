package http

import (
	"context"
	"time"

	"github.com/stretchr/testify/mock"
	"fintrack-go/internal/models"
)

type MockDBForHandler struct {
	mock.Mock
}

func (m *MockDBForHandler) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockDBForHandler) CreateUser(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockDBForHandler) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockDBForHandler) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockDBForHandler) CreateCategory(ctx context.Context, userID, name string) (*models.Category, error) {
	args := m.Called(ctx, userID, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Category), args.Error(1)
}

func (m *MockDBForHandler) ListCategories(ctx context.Context, userID string) ([]models.Category, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Category), args.Error(1)
}

func (m *MockDBForHandler) GetCategoryByID(ctx context.Context, id string) (*models.Category, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Category), args.Error(1)
}

func (m *MockDBForHandler) CreateTransaction(ctx context.Context, userID string, categoryID *string, amount float64, description *string, occurredAt time.Time) (*models.Transaction, error) {
	args := m.Called(ctx, userID, categoryID, amount, description, occurredAt)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Transaction), args.Error(1)
}

func (m *MockDBForHandler) ListTransactions(ctx context.Context, userID string, from, to *time.Time) ([]models.Transaction, error) {
	args := m.Called(ctx, userID, from, to)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Transaction), args.Error(1)
}

func (m *MockDBForHandler) GetTransactionByID(ctx context.Context, id string) (*models.Transaction, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Transaction), args.Error(1)
}

func (m *MockDBForHandler) ValidateCategoryOwnership(ctx context.Context, categoryID, userID string) error {
	args := m.Called(ctx, categoryID, userID)
	return args.Error(0)
}

func (m *MockDBForHandler) GetSummary(ctx context.Context, userID string, from, to *time.Time) (*models.Summary, error) {
	args := m.Called(ctx, userID, from, to)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Summary), args.Error(1)
}
