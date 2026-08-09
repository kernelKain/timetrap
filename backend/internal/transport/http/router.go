package httpapi

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/kernelKain/timetrap/backend/internal/domain"
	"github.com/kernelKain/timetrap/backend/internal/store"
)

const maxRequestBodyBytes int64 = 1 << 20

// Dependencies contains application services used by HTTP handlers.
type Dependencies struct {
	Scenarios     store.ScenarioRepository
	Analyses      store.AnalysisRepository
	Analyze       func(domain.Scenario) (domain.Analysis, error)
	DatabaseReady func(context.Context) error
	Logger        *slog.Logger
}

type healthResponse struct {
	Status    string `json:"status"`
	Service   string `json:"service"`
	Database  string `json:"database"`
	Timestamp string `json:"timestamp"`
}

type validationSuccessResponse struct {
	Valid     bool   `json:"valid"`
	RequestID string `json:"requestId"`
}

// NewRouter constructs the complete TimeTrap HTTP handler.
//
// Application dependencies and configuration are supplied by main so
// environment reading, concrete storage and analyzer selection remain outside
// the transport package.
func NewRouter(
	dependencies Dependencies,
	allowedOrigins []string,
) http.Handler {
	logger := dependencies.Logger
	mux := http.NewServeMux()

	mux.HandleFunc(
		"/api/v1/health",
		healthHandler(dependencies),
	)

	mux.HandleFunc(
		"/api/v1/scenarios",
		scenarioCollectionHandler(dependencies),
	)

	mux.HandleFunc(
		"/api/v1/scenarios/validate",
		validateScenarioHandler(logger),
	)

	mux.HandleFunc(
		"/api/v1/scenarios/{scenarioId}/analyses",
		createAnalysisHandler(dependencies),
	)

	mux.HandleFunc(
		"/api/v1/scenarios/{scenarioId}",
		scenarioResourceHandler(dependencies),
	)

	mux.HandleFunc(
		"/api/v1/analyses/{analysisId}",
		getAnalysisHandler(dependencies),
	)

	mux.HandleFunc(
		"/",
		notFoundHandler(logger),
	)

	var handler http.Handler = mux

	handler = corsMiddleware(logger, allowedOrigins, handler)
	handler = Recovery(logger)(handler)
	handler = RequestLogger(logger)(handler)
	handler = RequestID(handler)

	return handler
}

func healthHandler(
	dependencies Dependencies,
) http.HandlerFunc {
	return func(
		responseWriter http.ResponseWriter,
		request *http.Request,
	) {
		if request.Method != http.MethodGet {
			responseWriter.Header().Set("Allow", http.MethodGet)

			respondError(
				dependencies.Logger,
				responseWriter,
				request,
				http.StatusMethodNotAllowed,
				ErrorCodeMethodNotAllowed,
				"Method not allowed.",
				nil,
			)
			return
		}

		healthContext, cancel := context.WithTimeout(
			request.Context(),
			time.Second,
		)
		defer cancel()

		statusCode := http.StatusOK
		status := "ok"
		databaseStatus := "connected"

		if dependencies.DatabaseReady == nil ||
			dependencies.DatabaseReady(healthContext) != nil {
			statusCode = http.StatusServiceUnavailable
			status = "degraded"
			databaseStatus = "unavailable"

			// Deliberately omit the database error. It may contain internal
			// addressing or identity details.
			dependencies.Logger.Warn(
				"Database health check failed",
				"request_id",
				requestIDFromContext(request.Context()),
			)
		}

		response := healthResponse{
			Status:    status,
			Service:   "timetrap-api",
			Database:  databaseStatus,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		}

		if err := writeJSON(
			responseWriter,
			statusCode,
			response,
		); err != nil {
			dependencies.Logger.Error(
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

		_, accepted := validatedScenarioFromRequest(
			logger,
			responseWriter,
			request,
		)
		if !accepted {
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
	allowedOriginSet := make(
		map[string]struct{},
		len(allowedOrigins),
	)

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
				"GET, POST, PUT, OPTIONS",
			)

			responseWriter.Header().Set(
				"Access-Control-Allow-Headers",
				"Content-Type",
			)

			responseWriter.Header().Set(
				"Access-Control-Expose-Headers",
				"X-Request-ID",
			)

			responseWriter.Header().Add(
				"Vary",
				"Origin",
			)
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
