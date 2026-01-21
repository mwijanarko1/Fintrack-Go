package load

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"fintrack-go/tests/testutil"
)

func TestLoad_ConcurrentRequests(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping load test in short mode")
	}

	server := testutil.SetupTestServer(t)
	defer server.Cleanup(t)

	concurrentRequests := 200
	errors := make(chan error, concurrentRequests)
	startTime := make(chan struct{}, concurrentRequests)
	successes := 0

	for i := 0; i < concurrentRequests; i++ {
		go func(requestNum int) {
			startTime <- struct{}{}

			userReq := map[string]string{
				"email": fmt.Sprintf("load%d@example.com", requestNum),
			}
			resp := server.PostJSON(t, "/api/v1/users", userReq)

			if resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusConflict {
				<-startTime
			} else {
				<-startTime
				errors <- fmt.Errorf("request %d failed with status %d", requestNum, resp.StatusCode)
			}
		}(i)
	}

	completed := 0
	timeout := time.After(60 * time.Second)

	for completed < concurrentRequests {
		select {
		case err := <-errors:
			t.Error(err)
			completed++
		case <-timeout:
			t.Errorf("Test timed out after 60 seconds. Completed: %d/%d", completed, concurrentRequests)
			return
		default:
			time.Sleep(10 * time.Millisecond)
		}
	}

	successes = concurrentRequests - completed
	successRate := float64(successes) / float64(concurrentRequests) * 100
	t.Logf("Load test completed: %d/%d requests succeeded (%.1f%%)",
		successes, concurrentRequests, successRate)

	assert.Greater(t, successes, 0, "At least some requests should succeed")
	assert.GreaterOrEqual(t, int(successRate), 90, "Success rate should be at least 90%%")
}

func TestLoad_StressTransactionCreation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping load test in short mode")
	}

	server := testutil.SetupTestServer(t)
	defer server.Cleanup(t)

	userReq := map[string]string{"email": "stress@example.com"}
	resp := server.PostJSON(t, "/api/v1/users", userReq)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var user map[string]interface{}
	assert.NoError(t, json.NewDecoder(resp.Body).Decode(&user))
	userID := user["id"].(string)

	categoryReq := map[string]interface{}{"user_id": userID, "name": "Stress Category"}
	resp = server.PostJSON(t, "/api/v1/categories", categoryReq)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var category map[string]interface{}
	assert.NoError(t, json.NewDecoder(resp.Body).Decode(&category))
	categoryID := category["id"].(string)

	concurrentTransactions := 250
	errors := make(chan error, concurrentTransactions)

	for i := 0; i < concurrentTransactions; i++ {
		go func(txNum int) {
			txnReq := map[string]interface{}{
				"user_id":     userID,
				"category_id": categoryID,
				"amount":      float64(txNum + 1),
			}
			resp := server.PostJSON(t, "/api/v1/transactions", txnReq)

			if resp.StatusCode == http.StatusCreated {
			} else {
				errors <- fmt.Errorf("transaction %d failed with status %d", txNum, resp.StatusCode)
			}
		}(i)
	}

	completed := 0
	errorsCount := 0
	start := time.Now()
	timeout := time.After(90 * time.Second)

	for completed < concurrentTransactions {
		select {
		case <-errors:
			errorsCount++
			completed++
		case <-timeout:
			duration := time.Since(start)
			t.Errorf("Stress test timed out after 90 seconds. Completed: %d/%d in %v",
				completed, concurrentTransactions, duration)
			return
		default:
			if completed+errorsCount >= concurrentTransactions {
				break
			}
			time.Sleep(10 * time.Millisecond)
		}
	}

	completed = concurrentTransactions - errorsCount
	duration := time.Since(start)
	successRate := float64(completed) / float64(concurrentTransactions) * 100
	t.Logf("Stress test completed: %d transactions in %v (%.2f txn/sec, %.1f%% success)",
		completed, duration, float64(completed)/duration.Seconds(), successRate)

	assert.Greater(t, completed, 0, "At least some transactions should be created")
	assert.Less(t, duration.Milliseconds(), int64(90000), "Should complete within 90 seconds")
}

func TestLoad_BurstRequests(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping load test in short mode")
	}

	server := testutil.SetupTestServer(t)
	defer server.Cleanup(t)

	burstSize := 50
	startCh := make(chan struct{}, burstSize)

	for i := 0; i < burstSize; i++ {
		go func() {
			startCh <- struct{}{}
			resp := server.Get(t, "/health")
			<-startCh
			if resp.StatusCode != http.StatusOK {
				t.Errorf("Health check failed with status %d", resp.StatusCode)
			}
		}()
	}

	time.Sleep(100 * time.Millisecond)
	close(startCh)
	time.Sleep(1 * time.Second)

	startTime := time.Now()
	for i := 0; i < burstSize; i++ {
		go func() {
			resp := server.Get(t, "/health")
			if resp.StatusCode != http.StatusOK {
				t.Errorf("Health check failed in burst with status %d", resp.StatusCode)
			}
		}()
	}

	time.Sleep(2 * time.Second)
	duration := time.Since(startTime)

	t.Logf("Burst test completed: %d requests in %v (%.2f req/sec)",
		burstSize, duration, float64(burstSize)/duration.Seconds())

	assert.Less(t, duration.Milliseconds(), int64(5000), "Burst should complete within 5 seconds")
}

func TestLoad_RepeatedSummaryGeneration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping load test in short mode")
	}

	server := testutil.SetupTestServer(t)
	defer server.Cleanup(t)

	userReq := map[string]string{"email": "summary-load@example.com"}
	resp := server.PostJSON(t, "/api/v1/users", userReq)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var user map[string]interface{}
	assert.NoError(t, json.NewDecoder(resp.Body).Decode(&user))
	userID := user["id"].(string)

	numTransactions := 100
	for i := 0; i < numTransactions; i++ {
		txnReq := map[string]interface{}{
			"user_id": userID,
			"amount":  10.0,
		}
		resp := server.PostJSON(t, "/api/v1/transactions", txnReq)
		if resp.StatusCode != http.StatusCreated {
			t.Logf("Failed to create transaction %d/%d: %d", i+1, numTransactions, resp.StatusCode)
		}
	}

	summaryRequests := 200
	startTime := time.Now()
	errors := 0

	for i := 0; i < summaryRequests; i++ {
		resp := server.Get(t, fmt.Sprintf("/api/v1/summary?user_id=%s", userID))
		if resp.StatusCode != http.StatusOK {
			errors++
		}
	}

	duration := time.Since(startTime)
	avgLatency := duration.Milliseconds() / int64(summaryRequests)

	t.Logf("Summary load test: %d requests in %v (avg latency: %dms, errors: %d)",
		summaryRequests, duration, avgLatency, errors)

	assert.Less(t, avgLatency, int64(500), "Average latency should be less than 500ms")
	assert.Less(t, errors, summaryRequests/10, "Error rate should be less than 10%%")
}
