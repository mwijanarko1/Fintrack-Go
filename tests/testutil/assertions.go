package testutil

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func AssertJSONResponse(t *testing.T, w *http.Response, expectedCode int, expectedBody interface{}) {
	t.Helper()
	
	require.Equal(t, expectedCode, w.StatusCode, "expected status code %d, got %d", expectedCode, w.StatusCode)
	
	var actualBody interface{}
	err := json.NewDecoder(w.Body).Decode(&actualBody)
	require.NoError(t, err, "failed to decode response body")
	
	assert.Equal(t, expectedBody, actualBody, "response bodies don't match")
}

func AssertErrorResponse(t *testing.T, w *http.Response, expectedCode int, expectedMessage string) {
	t.Helper()
	
	require.Equal(t, expectedCode, w.StatusCode, "expected status code %d, got %d", expectedCode, w.StatusCode)
	
	var resp map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err, "failed to decode error response")
	
	errObj, ok := resp["error"].(map[string]interface{})
	require.True(t, ok, "error field not present or not an object")
	
	actualMsg := errObj["message"].(string)
	assert.Contains(t, actualMsg, expectedMessage, "error message doesn't contain expected text")
}

func AssertJSONContentType(t *testing.T, resp *http.Response) {
	t.Helper()
	
	contentType := resp.Header.Get("Content-Type")
	assert.Equal(t, "application/json", contentType, "expected JSON content type")
}

func AssertRequestIDHeader(t *testing.T, resp *http.Response) {
	t.Helper()
	
	requestID := resp.Header.Get("X-Request-ID")
	assert.NotEmpty(t, requestID, "expected X-Request-ID header to be present")
}

func AssertSuccessResponse(t *testing.T, w *http.Response) {
	t.Helper()
	
	assert.Contains(t, []int{http.StatusOK, http.StatusCreated}, w.StatusCode, 
		"expected success status code (200 or 201), got %d", w.StatusCode)
}

func AssertError(t *testing.T, err error, expectedMessage string) {
	t.Helper()
	
	require.Error(t, err, "expected error but got nil")
	assert.Contains(t, err.Error(), expectedMessage, "error message doesn't contain expected text")
}
