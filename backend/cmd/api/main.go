package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/kernelKain/timetrap/backend/internal/analyzer"
	"github.com/kernelKain/timetrap/backend/internal/config"
	"github.com/kernelKain/timetrap/backend/internal/store/postgres"
	httpapi "github.com/kernelKain/timetrap/backend/internal/transport/http"
)

func main() {
	if err := run(); err != nil {
		slog.Error("API stopped with an error", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load configuration: %w", err)
	}

	logger := configureLogger(cfg.LogLevel)

	logger.Info("database configuration loaded")

	startupContext, cancelStartup := context.WithTimeout(
		context.Background(),
		cfg.DatabaseConnectTimeout,
	)

	database, err := postgres.Open(
		startupContext,
		postgres.Config{
			DatabaseURL:    cfg.DatabaseURL,
			MaxConns:       cfg.DatabaseMaxConns,
			MinConns:       cfg.DatabaseMinConns,
			ConnectTimeout: cfg.DatabaseConnectTimeout,
			QueryTimeout:   cfg.DatabaseQueryTimeout,
		},
	)
	cancelStartup()

	if err != nil {
		logger.Error("database connectivity check failed")

		// Return only an operation-level error because connection failures may
		// contain database addressing or identity details.
		return errors.New("database startup failed")
	}
	defer database.Close()

	logger.Info("database connection pool ready")

	handler := httpapi.NewRouter(
		httpapi.Dependencies{
			Scenarios:     database,
			Analyses:      database,
			Analyze:       analyzer.Analyze,
			DatabaseReady: database.Ping,
			Logger:        logger,
		},
		cfg.AllowedOrigins,
	)

	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	shutdownContext, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	serverErrors := make(chan error, 1)

	go func() {
		logger.Info(
			"TimeTrap API starting",
			"environment",
			cfg.AppEnvironment,
			"port",
			cfg.Port,
			"version",
			cfg.EngineVersion,
		)

		serverErrors <- server.ListenAndServe()
	}()

	select {
	case err := <-serverErrors:
		if !errors.Is(err, http.ErrServerClosed) {
			return err
		}

	case <-shutdownContext.Done():
		logger.Info("Shutdown signal received")

		timeoutContext, cancel := context.WithTimeout(
			context.Background(),
			10*time.Second,
		)
		defer cancel()

		if err := server.Shutdown(timeoutContext); err != nil {
			return err
		}

		logger.Info("TimeTrap API stopped")
	}

	return nil
}

func configureLogger(levelName string) *slog.Logger {
	level := slog.LevelInfo
	if strings.EqualFold(levelName, "debug") {
		level = slog.LevelDebug
	}

	logger := slog.New(slog.NewJSONHandler(
		os.Stdout,
		&slog.HandlerOptions{
			Level: level,
		},
	))

	slog.SetDefault(logger)

	return logger
}
