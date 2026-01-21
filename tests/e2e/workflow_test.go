package e2e

import (
	"encoding/json"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"fintrack-go/tests/testutil"
)

func TestUserWorkflow_Lifecycle(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	server := testutil.SetupTestServer(t)
	defer server.Cleanup(t)

	t.Run("create and retrieve user", func(t *testing.T) {
		userReq := map[string]string{"email": "lifecycle@example.com"}
		resp := server.PostJSON(t, "/api/v1/users", userReq)

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var createdUser map[string]interface{}
		err := json.NewDecoder(resp.Body).Decode(&createdUser)
		require.NoError(t, err)

		userID := createdUser["id"].(string)
		assert.NotEmpty(t, userID)
		assert.Equal(t, "lifecycle@example.com", createdUser["email"])
		assert.NotEmpty(t, createdUser["created_at"])

		t.Run("verify email uniqueness", func(t *testing.T) {
			resp2 := server.PostJSON(t, "/api/v1/users", userReq)

			assert.Equal(t, http.StatusConflict, resp2.StatusCode)

			var errResp map[string]interface{}
			err := json.NewDecoder(resp2.Body).Decode(&errResp)
			assert.NoError(t, err)

			errObj := errResp["error"].(map[string]interface{})
			assert.Contains(t, errObj["message"], "already exists")
		})
	})
}

func TestCategoryWorkflow_CompleteLifecycle(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	server := testutil.SetupTestServer(t)
	defer server.Cleanup(t)

	t.Run("create user and manage categories", func(t *testing.T) {
		userReq := map[string]string{"email": "category-workflow@example.com"}
		resp := server.PostJSON(t, "/api/v1/users", userReq)
		require.Equal(t, http.StatusCreated, resp.StatusCode)

		var user map[string]interface{}
		assert.NoError(t, json.NewDecoder(resp.Body).Decode(&user))
		userID := user["id"].(string)

		t.Run("create multiple categories", func(t *testing.T) {
			categoryNames := []string{"Food", "Transport", "Entertainment", "Utilities"}
			createdIDs := make([]string, 0)

			for _, name := range categoryNames {
				catReq := map[string]interface{}{"user_id": userID, "name": name}
				resp := server.PostJSON(t, "/api/v1/categories", catReq)

				assert.Equal(t, http.StatusCreated, resp.StatusCode)

				var category map[string]interface{}
				err := json.NewDecoder(resp.Body).Decode(&category)
				assert.NoError(t, err)
				createdIDs = append(createdIDs, category["id"].(string))
			}

			t.Run("list all categories", func(t *testing.T) {
				q := url.Values{}
				q.Set("user_id", userID)
				resp := server.Get(t, "/api/v1/categories?"+q.Encode())

				assert.Equal(t, http.StatusOK, resp.StatusCode)

				var categories []map[string]interface{}
				err := json.NewDecoder(resp.Body).Decode(&categories)
				assert.NoError(t, err)
				assert.Len(t, categories, 4)

				for _, cat := range categories {
					assert.Equal(t, userID, cat["user_id"])
					assert.Contains(t, categoryNames, cat["name"])
				}
			})

			t.Run("prevent duplicate category names", func(t *testing.T) {
				dupReq := map[string]interface{}{"user_id": userID, "name": "Food"}
				resp := server.PostJSON(t, "/api/v1/categories", dupReq)

				assert.Equal(t, http.StatusConflict, resp.StatusCode)

				var errResp map[string]interface{}
				err := json.NewDecoder(resp.Body).Decode(&errResp)
				assert.NoError(t, err)

				errObj := errResp["error"].(map[string]interface{})
				assert.Contains(t, errObj["message"], "already exists")
			})

			t.Run("delete all categories", func(t *testing.T) {
				testutil.TeardownTestDB(t, server.DB.GetPool())

				q := url.Values{}
				q.Set("user_id", userID)
				resp := server.Get(t, "/api/v1/categories?"+q.Encode())

				assert.Equal(t, http.StatusOK, resp.StatusCode)

				var categories []interface{}
				err := json.NewDecoder(resp.Body).Decode(&categories)
				assert.NoError(t, err)
				assert.Empty(t, categories, "All categories should be deleted")
			})
		})
	})
}

