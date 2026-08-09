package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kernelKain/timetrap/backend/internal/domain"
)

func TestValidateScenarioRoute(t *testing.T) {
	validBody := validScenarioJSON(t)

	tests := []struct {
		name               string
		method             string
		body               []byte
		expectedStatus     int
		expectedErrorCode  ErrorCode
		forbiddenResponse  string
		expectedAllowValue string
	}{
		{
			name:           "valid scenario",
			method:         http.MethodPost,
			body:           validBody,
			expectedStatus: http.StatusOK,
		},
		{
			name:              "broken JSON",
			method:            http.MethodPost,
			body:              []byte(`{"name":`),
			expectedStatus:    http.StatusBadRequest,
			expectedErrorCode: ErrorCodeInvalidJSON,
			forbiddenResponse: "unexpected EOF",
		},
		{
			name:              "unknown JSON field",
			method:            http.MethodPost,
			body:              []byte(`{"unknownField":true}`),
			expectedStatus:    http.StatusBadRequest,
			expectedErrorCode: ErrorCodeInvalidJSON,
			forbiddenResponse: `unknown field "unknownField"`,
		},
		{
			name:              "two JSON objects",
			method:            http.MethodPost,
			body:              append(append([]byte{}, validBody...), []byte("\n{}")...),
			expectedStatus:    http.StatusBadRequest,
			expectedErrorCode: ErrorCodeInvalidJSON,
		},
		{
			name:              "valid JSON with invalid model",
			method:            http.MethodPost,
			body:              []byte(`{}`),
			expectedStatus:    http.StatusUnprocessableEntity,
			expectedErrorCode: ErrorCodeValidationFailed,
		},
		{
			name:              "oversized body",
			method:            http.MethodPost,
			body:              bytes.Repeat([]byte(" "), int(maxRequestBodyBytes)+1),
			expectedStatus:    http.StatusBadRequest,
			expectedErrorCode: ErrorCodeInvalidJSON,
			forbiddenResponse: "http: request body too large",
		},
		{
			name:               "wrong method",
			method:             http.MethodGet,
			body:               nil,
			expectedStatus:     http.StatusMethodNotAllowed,
			expectedErrorCode:  ErrorCodeMethodNotAllowed,
			expectedAllowValue: http.MethodPost,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			handler := newTestRouter()

			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(
				test.method,
				"/api/v1/scenarios/validate",
				bytes.NewReader(test.body),
			)
			request.Header.Set("Content-Type", "application/json")

			handler.ServeHTTP(recorder, request)

			if recorder.Code != test.expectedStatus {
				t.Fatalf(
					"expected status %d, got %d; body=%s",
					test.expectedStatus,
					recorder.Code,
					recorder.Body.String(),
				)
			}

			if test.expectedAllowValue != "" {
				if allow := recorder.Header().Get("Allow"); allow != test.expectedAllowValue {
					t.Fatalf(
						"expected Allow header %q, got %q",
						test.expectedAllowValue,
						allow,
					)
				}
			}

			if test.expectedErrorCode == "" {
				assertValidationSuccess(t, recorder)
				return
			}

			assertJSONErrorResponse(
				t,
				recorder,
				test.expectedErrorCode,
				test.forbiddenResponse,
			)
		})
	}
}

func TestCORSAllowedOrigin(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	request.Header.Set("Origin", "http://localhost:5173")
	newTestRouter().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}
	if got := recorder.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:5173" {
		t.Fatalf("Access-Control-Allow-Origin = %q", got)
	}
	if got := recorder.Header().Get("Access-Control-Expose-Headers"); got != "X-Request-ID" {
		t.Fatalf("Access-Control-Expose-Headers = %q", got)
	}
	if !strings.Contains(recorder.Header().Get("Vary"), "Origin") {
		t.Fatalf("Vary = %q; want Origin", recorder.Header().Get("Vary"))
	}
}

func TestCORSRejectedOrigin(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	request.Header.Set("Origin", "https://untrusted.example")
	newTestRouter().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d", recorder.Code)
	}
	if got := recorder.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("unexpected Access-Control-Allow-Origin %q", got)
	}
	assertJSONErrorResponse(t, recorder, ErrorCodeOriginNotAllowed, "")
}

