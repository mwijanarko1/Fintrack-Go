package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func assertJSONContentType(t *testing.T, w *httptest.ResponseRecorder) {
	t.Helper()
	contentType := w.Header().Get("Content-Type")
	assert.Equal(t, "application/json", contentType, "expected JSON content type")
}

func TestHandler_respondWithJSON(t *testing.T) {
	logger := zerolog.Nop()
	handler := NewHandler(logger)

	t.Run("success with payload", func(t *testing.T) {
		payload := map[string]string{"key": "value"}
		w := httptest.NewRecorder()

		handler.respondWithJSON(w, http.StatusOK, payload)

		assert.Equal(t, http.StatusOK, w.Code)
		assertJSONContentType(t, w)

		var resp map[string]string
		err := json.NewDecoder(w.Body).Decode(&resp)
		assert.NoError(t, err)
		assert.Equal(t, payload, resp)
	})

	t.Run("success with nil payload", func(t *testing.T) {
		w := httptest.NewRecorder()

		handler.respondWithJSON(w, http.StatusOK, nil)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Empty(t, w.Body.Bytes())
	})

	t.Run("created status", func(t *testing.T) {
		payload := map[string]string{"id": "123"}
		w := httptest.NewRecorder()

		handler.respondWithJSON(w, http.StatusCreated, payload)

		assert.Equal(t, http.StatusCreated, w.Code)

		var resp map[string]string
		err := json.NewDecoder(w.Body).Decode(&resp)
		assert.NoError(t, err)
		assert.Equal(t, payload, resp)
	})
}

func TestHandler_respondWithError(t *testing.T) {
	logger := zerolog.Nop()
	handler := NewHandler(logger)

	t.Run("bad request", func(t *testing.T) {
		w := httptest.NewRecorder()

		handler.respondWithError(w, http.StatusBadRequest, "Invalid input", nil)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assertJSONContentType(t, w)

		var resp map[string]interface{}
		err := json.NewDecoder(w.Body).Decode(&resp)
		assert.NoError(t, err)
		
		errObj := resp["error"].(map[string]interface{})
		assert.Equal(t, "Bad Request", errObj["code"])
		assert.Equal(t, "Invalid input", errObj["message"])
		assert.Nil(t, errObj["details"])
	})

	t.Run("with details", func(t *testing.T) {
		w := httptest.NewRecorder()
		// Use map[string]interface{} to match what the JSON decoder produces
		details := map[string]interface{}{"field": "email", "value": "invalid"}

		handler.respondWithError(w, http.StatusBadRequest, "Invalid email", details)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var resp map[string]interface{}
		err := json.NewDecoder(w.Body).Decode(&resp)
		assert.NoError(t, err)
		
		errObj := resp["error"].(map[string]interface{})
		assert.Equal(t, "Invalid email", errObj["message"])
		assert.Equal(t, details, errObj["details"])
	})

	t.Run("not found", func(t *testing.T) {
		w := httptest.NewRecorder()

		handler.respondWithError(w, http.StatusNotFound, "User not found", nil)

		assert.Equal(t, http.StatusNotFound, w.Code)

		var resp map[string]interface{}
		err := json.NewDecoder(w.Body).Decode(&resp)
		assert.NoError(t, err)
		
		errObj := resp["error"].(map[string]interface{})
		assert.Equal(t, "Not Found", errObj["code"])
		assert.Equal(t, "User not found", errObj["message"])
	})

	t.Run("internal server error", func(t *testing.T) {
		w := httptest.NewRecorder()

		handler.respondWithError(w, http.StatusInternalServerError, "Database error", nil)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var resp map[string]interface{}
		err := json.NewDecoder(w.Body).Decode(&resp)
		assert.NoError(t, err)
		
		errObj := resp["error"].(map[string]interface{})
		assert.Equal(t, "Internal Server Error", errObj["code"])
		assert.Equal(t, "Database error", errObj["message"])
	})
}

func TestHandler_JSONEncoding(t *testing.T) {
	logger := zerolog.Nop()
	handler := NewHandler(logger)

	t.Run("handles special characters in JSON", func(t *testing.T) {
		payload := map[string]string{
			"message": "Test with \"quotes\" and 'apostrophes'",
		}
		w := httptest.NewRecorder()

		handler.respondWithJSON(w, http.StatusOK, payload)

		var resp map[string]string
		err := json.NewDecoder(w.Body).Decode(&resp)
		assert.NoError(t, err)
		assert.Equal(t, payload, resp)
	})

	t.Run("handles unicode in JSON", func(t *testing.T) {
		payload := map[string]string{
			"message": "Hello 世界 🌍",
		}
		w := httptest.NewRecorder()

		handler.respondWithJSON(w, http.StatusOK, payload)

		var resp map[string]string
		err := json.NewDecoder(w.Body).Decode(&resp)
		assert.NoError(t, err)
		assert.Equal(t, payload, resp)
	})

	t.Run("handles null values in JSON", func(t *testing.T) {
		payload := map[string]interface{}{
			"string": nil,
			"number": 123,
		}
		w := httptest.NewRecorder()

		handler.respondWithJSON(w, http.StatusOK, payload)

		var resp map[string]interface{}
		err := json.NewDecoder(w.Body).Decode(&resp)
		assert.NoError(t, err)
		assert.Nil(t, resp["string"])
		// JSON numbers are decoded as float64 into interface{}
		assert.EqualValues(t, 123, resp["number"])
	})
}