func TestTransactionWorkflow_CompleteTracking(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	server := testutil.SetupTestServer(t)
	defer server.Cleanup(t)

	t.Run("create user, categories, and manage transactions", func(t *testing.T) {
		userReq := map[string]string{"email": "txn-workflow@example.com"}
		resp := server.PostJSON(t, "/api/v1/users", userReq)
		require.Equal(t, http.StatusCreated, resp.StatusCode)

		var user map[string]interface{}
		assert.NoError(t, json.NewDecoder(resp.Body).Decode(&user))
		userID := user["id"].(string)

		catReq := map[string]interface{}{"user_id": userID, "name": "Expenses"}
		resp = server.PostJSON(t, "/api/v1/categories", catReq)
		require.Equal(t, http.StatusCreated, resp.StatusCode)

		var category map[string]interface{}
		assert.NoError(t, json.NewDecoder(resp.Body).Decode(&category))
		categoryID := category["id"].(string)

		t.Run("create categorized transaction", func(t *testing.T) {
			now := time.Now()
			txnReq := map[string]interface{}{
				"user_id":     userID,
				"category_id": categoryID,
				"amount":      50.25,
				"description": "Grocery shopping",
				"occurred_at": now.Format(time.RFC3339),
			}
			resp := server.PostJSON(t, "/api/v1/transactions", txnReq)

			assert.Equal(t, http.StatusCreated, resp.StatusCode)

			var transaction map[string]interface{}
			err := json.NewDecoder(resp.Body).Decode(&transaction)
			assert.NoError(t, err)
			assert.Equal(t, 50.25, transaction["amount"])
			assert.Equal(t, "Grocery shopping", transaction["description"])
		})

		t.Run("create uncategorized transaction", func(t *testing.T) {
			txnReq := map[string]interface{}{
				"user_id": userID,
				"amount":  25.00,
				"description": "Uncategorized expense",
			}
			resp := server.PostJSON(t, "/api/v1/transactions", txnReq)

			assert.Equal(t, http.StatusCreated, resp.StatusCode)

			var transaction map[string]interface{}
			err := json.NewDecoder(resp.Body).Decode(&transaction)
			assert.NoError(t, err)
			assert.Nil(t, transaction["category_id"])
		})

		t.Run("list all transactions", func(t *testing.T) {
			q := url.Values{}
			q.Set("user_id", userID)
			resp := server.Get(t, "/api/v1/transactions?"+q.Encode())

			assert.Equal(t, http.StatusOK, resp.StatusCode)

			var transactions []map[string]interface{}
			err := json.NewDecoder(resp.Body).Decode(&transactions)
			assert.NoError(t, err)
			assert.Len(t, transactions, 2)

			total := 0.0
			for _, txn := range transactions {
				total += txn["amount"].(float64)
			}
			assert.Equal(t, 75.25, total)
		})

		t.Run("list transactions by date range", func(t *testing.T) {
			now := time.Now()
			startDate := now.Add(-48 * time.Hour)
			endDate := now.Add(-24 * time.Hour)

			txnReq := map[string]interface{}{
				"user_id":     userID,
				"category_id": categoryID,
				"amount":      100.00,
				"description": "Past transaction",
			}
			resp := server.PostJSON(t, "/api/v1/transactions", txnReq)
			require.Equal(t, http.StatusCreated, resp.StatusCode)

			q := url.Values{}
			q.Set("user_id", userID)
			q.Set("from", startDate.Format(time.RFC3339))
			q.Set("to", endDate.Format(time.RFC3339))
			resp = server.Get(t, "/api/v1/transactions?"+q.Encode())

			assert.Equal(t, http.StatusOK, resp.StatusCode)

			var transactions []map[string]interface{}
			err := json.NewDecoder(resp.Body).Decode(&transactions)
			assert.NoError(t, err)
			assert.Len(t, transactions, 1)
			assert.Equal(t, 100.00, transactions[0]["amount"])
		})
	})
}

