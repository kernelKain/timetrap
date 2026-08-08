package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/kernelKain/timetrap/backend/internal/domain"
)

const maxRequestBodyBytes int64 = 1 << 20

type healthResponse struct {
	Status    string `json:"status"`
	Service   string `json:"service"`
	Version   string `json:"version"`
	Database  string `json:"database"`
	Timestamp string `json:"timestamp"`
}

type validationSuccessResponse struct {
	Valid     bool   `json:"valid"`
	RequestID string `json:"requestId"`
}

// NewRouter constructs the complete Phase 2 HTTP handler.
//
// Configuration values are supplied by main so environment-variable reading
// remains outside the transport package.
func NewRouter(
	logger *slog.Logger,
	allowedOrigins []string,
	version string,
) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/v1/health", healthHandler(logger, version))
	mux.HandleFunc(
		"/api/v1/scenarios/validate",
		validateScenarioHandler(logger),
	)
	mux.HandleFunc("/", notFoundHandler(logger))

	var handler http.Handler = mux

	handler = corsMiddleware(logger, allowedOrigins, handler)
	handler = Recovery(logger)(handler)
	handler = RequestLogger(logger)(handler)
	handler = RequestID(handler)

	return handler
}

func healthHandler(
	logger *slog.Logger,
	version string,
) http.HandlerFunc {
	return func(
		responseWriter http.ResponseWriter,
		request *http.Request,
	) {
		if request.Method != http.MethodGet {
			responseWriter.Header().Set("Allow", http.MethodGet)

			respondError(
				logger,
				responseWriter,
				request,
				http.StatusMethodNotAllowed,
				ErrorCodeMethodNotAllowed,
				"Method not allowed.",
				nil,
			)
			return
		}

		response := healthResponse{
			Status:    "ok",
			Service:   "timetrap-api",
			Version:   version,
			Database:  "not_configured",
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		}

		if err := writeJSON(
			responseWriter,
			http.StatusOK,
			response,
		); err != nil {
			logger.Error(
				"Failed to write health response",
				"request_id",
				requestIDFromContext(request.Context()),
				"error",
				err,
			)
		}
	}
}

func validateScenarioHandler(
	logger *slog.Logger,
) http.HandlerFunc {
	return func(
		responseWriter http.ResponseWriter,
		request *http.Request,
	) {
		if request.Method != http.MethodPost {
			responseWriter.Header().Set("Allow", http.MethodPost)

			respondError(
				logger,
				responseWriter,
				request,
				http.StatusMethodNotAllowed,
				ErrorCodeMethodNotAllowed,
				"Method not allowed.",
				nil,
			)
			return
		}

		request.Body = http.MaxBytesReader(
			responseWriter,
			request.Body,
			maxRequestBodyBytes,
		)

		decoder := json.NewDecoder(request.Body)
		decoder.DisallowUnknownFields()

		var scenario domain.Scenario

		if err := decoder.Decode(&scenario); err != nil {
			message := "Request body must contain one valid JSON object."

			var maxBytesError *http.MaxBytesError
			if errors.As(err, &maxBytesError) {
				message = "Request body must not exceed 1 MiB."
			}

			logger.Debug(
				"Scenario JSON rejected",
				"request_id",
				requestIDFromContext(request.Context()),
				"error",
				err,
			)

			respondError(
				logger,
				responseWriter,
				request,
				http.StatusBadRequest,
				ErrorCodeInvalidJSON,
				message,
				nil,
			)
			return
		}

		// A second decode must reach EOF. Any other result means the request
		// contains trailing data or more than one JSON object.
		if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
			logger.Debug(
				"Scenario JSON contains trailing data",
				"request_id",
				requestIDFromContext(request.Context()),
				"error",
				err,
			)

			respondError(
				logger,
				responseWriter,
				request,
				http.StatusBadRequest,
				ErrorCodeInvalidJSON,
				"Request body must contain exactly one JSON object.",
				nil,
			)
			return
		}

		validationErrors := scenario.Validate()
		if !validationErrors.Empty() {
			respondError(
				logger,
				responseWriter,
				request,
				http.StatusUnprocessableEntity,
				ErrorCodeValidationFailed,
				"Scenario validation failed.",
				validationErrors,
			)
			return
		}

		response := validationSuccessResponse{
			Valid:     true,
			RequestID: requestIDFromContext(request.Context()),
		}

		if err := writeJSON(
			responseWriter,
			http.StatusOK,
			response,
		); err != nil {
			logger.Error(
				"Failed to write validation response",
				"request_id",
				requestIDFromContext(request.Context()),
				"error",
				err,
			)
		}
	}
}

func notFoundHandler(logger *slog.Logger) http.HandlerFunc {
	return func(
		responseWriter http.ResponseWriter,
		request *http.Request,
	) {
		respondError(
			logger,
			responseWriter,
			request,
			http.StatusNotFound,
			ErrorCodeNotFound,
			"Resource not found.",
			nil,
		)
	}
}

func corsMiddleware(
	logger *slog.Logger,
	allowedOrigins []string,
	next http.Handler,
) http.Handler {
	allowedOriginSet := make(map[string]struct{}, len(allowedOrigins))

	for _, origin := range allowedOrigins {
		origin = strings.TrimSpace(origin)
		if origin != "" {
			allowedOriginSet[origin] = struct{}{}
		}
	}

	return http.HandlerFunc(func(
		responseWriter http.ResponseWriter,
		request *http.Request,
	) {
		origin := request.Header.Get("Origin")

		if origin != "" {
			if _, allowed := allowedOriginSet[origin]; !allowed {
				respondError(
					logger,
					responseWriter,
					request,
					http.StatusForbidden,
					ErrorCodeOriginNotAllowed,
					"Origin is not allowed.",
					nil,
				)
				return
			}

			responseWriter.Header().Set(
				"Access-Control-Allow-Origin",
				origin,
			)
			responseWriter.Header().Set(
				"Access-Control-Allow-Methods",
				"GET, POST, OPTIONS",
			)
			responseWriter.Header().Set(
				"Access-Control-Allow-Headers",
				"Content-Type",
			)
			responseWriter.Header().Set(
				"Access-Control-Expose-Headers",
				"X-Request-ID",
			)
			responseWriter.Header().Add("Vary", "Origin")
		}

		if request.Method == http.MethodOptions {
			responseWriter.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(responseWriter, request)
	})
}

func respondError(
	logger *slog.Logger,
	responseWriter http.ResponseWriter,
	request *http.Request,
	status int,
	code ErrorCode,
	message string,
	fields domain.ValidationErrors,
) {
	if err := writeError(
		responseWriter,
		request,
		status,
		code,
		message,
		fields,
	); err != nil {
		logger.Error(
			"Failed to write error response",
			"request_id",
			requestIDFromContext(request.Context()),
			"error",
			err,
		)
	}
}
