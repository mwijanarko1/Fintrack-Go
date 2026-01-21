package security

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"fintrack-go/tests/testutil"
)

func TestSQLInjection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping security test in short mode")
	}

	server := testutil.SetupTestServer(t)
	defer server.Cleanup(t)

	userReq := map[string]string{"email": "sql-inj@example.com"}
	resp := server.PostJSON(t, "/api/v1/users", userReq)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var user map[string]interface{}
	assert.NoError(t, json.NewDecoder(resp.Body).Decode(&user))
	userID := user["id"].(string)

	injectionAttempts := []string{
		"'; DROP TABLE users; --",
		"' OR '1'='1",
		"admin'--",
		"' UNION SELECT * FROM users--",
		"1'; DELETE FROM users WHERE '1'='1",
	}

	for _, injection := range injectionAttempts {
		t.Run("email injection: "+injection, func(t *testing.T) {
			reqBody := map[string]string{"email": injection}
			resp := server.PostJSON(t, "/api/v1/users", reqBody)

			assert.Equal(t, http.StatusBadRequest, resp.StatusCode, "API should reject SQL injection attempts")
			assertJSONContentType(t, resp)

			var errResp map[string]interface{}
			err := json.NewDecoder(resp.Body).Decode(&errResp)
			assert.NoError(t, err)

			errObj := errResp["error"].(map[string]interface{})
			assert.NotContains(t, errObj["message"].(string), "DROP")
			assert.NotContains(t, errObj["message"].(string), "DELETE")
			assert.NotContains(t, errObj["message"].(string), "UNION")

			testutil.AssertRowCount(t, server.DB.GetPool(), 0, "SELECT COUNT(*) FROM users WHERE email = $1", injection)
		})
	}

	t.Run("category name injection", func(t *testing.T) {
		injection := "'; DROP TABLE categories; --"
		reqBody := map[string]interface{}{
			"user_id": userID,
			"name":    injection,
		}
		resp := server.PostJSON(t, "/api/v1/categories", reqBody)

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode, "API should reject SQL injection attempts in category name")
		assertJSONContentType(t, resp)

		var errResp map[string]interface{}
		err := json.NewDecoder(resp.Body).Decode(&errResp)
		assert.NoError(t, err)
		assert.NotContains(t, errResp["error"].(map[string]interface{})["message"].(string), "DROP")

		testutil.AssertRowCount(t, server.DB.GetPool(), 0, "SELECT COUNT(*) FROM categories WHERE name = $1", injection)
	})

	t.Run("transaction description injection", func(t *testing.T) {
		injection := "'; DELETE FROM transactions; --"
		reqBody := map[string]interface{}{
			"user_id":     userID,
			"amount":      10.0,
			"description": injection,
		}
		resp := server.PostJSON(t, "/api/v1/transactions", reqBody)

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode, "API should reject SQL injection attempts in transaction description")
		assertJSONContentType(t, resp)

		var errResp map[string]interface{}
		err := json.NewDecoder(resp.Body).Decode(&errResp)
		assert.NoError(t, err)
		assert.NotContains(t, errResp["error"].(map[string]interface{})["message"].(string), "DELETE")

		testutil.AssertRowCount(t, server.DB.GetPool(), 0, "SELECT COUNT(*) FROM transactions WHERE description = $1", injection)
	})

	t.Run("user_id parameter injection", func(t *testing.T) {
		maliciousUserID := "550e8400' OR '1'='1"
		q := url.Values{}
		q.Set("user_id", maliciousUserID)
		resp := server.Get(t, "/api/v1/categories?"+q.Encode())

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

		var errResp map[string]interface{}
		err := json.NewDecoder(resp.Body).Decode(&errResp)
		assert.NoError(t, err)
		assert.Contains(t, errResp["error"].(map[string]interface{})["message"].(string), "invalid UUID format")
	})
}

