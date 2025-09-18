package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/svanhalla/prompt-lab/greetd/internal/api"
	"github.com/svanhalla/prompt-lab/greetd/internal/storage"
)

var (
	host string
	port int
)

var apiCmd = &cobra.Command{
	Use:   "api",
	Short: "Start the HTTP API and Web server",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := loadConfigAndLogger()
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			os.Exit(1)
		}

		logger := globalLogger.(*logrus.Logger)

		// Override with flags if provided
		if host != "" {
			cfg.Server.Host = host
		}
		if port != 0 {
			cfg.Server.Port = port
		}

		// Initialize message store
		store := storage.NewMessageStore(cfg.DataPath)
		if err := store.Load(); err != nil {
			logger.WithError(err).Fatal("Failed to load message store")
		}

		// Create and start server
		server, err := api.NewServer(cfg, store, logger)
		if err != nil {
			logger.WithError(err).Fatal("Failed to create server")
		}

		// Graceful shutdown
		go func() {
			if err := server.Start(); err != nil {
				logger.WithError(err).Fatal("Server failed to start")
			}
		}()

		// Wait for interrupt signal
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit

		// Shutdown with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			logger.WithError(err).Error("Server shutdown error")
		}
	},
}

func init() {
	apiCmd.Flags().StringVar(&host, "host", "", "server host")
	apiCmd.Flags().IntVar(&port, "port", 0, "server port")

	viper.BindPFlag("server.host", apiCmd.Flags().Lookup("host"))
	viper.BindPFlag("server.port", apiCmd.Flags().Lookup("port"))

	rootCmd.AddCommand(apiCmd)
}
