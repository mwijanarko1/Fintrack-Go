package integration

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"fintrack-go/tests/testutil"
)

func TestConcurrency_CreateUsers(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	server := testutil.SetupTestServer(t)
	defer server.Cleanup(t)

	email := "concurrent-" + time.Now().Format("20060102-150405") + "@example.com"

	errors := make(chan error, 100)
	successes := make(chan bool, 100)

	for i := 0; i < 100; i++ {
		go func() {
			reqBody := map[string]string{"email": email}
			resp := server.PostJSON(t, "/api/v1/users", reqBody)

			if resp.StatusCode == http.StatusConflict || resp.StatusCode == http.StatusCreated {
				successes <- true
			} else {
				errors <- assert.AnError
			}
		}()
	}

	createdCount := 0
	conflictCount := 0

	for i := 0; i < 100; i++ {
		select {
		case <-successes:
			createdCount++
		case <-errors:
		case <-time.After(5 * time.Second):
			break
		}
	}

	assert.Equal(t, 1, createdCount, "Only one user should be created due to unique constraint")
	assert.Equal(t, 99, conflictCount, "99 requests should get conflict")
}

func TestTransactionEdgeCases_DateBoundaries(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	server := testutil.SetupTestServer(t)
	defer server.Cleanup(t)

	userReq := map[string]string{"email": "boundary-test@example.com"}
	resp := server.PostJSON(t, "/api/v1/users", userReq)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var user map[string]interface{}
	assert.NoError(t, json.NewDecoder(resp.Body).Decode(&user))
	userID := user["id"].(string)

	now := time.Now()

	t.Run("start boundary inclusive", func(t *testing.T) {
		startDate := now.Add(-24 * time.Hour)
		endDate := now

		_, err := server.DB.CreateTransaction(context.Background(), userID, nil, 10.0, nil, startDate)
		require.NoError(t, err)

		q := url.Values{}
		q.Set("user_id", userID)
		q.Set("from", startDate.Format(time.RFC3339))
		q.Set("to", endDate.Format(time.RFC3339))
		resp := server.Get(t, "/api/v1/transactions?"+q.Encode())

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var txns []interface{}
		err = json.NewDecoder(resp.Body).Decode(&txns)
		assert.NoError(t, err)
		assert.Len(t, txns, 1)
	})

	t.Run("end boundary inclusive", func(t *testing.T) {
		startDate := now.Add(-48 * time.Hour)
		endDate := now.Add(-24 * time.Hour)

		_, err := server.DB.CreateTransaction(context.Background(), userID, nil, 20.0, nil, endDate)
		require.NoError(t, err)

		q := url.Values{}
		q.Set("user_id", userID)
		q.Set("from", startDate.Format(time.RFC3339))
		q.Set("to", endDate.Format(time.RFC3339))
		resp := server.Get(t, "/api/v1/transactions?"+q.Encode())

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var txns []interface{}
		err = json.NewDecoder(resp.Body).Decode(&txns)
		assert.NoError(t, err)
		assert.Len(t, txns, 1)
	})
}

func TestTransactionEdgeCases_LargeAmount(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	server := testutil.SetupTestServer(t)
	defer server.Cleanup(t)

	userReq := map[string]string{"email": "large-amt@example.com"}
	resp := server.PostJSON(t, "/api/v1/users", userReq)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var user map[string]interface{}
	assert.NoError(t, json.NewDecoder(resp.Body).Decode(&user))
	userID := user["id"].(string)

	t.Run("maximum valid amount", func(t *testing.T) {
		amount := 99999999.99
		txnReq := map[string]interface{}{
			"user_id": userID,
			"amount":  amount,
		}
		resp := server.PostJSON(t, "/api/v1/transactions", txnReq)

		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	})

	t.Run("amount exceeding maximum", func(t *testing.T) {
		amount := 100000000.00
		txnReq := map[string]interface{}{
			"user_id": userID,
			"amount":  amount,
		}
		resp := server.PostJSON(t, "/api/v1/transactions", txnReq)

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

		var errResp map[string]interface{}
		err := json.NewDecoder(resp.Body).Decode(&errResp)
		assert.NoError(t, err)

		errObj := errResp["error"].(map[string]interface{})
		assert.Contains(t, errObj["message"], "exceeds maximum value")
	})
}

