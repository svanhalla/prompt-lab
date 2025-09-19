package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/svanhalla/prompt-lab/greetd/internal/storage"
)

func setupTestHandlers(t *testing.T) (*Handlers, string) {
	tmpDir, err := os.MkdirTemp("", "greetd-test")
	require.NoError(t, err)

	store := storage.NewMessageStore(tmpDir)
	err = store.Load()
	require.NoError(t, err)

	logger := logrus.New()
	logger.SetOutput(os.Stderr)

	handlers, err := NewHandlers(store, logger, tmpDir)
	if err != nil {
		t.Fatalf("Failed to create handlers: %v", err)
	}

	return handlers, tmpDir
}

func TestHealthHandler(t *testing.T) {
	handlers, tmpDir := setupTestHandlers(t)
	defer os.RemoveAll(tmpDir)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handlers.Health(c)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response HealthResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "ok", response.Status)
	assert.NotEmpty(t, response.Version.Version)
}

func TestHelloHandler(t *testing.T) {
	handlers, tmpDir := setupTestHandlers(t)
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name     string
		query    string
		expected string
	}{
		{
			name:     "default name",
			query:    "",
			expected: "Hello, World!",
		},
		{
			name:     "custom name",
			query:    "?name=Alice",
			expected: "Hello, Alice!",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/hello"+tt.query, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := handlers.Hello(c)
			require.NoError(t, err)

			assert.Equal(t, http.StatusOK, rec.Code)

			var response HelloResponse
			err = json.Unmarshal(rec.Body.Bytes(), &response)
			require.NoError(t, err)

			assert.Equal(t, tt.expected, response.Message)
		})
	}
}

func TestMessageHandlers(t *testing.T) {
	handlers, tmpDir := setupTestHandlers(t)
	defer os.RemoveAll(tmpDir)

	e := echo.New()

	// Test GET message (default)
	req := httptest.NewRequest(http.MethodGet, "/message", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handlers.GetMessage(c)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response MessageResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "Hello, World!", response.Message)

	// Test POST message
	newMessage := "Hello, Universe!"
	reqBody := MessageRequest{Message: newMessage}
	jsonBody, _ := json.Marshal(reqBody)

	req = httptest.NewRequest(http.MethodPost, "/message", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)

	err = handlers.SetMessage(c)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, rec.Code)

	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, newMessage, response.Message)

	// Test GET message (updated)
	req = httptest.NewRequest(http.MethodGet, "/message", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)

	err = handlers.GetMessage(c)
	require.NoError(t, err)

	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, newMessage, response.Message)
}

func TestSetMessageValidation(t *testing.T) {
	handlers, tmpDir := setupTestHandlers(t)
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name       string
		body       string
		statusCode int
	}{
		{
			name:       "empty message",
			body:       `{"message": ""}`,
			statusCode: http.StatusBadRequest,
		},
		{
			name:       "whitespace only message",
			body:       `{"message": "   "}`,
			statusCode: http.StatusBadRequest,
		},
		{
			name:       "invalid JSON",
			body:       `{"message": }`,
			statusCode: http.StatusBadRequest,
		},
		{
			name:       "valid message",
			body:       `{"message": "Valid message"}`,
			statusCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/message", bytes.NewReader([]byte(tt.body)))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := handlers.SetMessage(c)
			require.NoError(t, err)

			assert.Equal(t, tt.statusCode, rec.Code)
		})
	}
}
