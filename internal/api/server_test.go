package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/svanhalla/prompt-lab/greetd/internal/config"
	"github.com/svanhalla/prompt-lab/greetd/internal/storage"
)

func TestServerE2E(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "greetd-e2e-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create api directory and OpenAPI spec for testing
	apiDir := filepath.Join(tmpDir, "api")
	err = os.MkdirAll(apiDir, 0755)
	require.NoError(t, err)

	openAPISpec := `openapi: 3.1.0
info:
  title: Greetd API
  version: 1.0.0
paths:
  /health:
    get:
      summary: Health check
`
	err = os.WriteFile(filepath.Join(apiDir, "openapi.yaml"), []byte(openAPISpec), 0644)
	require.NoError(t, err)

	// Change to temp directory for test
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tmpDir)

	// Setup config
	cfg := config.DefaultConfig()
	cfg.DataPath = tmpDir
	cfg.Server.Host = "127.0.0.1"
	cfg.Server.Port = 0 // Use ephemeral port

	// Setup logger
	logger := logrus.New()
	logger.SetOutput(os.Stderr)

	// Setup storage
	store := storage.NewMessageStore(tmpDir)
	err = store.Load()
	require.NoError(t, err)

	// Create server
	server := NewServer(cfg, store, logger)

	// Start server on ephemeral port
	testServer := httptest.NewServer(server.echo)
	defer testServer.Close()

	baseURL := testServer.URL

	// Test health endpoint
	resp, err := http.Get(baseURL + "/health")
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var healthResp HealthResponse
	err = json.NewDecoder(resp.Body).Decode(&healthResp)
	require.NoError(t, err)
	assert.Equal(t, "ok", healthResp.Status)

	// Test hello endpoint
	resp, err = http.Get(baseURL + "/hello?name=E2ETest")
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var helloResp HelloResponse
	err = json.NewDecoder(resp.Body).Decode(&helloResp)
	require.NoError(t, err)
	assert.Equal(t, "Hello, E2ETest!", helloResp.Message)

	// Test get message endpoint
	resp, err = http.Get(baseURL + "/message")
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var messageResp MessageResponse
	err = json.NewDecoder(resp.Body).Decode(&messageResp)
	require.NoError(t, err)
	assert.Equal(t, "Hello, World!", messageResp.Message)

	// Test set message endpoint
	newMessage := "E2E Test Message"
	reqBody := MessageRequest{Message: newMessage}
	jsonBody, _ := json.Marshal(reqBody)

	resp, err = http.Post(baseURL+"/message", "application/json", bytes.NewReader(jsonBody))
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	err = json.NewDecoder(resp.Body).Decode(&messageResp)
	require.NoError(t, err)
	assert.Equal(t, newMessage, messageResp.Message)

	// Verify message was persisted
	resp, err = http.Get(baseURL + "/message")
	require.NoError(t, err)
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&messageResp)
	require.NoError(t, err)
	assert.Equal(t, newMessage, messageResp.Message)

	// Test UI endpoint
	resp, err = http.Get(baseURL + "/ui")
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, resp.Header.Get("Content-Type"), "text/html")

	// Test logs endpoint
	resp, err = http.Get(baseURL + "/logs")
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, resp.Header.Get("Content-Type"), "text/html")

	// Test Swagger UI endpoint
	resp, err = http.Get(baseURL + "/swagger/")
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, resp.Header.Get("Content-Type"), "text/html")

	// Test OpenAPI spec endpoint
	resp, err = http.Get(baseURL + "/swagger/openapi.yaml")
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, resp.Header.Get("Content-Type"), "application/yaml")

	// Test Redoc endpoint
	resp, err = http.Get(baseURL + "/docs")
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, resp.Header.Get("Content-Type"), "text/html")
}

func TestServerGracefulShutdown(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "greetd-shutdown-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Setup config
	cfg := config.DefaultConfig()
	cfg.DataPath = tmpDir

	// Setup logger
	logger := logrus.New()
	logger.SetOutput(os.Stderr)

	// Setup storage
	store := storage.NewMessageStore(tmpDir)
	err = store.Load()
	require.NoError(t, err)

	// Create server
	server := NewServer(cfg, store, logger)

	// Test graceful shutdown
	done := make(chan error, 1)
	go func() {
		done <- server.echo.Start(":0")
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Shutdown server
	err = server.echo.Close()
	assert.NoError(t, err)

	// Wait for server to stop
	select {
	case err := <-done:
		assert.Error(t, err) // Server should return error when closed
	case <-time.After(5 * time.Second):
		t.Fatal("Server did not shutdown within timeout")
	}
}
