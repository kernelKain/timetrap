package httpapi

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"

	"github.com/kernelKain/timetrap/backend/internal/domain"
)

// ErrorCode is a stable machine-readable API error identifier.
type ErrorCode string

const (
	ErrorCodeInvalidJSON      ErrorCode = "INVALID_JSON"
	ErrorCodeValidationFailed ErrorCode = "VALIDATION_FAILED"
	ErrorCodeMethodNotAllowed ErrorCode = "METHOD_NOT_ALLOWED"
	ErrorCodeNotFound         ErrorCode = "NOT_FOUND"
	ErrorCodeInternal         ErrorCode = "INTERNAL_ERROR"
	ErrorCodeOriginNotAllowed ErrorCode = "ORIGIN_NOT_ALLOWED"
)

type errorResponse struct {
	Error     errorBody `json:"error"`
	RequestID string    `json:"requestId"`
}

type errorBody struct {
	Code    ErrorCode               `json:"code"`
	Message string                  `json:"message"`
	Fields  domain.ValidationErrors `json:"fields,omitempty"`
}

// requestIDContextKey is shared with the request-ID middleware.
//
// The later middleware must store the generated request ID using this exact
// key instead of defining another context key.
type requestIDContextKey struct{}

var requestIDKey requestIDContextKey

// writeJSON writes one JSON response.
//
// It returns encoding or response-writer errors so the caller can log internal
// details separately. Once the status and body begin writing, callers must not
// attempt to send a second response.
func writeJSON(
	responseWriter http.ResponseWriter,
	status int,
	value any,
) error {
	responseWriter.Header().Set("Content-Type", "application/json")
	responseWriter.WriteHeader(status)

	return json.NewEncoder(responseWriter).Encode(value)
}

// writeError writes a safe API error envelope.
//
// message must be safe for clients. Callers must never pass a raw internal
// error, panic value, stack trace, database error or credential-bearing value.
func writeError(
	responseWriter http.ResponseWriter,
	request *http.Request,
	status int,
	code ErrorCode,
	message string,
	fields domain.ValidationErrors,
) error {
	requestID := requestIDFromContext(request.Context())

	// Middleware will normally provide the request ID. This fallback ensures
	// errors still have an identifier if the helper is used without middleware.
	if requestID == "" {
		requestID = newRequestID()
	}

	responseWriter.Header().Set("X-Request-ID", requestID)

	return writeJSON(responseWriter, status, errorResponse{
		Error: errorBody{
			Code:    code,
			Message: message,
			Fields:  fields,
		},
		RequestID: requestID,
	})
}

func requestIDFromContext(context context.Context) string {
	requestID, _ := context.Value(requestIDKey).(string)
	return requestID
}

func newRequestID() string {
	randomBytes := make([]byte, 16)

	if _, err := rand.Read(randomBytes); err != nil {
		// This fallback deliberately contains no internal error information.
		return "request-id-unavailable"
	}

	return hex.EncodeToString(randomBytes)
}
