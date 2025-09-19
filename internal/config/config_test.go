package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	assert.Equal(t, "0.0.0.0", cfg.Server.Host)
	assert.Equal(t, 8080, cfg.Server.Port)
	assert.Equal(t, "info", cfg.Logging.Level)
	assert.Equal(t, "text", cfg.Logging.Format)
	assert.NotEmpty(t, cfg.DataPath)
}

func TestConfigSaveAndLoad(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "greetd-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	configPath := filepath.Join(tmpDir, "config.json")

	// Create and save config
	cfg := DefaultConfig()
	cfg.DataPath = tmpDir
	cfg.Server.Port = 9090

	err = cfg.Save(configPath)
	require.NoError(t, err)

	// Load config
	loadedCfg, err := Load(configPath)
	require.NoError(t, err)

	assert.Equal(t, cfg.Server.Host, loadedCfg.Server.Host)
	assert.Equal(t, cfg.Server.Port, loadedCfg.Server.Port)
	assert.Equal(t, cfg.Logging.Level, loadedCfg.Logging.Level)
	assert.Equal(t, cfg.Logging.Format, loadedCfg.Logging.Format)
}

func TestLoadNonExistentConfig(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "greetd-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	configPath := filepath.Join(tmpDir, "config.json")

	// Load non-existent config (should create default)
	cfg, err := Load(configPath)
	require.NoError(t, err)

	// Verify default values
	assert.Equal(t, "0.0.0.0", cfg.Server.Host)
	assert.Equal(t, 8080, cfg.Server.Port)

	// Verify config file was created
	_, err = os.Stat(configPath)
	assert.NoError(t, err)
}
