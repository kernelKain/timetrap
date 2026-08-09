package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestHealthRouteConnected(t *testing.T) {
	calls := 0
	deadlineObserved := false

	handler := newHealthTestRouter(func(ctx context.Context) error {
		calls++

		deadline, exists := ctx.Deadline()
		if !exists {
			t.Fatal("database health context has no deadline")
		}

		remaining := time.Until(deadline)
		if remaining <= 0 || remaining > time.Second {
			t.Fatalf(
				"expected deadline within one second, remaining=%s",
				remaining,
			)
		}

		deadlineObserved = true
		return nil
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/health",
		nil,
	)

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d; body=%s",
			http.StatusOK,
			recorder.Code,
			recorder.Body.String(),
		)
	}

	if calls != 1 {
		t.Fatalf("expected one database check, got %d", calls)
	}

	if !deadlineObserved {
		t.Fatal("database check did not observe the health deadline")
	}

	var response healthResponse
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("decode health response: %v", err)
	}

	if response.Status != "ok" {
		t.Fatalf("expected status %q, got %q", "ok", response.Status)
	}

	if response.Service != "timetrap-api" {
		t.Fatalf(
			"expected service %q, got %q",
			"timetrap-api",
			response.Service,
		)
	}

	if response.Database != "connected" {
		t.Fatalf(
			"expected database %q, got %q",
			"connected",
			response.Database,
		)
	}

	if _, err := time.Parse(time.RFC3339, response.Timestamp); err != nil {
		t.Fatalf(
			"health timestamp is not RFC3339: %q",
			response.Timestamp,
		)
	}

	if strings.Contains(recorder.Body.String(), `"version"`) {
		t.Fatal("Phase 4 health response unexpectedly contains version")
	}
}

func TestHealthRouteUnavailable(t *testing.T) {
	const privateDetail = "postgres://user:password@internal-db-host/database"

	handler := newHealthTestRouter(func(context.Context) error {
		return errors.New(privateDetail)
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/health",
		nil,
	)

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf(
			"expected status %d, got %d; body=%s",
			http.StatusServiceUnavailable,
			recorder.Code,
			recorder.Body.String(),
		)
	}

	var response healthResponse
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("decode health response: %v", err)
	}

	if response.Status != "degraded" {
		t.Fatalf(
			"expected status %q, got %q",
			"degraded",
			response.Status,
		)
	}

	if response.Database != "unavailable" {
		t.Fatalf(
			"expected database %q, got %q",
			"unavailable",
			response.Database,
		)
	}

	if strings.Contains(recorder.Body.String(), privateDetail) ||
		strings.Contains(recorder.Body.String(), "internal-db-host") ||
		strings.Contains(recorder.Body.String(), "password") {
		t.Fatal("health response leaked private database information")
	}
}

func TestHealthRouteWithoutCheckerIsUnavailable(t *testing.T) {
	handler := newHealthTestRouter(nil)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/health",
		nil,
	)

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf(
			"expected status %d, got %d; body=%s",
			http.StatusServiceUnavailable,
			recorder.Code,
			recorder.Body.String(),
		)
	}

	var response healthResponse
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("decode health response: %v", err)
	}

	if response.Status != "degraded" ||
		response.Database != "unavailable" {
		t.Fatalf("unexpected degraded response: %#v", response)
	}
}

func TestHealthRouteRejectsWrongMethod(t *testing.T) {
	calls := 0

	handler := newHealthTestRouter(func(context.Context) error {
		calls++
		return nil
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/health",
		nil,
	)

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusMethodNotAllowed {
		t.Fatalf(
			"expected status %d, got %d; body=%s",
			http.StatusMethodNotAllowed,
			recorder.Code,
			recorder.Body.String(),
		)
	}

	if calls != 0 {
		t.Fatalf(
			"wrong method unexpectedly checked database %d times",
			calls,
		)
	}

	if allow := recorder.Header().Get("Allow"); allow != http.MethodGet {
		t.Fatalf(
			"expected Allow header %q, got %q",
			http.MethodGet,
			allow,
		)
	}

	assertJSONErrorResponse(
		t,
		recorder,
		ErrorCodeMethodNotAllowed,
		"",
	)
}

func newHealthTestRouter(
	databaseReady func(context.Context) error,
) http.Handler {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	return NewRouter(
		Dependencies{
			DatabaseReady: databaseReady,
			Logger:        logger,
		},
		[]string{"http://localhost:5173"},
	)
}
