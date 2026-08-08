package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kernelKain/timetrap/backend/internal/domain"
)

func TestWriteJSON(t *testing.T) {
	recorder := httptest.NewRecorder()

	payload := struct {
		Valid bool `json:"valid"`
	}{
		Valid: true,
	}

	if err := writeJSON(recorder, http.StatusOK, payload); err != nil {
		t.Fatalf("writeJSON returned an error: %v", err)
	}

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusOK,
			recorder.Code,
		)
	}

	if contentType := recorder.Header().Get("Content-Type"); contentType != "application/json" {
		t.Fatalf(
			"expected Content-Type application/json, got %q",
			contentType,
		)
	}

	var response struct {
		Valid bool `json:"valid"`
	}

	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !response.Valid {
		t.Fatal("expected response valid field to be true")
	}
}

func TestWriteErrorUsesRequestIDFromContext(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/scenarios/validate",
		nil,
	)

	const requestID = "request-123"

	request = request.WithContext(
		context.WithValue(request.Context(), requestIDKey, requestID),
	)

	fields := domain.ValidationErrors{
		{
			Field:   "objects[0].events[1].atMinute",
			Message: "must be between 0 and horizonMinutes",
		},
	}

	if err := writeError(
		recorder,
		request,
		http.StatusUnprocessableEntity,
		ErrorCodeValidationFailed,
		"Scenario validation failed.",
		fields,
	); err != nil {
		t.Fatalf("writeError returned an error: %v", err)
	}

	if recorder.Code != http.StatusUnprocessableEntity {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusUnprocessableEntity,
			recorder.Code,
		)
	}

	if headerRequestID := recorder.Header().Get("X-Request-ID"); headerRequestID != requestID {
		t.Fatalf(
			"expected X-Request-ID %q, got %q",
			requestID,
			headerRequestID,
		)
	}

	var response errorResponse

	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}

	if response.RequestID != requestID {
		t.Fatalf(
			"expected response requestId %q, got %q",
			requestID,
			response.RequestID,
		)
	}

	if response.Error.Code != ErrorCodeValidationFailed {
		t.Fatalf(
			"expected error code %q, got %q",
			ErrorCodeValidationFailed,
			response.Error.Code,
		)
	}

	if response.Error.Message != "Scenario validation failed." {
		t.Fatalf(
			"unexpected error message %q",
			response.Error.Message,
		)
	}

	if len(response.Error.Fields) != 1 {
		t.Fatalf(
			"expected one field error, got %d",
			len(response.Error.Fields),
		)
	}

	if response.Error.Fields[0] != fields[0] {
		t.Fatalf(
			"expected field error %#v, got %#v",
			fields[0],
			response.Error.Fields[0],
		)
	}
}

func TestWriteErrorGeneratesFallbackRequestID(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(
		http.MethodGet,
		"/missing",
		nil,
	)

	if err := writeError(
		recorder,
		request,
		http.StatusNotFound,
		ErrorCodeNotFound,
		"Resource not found.",
		nil,
	); err != nil {
		t.Fatalf("writeError returned an error: %v", err)
	}

	var response errorResponse

	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}

	if response.RequestID == "" {
		t.Fatal("expected a generated request ID")
	}

	if headerRequestID := recorder.Header().Get("X-Request-ID"); headerRequestID != response.RequestID {
		t.Fatalf(
			"expected matching request IDs, header=%q body=%q",
			headerRequestID,
			response.RequestID,
		)
	}
}
