package logging

import (
	"testing"
)

func TestSetup(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name     string
		level    string
		format   string
		dataPath string
	}{
		{"info_text", "info", "text", tmpDir},
		{"debug_json", "debug", "json", tmpDir},
		{"warn_text", "warn", "text", tmpDir},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, err := Setup(tt.level, tt.format, tt.dataPath)
			if err != nil {
				t.Fatalf("Setup failed: %v", err)
			}
			if logger == nil {
				t.Fatal("Setup returned nil logger")
			}

			// Test logging works
			logger.Info("test message")
		})
	}
}

func TestSetupInvalidLevel(t *testing.T) {
	tmpDir := t.TempDir()

	logger, err := Setup("invalid", "text", tmpDir)
	if err == nil {
		t.Error("Setup should fail with invalid level")
	}
	if logger != nil {
		t.Error("Setup should return nil logger on error")
	}
}