func TestCORSPreflight(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodOptions, "/api/v1/scenarios", nil)
	request.Header.Set("Origin", "http://localhost:5173")
	request.Header.Set("Access-Control-Request-Method", http.MethodPost)
	request.Header.Set("Access-Control-Request-Headers", "Content-Type, X-Request-ID")
	newTestRouter().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", recorder.Code)
	}
	if got := recorder.Header().Get("Access-Control-Allow-Methods"); got != "GET, POST, PUT, OPTIONS" {
		t.Fatalf("Access-Control-Allow-Methods = %q", got)
	}
	if got := recorder.Header().Get("Access-Control-Allow-Headers"); got != "Content-Type, X-Request-ID" {
		t.Fatalf("Access-Control-Allow-Headers = %q", got)
	}
}

func newTestRouter() http.Handler {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	return NewRouter(
		Dependencies{
			Analyze: func(
				domain.Scenario,
			) (domain.Analysis, error) {
				return domain.Analysis{}, nil
			},
			DatabaseReady: func(context.Context) error {
				return nil
			},
			Logger: logger,
		},
		[]string{"http://localhost:5173"},
	)
}

func validScenarioJSON(t *testing.T) []byte {
	t.Helper()

	graceMinutes := 5

	scenario := domain.Scenario{
		Name:           "Subscription cancellation leak",
		Description:    "Premium cache survives cancellation.",
		HorizonMinutes: 90,
		Objects: []domain.TimedObject{
			{
				ClientID: "subscription",
				Name:     "Premium subscription",
				Kind:     domain.ObjectKindSubscription,
				Events: []domain.Event{
					{
						Type:     domain.EventTypeIssue,
						AtMinute: 0,
					},
					{
						Type:     domain.EventTypeRevoke,
						AtMinute: 10,
					},
				},
			},
			{
				ClientID: "premium-cache",
				Name:     "Premium entitlement cache",
				Kind:     domain.ObjectKindCachedEntitlement,
				Events: []domain.Event{
					{
						Type:     domain.EventTypeIssue,
						AtMinute: 0,
					},
					{
						Type:     domain.EventTypeExpire,
						AtMinute: 60,
					},
				},
			},
		},
		Invariants: []domain.Invariant{
			{
				Type:              domain.InvariantTypeRevokedAccessGrace,
				SourceObjectID:    "subscription",
				DependentObjectID: "premium-cache",
				GraceMinutes:      &graceMinutes,
			},
		},
	}

	body, err := json.Marshal(scenario)
	if err != nil {
		t.Fatalf("failed to encode valid scenario fixture: %v", err)
	}

	return body
}

func assertValidationSuccess(
	t *testing.T,
	recorder *httptest.ResponseRecorder,
) {
	t.Helper()

	if contentType := recorder.Header().Get("Content-Type"); contentType != "application/json" {
		t.Fatalf(
			"expected application/json Content-Type, got %q",
			contentType,
		)
	}

	var response validationSuccessResponse

	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode success response: %v", err)
	}

	if !response.Valid {
		t.Fatal("expected valid=true")
	}

	if response.RequestID == "" {
		t.Fatal("expected request ID in success response")
	}

	if headerRequestID := recorder.Header().Get("X-Request-ID"); headerRequestID != response.RequestID {
		t.Fatalf(
			"expected matching request IDs, header=%q body=%q",
			headerRequestID,
			response.RequestID,
		)
	}
}

func assertJSONErrorResponse(
	t *testing.T,
	recorder *httptest.ResponseRecorder,
	expectedCode ErrorCode,
	forbiddenText string,
) {
	t.Helper()

	if contentType := recorder.Header().Get("Content-Type"); contentType != "application/json" {
		t.Fatalf(
			"expected application/json Content-Type, got %q",
			contentType,
		)
	}

	var response errorResponse

	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}

	if response.Error.Code != expectedCode {
		t.Fatalf(
			"expected error code %q, got %q",
			expectedCode,
			response.Error.Code,
		)
	}

	if response.RequestID == "" {
		t.Fatal("expected request ID in error response")
	}

	if headerRequestID := recorder.Header().Get("X-Request-ID"); headerRequestID != response.RequestID {
		t.Fatalf(
			"expected matching request IDs, header=%q body=%q",
			headerRequestID,
			response.RequestID,
		)
	}

	if forbiddenText != "" &&
		strings.Contains(recorder.Body.String(), forbiddenText) {
		t.Fatalf(
			"response exposed internal detail %q: %s",
			forbiddenText,
			recorder.Body.String(),
		)
	}
}
