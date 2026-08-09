package httpapi

import (
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/kernelKain/timetrap/backend/internal/domain"
	"github.com/kernelKain/timetrap/backend/internal/store"
)

type analysisResponse struct {
	Analysis  analysisResource `json:"analysis"`
	RequestID string           `json:"requestId"`
}

type analysisResource struct {
	ID               uuid.UUID        `json:"id"`
	ScenarioID       uuid.UUID        `json:"scenarioId"`
	ScenarioSnapshot *domain.Scenario `json:"scenarioSnapshot,omitempty"`
	Result           domain.Analysis  `json:"result"`
	EngineVersion    string           `json:"engineVersion"`
	CreatedAt        time.Time        `json:"createdAt"`
}

func createAnalysisHandler(
	dependencies Dependencies,
) http.HandlerFunc {
	return func(
		responseWriter http.ResponseWriter,
		request *http.Request,
	) {
		if request.Method != http.MethodPost {
			responseWriter.Header().Set("Allow", http.MethodPost)

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

		scenarioID, err := uuid.Parse(
			request.PathValue("scenarioId"),
		)
		if err != nil {
			respondError(
				dependencies.Logger,
				responseWriter,
				request,
				http.StatusBadRequest,
				ErrorCodeInvalidID,
				"Scenario ID must be a valid UUID.",
				nil,
			)
			return
		}

		scenarioRecord, err := dependencies.Scenarios.GetScenario(
			request.Context(),
			scenarioID,
		)
		if errors.Is(err, store.ErrNotFound) {
			respondError(
				dependencies.Logger,
				responseWriter,
				request,
				http.StatusNotFound,
				ErrorCodeScenarioNotFound,
				"Scenario not found.",
				nil,
			)
			return
		}
		if err != nil {
			respondAnalysisStorageError(
				dependencies.Logger,
				responseWriter,
				request,
				"load scenario",
				err,
			)
			return
		}

		analysis, err := dependencies.Analyze(
			scenarioRecord.Scenario,
		)
		if err != nil {
			var validationErrors domain.ValidationErrors
			if errors.As(err, &validationErrors) {
				dependencies.Logger.Warn(
					"Stored scenario cannot be analyzed",
					"request_id",
					requestIDFromContext(request.Context()),
					"scenario_id",
					scenarioID,
				)

				respondError(
					dependencies.Logger,
					responseWriter,
					request,
					http.StatusConflict,
					ErrorCodeScenarioNotAnalyzable,
					"Stored scenario cannot be analyzed.",
					validationErrors,
				)
				return
			}

			dependencies.Logger.Error(
				"Scenario analysis failed",
				"request_id",
				requestIDFromContext(request.Context()),
				"scenario_id",
				scenarioID,
				"error",
				err,
			)

			respondError(
				dependencies.Logger,
				responseWriter,
				request,
				http.StatusInternalServerError,
				ErrorCodeAnalysisFailed,
				"Unable to analyze the scenario.",
				nil,
			)
			return
		}

		analysisRecord, err := dependencies.Analyses.SaveAnalysis(
			request.Context(),
			scenarioID,
			scenarioRecord.Scenario,
			analysis,
		)
		if errors.Is(err, store.ErrNotFound) {
			// The scenario may have been deleted after it was loaded but
			// before the analysis insert reached the foreign-key check.
			respondError(
				dependencies.Logger,
				responseWriter,
				request,
				http.StatusNotFound,
				ErrorCodeScenarioNotFound,
				"Scenario not found.",
				nil,
			)
			return
		}
		if err != nil {
			respondAnalysisStorageError(
				dependencies.Logger,
				responseWriter,
				request,
				"save analysis",
				err,
			)
			return
		}

		writeAnalysisResponse(
			dependencies.Logger,
			responseWriter,
			request,
			http.StatusCreated,
			analysisRecord,
			false,
		)
	}
}

func getAnalysisHandler(
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

		analysisID, err := uuid.Parse(
			request.PathValue("analysisId"),
		)
		if err != nil {
			respondError(
				dependencies.Logger,
				responseWriter,
				request,
				http.StatusBadRequest,
				ErrorCodeInvalidID,
				"Analysis ID must be a valid UUID.",
				nil,
			)
			return
		}

		record, err := dependencies.Analyses.GetAnalysis(
			request.Context(),
			analysisID,
		)
		if errors.Is(err, store.ErrNotFound) {
			respondError(
				dependencies.Logger,
				responseWriter,
				request,
				http.StatusNotFound,
				ErrorCodeAnalysisNotFound,
				"Analysis not found.",
				nil,
			)
			return
		}
		if err != nil {
			respondAnalysisStorageError(
				dependencies.Logger,
				responseWriter,
				request,
				"get analysis",
				err,
			)
			return
		}

		writeAnalysisResponse(
			dependencies.Logger,
			responseWriter,
			request,
			http.StatusOK,
			record,
			true,
		)
	}
}

func writeAnalysisResponse(
	logger *slog.Logger,
	responseWriter http.ResponseWriter,
	request *http.Request,
	status int,
	record store.AnalysisRecord,
	includeSnapshot bool,
) {
	resource := analysisResource{
		ID:            record.ID,
		ScenarioID:    record.ScenarioID,
		Result:        record.Analysis,
		EngineVersion: record.EngineVersion,
		CreatedAt:     record.CreatedAt,
	}

	if includeSnapshot {
		snapshot := record.ScenarioSnapshot
		resource.ScenarioSnapshot = &snapshot
	}

	response := analysisResponse{
		Analysis:  resource,
		RequestID: requestIDFromContext(request.Context()),
	}

	if err := writeJSON(responseWriter, status, response); err != nil {
		logger.Error(
			"Failed to write analysis response",
			"request_id",
			requestIDFromContext(request.Context()),
			"error",
			err,
		)
	}
}

func respondAnalysisStorageError(
	logger *slog.Logger,
	responseWriter http.ResponseWriter,
	request *http.Request,
	operation string,
	err error,
) {
	logger.Error(
		"Analysis storage operation failed",
		"request_id",
		requestIDFromContext(request.Context()),
		"operation",
		operation,
		"error",
		err,
	)

	respondError(
		logger,
		responseWriter,
		request,
		http.StatusInternalServerError,
		ErrorCodeStorage,
		"Unable to complete analysis storage operation.",
		nil,
	)
}
