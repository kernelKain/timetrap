package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kernelKain/timetrap/backend/internal/domain"
	"github.com/kernelKain/timetrap/backend/internal/store"
)

type fakeScenarioRepository struct {
	createCalls  int
	createInput  domain.Scenario
	createRecord store.ScenarioRecord
	createError  error
	getCalls     int
	getID        uuid.UUID
	getRecord    store.ScenarioRecord
	getError     error
	updateCalls  int
	updateID     uuid.UUID
	updateInput  domain.Scenario
	updateRecord store.ScenarioRecord
	updateError  error
}

func (repository *fakeScenarioRepository) CreateScenario(
	_ context.Context,
	scenario domain.Scenario,
) (store.ScenarioRecord, error) {
	repository.createCalls++
	repository.createInput = scenario

	return repository.createRecord, repository.createError
}

func (repository *fakeScenarioRepository) GetScenario(
	_ context.Context,
	id uuid.UUID,
) (store.ScenarioRecord, error) {
	repository.getCalls++
	repository.getID = id

	return repository.getRecord, repository.getError
}

func (repository *fakeScenarioRepository) UpdateScenario(
	_ context.Context,
	id uuid.UUID,
	scenario domain.Scenario,
) (store.ScenarioRecord, error) {
	repository.updateCalls++
	repository.updateID = id
	repository.updateInput = scenario

	return repository.updateRecord, repository.updateError
}

func TestCreateScenarioRoute(t *testing.T) {
	scenario := validScenarioValue(t)
	scenarioID := uuid.New()
	createdAt := time.Date(
		2026,
		time.August,
		9,
		9,
		0,
		0,
		0,
		time.UTC,
	)
	updatedAt := createdAt.Add(time.Minute)

	repository := &fakeScenarioRepository{
		createRecord: store.ScenarioRecord{
			ID:        scenarioID,
			Scenario:  scenario,
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
		},
	}

	handler := newScenarioTestRouter(repository)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/scenarios",
		bytes.NewReader(validScenarioJSON(t)),
	)
	request.Header.Set("Content-Type", "application/json")

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf(
			"expected status %d, got %d; body=%s",
			http.StatusCreated,
			recorder.Code,
			recorder.Body.String(),
		)
	}

	if repository.createCalls != 1 {
		t.Fatalf(
			"expected one CreateScenario call, got %d",
			repository.createCalls,
		)
	}

	if !reflect.DeepEqual(repository.createInput, scenario) {
		t.Fatalf(
			"repository received unexpected scenario: %#v",
			repository.createInput,
		)
	}

	var response scenarioResponse
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if response.Scenario.ID != scenarioID {
		t.Fatalf(
			"expected scenario ID %s, got %s",
			scenarioID,
			response.Scenario.ID,
		)
	}

	if !reflect.DeepEqual(response.Scenario.Definition, scenario) {
		t.Fatalf(
			"response contained unexpected definition: %#v",
			response.Scenario.Definition,
		)
	}

	if !response.Scenario.CreatedAt.Equal(createdAt) {
		t.Fatalf(
			"expected createdAt %s, got %s",
			createdAt,
			response.Scenario.CreatedAt,
		)
	}

	if !response.Scenario.UpdatedAt.Equal(updatedAt) {
		t.Fatalf(
			"expected updatedAt %s, got %s",
			updatedAt,
			response.Scenario.UpdatedAt,
		)
	}

	if response.RequestID == "" {
		t.Fatal("expected non-empty request ID")
	}

	if response.RequestID != recorder.Header().Get("X-Request-ID") {
		t.Fatal("response and header request IDs do not match")
	}
}

