package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/kernelKain/timetrap/backend/internal/domain"
	"github.com/kernelKain/timetrap/backend/internal/store"
)

type scenarioResponse struct {
	Scenario  scenarioResource `json:"scenario"`
	RequestID string           `json:"requestId"`
}

type scenarioResource struct {
	ID         uuid.UUID       `json:"id"`
	Definition domain.Scenario `json:"definition"`
	CreatedAt  time.Time       `json:"createdAt"`
	UpdatedAt  time.Time       `json:"updatedAt"`
}

func scenarioCollectionHandler(
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

		createScenario(
			dependencies,
			responseWriter,
			request,
		)
	}
}

func scenarioResourceHandler(
	dependencies Dependencies,
) http.HandlerFunc {
	return func(
		responseWriter http.ResponseWriter,
		request *http.Request,
	) {
		switch request.Method {
		case http.MethodGet, http.MethodPut:
			// Continue below.
		default:
			responseWriter.Header().Set(
				"Allow",
				http.MethodGet+", "+http.MethodPut,
			)

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

		if request.Method == http.MethodGet {
			getScenario(
				dependencies,
				responseWriter,
				request,
				scenarioID,
			)
			return
		}

		updateScenario(
			dependencies,
			responseWriter,
			request,
			scenarioID,
		)
	}
}

func createScenario(
	dependencies Dependencies,
	responseWriter http.ResponseWriter,
	request *http.Request,
) {
	scenario, accepted := validatedScenarioFromRequest(
		dependencies.Logger,
		responseWriter,
		request,
	)
	if !accepted {
		return
	}

	record, err := dependencies.Scenarios.CreateScenario(
		request.Context(),
		scenario,
	)
	if err != nil {
		respondScenarioStorageError(
			dependencies.Logger,
			responseWriter,
			request,
			"create",
			err,
		)
		return
	}

	writeScenarioResponse(
		dependencies.Logger,
		responseWriter,
		request,
		http.StatusCreated,
		record,
	)
}

func getScenario(
	dependencies Dependencies,
	responseWriter http.ResponseWriter,
	request *http.Request,
	scenarioID uuid.UUID,
) {
	record, err := dependencies.Scenarios.GetScenario(
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
		respondScenarioStorageError(
			dependencies.Logger,
			responseWriter,
			request,
			"get",
			err,
		)
		return
	}

	writeScenarioResponse(
		dependencies.Logger,
		responseWriter,
		request,
		http.StatusOK,
		record,
	)
}

func updateScenario(
	dependencies Dependencies,
	responseWriter http.ResponseWriter,
	request *http.Request,
	scenarioID uuid.UUID,
) {
	scenario, accepted := validatedScenarioFromRequest(
		dependencies.Logger,
		responseWriter,
		request,
	)
	if !accepted {
		return
	}

	record, err := dependencies.Scenarios.UpdateScenario(
		request.Context(),
		scenarioID,
		scenario,
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
		respondScenarioStorageError(
			dependencies.Logger,
			responseWriter,
			request,
			"update",
			err,
		)
		return
	}

	writeScenarioResponse(
		dependencies.Logger,
		responseWriter,
		request,
		http.StatusOK,
		record,
	)
}

func validatedScenarioFromRequest(
	logger *slog.Logger,
	responseWriter http.ResponseWriter,
	request *http.Request,
) (domain.Scenario, bool) {
	scenario, message, err := decodeScenarioRequest(
		responseWriter,
		request,
	)
	if err != nil {
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

		return domain.Scenario{}, false
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

		return domain.Scenario{}, false
	}

	return scenario, true
}

func decodeScenarioRequest(
	responseWriter http.ResponseWriter,
	request *http.Request,
) (domain.Scenario, string, error) {
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

		return domain.Scenario{}, message, err
	}

	trailingError := decoder.Decode(&struct{}{})
	if errors.Is(trailingError, io.EOF) {
		return scenario, "", nil
	}

	if trailingError == nil {
		trailingError = errors.New(
			"request body contains more than one JSON object",
		)
	}

	return domain.Scenario{},
		"Request body must contain exactly one JSON object.",
		trailingError
}

func writeScenarioResponse(
	logger *slog.Logger,
	responseWriter http.ResponseWriter,
	request *http.Request,
	status int,
	record store.ScenarioRecord,
) {
	response := scenarioResponse{
		Scenario: scenarioResource{
			ID:         record.ID,
			Definition: record.Scenario,
			CreatedAt:  record.CreatedAt,
			UpdatedAt:  record.UpdatedAt,
		},
		RequestID: requestIDFromContext(request.Context()),
	}

	if err := writeJSON(responseWriter, status, response); err != nil {
		logger.Error(
			"Failed to write scenario response",
			"request_id",
			requestIDFromContext(request.Context()),
			"error",
			err,
		)
	}
}

func respondScenarioStorageError(
	logger *slog.Logger,
	responseWriter http.ResponseWriter,
	request *http.Request,
	operation string,
	err error,
) {
	logger.Error(
		"Scenario storage operation failed",
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
		"Unable to complete scenario storage operation.",
		nil,
	)
}