func TestXSSPrevention(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping security test in short mode")
	}

	server := testutil.SetupTestServer(t)
	defer server.Cleanup(t)

	xssPayloads := []string{
		"<script>alert('XSS')</script>",
		"<img src=x onerror=alert('XSS')>",
		"<svg onload=alert('XSS')>",
		"javascript:alert('XSS')",
		"<script>document.location='http://evil.com'</script>",
	}

	for _, payload := range xssPayloads {
		t.Run("XSS in email: "+payload, func(t *testing.T) {
			reqBody := map[string]string{"email": payload}
			resp := server.PostJSON(t, "/api/v1/users", reqBody)

			assert.Equal(t, http.StatusBadRequest, resp.StatusCode, "API should reject XSS payloads in email")

			var errResp map[string]interface{}
			err := json.NewDecoder(resp.Body).Decode(&errResp)
			assert.NoError(t, err)

			errMsg := errResp["error"].(map[string]interface{})["message"].(string)
			assert.NotContains(t, errMsg, "<script>")
			assert.NotContains(t, errMsg, "<img")
			assert.NotContains(t, errMsg, "javascript:")
		})

		t.Run("XSS in category name: "+payload, func(t *testing.T) {
			userReq := map[string]string{"email": "xss-test@example.com"}
			resp := server.PostJSON(t, "/api/v1/users", userReq)
			require.Equal(t, http.StatusCreated, resp.StatusCode)

			var user map[string]interface{}
			assert.NoError(t, json.NewDecoder(resp.Body).Decode(&user))
			userID := user["id"].(string)

			catReq := map[string]interface{}{"user_id": userID, "name": payload}
			resp = server.PostJSON(t, "/api/v1/categories", catReq)

			assert.Equal(t, http.StatusBadRequest, resp.StatusCode, "API should reject XSS payloads in category name")

			var errResp map[string]interface{}
			err := json.NewDecoder(resp.Body).Decode(&errResp)
			assert.NoError(t, err)

			errMsg := errResp["error"].(map[string]interface{})["message"].(string)
			assert.NotContains(t, errMsg, "<script>")
		})

		t.Run("XSS in transaction description: "+payload, func(t *testing.T) {
			userReq := map[string]string{"email": "xss-desc@example.com"}
			resp := server.PostJSON(t, "/api/v1/users", userReq)
			require.Equal(t, http.StatusCreated, resp.StatusCode)

			var user map[string]interface{}
			assert.NoError(t, json.NewDecoder(resp.Body).Decode(&user))
			userID := user["id"].(string)

			txnReq := map[string]interface{}{
				"user_id":     userID,
				"amount":      10.0,
				"description": payload,
			}
			resp = server.PostJSON(t, "/api/v1/transactions", txnReq)

			assert.Equal(t, http.StatusBadRequest, resp.StatusCode, "API should reject XSS payloads in transaction description")

			var errResp map[string]interface{}
			err := json.NewDecoder(resp.Body).Decode(&errResp)
			assert.NoError(t, err)

			errMsg := errResp["error"].(map[string]interface{})["message"].(string)
			assert.NotContains(t, errMsg, "<script>")
		})
	}
}