func TestSummaryEdgeCases_EmptyResults(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	server := testutil.SetupTestServer(t)
	defer server.Cleanup(t)

	userReq := map[string]string{"email": "empty-summary@example.com"}
	resp := server.PostJSON(t, "/api/v1/users", userReq)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var user map[string]interface{}
	assert.NoError(t, json.NewDecoder(resp.Body).Decode(&user))
	userID := user["id"].(string)

	t.Run("no transactions", func(t *testing.T) {
		q := url.Values{}
		q.Set("user_id", userID)
		resp := server.Get(t, "/api/v1/summary?"+q.Encode())

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var summary map[string]interface{}
		err := json.NewDecoder(resp.Body).Decode(&summary)
		assert.NoError(t, err)

		categories := summary["categories"].([]interface{})
		assert.Len(t, categories, 0)
	})

	t.Run("only uncategorized transactions", func(t *testing.T) {
		for i := 0; i < 3; i++ {
			_, err := server.DB.CreateTransaction(context.Background(), userID, nil, 10.0, nil, time.Now())
			assert.NoError(t, err)
		}

		q := url.Values{}
		q.Set("user_id", userID)
		resp := server.Get(t, "/api/v1/summary?"+q.Encode())

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var summary map[string]interface{}
		err := json.NewDecoder(resp.Body).Decode(&summary)
		assert.NoError(t, err)

		categories := summary["categories"].([]interface{})
		assert.Len(t, categories, 1)

		cat := categories[0].(map[string]interface{})
		assert.Equal(t, "Uncategorized", cat["category_name"])
		assert.Equal(t, 30.0, cat["total"])
	})
}

func TestErrorResponses_AllEndpoints(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	server := testutil.SetupTestServer(t)
	defer server.Cleanup(t)

	t.Run("malformed JSON", func(t *testing.T) {
		resp := server.Post(t, "/api/v1/users", []byte("invalid json"))
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

		var errResp map[string]interface{}
		err := json.NewDecoder(resp.Body).Decode(&errResp)
		assert.NoError(t, err)
		assert.Contains(t, errResp["error"].(map[string]interface{})["message"], "Invalid request body")
	})

	t.Run("invalid content type", func(t *testing.T) {
		resp := server.MakeRequestWithHeaders(t, http.MethodPost, "/api/v1/users", []byte("{}"), map[string]string{
			"Content-Type": "text/plain",
		})

		assert.Equal(t, http.StatusUnsupportedMediaType, resp.StatusCode)
	})
}

func TestLargeDataset_Performance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	server := testutil.SetupTestServer(t)
	defer server.Cleanup(t)

	userReq := map[string]string{"email": "perf-test@example.com"}
	resp := server.PostJSON(t, "/api/v1/users", userReq)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var user map[string]interface{}
	assert.NoError(t, json.NewDecoder(resp.Body).Decode(&user))
	userID := user["id"].(string)

	batchSize := 100
	startTime := time.Now()

	for i := 0; i < batchSize; i++ {
		_, err := server.DB.CreateTransaction(context.Background(), userID, nil, 10.0, nil, time.Now())
		assert.NoError(t, err)
	}

	createDuration := time.Since(startTime)
	t.Logf("Created %d transactions in %v", batchSize, createDuration)

	listStartTime := time.Now()
	q := url.Values{}
	q.Set("user_id", userID)
	resp = server.Get(t, "/api/v1/transactions?"+q.Encode())
	listDuration := time.Since(listStartTime)

	t.Logf("Listed %d transactions in %v", batchSize, listDuration)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Less(t, listDuration.Milliseconds(), int64(100), "Should list 100 transactions in < 100ms")
}
