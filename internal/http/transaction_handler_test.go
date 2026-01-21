package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"fintrack-go/internal/models"
)

func TestTransactionHandler_CreateTransaction(t *testing.T) {
	logger := zerolog.Nop()

	t.Run("success with category", func(t *testing.T) {
		mockDB := new(MockDBForHandler)
		handler := NewTransactionHandler(logger, mockDB)

		categoryID := "660e8400-e29b-41d4-a716-446655440001"
		userID := "550e8400-e29b-41d4-a716-446655440000"
		expectedTxn := &models.Transaction{
			ID:          "770e8400-e29b-41d4-a716-446655440002",
			UserID:      userID,
			CategoryID:  &categoryID,
			Amount:      25.50,
			Description:  strPtr("Lunch"),
			OccurredAt:  time.Now(),
		}
		mockDB.On("CreateTransaction", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(expectedTxn, nil)

		reqBody := map[string]interface{}{
			"user_id":     userID,
			"category_id": categoryID,
			"amount":      25.50,
			"description": "Lunch",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/transactions", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.CreateTransaction(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		assertJSONContentType(t, w)

		var resp models.Transaction
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Equal(t, expectedTxn.ID, resp.ID)
		assert.Equal(t, expectedTxn.Amount, resp.Amount)

		mockDB.AssertExpectations(t)
	})

	t.Run("success without category", func(t *testing.T) {
		mockDB := new(MockDBForHandler)
		handler := NewTransactionHandler(logger, mockDB)

		userID := "550e8400-e29b-41d4-a716-446655440000"
		expectedTxn := &models.Transaction{
			ID:          "770e8400-e29b-41d4-a716-446655440002",
			UserID:      userID,
			CategoryID:  nil,
			Amount:      15.00,
			Description:  nil,
			OccurredAt:  time.Now(),
		}
		mockDB.On("CreateTransaction", mock.Anything, mock.Anything, (*string)(nil), mock.Anything, (*string)(nil), mock.Anything).Return(expectedTxn, nil)

		reqBody := map[string]interface{}{
			"user_id": userID,
			"amount":  15.00,
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/transactions", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.CreateTransaction(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		assertJSONContentType(t, w)

		var resp models.Transaction
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Nil(t, resp.CategoryID)

		mockDB.AssertExpectations(t)
	})

	t.Run("invalid amount (negative)", func(t *testing.T) {
		mockDB := new(MockDBForHandler)
		handler := NewTransactionHandler(logger, mockDB)

		reqBody := map[string]interface{}{
			"user_id": "550e8400-e29b-41d4-a716-446655440000",
			"amount":  -10.0,
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/transactions", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.CreateTransaction(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var resp map[string]interface{}
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		
		errObj := resp["error"].(map[string]interface{})
		assert.Contains(t, errObj["message"], "amount must be greater than 0")
	})

	t.Run("invalid amount (zero)", func(t *testing.T) {
		mockDB := new(MockDBForHandler)
		handler := NewTransactionHandler(logger, mockDB)

		reqBody := map[string]interface{}{
			"user_id": "550e8400-e29b-41d4-a716-446655440000",
			"amount":  0.0,
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/transactions", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.CreateTransaction(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var resp map[string]interface{}
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		
		errObj := resp["error"].(map[string]interface{})
		assert.Contains(t, errObj["message"], "amount must be greater than 0")
	})

	t.Run("invalid user_id", func(t *testing.T) {
		mockDB := new(MockDBForHandler)
		handler := NewTransactionHandler(logger, mockDB)

		reqBody := map[string]interface{}{
			"user_id": "invalid-uuid",
			"amount":  10.0,
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/transactions", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.CreateTransaction(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var resp map[string]interface{}
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		
		errObj := resp["error"].(map[string]interface{})
		assert.Contains(t, errObj["message"], "invalid UUID format")
	})

	t.Run("invalid category_id", func(t *testing.T) {
		mockDB := new(MockDBForHandler)
		handler := NewTransactionHandler(logger, mockDB)

		categoryID := "invalid-uuid"
		reqBody := map[string]interface{}{
			"user_id":     "550e8400-e29b-41d4-a716-446655440000",
			"category_id": categoryID,
			"amount":      10.0,
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/transactions", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.CreateTransaction(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var resp map[string]interface{}
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		
		errObj := resp["error"].(map[string]interface{})
		assert.Contains(t, errObj["message"], "invalid UUID format")
	})

	t.Run("database error", func(t *testing.T) {
		mockDB := new(MockDBForHandler)
		handler := NewTransactionHandler(logger, mockDB)

		mockDB.On("CreateTransaction", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, assert.AnError)

		reqBody := map[string]interface{}{
			"user_id": "550e8400-e29b-41d4-a716-446655440000",
			"amount":  10.0,
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/transactions", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.CreateTransaction(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var resp map[string]interface{}
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		
		errObj := resp["error"].(map[string]interface{})
		assert.Contains(t, errObj["message"], "Failed to create transaction")
		mockDB.AssertExpectations(t)
	})
}

func TestTransactionHandler_ListTransactions(t *testing.T) {
	logger := zerolog.Nop()

	t.Run("success", func(t *testing.T) {
		mockDB := new(MockDBForHandler)
		handler := NewTransactionHandler(logger, mockDB)

		userID := "550e8400-e29b-41d4-a716-446655440000"
		expectedTxns := []models.Transaction{
			{ID: "770e8400-e29b-41d4-a716-446655440002", UserID: userID, Amount: 10.0},
			{ID: "770e8400-e29b-41d4-a716-446655440003", UserID: userID, Amount: 20.0},
		}
		mockDB.On("ListTransactions", mock.Anything, userID, (*time.Time)(nil), (*time.Time)(nil)).Return(expectedTxns, nil)

		q := url.Values{}
		q.Set("user_id", userID)
		req := httptest.NewRequest(http.MethodGet, "/transactions?"+q.Encode(), nil)
		w := httptest.NewRecorder()

		handler.ListTransactions(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp []models.Transaction
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Len(t, resp, 2)
		mockDB.AssertExpectations(t)
	})

	t.Run("with date range", func(t *testing.T) {
		mockDB := new(MockDBForHandler)
		handler := NewTransactionHandler(logger, mockDB)

		userID := "550e8400-e29b-41d4-a716-446655440000"
		expectedTxns := []models.Transaction{
			{ID: "770e8400-e29b-41d4-a716-446655440002", UserID: userID, Amount: 10.0},
		}
		// Truncate to seconds to match RFC3339 precision used in query params
		startDate := time.Now().Add(-48 * time.Hour).Truncate(time.Second).UTC()
		endDate := time.Now().Add(-24 * time.Hour).Truncate(time.Second).UTC()
		mockDB.On("ListTransactions", mock.Anything, userID, &startDate, &endDate).Return(expectedTxns, nil)

		q := url.Values{}
		q.Set("user_id", userID)
		q.Set("from", startDate.Format(time.RFC3339))
		q.Set("to", endDate.Format(time.RFC3339))
		req := httptest.NewRequest(http.MethodGet, "/transactions?"+q.Encode(), nil)
		w := httptest.NewRecorder()

		handler.ListTransactions(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp []models.Transaction
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Len(t, resp, 1)
		mockDB.AssertExpectations(t)
	})

	t.Run("missing user_id", func(t *testing.T) {
		mockDB := new(MockDBForHandler)
		handler := NewTransactionHandler(logger, mockDB)

		req := httptest.NewRequest(http.MethodGet, "/transactions", nil)
		w := httptest.NewRecorder()

		handler.ListTransactions(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var resp map[string]interface{}
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		
		errObj := resp["error"].(map[string]interface{})
		assert.Contains(t, errObj["message"], "user_id query parameter is required")
	})

	t.Run("invalid date format", func(t *testing.T) {
		mockDB := new(MockDBForHandler)
		handler := NewTransactionHandler(logger, mockDB)

		userID := "550e8400-e29b-41d4-a716-446655440000"
		q := url.Values{}
		q.Set("user_id", userID)
		q.Set("from", "invalid-date")
		req := httptest.NewRequest(http.MethodGet, "/transactions?"+q.Encode(), nil)
		w := httptest.NewRecorder()

		handler.ListTransactions(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var resp map[string]interface{}
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		
		errObj := resp["error"].(map[string]interface{})
		assert.Contains(t, errObj["message"], "Invalid 'from' date format")
	})

	t.Run("invalid date range (from > to)", func(t *testing.T) {
		mockDB := new(MockDBForHandler)
		handler := NewTransactionHandler(logger, mockDB)

		userID := "550e8400-e29b-41d4-a716-446655440000"
		q := url.Values{}
		q.Set("user_id", userID)
		q.Set("from", time.Now().Add(24*time.Hour).Format(time.RFC3339))
		q.Set("to", time.Now().Add(-24*time.Hour).Format(time.RFC3339))
		req := httptest.NewRequest(http.MethodGet, "/transactions?"+q.Encode(), nil)
		w := httptest.NewRecorder()

		handler.ListTransactions(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var resp map[string]interface{}
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		
		errObj := resp["error"].(map[string]interface{})
		assert.Contains(t, errObj["message"], "'from' date must be before or equal to 'to' date")
	})
}

func strPtr(s string) *string {
	return &s
}
