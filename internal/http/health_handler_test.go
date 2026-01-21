package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"fintrack-go/internal/models"
)

type MockPoolForHealth struct {
	mock.Mock
}

func (m *MockPoolForHealth) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockPoolForHealth) Close() {}

func (m *MockPoolForHealth) GetUserByEmail(ctx context.Context, email string) (*models.User, error) { return nil, nil }
func (m *MockPoolForHealth) GetUserByID(ctx context.Context, id string) (*models.User, error) { return nil, nil }
func (m *MockPoolForHealth) CreateUser(ctx context.Context, email string) (*models.User, error) { return nil, nil }
func (m *MockPoolForHealth) CreateCategory(ctx context.Context, userID, name string) (*models.Category, error) { return nil, nil }
func (m *MockPoolForHealth) ListCategories(ctx context.Context, userID string) ([]models.Category, error) { return nil, nil }
func (m *MockPoolForHealth) GetCategoryByID(ctx context.Context, id string) (*models.Category, error) { return nil, nil }
func (m *MockPoolForHealth) CreateTransaction(ctx context.Context, userID string, categoryID *string, amount float64, description *string, occurredAt time.Time) (*models.Transaction, error) { return nil, nil }
func (m *MockPoolForHealth) ListTransactions(ctx context.Context, userID string, from, to *time.Time) ([]models.Transaction, error) { return nil, nil }
func (m *MockPoolForHealth) GetTransactionByID(ctx context.Context, id string) (*models.Transaction, error) { return nil, nil }
func (m *MockPoolForHealth) ValidateCategoryOwnership(ctx context.Context, categoryID, userID string) error { return nil }
func (m *MockPoolForHealth) GetSummary(ctx context.Context, userID string, from, to *time.Time) (*models.Summary, error) { return nil, nil }

func TestHealthHandler_Health(t *testing.T) {
	logger := zerolog.Nop()

	t.Run("success", func(t *testing.T) {
		mockPool := new(MockPoolForHealth)
		mockPool.On("Ping", mock.Anything).Return(nil)

		handler := NewHealthHandler(logger, mockPool)

		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		w := httptest.NewRecorder()

		handler.Health(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assertJSONContentType(t, w)

		var resp map[string]string
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Equal(t, "healthy", resp["status"])
		mockPool.AssertExpectations(t)
	})

	t.Run("database error", func(t *testing.T) {
		mockPool := new(MockPoolForHealth)
		mockPool.On("Ping", mock.Anything).Return(assert.AnError)

		handler := NewHealthHandler(logger, mockPool)

		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		w := httptest.NewRecorder()

		handler.Health(w, req)

		assert.Equal(t, http.StatusServiceUnavailable, w.Code)

		var resp map[string]interface{}
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		
		errObj := resp["error"].(map[string]interface{})
		assert.Contains(t, errObj["message"], "Service unavailable")
		mockPool.AssertExpectations(t)
	})

	t.Run("integration test with real DB", func(t *testing.T) {
		t.Skip("Skipping integration test in unit test file to avoid import cycle")
	})
}
