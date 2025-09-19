package web

import (
	"testing"
)

func TestNewTemplates(t *testing.T) {
	// Test with dev mode false (embedded templates)
	templates, err := NewTemplates(false)
	if err != nil {
		t.Fatalf("NewTemplates(false) failed: %v", err)
	}
	if templates == nil {
		t.Fatal("NewTemplates(false) returned nil templates")
	}

	// Test template getters
	if templates.GetUI() == nil {
		t.Error("GetUI() returned nil")
	}
	if templates.GetLogs() == nil {
		t.Error("GetLogs() returned nil")
	}
	if templates.GetNotFound() == nil {
		t.Error("GetNotFound() returned nil")
	}
	if templates.GetSwagger() == nil {
		t.Error("GetSwagger() returned nil")
	}
	if templates.GetRedoc() == nil {
		t.Error("GetRedoc() returned nil")
	}
}

func TestNewTemplatesDevMode(t *testing.T) {
	// Test with dev mode true (filesystem templates if available)
	templates, err := NewTemplates(true)
	if err != nil {
		t.Fatalf("NewTemplates(true) failed: %v", err)
	}
	if templates == nil {
		t.Fatal("NewTemplates(true) returned nil templates")
	}
}
