package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/svanhalla/prompt-lab/greetd/internal/config"
	"github.com/svanhalla/prompt-lab/greetd/internal/storage"
)

// TestRuntimeVerification validates all documented endpoints return expected responses
func TestRuntimeVerification(t *testing.T) {
	// Setup test environment
	tmpDir, err := os.MkdirTemp("", "greetd-runtime-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create OpenAPI spec for testing
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

	// Change to temp directory
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tmpDir)

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

	// Start test server
	testServer := httptest.NewServer(server.echo)
	defer testServer.Close()

	baseURL := testServer.URL

	t.Run("Server starts successfully", func(t *testing.T) {
		// Server should be running without errors
		resp, err := http.Get(baseURL + "/health")
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("/health returns status JSON with ok and version info", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/health")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Contains(t, resp.Header.Get("Content-Type"), "application/json")

		var healthResp HealthResponse
		err = json.NewDecoder(resp.Body).Decode(&healthResp)
		require.NoError(t, err)

		assert.Equal(t, "ok", healthResp.Status)
		assert.NotEmpty(t, healthResp.Version.Version)
		assert.NotEmpty(t, healthResp.Version.GoVersion)
		assert.NotZero(t, healthResp.Timestamp)
	})

	t.Run("/hello?name=Test returns greeting JSON with Hello, Test!", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/hello?name=Test")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Contains(t, resp.Header.Get("Content-Type"), "application/json")

		var helloResp HelloResponse
		err = json.NewDecoder(resp.Body).Decode(&helloResp)
		require.NoError(t, err)

		assert.Equal(t, "Hello, Test!", helloResp.Message)
	})

	t.Run("/message returns latest stored message", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/message")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Contains(t, resp.Header.Get("Content-Type"), "application/json")

		var messageResp MessageResponse
		err = json.NewDecoder(resp.Body).Decode(&messageResp)
		require.NoError(t, err)

		assert.Equal(t, "Hello, World!", messageResp.Message)
	})

	t.Run("POST /message updates and returns message JSON", func(t *testing.T) {
		newMessage := "Updated message from runtime test"
		reqBody := MessageRequest{Message: newMessage}
		jsonBody, _ := json.Marshal(reqBody)

		resp, err := http.Post(baseURL+"/message", "application/json", bytes.NewReader(jsonBody))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Contains(t, resp.Header.Get("Content-Type"), "application/json")

		var messageResp MessageResponse
		err = json.NewDecoder(resp.Body).Decode(&messageResp)
		require.NoError(t, err)

		assert.Equal(t, newMessage, messageResp.Message)

		// Verify persistence
		resp, err = http.Get(baseURL + "/message")
		require.NoError(t, err)
		defer resp.Body.Close()

		err = json.NewDecoder(resp.Body).Decode(&messageResp)
		require.NoError(t, err)
		assert.Equal(t, newMessage, messageResp.Message)
	})

	t.Run("/ui renders HTML page with current message and update form", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/ui")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Contains(t, resp.Header.Get("Content-Type"), "text/html")

		body := make([]byte, 4096)
		n, _ := resp.Body.Read(body)
		content := string(body[:n])

		// Check for Tailwind CSS
		assert.Contains(t, content, "tailwindcss.com")
		// Check for message display
		assert.Contains(t, content, "Current Message")
		// Check for update form
		assert.Contains(t, content, "form")
		assert.Contains(t, content, "Update")
	})

	t.Run("/logs renders HTML page with recent logs", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/logs")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Contains(t, resp.Header.Get("Content-Type"), "text/html")

		body := make([]byte, 1024)
		n, _ := resp.Body.Read(body)
		content := string(body[:n])

		// Check for Tailwind CSS
		assert.Contains(t, content, "tailwindcss.com")
		// Check for logs display
		assert.Contains(t, content, "Application Logs")
	})

	t.Run("/swagger/ loads Swagger UI with API documentation", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/swagger/")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Contains(t, resp.Header.Get("Content-Type"), "text/html")

		body := make([]byte, 2048)
		n, _ := resp.Body.Read(body)
		content := string(body[:n])

		// Check for Swagger UI elements
		assert.Contains(t, content, "swagger-ui")
		assert.Contains(t, content, "SwaggerUIBundle")
		assert.Contains(t, content, "Greetd API")
	})

	t.Run("/docs renders Redoc page cleanly", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/docs")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Contains(t, resp.Header.Get("Content-Type"), "text/html")

		body := make([]byte, 2048)
		n, _ := resp.Body.Read(body)
		content := string(body[:n])

		// Check for Redoc elements
		assert.Contains(t, content, "redoc")
		assert.Contains(t, content, "redoc.standalone.js")
		assert.Contains(t, content, "Greetd API")
	})

	t.Run("OpenAPI spec endpoint serves YAML", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/swagger/openapi.yaml")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Contains(t, resp.Header.Get("Content-Type"), "application/yaml")

		body := make([]byte, 1024)
		n, _ := resp.Body.Read(body)
		content := string(body[:n])

		// Check for OpenAPI spec content
		assert.Contains(t, content, "openapi:")
		assert.Contains(t, content, "Greetd API")
	})
}