func TestCreateScenarioRouteErrors(t *testing.T) {
	tests := []struct {
		name           string
		body           []byte
		repositoryErr  error
		expectedStatus int
		expectedCode   ErrorCode
		expectedCalls  int
		forbiddenText  string
	}{
		{
			name:           "invalid JSON",
			body:           []byte(`{"name":`),
			expectedStatus: http.StatusBadRequest,
			expectedCode:   ErrorCodeInvalidJSON,
		},
		{
			name:           "invalid scenario",
			body:           []byte(`{}`),
			expectedStatus: http.StatusUnprocessableEntity,
			expectedCode:   ErrorCodeValidationFailed,
		},
		{
			name:           "storage failure",
			body:           validScenarioJSON(t),
			repositoryErr:  errors.New("private database detail"),
			expectedStatus: http.StatusInternalServerError,
			expectedCode:   ErrorCodeStorage,
			expectedCalls:  1,
			forbiddenText:  "private database detail",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repository := &fakeScenarioRepository{
				createError: test.repositoryErr,
			}
			handler := newScenarioTestRouter(repository)
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(
				http.MethodPost,
				"/api/v1/scenarios",
				bytes.NewReader(test.body),
			)

			handler.ServeHTTP(recorder, request)

			if recorder.Code != test.expectedStatus {
				t.Fatalf(
					"expected status %d, got %d; body=%s",
					test.expectedStatus,
					recorder.Code,
					recorder.Body.String(),
				)
			}

			if repository.createCalls != test.expectedCalls {
				t.Fatalf(
					"expected %d repository calls, got %d",
					test.expectedCalls,
					repository.createCalls,
				)
			}

			assertJSONErrorResponse(
				t,
				recorder,
				test.expectedCode,
				test.forbiddenText,
			)
		})
	}
}

func TestGetScenarioRoute(t *testing.T) {
	scenario := validScenarioValue(t)
	scenarioID := uuid.New()
	timestamp := time.Date(
		2026,
		time.August,
		9,
		9,
		15,
		0,
		0,
		time.UTC,
	)

	repository := &fakeScenarioRepository{
		getRecord: store.ScenarioRecord{
			ID:        scenarioID,
			Scenario:  scenario,
			CreatedAt: timestamp,
			UpdatedAt: timestamp,
		},
	}

	handler := newScenarioTestRouter(repository)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/scenarios/"+scenarioID.String(),
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

	if repository.getCalls != 1 || repository.getID != scenarioID {
		t.Fatalf(
			"repository received unexpected ID: %s",
			repository.getID,
		)
	}

	var response scenarioResponse
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if response.Scenario.ID != scenarioID {
		t.Fatalf(
			"expected scenario ID %s, got %s",
			scenarioID,
			response.Scenario.ID,
		)
	}
}

func TestGetScenarioRouteErrors(t *testing.T) {
	tests := []struct {
		name           string
		id             string
		repositoryErr  error
		expectedStatus int
		expectedCode   ErrorCode
		expectedCalls  int
		forbiddenText  string
	}{
		{
			name:           "invalid UUID",
			id:             "not-a-uuid",
			expectedStatus: http.StatusBadRequest,
			expectedCode:   ErrorCodeInvalidID,
		},
		{
			name:           "missing scenario",
			id:             uuid.NewString(),
			repositoryErr:  store.ErrNotFound,
			expectedStatus: http.StatusNotFound,
			expectedCode:   ErrorCodeScenarioNotFound,
			expectedCalls:  1,
		},
		{
			name:           "storage failure",
			id:             uuid.NewString(),
			repositoryErr:  errors.New("private database detail"),
			expectedStatus: http.StatusInternalServerError,
			expectedCode:   ErrorCodeStorage,
			expectedCalls:  1,
			forbiddenText:  "private database detail",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repository := &fakeScenarioRepository{
				getError: test.repositoryErr,
			}
			handler := newScenarioTestRouter(repository)
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(
				http.MethodGet,
				"/api/v1/scenarios/"+test.id,
				nil,
			)

			handler.ServeHTTP(recorder, request)

			if recorder.Code != test.expectedStatus {
				t.Fatalf(
					"expected status %d, got %d; body=%s",
					test.expectedStatus,
					recorder.Code,
					recorder.Body.String(),
				)
			}

			if repository.getCalls != test.expectedCalls {
				t.Fatalf(
					"expected %d repository calls, got %d",
					test.expectedCalls,
					repository.getCalls,
				)
			}

			assertJSONErrorResponse(
				t,
				recorder,
				test.expectedCode,
				test.forbiddenText,
			)
		})
	}
}

