package web

import (
	"embed"
	"html/template"
	"os"
	"path/filepath"
)

//go:embed templates/*.html
var templateFS embed.FS

type Templates struct {
	UI      *template.Template
	Logs    *template.Template
	NotFound *template.Template
	Swagger *template.Template
	Redoc   *template.Template
	devMode bool
}

// parseTemplate tries to load from filesystem first, falls back to embedded
func parseTemplate(name string, devMode bool) (*template.Template, error) {
	// In development mode, always try filesystem first
	if devMode {
		fsPath := filepath.Join("internal", "web", "templates", name)
		if _, err := os.Stat(fsPath); err == nil {
			return template.ParseFiles(fsPath)
		}
	}
	
	// Fallback to embedded (for production or when filesystem not available)
	return template.ParseFS(templateFS, "templates/"+name)
}

// reloadTemplate reloads a template from filesystem if in dev mode
func (t *Templates) reloadTemplate(name string) *template.Template {
	if !t.devMode {
		return nil // Don't reload in production
	}
	
	fsPath := filepath.Join("internal", "web", "templates", name)
	if _, err := os.Stat(fsPath); err == nil {
		if tmpl, err := template.ParseFiles(fsPath); err == nil {
			return tmpl
		}
	}
	return nil
}

// GetUI returns UI template, reloading from filesystem if in dev mode
func (t *Templates) GetUI() *template.Template {
	if reloaded := t.reloadTemplate("ui.html"); reloaded != nil {
		return reloaded
	}
	return t.UI
}

// GetLogs returns Logs template, reloading from filesystem if in dev mode
func (t *Templates) GetLogs() *template.Template {
	if reloaded := t.reloadTemplate("logs.html"); reloaded != nil {
		return reloaded
	}
	return t.Logs
}

// GetNotFound returns NotFound template, reloading from filesystem if in dev mode
func (t *Templates) GetNotFound() *template.Template {
	if reloaded := t.reloadTemplate("404.html"); reloaded != nil {
		return reloaded
	}
	return t.NotFound
}

// GetSwagger returns Swagger template, reloading from filesystem if in dev mode
func (t *Templates) GetSwagger() *template.Template {
	if reloaded := t.reloadTemplate("swagger.html"); reloaded != nil {
		return reloaded
	}
	return t.Swagger
}

// GetRedoc returns Redoc template, reloading from filesystem if in dev mode
func (t *Templates) GetRedoc() *template.Template {
	if reloaded := t.reloadTemplate("redoc.html"); reloaded != nil {
		return reloaded
	}
	return t.Redoc
}

func NewTemplates(devMode bool) (*Templates, error) {
	ui, err := parseTemplate("ui.html", devMode)
	if err != nil {
		return nil, err
	}

	logs, err := parseTemplate("logs.html", devMode)
	if err != nil {
		return nil, err
	}

	notFound, err := parseTemplate("404.html", devMode)
	if err != nil {
		return nil, err
	}

	swagger, err := parseTemplate("swagger.html", devMode)
	if err != nil {
		return nil, err
	}

	redoc, err := parseTemplate("redoc.html", devMode)
	if err != nil {
		return nil, err
	}

	return &Templates{
		UI:       ui,
		Logs:     logs,
		NotFound: notFound,
		Swagger:  swagger,
		Redoc:    redoc,
		devMode:  devMode,
	}, nil
}
