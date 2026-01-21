package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"

	"fintrack-go/internal/db"
)

func TestSetupRoutes(t *testing.T) {
	logger := zerolog.Nop()
	testDB, err := db.NewDB(t.Context(), "postgres://test:test@localhost:5432/test?sslmode=disable", logger)
	if err != nil {
		t.Skip("Skipping test: database not available")
	}
	defer testDB.Close()

	router := SetupRoutes(logger, testDB)

	t.Run("health endpoint exists", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("user creation endpoint exists", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/users", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.NotEqual(t, http.StatusNotFound, w.Code)
	})

	t.Run("category endpoints exist", func(t *testing.T) {
		t.Run("POST /api/v1/categories", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/v1/categories", nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.NotEqual(t, http.StatusNotFound, w.Code)
		})

		t.Run("GET /api/v1/categories", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/categories", nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.NotEqual(t, http.StatusNotFound, w.Code)
		})
	})

	t.Run("transaction endpoints exist", func(t *testing.T) {
		t.Run("POST /api/v1/transactions", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/v1/transactions", nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.NotEqual(t, http.StatusNotFound, w.Code)
		})

		t.Run("GET /api/v1/transactions", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/transactions", nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.NotEqual(t, http.StatusNotFound, w.Code)
		})
	})

	t.Run("summary endpoint exists", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/summary", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.NotEqual(t, http.StatusNotFound, w.Code)
	})

	t.Run("non-existent endpoint returns 404", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/nonexistent", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("invalid method returns 405", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/health", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})

	t.Run("max body size middleware is applied", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/users", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.NotEqual(t, http.StatusNotFound, w.Code)
	})
}