func TestUpdateScenarioRoute(t *testing.T) {
	scenario := validScenarioValue(t)
	pathID := uuid.New()
	timestamp := time.Date(
		2026,
		time.August,
		9,
		9,
		30,
		0,
		0,
		time.UTC,
	)

	repository := &fakeScenarioRepository{
		updateRecord: store.ScenarioRecord{
			ID:        pathID,
			Scenario:  scenario,
			CreatedAt: timestamp,
			UpdatedAt: timestamp.Add(time.Minute),
		},
	}

	handler := newScenarioTestRouter(repository)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(
		http.MethodPut,
		"/api/v1/scenarios/"+pathID.String(),
		bytes.NewReader(validScenarioJSON(t)),
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

	if repository.updateCalls != 1 {
		t.Fatalf(
			"expected one UpdateScenario call, got %d",
			repository.updateCalls,
		)
	}

	if repository.updateID != pathID {
		t.Fatalf(
			"expected path ID %s, got %s",
			pathID,
			repository.updateID,
		)
	}

	if !reflect.DeepEqual(repository.updateInput, scenario) {
		t.Fatalf(
			"repository received unexpected scenario: %#v",
			repository.updateInput,
		)
	}
}

func TestUpdateScenarioRouteErrors(t *testing.T) {
	embeddedIDBody := append(
		[]byte(`{"id":"`+uuid.NewString()+`",`),
		validScenarioJSON(t)[1:]...,
	)

	tests := []struct {
		name           string
		id             string
		body           []byte
		repositoryErr  error
		expectedStatus int
		expectedCode   ErrorCode
		expectedCalls  int
		forbiddenText  string
	}{
		{
			name:           "invalid UUID",
			id:             "not-a-uuid",
			body:           validScenarioJSON(t),
			expectedStatus: http.StatusBadRequest,
			expectedCode:   ErrorCodeInvalidID,
		},
		{
			name:           "invalid JSON",
			id:             uuid.NewString(),
			body:           []byte(`{"name":`),
			expectedStatus: http.StatusBadRequest,
			expectedCode:   ErrorCodeInvalidJSON,
		},
		{
			name:           "invalid scenario",
			id:             uuid.NewString(),
			body:           []byte(`{}`),
			expectedStatus: http.StatusUnprocessableEntity,
			expectedCode:   ErrorCodeValidationFailed,
		},
		{
			name:           "body cannot replace path ID",
			id:             uuid.NewString(),
			body:           embeddedIDBody,
			expectedStatus: http.StatusBadRequest,
			expectedCode:   ErrorCodeInvalidJSON,
		},
		{
			name:           "missing scenario",
			id:             uuid.NewString(),
			body:           validScenarioJSON(t),
			repositoryErr:  store.ErrNotFound,
			expectedStatus: http.StatusNotFound,
			expectedCode:   ErrorCodeScenarioNotFound,
			expectedCalls:  1,
		},
		{
			name:           "storage failure",
			id:             uuid.NewString(),
			body:           validScenarioJSON(t),
			repositoryErr:  errors.New("private database detail"),
			expectedStatus: http.StatusInternalServerError,
			expectedCode:   ErrorCodeStorage,
			expectedCalls:  1,
			forbiddenText:  "private database detail",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repository := &fakeScenarioRepository{
				updateError: test.repositoryErr,
			}
			handler := newScenarioTestRouter(repository)
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(
				http.MethodPut,
				"/api/v1/scenarios/"+test.id,
				bytes.NewReader(test.body),
			)

			handler.ServeHTTP(recorder, request)

			if recorder.Code != test.expectedStatus {
				t.Fatalf(
					"expected status %d, got %d; body=%s",
					test.expectedStatus,
					recorder.Code,
					recorder.Body.String(),
				)
			}

			if repository.updateCalls != test.expectedCalls {
				t.Fatalf(
					"expected %d repository calls, got %d",
					test.expectedCalls,
					repository.updateCalls,
				)
			}

			assertJSONErrorResponse(
				t,
				recorder,
				test.expectedCode,
				test.forbiddenText,
			)
		})
	}
}

func newScenarioTestRouter(
	repository store.ScenarioRepository,
) http.Handler {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	return NewRouter(
		Dependencies{
			Scenarios: repository,
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

func validScenarioValue(t *testing.T) domain.Scenario {
	t.Helper()

	var scenario domain.Scenario
	if err := json.Unmarshal(validScenarioJSON(t), &scenario); err != nil {
		t.Fatalf("decode valid test scenario: %v", err)
	}

	return scenario
}
