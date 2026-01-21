package testutil

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"

	"fintrack-go/internal/db"
	apphttp "fintrack-go/internal/http"
)

type TestServer struct {
	Server    *httptest.Server
	Router    http.Handler
	DB        *db.DB
	Logger    zerolog.Logger
	Shutdown  func()
}

func SetupTestServer(t *testing.T) *TestServer {
	t.Helper()
	
	logger := CreateTestLogger(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	database, err := db.NewDB(ctx, GetTestDatabaseURL(), logger)
	require.NoError(t, err)

	router := apphttp.SetupRoutes(logger, database)

	server := httptest.NewServer(router)

	shutdown := func() {
		server.Close()
		database.Close()
	}

	return &TestServer{
		Server:   server,
		Router:   router,
		DB:       database,
		Logger:   logger,
		Shutdown: shutdown,
	}
}

func (ts *TestServer) URL() string {
	return ts.Server.URL
}

func (ts *TestServer) MakeRequest(t *testing.T, method, path string, body []byte) *http.Response {
	return ts.MakeRequestWithHeaders(t, method, path, body, nil)
}

func (ts *TestServer) MakeRequestWithHeaders(t *testing.T, method, path string, body []byte, headers map[string]string) *http.Response {
	t.Helper()
	
	var req *http.Request
	if body != nil {
		req = httptest.NewRequest(method, ts.Server.URL+path, bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, ts.Server.URL+path, nil)
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}
	
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err, "request failed")
	
	return resp
}

func (ts *TestServer) Get(t *testing.T, path string) *http.Response {
	return ts.MakeRequest(t, http.MethodGet, path, nil)
}

func (ts *TestServer) Post(t *testing.T, path string, body []byte) *http.Response {
	return ts.MakeRequest(t, http.MethodPost, path, body)
}

func (ts *TestServer) PostJSON(t *testing.T, path string, v interface{}) *http.Response {
	t.Helper()
	
	body, err := json.Marshal(v)
	require.NoError(t, err, "failed to marshal JSON body")
	
	return ts.Post(t, path, body)
}

func (ts *TestServer) Cleanup(t *testing.T) {
	t.Helper()
	
	TeardownTestDB(t, ts.DB.GetPool())
	ts.Shutdown()
}
