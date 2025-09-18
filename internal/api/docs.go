package api

import (
	"html/template"
	"net/http"
	"os"
	"path/filepath"

	"github.com/labstack/echo/v4"
	"gopkg.in/yaml.v3"
)

func (h *Handlers) SwaggerUI(c echo.Context) error {
	tmpl := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Greetd API - Swagger UI</title>
    <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@5.9.0/swagger-ui.css" />
    <style>
        html { box-sizing: border-box; overflow: -moz-scrollbars-vertical; overflow-y: scroll; }
        *, *:before, *:after { box-sizing: inherit; }
        body { margin:0; background: #fafafa; }
    </style>
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@5.9.0/swagger-ui-bundle.js"></script>
    <script src="https://unpkg.com/swagger-ui-dist@5.9.0/swagger-ui-standalone-preset.js"></script>
    <script>
        window.onload = function() {
            const ui = SwaggerUIBundle({
                url: '/swagger/openapi.yaml',
                dom_id: '#swagger-ui',
                deepLinking: true,
                presets: [
                    SwaggerUIBundle.presets.apis,
                    SwaggerUIStandalonePreset
                ],
                plugins: [
                    SwaggerUIBundle.plugins.DownloadUrl
                ],
                layout: "StandaloneLayout"
            });
        };
    </script>
</body>
</html>`

	t, err := template.New("swagger").Parse(tmpl)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Template error"})
	}

	return t.Execute(c.Response().Writer, nil)
}

func (h *Handlers) SwaggerSpec(c echo.Context) error {
	// Try relative path first, then absolute
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
	// Try relative path first, then absolute
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

	// Parse YAML to get title and version
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

	tmpl := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}} - Documentation</title>
    <link href="https://fonts.googleapis.com/css?family=Montserrat:300,400,700|Roboto:300,400,700" rel="stylesheet">
    <style>
        body { margin: 0; padding: 0; }
    </style>
</head>
<body>
    <redoc spec-url='/swagger/openapi.yaml'></redoc>
    <script src="https://cdn.redoc.ly/redoc/latest/bundles/redoc.standalone.js"></script>
</body>
</html>`

	t, err := template.New("redoc").Parse(tmpl)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Template error"})
	}

	data_struct := struct {
		Title string
	}{
		Title: title,
	}

	return t.Execute(c.Response().Writer, data_struct)
}
