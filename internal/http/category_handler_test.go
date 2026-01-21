package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"fintrack-go/internal/db"
	"fintrack-go/internal/models"
)

func TestCategoryHandler_CreateCategory(t *testing.T) {
	logger := zerolog.Nop()

	t.Run("success", func(t *testing.T) {
		mockDB := new(MockDBForHandler)
		handler := NewCategoryHandler(logger, mockDB)

		expectedCat := &models.Category{
			ID:     "660e8400-e29b-41d4-a716-446655440001",
			UserID: "550e8400-e29b-41d4-a716-446655440000",
			Name:   "Food",
		}
		mockDB.On("CreateCategory", mock.Anything, mock.Anything, "Food").Return(expectedCat, nil)

		reqBody := map[string]interface{}{
			"user_id": "550e8400-e29b-41d4-a716-446655440000",
			"name":    "Food",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/categories", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.CreateCategory(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		assertJSONContentType(t, w)

		var resp models.Category
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Equal(t, expectedCat.ID, resp.ID)
		assert.Equal(t, expectedCat.Name, resp.Name)

		mockDB.AssertExpectations(t)
	})

	t.Run("invalid user_id", func(t *testing.T) {
		mockDB := new(MockDBForHandler)
		handler := NewCategoryHandler(logger, mockDB)

		reqBody := map[string]interface{}{
			"user_id": "invalid-uuid",
			"name":    "Food",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/categories", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.CreateCategory(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var resp map[string]interface{}
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		
		errObj := resp["error"].(map[string]interface{})
		assert.Contains(t, errObj["message"], "invalid UUID format")
	})

	t.Run("invalid name (empty)", func(t *testing.T) {
		mockDB := new(MockDBForHandler)
		handler := NewCategoryHandler(logger, mockDB)

		reqBody := map[string]interface{}{
			"user_id": "550e8400-e29b-41d4-a716-446655440000",
			"name":    "",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/categories", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.CreateCategory(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var resp map[string]interface{}
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		
		errObj := resp["error"].(map[string]interface{})
		assert.Contains(t, errObj["message"], "category name is required")
	})

	t.Run("user not found", func(t *testing.T) {
		mockDB := new(MockDBForHandler)
		handler := NewCategoryHandler(logger, mockDB)

		mockDB.On("CreateCategory", mock.Anything, mock.Anything, mock.Anything).Return(nil, db.ErrUserNotFound)

		reqBody := map[string]interface{}{
			"user_id": "550e8400-e29b-41d4-a716-446655440000",
			"name":    "Food",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/categories", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.CreateCategory(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)

		var resp map[string]interface{}
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		
		errObj := resp["error"].(map[string]interface{})
		assert.Contains(t, errObj["message"], "User not found")
		mockDB.AssertExpectations(t)
	})

	t.Run("duplicate category name", func(t *testing.T) {
		mockDB := new(MockDBForHandler)
		handler := NewCategoryHandler(logger, mockDB)

		mockDB.On("CreateCategory", mock.Anything, mock.Anything, mock.Anything).Return(nil, db.ErrDuplicateCategory)

		reqBody := map[string]interface{}{
			"user_id": "550e8400-e29b-41d4-a716-446655440000",
			"name":    "Food",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/categories", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.CreateCategory(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)

		var resp map[string]interface{}
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		
		errObj := resp["error"].(map[string]interface{})
		assert.Contains(t, errObj["message"], "already exists")
		mockDB.AssertExpectations(t)
	})
}

func TestCategoryHandler_ListCategories(t *testing.T) {
	logger := zerolog.Nop()

	t.Run("success", func(t *testing.T) {
		mockDB := new(MockDBForHandler)
		handler := NewCategoryHandler(logger, mockDB)

		userID := "550e8400-e29b-41d4-a716-446655440000"
		expectedCats := []models.Category{
			{ID: "660e8400-e29b-41d4-a716-446655440001", UserID: userID, Name: "Food"},
			{ID: "660e8400-e29b-41d4-a716-446655440002", UserID: userID, Name: "Transport"},
		}
		mockDB.On("ListCategories", mock.Anything, userID).Return(expectedCats, nil)

		q := url.Values{}
		q.Set("user_id", userID)
		req := httptest.NewRequest(http.MethodGet, "/categories?"+q.Encode(), nil)
		w := httptest.NewRecorder()

		handler.ListCategories(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp []models.Category
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Len(t, resp, 2)
		mockDB.AssertExpectations(t)
	})

	t.Run("missing user_id", func(t *testing.T) {
		mockDB := new(MockDBForHandler)
		handler := NewCategoryHandler(logger, mockDB)

		req := httptest.NewRequest(http.MethodGet, "/categories", nil)
		w := httptest.NewRecorder()

		handler.ListCategories(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var resp map[string]interface{}
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		
		errObj := resp["error"].(map[string]interface{})
		assert.Contains(t, errObj["message"], "user_id query parameter is required")
	})

	t.Run("invalid user_id", func(t *testing.T) {
		mockDB := new(MockDBForHandler)
		handler := NewCategoryHandler(logger, mockDB)

		q := url.Values{}
		q.Set("user_id", "invalid-uuid")
		req := httptest.NewRequest(http.MethodGet, "/categories?"+q.Encode(), nil)
		w := httptest.NewRecorder()

		handler.ListCategories(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var resp map[string]interface{}
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		
		errObj := resp["error"].(map[string]interface{})
		assert.Contains(t, errObj["message"], "invalid UUID format")
	})

	t.Run("database error", func(t *testing.T) {
		mockDB := new(MockDBForHandler)
		handler := NewCategoryHandler(logger, mockDB)

		mockDB.On("ListCategories", mock.Anything, mock.Anything).Return(nil, assert.AnError)

		q := url.Values{}
		q.Set("user_id", "550e8400-e29b-41d4-a716-446655440000")
		req := httptest.NewRequest(http.MethodGet, "/categories?"+q.Encode(), nil)
		w := httptest.NewRecorder()

		handler.ListCategories(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var resp map[string]interface{}
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		
		errObj := resp["error"].(map[string]interface{})
		assert.Contains(t, errObj["message"], "Failed to list categories")
		mockDB.AssertExpectations(t)
	})
}
