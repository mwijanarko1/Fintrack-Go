package http

import (
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

func TestSummaryHandler_GetSummary(t *testing.T) {
	logger := zerolog.Nop()

	t.Run("success", func(t *testing.T) {
		mockDB := new(MockDBForHandler)
		handler := NewSummaryHandler(logger, mockDB)

		userID := "550e8400-e29b-41d4-a716-446655440000"
		expectedSummary := &models.Summary{
			UserID: userID,
			Categories: []models.CategorySummary{
				{CategoryName: "Food", Total: 100.0},
				{CategoryName: "Transport", Total: 50.5},
			},
		}
		mockDB.On("GetSummary", mock.Anything, userID, (*time.Time)(nil), (*time.Time)(nil)).Return(expectedSummary, nil)

		q := url.Values{}
		q.Set("user_id", userID)
		req := httptest.NewRequest(http.MethodGet, "/summary?"+q.Encode(), nil)
		w := httptest.NewRecorder()

		handler.GetSummary(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assertJSONContentType(t, w)

		var resp models.Summary
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Equal(t, expectedSummary.UserID, resp.UserID)
		assert.Len(t, resp.Categories, 2)

		mockDB.AssertExpectations(t)
	})

	t.Run("with date range", func(t *testing.T) {
		mockDB := new(MockDBForHandler)
		handler := NewSummaryHandler(logger, mockDB)

		userID := "550e8400-e29b-41d4-a716-446655440000"
		expectedSummary := &models.Summary{
			UserID: userID,
			Categories:       []models.CategorySummary{},
		}
		// Truncate to seconds to match RFC3339 precision used in query params
		startDate := time.Now().Add(-48 * time.Hour).Truncate(time.Second).UTC()
		endDate := time.Now().Add(-24 * time.Hour).Truncate(time.Second).UTC()
		mockDB.On("GetSummary", mock.Anything, userID, &startDate, &endDate).Return(expectedSummary, nil)

		q := url.Values{}
		q.Set("user_id", userID)
		q.Set("from", startDate.Format(time.RFC3339))
		q.Set("to", endDate.Format(time.RFC3339))
		req := httptest.NewRequest(http.MethodGet, "/summary?"+q.Encode(), nil)
		w := httptest.NewRecorder()

		handler.GetSummary(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		mockDB.AssertExpectations(t)
	})

	t.Run("missing user_id", func(t *testing.T) {
		mockDB := new(MockDBForHandler)
		handler := NewSummaryHandler(logger, mockDB)

		req := httptest.NewRequest(http.MethodGet, "/summary", nil)
		w := httptest.NewRecorder()

		handler.GetSummary(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var resp map[string]interface{}
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)

		errObj := resp["error"].(map[string]interface{})
		assert.Contains(t, errObj["message"], "user_id query parameter is required")
	})

	t.Run("invalid user_id", func(t *testing.T) {
		mockDB := new(MockDBForHandler)
		handler := NewSummaryHandler(logger, mockDB)

		q := url.Values{}
		q.Set("user_id", "invalid-uuid")
		req := httptest.NewRequest(http.MethodGet, "/summary?"+q.Encode(), nil)
		w := httptest.NewRecorder()

		handler.GetSummary(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var resp map[string]interface{}
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)

		errObj := resp["error"].(map[string]interface{})
		assert.Contains(t, errObj["message"], "invalid UUID format")
	})

	t.Run("invalid from date format", func(t *testing.T) {
		mockDB := new(MockDBForHandler)
		handler := NewSummaryHandler(logger, mockDB)

		userID := "550e8400-e29b-41d4-a716-446655440000"
		q := url.Values{}
		q.Set("user_id", userID)
		q.Set("from", "invalid-date")
		req := httptest.NewRequest(http.MethodGet, "/summary?"+q.Encode(), nil)
		w := httptest.NewRecorder()

		handler.GetSummary(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var resp map[string]interface{}
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)

		errObj := resp["error"].(map[string]interface{})
		assert.Contains(t, errObj["message"], "Invalid 'from' date format")
	})

	t.Run("invalid to date format", func(t *testing.T) {
		mockDB := new(MockDBForHandler)
		handler := NewSummaryHandler(logger, mockDB)

		userID := "550e8400-e29b-41d4-a716-446655440000"
		q := url.Values{}
		q.Set("user_id", userID)
		q.Set("to", "invalid-date")
		req := httptest.NewRequest(http.MethodGet, "/summary?"+q.Encode(), nil)
		w := httptest.NewRecorder()

		handler.GetSummary(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var resp map[string]interface{}
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)

		errObj := resp["error"].(map[string]interface{})
		assert.Contains(t, errObj["message"], "Invalid 'to' date format")
	})

	t.Run("invalid date range (from > to)", func(t *testing.T) {
		mockDB := new(MockDBForHandler)
		handler := NewSummaryHandler(logger, mockDB)

		userID := "550e8400-e29b-41d4-a716-446655440000"
		q := url.Values{}
		q.Set("user_id", userID)
		q.Set("from", time.Now().Add(24*time.Hour).Format(time.RFC3339))
		q.Set("to", time.Now().Add(-24*time.Hour).Format(time.RFC3339))
		req := httptest.NewRequest(http.MethodGet, "/summary?"+q.Encode(), nil)
		w := httptest.NewRecorder()

		handler.GetSummary(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var resp map[string]interface{}
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)

		errObj := resp["error"].(map[string]interface{})
		assert.Contains(t, errObj["message"], "'from' date must be before or equal to 'to' date")
	})

	t.Run("database error", func(t *testing.T) {
		mockDB := new(MockDBForHandler)
		handler := NewSummaryHandler(logger, mockDB)

		userID := "550e8400-e29b-41d4-a716-446655440000"
		mockDB.On("GetSummary", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, assert.AnError)

		q := url.Values{}
		q.Set("user_id", userID)
		req := httptest.NewRequest(http.MethodGet, "/summary?"+q.Encode(), nil)
		w := httptest.NewRecorder()

		handler.GetSummary(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var resp map[string]interface{}
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)

		errObj := resp["error"].(map[string]interface{})
		assert.Contains(t, errObj["message"], "Failed to get summary")

		mockDB.AssertExpectations(t)
	})
}
