package api

import (
	"bufio"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"github.com/svanhalla/prompt-lab/greetd/internal/storage"
	"github.com/svanhalla/prompt-lab/greetd/internal/version"
)

type Handlers struct {
	store     *storage.MessageStore
	logger    *logrus.Logger
	startTime time.Time
	dataPath  string
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

type MessageRequest struct {
	Message string `json:"message"`
}

type MessageResponse struct {
	Message string `json:"message"`
}

func NewHandlers(store *storage.MessageStore, logger *logrus.Logger, dataPath string) *Handlers {
	return &Handlers{
		store:     store,
		logger:    logger,
		startTime: time.Now(),
		dataPath:  dataPath,
	}
}

func (h *Handlers) Health(c echo.Context) error {
	response := HealthResponse{
		Status:    "ok",
		Version:   version.Get(),
		Uptime:    time.Since(h.startTime),
		Timestamp: time.Now(),
	}
	return c.JSON(http.StatusOK, response)
}

func (h *Handlers) Hello(c echo.Context) error {
	name := c.QueryParam("name")
	if name == "" {
		name = "World"
	}

	response := HelloResponse{
		Message: fmt.Sprintf("Hello, %s!", name),
	}
	return c.JSON(http.StatusOK, response)
}

func (h *Handlers) GetMessage(c echo.Context) error {
	message := h.store.GetMessage()
	response := MessageResponse{Message: message}
	return c.JSON(http.StatusOK, response)
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

	response := MessageResponse{Message: req.Message}
	return c.JSON(http.StatusOK, response)
}

func (h *Handlers) UI(c echo.Context) error {
	message := h.store.GetMessage()

	tmpl := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Greetd - Message Manager</title>
    <script src="https://cdn.tailwindcss.com"></script>
</head>
<body class="bg-gray-100 min-h-screen">
    <div class="container mx-auto px-4 py-8">
        <div class="max-w-md mx-auto bg-white rounded-lg shadow-md p-6">
            <h1 class="text-2xl font-bold text-gray-800 mb-6 text-center">Message Manager</h1>
            
            <div class="mb-6">
                <h2 class="text-lg font-semibold text-gray-700 mb-2">Current Message:</h2>
                <div class="bg-gray-50 p-4 rounded border">
                    <p class="text-gray-800">{{.Message}}</p>
                </div>
            </div>

            <form id="messageForm" class="space-y-4">
                <div>
                    <label for="message" class="block text-sm font-medium text-gray-700 mb-2">
                        Update Message:
                    </label>
                    <textarea 
                        id="message" 
                        name="message" 
                        rows="3" 
                        class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                        placeholder="Enter your message here..."
                        required
                    ></textarea>
                </div>
                <button 
                    type="submit" 
                    class="w-full bg-blue-500 hover:bg-blue-600 text-white font-medium py-2 px-4 rounded-md transition duration-200"
                >
                    Update Message
                </button>
            </form>

            <div class="mt-6 text-center">
                <a href="/logs" class="text-blue-500 hover:text-blue-600 text-sm">View Application Logs</a>
            </div>
        </div>
    </div>

    <script>
        document.getElementById('messageForm').addEventListener('submit', async (e) => {
            e.preventDefault();
            const message = document.getElementById('message').value;
            
            try {
                const response = await fetch('/message', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    body: JSON.stringify({ message }),
                });
                
                if (response.ok) {
                    location.reload();
                } else {
                    const error = await response.json();
                    alert('Error: ' + error.error);
                }
            } catch (err) {
                alert('Error updating message: ' + err.message);
            }
        });
    </script>
</body>
</html>`

	t, err := template.New("ui").Parse(tmpl)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Template error"})
	}

	data := struct {
		Message string
	}{
		Message: message,
	}

	return t.Execute(c.Response().Writer, data)
}

func (h *Handlers) NotFound(c echo.Context) error {
	// For API requests (JSON), return JSON error
	if c.Request().Header.Get("Accept") == "application/json" || 
	   c.Request().Header.Get("Content-Type") == "application/json" {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "Not Found",
			"message": "The requested endpoint does not exist",
		})
	}

	// For browser requests, return helpful HTML page
	tmpl := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Page Not Found - Greetd</title>
    <script src="https://cdn.tailwindcss.com"></script>
</head>
<body class="bg-gray-100 min-h-screen flex items-center justify-center">
    <div class="max-w-md mx-auto bg-white rounded-lg shadow-md p-8">
        <div class="text-center mb-6">
            <h1 class="text-4xl font-bold text-gray-800 mb-2">404</h1>
            <p class="text-gray-600">Page not found</p>
        </div>
        
        <div class="mb-6">
            <p class="text-gray-700 mb-4">The page you're looking for doesn't exist. Here are the available endpoints:</p>
        </div>

        <div class="space-y-2">
            <a href="/" class="block w-full text-left px-4 py-2 bg-blue-50 hover:bg-blue-100 rounded border text-blue-700 hover:text-blue-800 transition-colors">
                <strong>/</strong> - Home (redirects to UI)
            </a>
            <a href="/ui" class="block w-full text-left px-4 py-2 bg-green-50 hover:bg-green-100 rounded border text-green-700 hover:text-green-800 transition-colors">
                <strong>/ui</strong> - Web Interface
            </a>
            <a href="/health" class="block w-full text-left px-4 py-2 bg-purple-50 hover:bg-purple-100 rounded border text-purple-700 hover:text-purple-800 transition-colors">
                <strong>/health</strong> - Health Check
            </a>
            <a href="/hello" class="block w-full text-left px-4 py-2 bg-yellow-50 hover:bg-yellow-100 rounded border text-yellow-700 hover:text-yellow-800 transition-colors">
                <strong>/hello</strong> - Greeting API
            </a>
            <a href="/message" class="block w-full text-left px-4 py-2 bg-indigo-50 hover:bg-indigo-100 rounded border text-indigo-700 hover:text-indigo-800 transition-colors">
                <strong>/message</strong> - Message API
            </a>
            <a href="/logs" class="block w-full text-left px-4 py-2 bg-gray-50 hover:bg-gray-100 rounded border text-gray-700 hover:text-gray-800 transition-colors">
                <strong>/logs</strong> - Application Logs
            </a>
            <a href="/swagger/" class="block w-full text-left px-4 py-2 bg-red-50 hover:bg-red-100 rounded border text-red-700 hover:text-red-800 transition-colors">
                <strong>/swagger/</strong> - API Documentation (Swagger)
            </a>
            <a href="/docs" class="block w-full text-left px-4 py-2 bg-pink-50 hover:bg-pink-100 rounded border text-pink-700 hover:text-pink-800 transition-colors">
                <strong>/docs</strong> - API Documentation (Redoc)
            </a>
        </div>

        <div class="mt-6 text-center">
            <p class="text-sm text-gray-500">
                Greetd - A friendly CLI and API application
            </p>
        </div>
    </div>
</body>
</html>`

	t, err := template.New("404").Parse(tmpl)
	if err != nil {
		return c.String(http.StatusNotFound, "404 - Page Not Found")
	}

	c.Response().Header().Set("Content-Type", "text/html; charset=utf-8")
	c.Response().WriteHeader(http.StatusNotFound)
	return t.Execute(c.Response().Writer, nil)
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
		
		// Keep only last 100 lines
		if len(logs) > 100 {
			logs = logs[len(logs)-100:]
		}
	}

	tmpl := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Greetd - Application Logs</title>
    <script src="https://cdn.tailwindcss.com"></script>
</head>
<body class="bg-gray-100 min-h-screen">
    <div class="container mx-auto px-4 py-8">
        <div class="max-w-4xl mx-auto bg-white rounded-lg shadow-md p-6">
            <div class="flex justify-between items-center mb-6">
                <h1 class="text-2xl font-bold text-gray-800">Application Logs</h1>
                <a href="/ui" class="text-blue-500 hover:text-blue-600">← Back to UI</a>
            </div>
            
            <div class="bg-gray-900 text-green-400 p-4 rounded-lg font-mono text-sm overflow-x-auto">
                {{range .Logs}}
                <div class="mb-1">{{.}}</div>
                {{end}}
            </div>
        </div>
    </div>
</body>
</html>`

	t, err := template.New("logs").Parse(tmpl)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Template error"})
	}

	data := struct {
		Logs []string
	}{
		Logs: logs,
	}

	return t.Execute(c.Response().Writer, data)
}
