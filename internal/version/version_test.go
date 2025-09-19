package version

import (
	"strings"
	"testing"
)

func TestGet(t *testing.T) {
	info := Get()

	if info.Version == "" {
		t.Error("Version should not be empty")
	}

	if info.Commit == "" {
		t.Error("Commit should not be empty")
	}

	if info.BuildTime == "" {
		t.Error("BuildTime should not be empty")
	}

	if info.GoVersion == "" {
		t.Error("GoVersion should not be empty")
	}

	if !strings.HasPrefix(info.GoVersion, "go") {
		t.Errorf("GoVersion should start with 'go', got: %s", info.GoVersion)
	}
}

func TestVariables(t *testing.T) {
	if Version == "" {
		t.Error("Version variable should not be empty")
	}

	if Commit == "" {
		t.Error("Commit variable should not be empty")
	}

	if BuildTime == "" {
		t.Error("BuildTime variable should not be empty")
	}
}
