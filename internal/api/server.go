package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sirupsen/logrus"
	"github.com/svanhalla/prompt-lab/greetd/internal/config"
	"github.com/svanhalla/prompt-lab/greetd/internal/storage"
)

type Server struct {
	echo   *echo.Echo
	config *config.Config
	logger *logrus.Logger
}

func NewServer(cfg *config.Config, store *storage.MessageStore, logger *logrus.Logger) *Server {
	e := echo.New()
	e.HideBanner = true

	// Middleware
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())
	e.Use(RequestLogger(logger))

	// Handlers
	handlers := NewHandlers(store, logger, cfg.DataPath)

	// Routes
	e.GET("/health", handlers.Health)
	e.GET("/hello", handlers.Hello)
	e.GET("/message", handlers.GetMessage)
	e.POST("/message", handlers.SetMessage)
	e.GET("/ui", handlers.UI)
	e.GET("/logs", handlers.Logs)

	return &Server{
		echo:   e,
		config: cfg,
		logger: logger,
	}
}

func (s *Server) Start() error {
	addr := fmt.Sprintf("%s:%d", s.config.Server.Host, s.config.Server.Port)
	s.logger.Infof("Starting server on %s", addr)
	return s.echo.Start(addr)
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("Shutting down server...")
	return s.echo.Shutdown(ctx)
}

func RequestLogger(logger *logrus.Logger) echo.MiddlewareFunc {
	return middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:    true,
		LogStatus: true,
		LogMethod: true,
		LogLatency: true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			logger.WithFields(logrus.Fields{
				"method":  v.Method,
				"uri":     v.URI,
				"status":  v.Status,
				"latency": v.Latency,
			}).Info("HTTP request")
			return nil
		},
	})
}
