package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"fintrack-go/internal/db"
	apphttp "fintrack-go/internal/http"
)

func getTestDatabaseURL() string {
	if url := os.Getenv("TEST_DATABASE_URL"); url != "" {
		return url
	}
	return "postgres://fintrack:fintrack@localhost:5432/fintrack_test"
}

func setupTestDB(t *testing.T) *db.DB {
	databaseURL := getTestDatabaseURL()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	testDB, err := db.NewDB(ctx, databaseURL, zerolog.Nop())
	require.NoError(t, err)

	return testDB
}

func cleanupTestDB(t *testing.T, database *db.DB) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := database.GetPool().Exec(ctx, "TRUNCATE TABLE transactions CASCADE")
	require.NoError(t, err)

	_, err = database.GetPool().Exec(ctx, "TRUNCATE TABLE categories CASCADE")
	require.NoError(t, err)

	_, err = database.GetPool().Exec(ctx, "TRUNCATE TABLE users CASCADE")
	require.NoError(t, err)
}

func setupRouter(logger zerolog.Logger, database *db.DB) http.Handler {
	r := apphttp.SetupRoutes(logger, database)
	return r
}

func TestCreateUser(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	database := setupTestDB(t)
	defer cleanupTestDB(t, database)
	defer database.Close()

	logger := zerolog.New(zerolog.NewTestWriter(t))
	router := setupRouter(logger, database)

	reqBody := map[string]string{"email": "integration-test@example.com"}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "integration-test@example.com", resp["email"])
	assert.NotEmpty(t, resp["id"])
}

func TestCreateUserDuplicateEmail(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	database := setupTestDB(t)
	defer cleanupTestDB(t, database)
	defer database.Close()

	logger := zerolog.New(zerolog.NewTestWriter(t))
	router := setupRouter(logger, database)

	email := "duplicate-test@example.com"

	reqBody := map[string]string{"email": email}
	body, _ := json.Marshal(reqBody)

	req1 := httptest.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewBuffer(body))
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusCreated, w1.Code)

	req2 := httptest.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewBuffer(body))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusConflict, w2.Code)
}

func TestCategoryFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	database := setupTestDB(t)
	defer cleanupTestDB(t, database)
	defer database.Close()

	logger := zerolog.New(zerolog.NewTestWriter(t))
	router := setupRouter(logger, database)

	userReqBody := map[string]string{"email": "category-test@example.com"}
	userBody, _ := json.Marshal(userReqBody)
	userReq := httptest.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewBuffer(userBody))
	userReq.Header.Set("Content-Type", "application/json")
	userW := httptest.NewRecorder()
	router.ServeHTTP(userW, userReq)

	require.Equal(t, http.StatusCreated, userW.Code)
	var userResp map[string]interface{}
	json.Unmarshal(userW.Body.Bytes(), &userResp)
	userID := userResp["id"].(string)

	catReqBody := map[string]interface{}{
		"user_id": userID,
		"name":    "Food",
	}
	catBody, _ := json.Marshal(catReqBody)
	catReq := httptest.NewRequest(http.MethodPost, "/api/v1/categories", bytes.NewBuffer(catBody))
	catReq.Header.Set("Content-Type", "application/json")
	catW := httptest.NewRecorder()
	router.ServeHTTP(catW, catReq)

	assert.Equal(t, http.StatusCreated, catW.Code)

	var catResp map[string]interface{}
	json.Unmarshal(catW.Body.Bytes(), &catResp)
	assert.Equal(t, "Food", catResp["name"])
	assert.Equal(t, userID, catResp["user_id"])

	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/categories?user_id="+userID, nil)
	listW := httptest.NewRecorder()
	router.ServeHTTP(listW, listReq)

	assert.Equal(t, http.StatusOK, listW.Code)

	var listResp []map[string]interface{}
	json.Unmarshal(listW.Body.Bytes(), &listResp)
	assert.Equal(t, 1, len(listResp))
	assert.Equal(t, "Food", listResp[0]["name"])
}

func TestTransactionFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	database := setupTestDB(t)
	defer cleanupTestDB(t, database)
	defer database.Close()

	logger := zerolog.New(zerolog.NewTestWriter(t))
	router := setupRouter(logger, database)

	userReqBody := map[string]string{"email": "transaction-test@example.com"}
	userBody, _ := json.Marshal(userReqBody)
	userReq := httptest.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewBuffer(userBody))
	userReq.Header.Set("Content-Type", "application/json")
	userW := httptest.NewRecorder()
	router.ServeHTTP(userW, userReq)

	var userResp map[string]interface{}
	json.Unmarshal(userW.Body.Bytes(), &userResp)
	userID := userResp["id"].(string)

	catReqBody := map[string]interface{}{
		"user_id": userID,
		"name":    "Food",
	}
	catBody, _ := json.Marshal(catReqBody)
	catReq := httptest.NewRequest(http.MethodPost, "/api/v1/categories", bytes.NewBuffer(catBody))
	catReq.Header.Set("Content-Type", "application/json")
	catW := httptest.NewRecorder()
	router.ServeHTTP(catW, catReq)

	var catResp map[string]interface{}
	json.Unmarshal(catW.Body.Bytes(), &catResp)
	categoryID := catResp["id"].(string)

	txReqBody := map[string]interface{}{
		"user_id":     userID,
		"category_id": categoryID,
		"amount":      25.50,
		"description": "Lunch",
	}
	txBody, _ := json.Marshal(txReqBody)
	txReq := httptest.NewRequest(http.MethodPost, "/api/v1/transactions", bytes.NewBuffer(txBody))
	txReq.Header.Set("Content-Type", "application/json")
	txW := httptest.NewRecorder()
	router.ServeHTTP(txW, txReq)

	assert.Equal(t, http.StatusCreated, txW.Code)

	var txResp map[string]interface{}
	json.Unmarshal(txW.Body.Bytes(), &txResp)
	assert.Equal(t, 25.50, txResp["amount"])
	assert.Equal(t, "Lunch", txResp["description"])

	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/transactions?user_id="+userID, nil)
	listW := httptest.NewRecorder()
	router.ServeHTTP(listW, listReq)

	assert.Equal(t, http.StatusOK, listW.Code)

	var listResp []map[string]interface{}
	json.Unmarshal(listW.Body.Bytes(), &listResp)
	assert.Equal(t, 1, len(listResp))
	assert.Equal(t, "Food", listResp[0]["category_name"])
}

func TestSummary(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	database := setupTestDB(t)
	defer cleanupTestDB(t, database)
	defer database.Close()

	logger := zerolog.New(zerolog.NewTestWriter(t))
	router := setupRouter(logger, database)

	userReqBody := map[string]string{"email": "summary-test@example.com"}
	userBody, _ := json.Marshal(userReqBody)
	userReq := httptest.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewBuffer(userBody))
	userReq.Header.Set("Content-Type", "application/json")
	userW := httptest.NewRecorder()
	router.ServeHTTP(userW, userReq)

	var userResp map[string]interface{}
	json.Unmarshal(userW.Body.Bytes(), &userResp)
	userID := userResp["id"].(string)

	for i := 0; i < 3; i++ {
		txReqBody := map[string]interface{}{
			"user_id":    userID,
			"amount":     10.0,
			"description": "Test transaction",
		}
		txBody, _ := json.Marshal(txReqBody)
		txReq := httptest.NewRequest(http.MethodPost, "/api/v1/transactions", bytes.NewBuffer(txBody))
		txReq.Header.Set("Content-Type", "application/json")
		txW := httptest.NewRecorder()
		router.ServeHTTP(txW, txReq)
	}

	summaryReq := httptest.NewRequest(http.MethodGet, "/api/v1/summary?user_id="+userID, nil)
	summaryW := httptest.NewRecorder()
	router.ServeHTTP(summaryW, summaryReq)

	assert.Equal(t, http.StatusOK, summaryW.Code)

	var summaryResp map[string]interface{}
	json.Unmarshal(summaryW.Body.Bytes(), &summaryResp)
	
	categories := summaryResp["categories"].([]interface{})
	assert.Equal(t, 1, len(categories))
	
	catSummary := categories[0].(map[string]interface{})
	assert.Equal(t, "Uncategorized", catSummary["category_name"])
	assert.Equal(t, 30.0, catSummary["total"])
}

func TestValidationErrors(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	database := setupTestDB(t)
	defer cleanupTestDB(t, database)
	defer database.Close()

	logger := zerolog.New(zerolog.NewTestWriter(t))
	router := setupRouter(logger, database)

	tests := []struct {
		name         string
		endpoint     string
		body         map[string]interface{}
		expectedCode int
		expectedMsg  string
	}{
		{
			name:     "invalid email",
			endpoint: "/api/v1/users",
			body:     map[string]interface{}{"email": "invalid-email"},
			expectedCode: http.StatusBadRequest,
			expectedMsg:  "invalid email format",
		},
		{
			name:     "missing email",
			endpoint: "/api/v1/users",
			body:     map[string]interface{}{},
			expectedCode: http.StatusBadRequest,
			expectedMsg:  "email is required",
		},
		{
			name:     "invalid amount",
			endpoint: "/api/v1/transactions",
			body: map[string]interface{}{
				"user_id": uuid.New().String(),
				"amount":  -10.0,
			},
			expectedCode: http.StatusBadRequest,
			expectedMsg:  "amount must be greater than 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPost, tt.endpoint, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedCode, w.Code)

			var resp map[string]interface{}
			json.Unmarshal(w.Body.Bytes(), &resp)
			errMsg := resp["error"].(map[string]interface{})["message"].(string)
			assert.Contains(t, errMsg, tt.expectedMsg)
		})
	}
}