func TestSummaryWorkflow_Analytics(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	server := testutil.SetupTestServer(t)
	defer server.Cleanup(t)

	userReq := map[string]string{"email": "summary-workflow@example.com"}
	resp := server.PostJSON(t, "/api/v1/users", userReq)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var user map[string]interface{}
	assert.NoError(t, json.NewDecoder(resp.Body).Decode(&user))
	userID := user["id"].(string)

	categoryReqs := []map[string]interface{}{
		{"user_id": userID, "name": "Food"},
		{"user_id": userID, "name": "Transport"},
	}
	for _, catReq := range categoryReqs {
		resp := server.PostJSON(t, "/api/v1/categories", catReq)
		require.Equal(t, http.StatusCreated, resp.StatusCode)
	}

	var categories []map[string]interface{}
	q := url.Values{}
	q.Set("user_id", userID)
	resp = server.Get(t, "/api/v1/categories?"+q.Encode())
	assert.NoError(t, json.NewDecoder(resp.Body).Decode(&categories))
	require.Len(t, categories, 2)

	now := time.Now()
	foodID := categories[0]["id"].(string)
	transportID := categories[1]["id"].(string)

	transactions := []map[string]interface{}{
		{"user_id": userID, "category_id": foodID, "amount": 30.0, "description": "Groceries", "occurred_at": now.Add(-48*time.Hour).Format(time.RFC3339)},
		{"user_id": userID, "category_id": foodID, "amount": 45.50, "description": "Restaurant", "occurred_at": now.Add(-36*time.Hour).Format(time.RFC3339)},
		{"user_id": userID, "category_id": transportID, "amount": 20.0, "description": "Gas", "occurred_at": now.Add(-24*time.Hour).Format(time.RFC3339)},
		{"user_id": userID, "amount": 15.0, "description": "Cash withdrawal", "occurred_at": now.Add(-12*time.Hour).Format(time.RFC3339)},
	}

	for _, txnReq := range transactions {
		resp := server.PostJSON(t, "/api/v1/transactions", txnReq)
		require.Equal(t, http.StatusCreated, resp.StatusCode)
	}

	t.Run("get overall summary", func(t *testing.T) {
		resp := server.Get(t, "/api/v1/summary?user_id="+userID)

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var summary map[string]interface{}
		err := json.NewDecoder(resp.Body).Decode(&summary)
		assert.NoError(t, err)

		categories := summary["categories"].([]interface{})
		assert.Len(t, categories, 3)

		categoryTotals := make(map[string]float64)
		for _, cat := range categories {
			catMap := cat.(map[string]interface{})
			categoryTotals[catMap["category_name"].(string)] = catMap["total"].(float64)
		}

		assert.Equal(t, 75.5, categoryTotals["Food"])
		assert.Equal(t, 20.0, categoryTotals["Transport"])
		assert.Equal(t, 15.0, categoryTotals["Uncategorized"])
	})

	t.Run("get summary with date range", func(t *testing.T) {
		startDate := now.Add(-48 * time.Hour)
		endDate := now.Add(-24 * time.Hour)

		q := url.Values{}
		q.Set("user_id", userID)
		q.Set("from", startDate.Format(time.RFC3339))
		q.Set("to", endDate.Format(time.RFC3339))
		resp := server.Get(t, "/api/v1/summary?"+q.Encode())

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var summary map[string]interface{}
		err := json.NewDecoder(resp.Body).Decode(&summary)
		assert.NoError(t, err)

		categories := summary["categories"].([]interface{})
		assert.Len(t, categories, 1)
		assert.Equal(t, 20.0, categories[0].(map[string]interface{})["total"])
	})
}

