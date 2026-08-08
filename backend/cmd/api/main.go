package main

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

type healthResponse struct {
	Status    string `json:"status"`
	Service   string `json:"service"`
	Version   string `json:"version"`
	Database  string `json:"database"`
	Timestamp string `json:"timestamp"`
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

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

	configureLogger(environment("LOG_LEVEL", "debug"))

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/health", healthHandler(version))

	handler := requestLogger(corsMiddleware(allowedOrigins, mux))

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
		slog.Info(
			"TimeTrap API starting",
			"port", port,
			"version", version,
		)

		serverErrors <- server.ListenAndServe()
	}()

	select {
	case err := <-serverErrors:
		if !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	case <-shutdownContext.Done():
		slog.Info("Shutdown signal received")

		timeoutContext, cancel := context.WithTimeout(
			context.Background(),
			10*time.Second,
		)
		defer cancel()

		if err := server.Shutdown(timeoutContext); err != nil {
			return err
		}

		slog.Info("TimeTrap API stopped")
	}

	return nil
}

func healthHandler(version string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", http.MethodGet)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		response := healthResponse{
			Status:    "ok",
			Service:   "timetrap-api",
			Version:   version,
			Database:  "not_configured",
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		}

		w.Header().Set("Content-Type", "application/json")

		if err := json.NewEncoder(w).Encode(response); err != nil {
			slog.Error("Unable to encode health response", "error", err)
		}
	}
}

func corsMiddleware(
	allowedOrigins map[string]struct{},
	next http.Handler,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		if origin != "" {
			if _, allowed := allowedOrigins[origin]; !allowed {
				http.Error(w, "origin not allowed", http.StatusForbidden)
				return
			}

			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			w.Header().Add("Vary", "Origin")
		}

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func requestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		recorder := &statusRecorder{
			ResponseWriter: w,
			status:         http.StatusOK,
		}

		next.ServeHTTP(recorder, r)

		slog.Info(
			"HTTP request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", recorder.status,
			"duration", time.Since(started),
		)
	})
}

func environment(name, fallback string) string {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return fallback
	}

	return value
}

func parseAllowedOrigins(value string) map[string]struct{} {
	origins := make(map[string]struct{})

	for _, origin := range strings.Split(value, ",") {
		origin = strings.TrimSpace(origin)
		if origin != "" {
			origins[origin] = struct{}{}
		}
	}

	return origins
}

func configureLogger(levelName string) {
	level := slog.LevelInfo
	if strings.EqualFold(levelName, "debug") {
		level = slog.LevelDebug
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	}))

	slog.SetDefault(logger)
}
