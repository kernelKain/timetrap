package httpapi

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRequestID(t *testing.T) {
	var contextRequestID string

	next := http.HandlerFunc(func(
		responseWriter http.ResponseWriter,
		request *http.Request,
	) {
		contextRequestID = requestIDFromContext(request.Context())
		responseWriter.WriteHeader(http.StatusNoContent)
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/health", nil)

	RequestID(next).ServeHTTP(recorder, request)

	headerRequestID := recorder.Header().Get("X-Request-ID")

	if headerRequestID == "" {
		t.Fatal("expected X-Request-ID response header")
	}

	if contextRequestID == "" {
		t.Fatal("expected request ID in request context")
	}

	if headerRequestID != contextRequestID {
		t.Fatalf(
			"expected matching request IDs, header=%q context=%q",
			headerRequestID,
			contextRequestID,
		)
	}

	if len(headerRequestID) != 32 {
		t.Fatalf(
			"expected a 32-character hexadecimal request ID, got %q",
			headerRequestID,
		)
	}

	if _, err := hex.DecodeString(headerRequestID); err != nil {
		t.Fatalf("request ID is not hexadecimal: %v", err)
	}
}

func TestRequestLogger(t *testing.T) {
	var logBuffer bytes.Buffer

	logger := slog.New(slog.NewJSONHandler(&logBuffer, nil))

	next := http.HandlerFunc(func(
		responseWriter http.ResponseWriter,
		_ *http.Request,
	) {
		responseWriter.WriteHeader(http.StatusAccepted)
	})

	handler := RequestID(RequestLogger(logger)(next))

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/scenarios/validate",
		nil,
	)
	request.RemoteAddr = "192.0.2.1:1234"

	handler.ServeHTTP(recorder, request)

	var logEntry map[string]any

	if err := json.Unmarshal(
		bytes.TrimSpace(logBuffer.Bytes()),
		&logEntry,
	); err != nil {
		t.Fatalf("failed to decode structured log: %v", err)
	}

	if logEntry["request_id"] == "" {
		t.Fatal("expected request_id in structured log")
	}

	if logEntry["method"] != http.MethodPost {
		t.Fatalf(
			"expected method %q, got %#v",
			http.MethodPost,
			logEntry["method"],
		)
	}

	if logEntry["path"] != "/api/v1/scenarios/validate" {
		t.Fatalf("unexpected path %#v", logEntry["path"])
	}

	if logEntry["status"] != float64(http.StatusAccepted) {
		t.Fatalf(
			"expected status %d, got %#v",
			http.StatusAccepted,
			logEntry["status"],
		)
	}

	if _, exists := logEntry["duration_ms"]; !exists {
		t.Fatal("expected duration_ms in structured log")
	}

	if logEntry["remote_addr"] != "192.0.2.1:1234" {
		t.Fatalf(
			"unexpected remote_addr %#v",
			logEntry["remote_addr"],
		)
	}
}

func TestRequestLoggerRecordsImplicitOK(t *testing.T) {
	var logBuffer bytes.Buffer

	logger := slog.New(slog.NewJSONHandler(&logBuffer, nil))

	next := http.HandlerFunc(func(
		responseWriter http.ResponseWriter,
		_ *http.Request,
	) {
		if _, err := responseWriter.Write([]byte("ok")); err != nil {
			t.Fatalf("failed to write test response: %v", err)
		}
	})

	handler := RequestID(RequestLogger(logger)(next))

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/implicit", nil)

	handler.ServeHTTP(recorder, request)

	var logEntry map[string]any

	if err := json.Unmarshal(
		bytes.TrimSpace(logBuffer.Bytes()),
		&logEntry,
	); err != nil {
		t.Fatalf("failed to decode structured log: %v", err)
	}

	if logEntry["status"] != float64(http.StatusOK) {
		t.Fatalf(
			"expected status %d, got %#v",
			http.StatusOK,
			logEntry["status"],
		)
	}
}

func TestRecovery(t *testing.T) {
	var logBuffer bytes.Buffer

	logger := slog.New(slog.NewJSONHandler(&logBuffer, nil))

	const panicMessage = "sensitive panic detail"

	next := http.HandlerFunc(func(
		http.ResponseWriter,
		*http.Request,
	) {
		panic(panicMessage)
	})

	handler := RequestID(
		RequestLogger(logger)(
			Recovery(logger)(next),
		),
	)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(
		http.MethodGet,
		"/panic",
		nil,
	)

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusInternalServerError,
			recorder.Code,
		)
	}

	if contentType := recorder.Header().Get("Content-Type"); contentType != "application/json" {
		t.Fatalf(
			"expected application/json Content-Type, got %q",
			contentType,
		)
	}

	if strings.Contains(recorder.Body.String(), panicMessage) {
		t.Fatal("panic detail was exposed in the client response")
	}

	var response errorResponse

	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode recovery response: %v", err)
	}

	if response.Error.Code != ErrorCodeInternal {
		t.Fatalf(
			"expected error code %q, got %q",
			ErrorCodeInternal,
			response.Error.Code,
		)
	}

	if response.RequestID == "" {
		t.Fatal("expected request ID in recovery response")
	}

	if headerRequestID := recorder.Header().Get("X-Request-ID"); headerRequestID != response.RequestID {
		t.Fatalf(
			"expected matching request IDs, header=%q body=%q",
			headerRequestID,
			response.RequestID,
		)
	}

	logOutput := logBuffer.String()

	if !strings.Contains(logOutput, "HTTP handler panic") {
		t.Fatal("expected internal panic log")
	}

	if !strings.Contains(logOutput, response.RequestID) {
		t.Fatal("expected request ID in panic and request logs")
	}

	if !strings.Contains(logOutput, `"status":500`) {
		t.Fatal("expected request log to record status 500")
	}
}
