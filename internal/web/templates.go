package web

import (
	"embed"
	"html/template"
)

//go:embed templates/*.html
var templateFS embed.FS

type Templates struct {
	UI      *template.Template
	Logs    *template.Template
	NotFound *template.Template
	Swagger *template.Template
	Redoc   *template.Template
}

func NewTemplates() (*Templates, error) {
	ui, err := template.ParseFS(templateFS, "templates/ui.html")
	if err != nil {
		return nil, err
	}

	logs, err := template.ParseFS(templateFS, "templates/logs.html")
	if err != nil {
		return nil, err
	}

	notFound, err := template.ParseFS(templateFS, "templates/404.html")
	if err != nil {
		return nil, err
	}

	swagger, err := template.ParseFS(templateFS, "templates/swagger.html")
	if err != nil {
		return nil, err
	}

	redoc, err := template.ParseFS(templateFS, "templates/redoc.html")
	if err != nil {
		return nil, err
	}

	return &Templates{
		UI:       ui,
		Logs:     logs,
		NotFound: notFound,
		Swagger:  swagger,
		Redoc:    redoc,
	}, nil
}