func TestMultiUserWorkflow_DataIsolation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	server := testutil.SetupTestServer(t)
	defer server.Cleanup(t)

	user1Req := map[string]string{"email": "user1@example.com"}
	resp1 := server.PostJSON(t, "/api/v1/users", user1Req)
	require.Equal(t, http.StatusCreated, resp1.StatusCode)

	var user1 map[string]interface{}
	assert.NoError(t, json.NewDecoder(resp1.Body).Decode(&user1))
	user1ID := user1["id"].(string)

	user2Req := map[string]string{"email": "user2@example.com"}
	resp2 := server.PostJSON(t, "/api/v1/users", user2Req)
	require.Equal(t, http.StatusCreated, resp2.StatusCode)

	var user2 map[string]interface{}
	assert.NoError(t, json.NewDecoder(resp2.Body).Decode(&user2))
	user2ID := user2["id"].(string)

	cat1Req := map[string]interface{}{"user_id": user1ID, "name": "User1 Category"}
	resp1 = server.PostJSON(t, "/api/v1/categories", cat1Req)
	require.Equal(t, http.StatusCreated, resp1.StatusCode)

	var cat1 map[string]interface{}
	assert.NoError(t, json.NewDecoder(resp1.Body).Decode(&cat1))
	cat1ID := cat1["id"].(string)

	cat2Req := map[string]interface{}{"user_id": user2ID, "name": "User2 Category"}
	resp2 = server.PostJSON(t, "/api/v1/categories", cat2Req)
	require.Equal(t, http.StatusCreated, resp2.StatusCode)

	var cat2 map[string]interface{}
	assert.NoError(t, json.NewDecoder(resp2.Body).Decode(&cat2))
	cat2ID := cat2["id"].(string)

	txn1Req := map[string]interface{}{"user_id": user1ID, "category_id": cat1ID, "amount": 100.0}
	resp1 = server.PostJSON(t, "/api/v1/transactions", txn1Req)
	require.Equal(t, http.StatusCreated, resp1.StatusCode)

	txn2Req := map[string]interface{}{"user_id": user2ID, "category_id": cat2ID, "amount": 50.0}
	resp2 = server.PostJSON(t, "/api/v1/transactions", txn2Req)
	require.Equal(t, http.StatusCreated, resp2.StatusCode)

	t.Run("user1 cannot see user2's data", func(t *testing.T) {
		q := url.Values{}
		q.Set("user_id", user1ID)
		resp := server.Get(t, "/api/v1/transactions?"+q.Encode())

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var transactions []map[string]interface{}
		err := json.NewDecoder(resp.Body).Decode(&transactions)
		assert.NoError(t, err)
		assert.Len(t, transactions, 1)
		assert.Equal(t, 100.0, transactions[0]["amount"])
	})

	t.Run("user2 cannot see user1's data", func(t *testing.T) {
		q := url.Values{}
		q.Set("user_id", user2ID)
		resp := server.Get(t, "/api/v1/transactions?"+q.Encode())

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var transactions []map[string]interface{}
		err := json.NewDecoder(resp.Body).Decode(&transactions)
		assert.NoError(t, err)
		assert.Len(t, transactions, 1)
		assert.Equal(t, 50.0, transactions[0]["amount"])
	})

	t.Run("summaries are isolated", func(t *testing.T) {
		q1 := url.Values{}
		q1.Set("user_id", user1ID)
		resp1 := server.Get(t, "/api/v1/summary?"+q1.Encode())
		var summary1 map[string]interface{}
		assert.NoError(t, json.NewDecoder(resp1.Body).Decode(&summary1))

		q2 := url.Values{}
		q2.Set("user_id", user2ID)
		resp2 := server.Get(t, "/api/v1/summary?"+q2.Encode())
		var summary2 map[string]interface{}
		assert.NoError(t, json.NewDecoder(resp2.Body).Decode(&summary2))

		categories1 := summary1["categories"].([]interface{})
		total1 := 0.0
		for _, cat := range categories1 {
			catMap := cat.(map[string]interface{})
			total1 += catMap["total"].(float64)
		}
		assert.Equal(t, 100.0, total1)

		categories2 := summary2["categories"].([]interface{})
		total2 := 0.0
		for _, cat := range categories2 {
			catMap := cat.(map[string]interface{})
			total2 += catMap["total"].(float64)
		}
		assert.Equal(t, 50.0, total2)
	})

	t.Run("user1 cannot create transaction with user2's category", func(t *testing.T) {
		txnReq := map[string]interface{}{
			"user_id":     user1ID,
			"category_id": cat2ID,
			"amount":      10.0,
		}
		resp := server.PostJSON(t, "/api/v1/transactions", txnReq)

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

		var errResp map[string]interface{}
		err := json.NewDecoder(resp.Body).Decode(&errResp)
		assert.NoError(t, err)

		errObj := errResp["error"].(map[string]interface{})
		assert.Contains(t, errObj["message"], "does not belong to user")
	})
}
