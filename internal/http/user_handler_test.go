package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"fintrack-go/internal/db"
	"fintrack-go/internal/models"
)

func TestUserHandler_CreateUser(t *testing.T) {
	logger := zerolog.Nop()

	t.Run("success", func(t *testing.T) {
		mockDB := new(MockDBForHandler)
		handler := NewUserHandler(logger, mockDB)

		expectedUser := &models.User{
			ID:    "550e8400-e29b-41d4-a716-446655440000",
			Email: "test@example.com",
		}
		mockDB.On("CreateUser", mock.Anything, "test@example.com").Return(expectedUser, nil)

		reqBody := map[string]string{"email": "test@example.com"}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.CreateUser(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		assertJSONContentType(t, w)

		var resp models.User
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Equal(t, expectedUser.ID, resp.ID)
		assert.Equal(t, expectedUser.Email, resp.Email)

		mockDB.AssertExpectations(t)
	})

	t.Run("invalid email", func(t *testing.T) {
		mockDB := new(MockDBForHandler)
		handler := NewUserHandler(logger, mockDB)

		reqBody := map[string]string{"email": "invalid-email"}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.CreateUser(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assertJSONContentType(t, w)

		var resp map[string]interface{}
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		
		errObj := resp["error"].(map[string]interface{})
		assert.Contains(t, errObj["message"], "invalid email format")
	})

	t.Run("missing email", func(t *testing.T) {
		mockDB := new(MockDBForHandler)
		handler := NewUserHandler(logger, mockDB)

		reqBody := map[string]string{}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.CreateUser(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assertJSONContentType(t, w)

		var resp map[string]interface{}
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		
		errObj := resp["error"].(map[string]interface{})
		assert.Contains(t, errObj["message"], "email is required")
	})

	t.Run("duplicate email", func(t *testing.T) {
		mockDB := new(MockDBForHandler)
		handler := NewUserHandler(logger, mockDB)

		mockDB.On("CreateUser", mock.Anything, "duplicate@example.com").Return(nil, db.ErrDuplicateEmail)

		reqBody := map[string]string{"email": "duplicate@example.com"}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.CreateUser(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)
		assertJSONContentType(t, w)

		var resp map[string]interface{}
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		
		errObj := resp["error"].(map[string]interface{})
		assert.Contains(t, errObj["message"], "already exists")
		mockDB.AssertExpectations(t)
	})

	t.Run("database error", func(t *testing.T) {
		mockDB := new(MockDBForHandler)
		handler := NewUserHandler(logger, mockDB)

		mockDB.On("CreateUser", mock.Anything, "error@example.com").Return(nil, assert.AnError)

		reqBody := map[string]string{"email": "error@example.com"}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.CreateUser(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assertJSONContentType(t, w)

		var resp map[string]interface{}
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		
		errObj := resp["error"].(map[string]interface{})
		assert.Contains(t, errObj["message"], "Failed to create user")
		mockDB.AssertExpectations(t)
	})

	t.Run("invalid JSON", func(t *testing.T) {
		mockDB := new(MockDBForHandler)
		handler := NewUserHandler(logger, mockDB)

		req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBuffer([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.CreateUser(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assertJSONContentType(t, w)

		var resp map[string]interface{}
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		
		errObj := resp["error"].(map[string]interface{})
		assert.Contains(t, errObj["message"], "Invalid request body")
	})
}