// TestServerStartupValidation ensures server starts without errors
func TestServerStartupValidation(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "greetd-startup-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	cfg := config.DefaultConfig()
	cfg.DataPath = tmpDir

	logger := logrus.New()
	logger.SetOutput(os.Stderr)

	store := storage.NewMessageStore(tmpDir)
	err = store.Load()
	require.NoError(t, err)

	// Server creation should not panic or error
	server := NewServer(cfg, store, logger)
	assert.NotNil(t, server)
	assert.NotNil(t, server.echo)

	// Test server can bind to ephemeral port
	testServer := httptest.NewServer(server.echo)
	defer testServer.Close()

	// Basic connectivity test
	resp, err := http.Get(testServer.URL + "/health")
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestCoreEndpointsSmoke validates core endpoints at runtime
func TestCoreEndpointsSmoke(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "greetd-smoke-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create OpenAPI spec for testing
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

	// Change to temp directory
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tmpDir)

	cfg := config.DefaultConfig()
	cfg.DataPath = tmpDir

	logger := logrus.New()
	logger.SetOutput(os.Stderr)

	store := storage.NewMessageStore(tmpDir)
	err = store.Load()
	require.NoError(t, err)

	server := NewServer(cfg, store, logger)
	testServer := httptest.NewServer(server.echo)
	defer testServer.Close()

	baseURL := testServer.URL

	// Core endpoints that must work
	endpoints := []struct {
		name   string
		method string
		path   string
		status int
	}{
		{"Health Check", "GET", "/health", 200},
		{"Hello Default", "GET", "/hello", 200},
		{"Hello Named", "GET", "/hello?name=Smoke", 200},
		{"Get Message", "GET", "/message", 200},
		{"Web UI", "GET", "/ui", 200},
		{"Logs Page", "GET", "/logs", 200},
		{"Swagger UI", "GET", "/swagger/", 200},
		{"Redoc Docs", "GET", "/docs", 200},
	}

	for _, endpoint := range endpoints {
		t.Run(endpoint.name, func(t *testing.T) {
			var resp *http.Response
			var err error

			switch endpoint.method {
			case "GET":
				resp, err = http.Get(baseURL + endpoint.path)
			default:
				t.Fatalf("Unsupported method: %s", endpoint.method)
			}

			require.NoError(t, err, "Failed to make request to %s", endpoint.path)
			defer resp.Body.Close()

			assert.Equal(t, endpoint.status, resp.StatusCode, 
				"Unexpected status for %s %s", endpoint.method, endpoint.path)
		})
	}

	// Test POST /message
	t.Run("POST Message", func(t *testing.T) {
		reqBody := MessageRequest{Message: "Smoke test message"}
		jsonBody, _ := json.Marshal(reqBody)

		resp, err := http.Post(baseURL+"/message", "application/json", bytes.NewReader(jsonBody))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var messageResp MessageResponse
		err = json.NewDecoder(resp.Body).Decode(&messageResp)
		require.NoError(t, err)
		assert.Equal(t, "Smoke test message", messageResp.Message)
	})
}
