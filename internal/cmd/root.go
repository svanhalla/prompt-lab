package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/svanhalla/prompt-lab/greetd/internal/config"
	"github.com/svanhalla/prompt-lab/greetd/internal/logging"
)

var (
	cfgFile   string
	logLevel  string
	logFormat string
)

var rootCmd = &cobra.Command{
	Use:   "greetd",
	Short: "A friendly greeting and message management CLI",
	Long: `Greetd is a production-quality CLI application that manages greetings and messages.
It provides both command-line interface and web API functionality.

The name "greetd" was chosen for its simplicity and memorability - it's short,
descriptive, and follows Unix naming conventions for daemon-like applications.`,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file path")
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "info", "log level (debug, info, warn, error)")
	rootCmd.PersistentFlags().StringVar(&logFormat, "log-format", "text", "log format (text, json)")

	viper.BindPFlag("logging.level", rootCmd.PersistentFlags().Lookup("log-level"))
	viper.BindPFlag("logging.format", rootCmd.PersistentFlags().Lookup("log-format"))
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	}
}

func loadConfigAndLogger() (*config.Config, error) {
	cfg, err := config.Load(cfgFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Override with flags if provided
	if logLevel != "" {
		cfg.Logging.Level = logLevel
	}
	if logFormat != "" {
		cfg.Logging.Format = logFormat
	}

	logger, err := logging.Setup(cfg.Logging.Level, cfg.Logging.Format, cfg.DataPath)
	if err != nil {
		return nil, fmt.Errorf("failed to setup logging: %w", err)
	}

	// Store logger globally for commands to use
	globalLogger = logger

	return cfg, nil
}

var globalLogger interface{}