func TestAuthorization_UserIsolation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping security test in short mode")
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

	catReq := map[string]interface{}{"user_id": user1ID, "name": "Private Category"}
	resp := server.PostJSON(t, "/api/v1/categories", catReq)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var cat map[string]interface{}
	assert.NoError(t, json.NewDecoder(resp.Body).Decode(&cat))
	catID := cat["id"].(string)

	t.Run("user2 cannot create transaction with user1's category", func(t *testing.T) {
		txnReq := map[string]interface{}{
			"user_id":     user2ID,
			"category_id": catID,
			"amount":      10.0,
		}
		resp := server.PostJSON(t, "/api/v1/transactions", txnReq)

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

		var errResp map[string]interface{}
		err := json.NewDecoder(resp.Body).Decode(&errResp)
		assert.NoError(t, err)

		errMsg := errResp["error"].(map[string]interface{})["message"].(string)
		assert.Contains(t, errMsg, "does not belong to user")
	})

	t.Run("user2 cannot list user1's categories", func(t *testing.T) {
		q := url.Values{}
		q.Set("user_id", user1ID)
		resp := server.Get(t, "/api/v1/categories?"+q.Encode())

		categories := make([]interface{}, 0)
		err := json.NewDecoder(resp.Body).Decode(&categories)
		assert.NoError(t, err)

		catNames := make([]string, 0)
		for _, cat := range categories {
			if catMap, ok := cat.(map[string]interface{}); ok {
				catNames = append(catNames, catMap["name"].(string))
			}
		}

		for _, name := range catNames {
			assert.NotEqual(t, "Private Category", name)
		}
	})

	t.Run("user2 cannot see user1's transactions", func(t *testing.T) {
		q := url.Values{}
		q.Set("user_id", user1ID)
		resp := server.Get(t, "/api/v1/transactions?"+q.Encode())

		txns := make([]interface{}, 0)
		err := json.NewDecoder(resp.Body).Decode(&txns)
		assert.NoError(t, err)

		assert.Empty(t, txns, "User2 should not see User1's transactions")
	})

	t.Run("user2's summary only shows their data", func(t *testing.T) {
		txn1Req := map[string]interface{}{"user_id": user2ID, "amount": 10.0}
		resp := server.PostJSON(t, "/api/v1/transactions", txn1Req)
		require.Equal(t, http.StatusCreated, resp.StatusCode)

		q1 := url.Values{}
		q1.Set("user_id", user1ID)
		resp1 := server.Get(t, "/api/v1/summary?"+q1.Encode())

		var summary1 map[string]interface{}
		err := json.NewDecoder(resp1.Body).Decode(&summary1)
		assert.NoError(t, err)

		total1 := 0.0
		categories1 := summary1["categories"].([]interface{})
		for _, cat := range categories1 {
			if catMap, ok := cat.(map[string]interface{}); ok {
				total1 += catMap["total"].(float64)
			}
		}
		assert.Equal(t, 0.0, total1, "User1's summary should be 0.0")

		q2 := url.Values{}
		q2.Set("user_id", user2ID)
		resp2 := server.Get(t, "/api/v1/summary?"+q2.Encode())

		var summary2 map[string]interface{}
		err = json.NewDecoder(resp2.Body).Decode(&summary2)
		assert.NoError(t, err)

		total2 := 0.0
		categories2 := summary2["categories"].([]interface{})
		for _, cat := range categories2 {
			if catMap, ok := cat.(map[string]interface{}); ok {
				total2 += catMap["total"].(float64)
			}
		}
		assert.Equal(t, 10.0, total2, "User2's summary should be 10.0")
	})
}

func TestUUIDEnumerationPrevention(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping security test in short mode")
	}

	server := testutil.SetupTestServer(t)
	defer server.Cleanup(t)

	t.Run("sequential UUID enumeration", func(t *testing.T) {
		foundUsers := 0
		for i := 0; i < 100; i++ {
			uuid := fmt.Sprintf("550e8400-e29b-41d4-a716-44665544%04d", i)

			q := url.Values{}
			q.Set("user_id", uuid)
			resp := server.Get(t, "/api/v1/categories?"+q.Encode())

			if resp.StatusCode == http.StatusOK {
				categories := make([]interface{}, 0)
				err := json.NewDecoder(resp.Body).Decode(&categories)
				if err == nil && len(categories) > 0 {
					foundUsers++
				}
			}
		}

		assert.Less(t, foundUsers, 10, "Should not find many users via sequential UUID enumeration")
	})

	t.Run("predictable UUID access", func(t *testing.T) {
		predictableUUIDs := []string{
			"00000000-0000-0000-0000-000000000000",
			"ffffffff-ffff-ffff-ffff-ffffffffffff",
			"550e8400-e29b-41d4-a716-4466554400000",
		}

		for _, uuid := range predictableUUIDs {
			t.Run("predictable UUID: "+uuid, func(t *testing.T) {
				q := url.Values{}
				q.Set("user_id", uuid)
				resp := server.Get(t, "/api/v1/categories?"+q.Encode())

				if resp.StatusCode == http.StatusOK {
					categories := make([]interface{}, 0)
					err := json.NewDecoder(resp.Body).Decode(&categories)
					assert.NoError(t, err)
					t.Logf("Found %d categories for UUID %s", len(categories), uuid)
				}
			})
		}
	})
}

func assertJSONContentType(t *testing.T, resp *http.Response) {
	t.Helper()
	contentType := resp.Header.Get("Content-Type")
	assert.Equal(t, "application/json", contentType, "expected JSON content type")
}
