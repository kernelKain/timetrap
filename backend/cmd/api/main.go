package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	httpapi "github.com/kernelKain/timetrap/backend/internal/transport/http"
)

func main() {
	if err := run(); err != nil {
		slog.Error("API stopped with an error", "error", err)
		os.Exit(1)
	}
}

func run() error {
	port := environment("PORT", "8080")
	version := environment("ENGINE_VERSION", "dev")
	allowedOrigins := parseAllowedOrigins(
		environment("ALLOWED_ORIGINS", "http://localhost:5173"),
	)

	logger := configureLogger(environment("LOG_LEVEL", "debug"))

	handler := httpapi.NewRouter(
		logger,
		allowedOrigins,
		version,
	)

	server := &http.Server{
		Addr:              ":" + port,
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
			"port",
			port,
			"version",
			version,
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

func environment(name, fallback string) string {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return fallback
	}

	return value
}

func parseAllowedOrigins(value string) []string {
	origins := make([]string, 0)

	for _, origin := range strings.Split(value, ",") {
		origin = strings.TrimSpace(origin)
		if origin != "" {
			origins = append(origins, origin)
		}
	}

	return origins
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
