package api

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"github.com/svanhalla/prompt-lab/greetd/internal/storage"
	"github.com/svanhalla/prompt-lab/greetd/internal/version"
	"github.com/svanhalla/prompt-lab/greetd/internal/web"
	"gopkg.in/yaml.v3"
)

type Handlers struct {
	store     *storage.MessageStore
	logger    *logrus.Logger
	startTime time.Time
	dataPath  string
	templates *web.Templates
}

type HealthResponse struct {
	Status    string        `json:"status"`
	Version   version.Info  `json:"version"`
	Uptime    time.Duration `json:"uptime"`
	Timestamp time.Time     `json:"timestamp"`
}

type HelloResponse struct {
	Message string `json:"message"`
}

type MessageResponse struct {
	Message string `json:"message"`
}

type MessageRequest struct {
	Message string `json:"message"`
}

func NewHandlers(store *storage.MessageStore, logger *logrus.Logger, dataPath string) (*Handlers, error) {
	// Detect development mode by checking if template files exist
	devMode := false
	if _, err := os.Stat(filepath.Join("internal", "web", "templates", "ui.html")); err == nil {
		devMode = true
		logger.Info("Development mode: Using filesystem templates with hot reload")
	} else {
		logger.Info("Production mode: Using embedded templates")
	}

	templates, err := web.NewTemplates(devMode)
	if err != nil {
		return nil, fmt.Errorf("failed to load templates: %w", err)
	}

	return &Handlers{
		store:     store,
		logger:    logger,
		startTime: time.Now(),
		dataPath:  dataPath,
		templates: templates,
	}, nil
}

func (h *Handlers) Health(c echo.Context) error {
	return c.JSON(http.StatusOK, HealthResponse{
		Status:    "ok",
		Version:   version.Get(),
		Uptime:    time.Since(h.startTime),
		Timestamp: time.Now(),
	})
}

func (h *Handlers) Hello(c echo.Context) error {
	name := c.QueryParam("name")
	if name == "" {
		name = "World"
	}

	return c.JSON(http.StatusOK, HelloResponse{
		Message: fmt.Sprintf("Hello, %s!", name),
	})
}

func (h *Handlers) GetMessage(c echo.Context) error {
	message := h.store.GetMessage()

	return c.JSON(http.StatusOK, MessageResponse{
		Message: message,
	})
}

func (h *Handlers) SetMessage(c echo.Context) error {
	var req MessageRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid JSON"})
	}

	if strings.TrimSpace(req.Message) == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Message cannot be empty"})
	}

	if err := h.store.SetMessage(req.Message); err != nil {
		h.logger.WithError(err).Error("Failed to save message")
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to save message"})
	}

	return c.JSON(http.StatusOK, MessageResponse(req))
}

func (h *Handlers) UI(c echo.Context) error {
	message := h.store.GetMessage()

	data := struct {
		Message string
	}{
		Message: message,
	}

	c.Response().Header().Set("Content-Type", "text/html; charset=utf-8")
	return h.templates.GetUI().Execute(c.Response().Writer, data)
}

func (h *Handlers) Logs(c echo.Context) error {
	logFile := filepath.Join(h.dataPath, "app.log")

	var logs []string
	file, err := os.Open(logFile)
	if err != nil {
		logs = []string{"No logs available"}
	} else {
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			logs = append(logs, scanner.Text())
		}

		// Keep only last 50 lines
		if len(logs) > 50 {
			logs = logs[len(logs)-50:]
		}
	}

	data := struct {
		Logs []string
	}{
		Logs: logs,
	}

	c.Response().Header().Set("Content-Type", "text/html; charset=utf-8")
	return h.templates.GetLogs().Execute(c.Response().Writer, data)
}

func (h *Handlers) SwaggerUI(c echo.Context) error {
	c.Response().Header().Set("Content-Type", "text/html; charset=utf-8")
	return h.templates.GetSwagger().Execute(c.Response().Writer, nil)
}

func (h *Handlers) SwaggerSpec(c echo.Context) error {
	specPaths := []string{
		"api/openapi.yaml",
		filepath.Join(".", "api", "openapi.yaml"),
		"../../../api/openapi.yaml", // For tests
	}

	var data []byte
	var err error

	for _, specPath := range specPaths {
		data, err = os.ReadFile(specPath)
		if err == nil {
			break
		}
	}

	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "OpenAPI spec not found"})
	}

	return c.Blob(http.StatusOK, "application/yaml", data)
}

func (h *Handlers) RedocDocs(c echo.Context) error {
	specPaths := []string{
		"api/openapi.yaml",
		filepath.Join(".", "api", "openapi.yaml"),
		"../../../api/openapi.yaml", // For tests
	}

	var data []byte
	var err error

	for _, specPath := range specPaths {
		data, err = os.ReadFile(specPath)
		if err == nil {
			break
		}
	}

	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "OpenAPI spec not found"})
	}

	var spec map[string]interface{}
	if err := yaml.Unmarshal(data, &spec); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Invalid OpenAPI spec"})
	}

	info, ok := spec["info"].(map[string]interface{})
	if !ok {
		info = map[string]interface{}{"title": "Greetd API", "version": "1.0.0"}
	}

	title, _ := info["title"].(string)
	if title == "" {
		title = "Greetd API"
	}

	data_struct := struct {
		Title string
	}{
		Title: title,
	}

	c.Response().Header().Set("Content-Type", "text/html; charset=utf-8")
	return h.templates.GetRedoc().Execute(c.Response().Writer, data_struct)
}

func (h *Handlers) NotFound(c echo.Context) error {
	// For API requests (JSON), return JSON error
	if c.Request().Header.Get("Accept") == "application/json" ||
		c.Request().Header.Get("Content-Type") == "application/json" {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error":   "Not Found",
			"message": "The requested endpoint does not exist",
		})
	}

	// For browser requests, return helpful HTML page
	c.Response().Header().Set("Content-Type", "text/html; charset=utf-8")
	c.Response().WriteHeader(http.StatusNotFound)
	return h.templates.GetNotFound().Execute(c.Response().Writer, nil)
}
