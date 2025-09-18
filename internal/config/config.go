package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig `json:"server" mapstructure:"server"`
	Logging  LogConfig    `json:"logging" mapstructure:"logging"`
	DataPath string       `json:"data_path" mapstructure:"data_path"`
}

type ServerConfig struct {
	Host string `json:"host" mapstructure:"host"`
	Port int    `json:"port" mapstructure:"port"`
}

type LogConfig struct {
	Level  string `json:"level" mapstructure:"level"`
	Format string `json:"format" mapstructure:"format"`
}

func DefaultConfig() *Config {
	homeDir, _ := os.UserHomeDir()
	dataPath := filepath.Join(homeDir, ".greetd")

	return &Config{
		Server: ServerConfig{
			Host: "0.0.0.0",
			Port: 8080,
		},
		Logging: LogConfig{
			Level:  "info",
			Format: "text",
		},
		DataPath: dataPath,
	}
}

func Load(configPath string) (*Config, error) {
	cfg := DefaultConfig()

	if configPath == "" {
		configPath = filepath.Join(cfg.DataPath, "config.json")
	}

	// Create data directory if it doesn't exist
	if err := os.MkdirAll(cfg.DataPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	// Create config file with defaults if it doesn't exist
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		if err := cfg.Save(configPath); err != nil {
			return nil, fmt.Errorf("failed to create default config: %w", err)
		}
	}

	viper.SetConfigFile(configPath)
	viper.SetEnvPrefix("GREETD")
	viper.AutomaticEnv()

	// Set defaults
	viper.SetDefault("server.host", cfg.Server.Host)
	viper.SetDefault("server.port", cfg.Server.Port)
	viper.SetDefault("logging.level", cfg.Logging.Level)
	viper.SetDefault("logging.format", cfg.Logging.Format)
	viper.SetDefault("data_path", cfg.DataPath)

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	if err := viper.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return cfg, nil
}

func (c *Config) Save(path string) error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
