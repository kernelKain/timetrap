package httpapi

import (
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

type fakeAnalysisRepository struct {
	saveCalls      int
	saveScenarioID uuid.UUID
	saveSnapshot   domain.Scenario
	saveResult     domain.Analysis
	saveRecord     store.AnalysisRecord
	saveError      error
	getCalls       int
	getID          uuid.UUID
	getRecord      store.AnalysisRecord
	getError       error
}

func (repository *fakeAnalysisRepository) SaveAnalysis(
	_ context.Context,
	scenarioID uuid.UUID,
	snapshot domain.Scenario,
	analysis domain.Analysis,
) (store.AnalysisRecord, error) {
	repository.saveCalls++
	repository.saveScenarioID = scenarioID
	repository.saveSnapshot = snapshot
	repository.saveResult = analysis

	return repository.saveRecord, repository.saveError
}

func (repository *fakeAnalysisRepository) GetAnalysis(
	_ context.Context,
	id uuid.UUID,
) (store.AnalysisRecord, error) {
	repository.getCalls++
	repository.getID = id

	return repository.getRecord, repository.getError
}

type fakeAnalyzer struct {
	calls  int
	input  domain.Scenario
	result domain.Analysis
	err    error
}

func (analyzer *fakeAnalyzer) Analyze(
	scenario domain.Scenario,
) (domain.Analysis, error) {
	analyzer.calls++
	analyzer.input = scenario

	return analyzer.result, analyzer.err
}

func TestCreateAnalysisRoute(t *testing.T) {
	scenario := validScenarioValue(t)
	scenarioID := uuid.New()
	analysisID := uuid.New()
	result := validAnalysisResult()
	createdAt := time.Date(
		2026,
		time.August,
		9,
		10,
		0,
		0,
		0,
		time.UTC,
	)

	scenarios := &fakeScenarioRepository{
		getRecord: store.ScenarioRecord{
			ID:       scenarioID,
			Scenario: scenario,
		},
	}
	analyses := &fakeAnalysisRepository{
		saveRecord: store.AnalysisRecord{
			ID:               analysisID,
			ScenarioID:       scenarioID,
			ScenarioSnapshot: scenario,
			Analysis:         result,
			EngineVersion:    result.EngineVersion,
			CreatedAt:        createdAt,
		},
	}
	analyze := &fakeAnalyzer{
		result: result,
	}

	handler := newAnalysisTestRouter(
		scenarios,
		analyses,
		analyze.Analyze,
	)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/scenarios/"+scenarioID.String()+"/analyses",
		nil,
	)

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf(
			"expected status %d, got %d; body=%s",
			http.StatusCreated,
			recorder.Code,
			recorder.Body.String(),
		)
	}

	if scenarios.getCalls != 1 || scenarios.getID != scenarioID {
		t.Fatalf(
			"scenario repository received unexpected ID: %s",
			scenarios.getID,
		)
	}

	if analyze.calls != 1 {
		t.Fatalf(
			"expected one analyzer call, got %d",
			analyze.calls,
		)
	}

	if !reflect.DeepEqual(analyze.input, scenario) {
		t.Fatal("analyzer did not receive the stored scenario")
	}

	if analyses.saveCalls != 1 {
		t.Fatalf(
			"expected one SaveAnalysis call, got %d",
			analyses.saveCalls,
		)
	}

	if analyses.saveScenarioID != scenarioID {
		t.Fatalf(
			"expected saved scenario ID %s, got %s",
			scenarioID,
			analyses.saveScenarioID,
		)
	}

	if !reflect.DeepEqual(analyses.saveSnapshot, scenario) {
		t.Fatal("analysis did not save the exact loaded scenario snapshot")
	}

	if !reflect.DeepEqual(analyses.saveResult, result) {
		t.Fatal("analysis repository received an unexpected result")
	}

	var response analysisResponse
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if response.Analysis.ID != analysisID {
		t.Fatalf(
			"expected analysis ID %s, got %s",
			analysisID,
			response.Analysis.ID,
		)
	}

	if response.Analysis.ScenarioID != scenarioID {
		t.Fatalf(
			"expected scenario ID %s, got %s",
			scenarioID,
			response.Analysis.ScenarioID,
		)
	}

	if response.Analysis.ScenarioSnapshot != nil {
		t.Fatal("POST response must not include the scenario snapshot")
	}

	if !reflect.DeepEqual(response.Analysis.Result, result) {
		t.Fatal("POST response contained an unexpected analysis result")
	}

	if response.Analysis.EngineVersion != result.EngineVersion {
		t.Fatalf(
			"expected engine version %q, got %q",
			result.EngineVersion,
			response.Analysis.EngineVersion,
		)
	}

	if response.RequestID == "" {
		t.Fatal("expected non-empty request ID")
	}
}

func TestCreateAnalysisRouteErrors(t *testing.T) {
	validationFailure := domain.ValidationErrors{
		{
			Field:   "objects",
			Message: "contains an invalid reference",
		},
	}

	tests := []struct {
		name                  string
		scenarioID            string
		scenarioError         error
		analyzerError         error
		saveError             error
		expectedStatus        int
		expectedCode          ErrorCode
		expectedScenarioCalls int
		expectedAnalyzerCalls int
		expectedSaveCalls     int
		forbiddenText         string
	}{
		{
			name:           "invalid scenario UUID",
			scenarioID:     "not-a-uuid",
			expectedStatus: http.StatusBadRequest,
			expectedCode:   ErrorCodeInvalidID,
		},
		{
			name:                  "missing scenario",
			scenarioID:            uuid.NewString(),
			scenarioError:         store.ErrNotFound,
			expectedStatus:        http.StatusNotFound,
			expectedCode:          ErrorCodeScenarioNotFound,
			expectedScenarioCalls: 1,
		},
		{
			name:                  "scenario storage failure",
			scenarioID:            uuid.NewString(),
			scenarioError:         errors.New("private scenario storage detail"),
			expectedStatus:        http.StatusInternalServerError,
			expectedCode:          ErrorCodeStorage,
			expectedScenarioCalls: 1,
			forbiddenText:         "private scenario storage detail",
		},
		{
			name:                  "scenario cannot be analyzed",
			scenarioID:            uuid.NewString(),
			analyzerError:         validationFailure,
			expectedStatus:        http.StatusConflict,
			expectedCode:          ErrorCodeScenarioNotAnalyzable,
			expectedScenarioCalls: 1,
			expectedAnalyzerCalls: 1,
		},
		{
			name:                  "unexpected analyzer failure",
			scenarioID:            uuid.NewString(),
			analyzerError:         errors.New("private analyzer detail"),
			expectedStatus:        http.StatusInternalServerError,
			expectedCode:          ErrorCodeAnalysisFailed,
			expectedScenarioCalls: 1,
			expectedAnalyzerCalls: 1,
			forbiddenText:         "private analyzer detail",
		},
		{
			name:                  "scenario deleted before save",
			scenarioID:            uuid.NewString(),
			saveError:             store.ErrNotFound,
			expectedStatus:        http.StatusNotFound,
			expectedCode:          ErrorCodeScenarioNotFound,
			expectedScenarioCalls: 1,
			expectedAnalyzerCalls: 1,
			expectedSaveCalls:     1,
		},
		{
			name:                  "analysis storage failure",
			scenarioID:            uuid.NewString(),
			saveError:             errors.New("private analysis storage detail"),
			expectedStatus:        http.StatusInternalServerError,
			expectedCode:          ErrorCodeStorage,
			expectedScenarioCalls: 1,
			expectedAnalyzerCalls: 1,
			expectedSaveCalls:     1,
			forbiddenText:         "private analysis storage detail",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			scenario := validScenarioValue(t)
			scenarios := &fakeScenarioRepository{
				getRecord: store.ScenarioRecord{
					Scenario: scenario,
				},
				getError: test.scenarioError,
			}
			analyses := &fakeAnalysisRepository{
				saveError: test.saveError,
			}
			analyze := &fakeAnalyzer{
				result: validAnalysisResult(),
				err:    test.analyzerError,
			}

			handler := newAnalysisTestRouter(
				scenarios,
				analyses,
				analyze.Analyze,
			)
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(
				http.MethodPost,
				"/api/v1/scenarios/"+test.scenarioID+"/analyses",
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

			if scenarios.getCalls != test.expectedScenarioCalls {
				t.Fatalf(
					"expected %d scenario calls, got %d",
					test.expectedScenarioCalls,
					scenarios.getCalls,
				)
			}

			if analyze.calls != test.expectedAnalyzerCalls {
				t.Fatalf(
					"expected %d analyzer calls, got %d",
					test.expectedAnalyzerCalls,
					analyze.calls,
				)
			}

			if analyses.saveCalls != test.expectedSaveCalls {
				t.Fatalf(
					"expected %d save calls, got %d",
					test.expectedSaveCalls,
					analyses.saveCalls,
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

func TestGetAnalysisRouteUsesStoredSnapshot(t *testing.T) {
	snapshot := validScenarioValue(t)
	snapshot.Name = "Original stored snapshot"

	analysisID := uuid.New()
	scenarioID := uuid.New()
	result := validAnalysisResult()
	createdAt := time.Date(
		2026,
		time.August,
		9,
		10,
		30,
		0,
		0,
		time.UTC,
	)

	scenarios := &fakeScenarioRepository{}
	analyses := &fakeAnalysisRepository{
		getRecord: store.AnalysisRecord{
			ID:               analysisID,
			ScenarioID:       scenarioID,
			ScenarioSnapshot: snapshot,
			Analysis:         result,
			EngineVersion:    result.EngineVersion,
			CreatedAt:        createdAt,
		},
	}

	handler := newAnalysisTestRouter(
		scenarios,
		analyses,
		(&fakeAnalyzer{}).Analyze,
	)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/analyses/"+analysisID.String(),
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

	if analyses.getCalls != 1 || analyses.getID != analysisID {
		t.Fatalf(
			"analysis repository received unexpected ID: %s",
			analyses.getID,
		)
	}

	if scenarios.getCalls != 0 {
		t.Fatalf(
			"GET analysis unexpectedly loaded the current scenario %d times",
			scenarios.getCalls,
		)
	}

	var response analysisResponse
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if response.Analysis.ScenarioSnapshot == nil {
		t.Fatal("GET response did not include the stored snapshot")
	}

	if !reflect.DeepEqual(
		*response.Analysis.ScenarioSnapshot,
		snapshot,
	) {
		t.Fatal("GET response did not use the stored scenario snapshot")
	}

	if !reflect.DeepEqual(response.Analysis.Result, result) {
		t.Fatal("GET response contained an unexpected analysis result")
	}

	if !response.Analysis.CreatedAt.Equal(createdAt) {
		t.Fatalf(
			"expected createdAt %s, got %s",
			createdAt,
			response.Analysis.CreatedAt,
		)
	}
}

func TestGetAnalysisRouteErrors(t *testing.T) {
	tests := []struct {
		name           string
		analysisID     string
		repositoryErr  error
		expectedStatus int
		expectedCode   ErrorCode
		expectedCalls  int
		forbiddenText  string
	}{
		{
			name:           "invalid analysis UUID",
			analysisID:     "not-a-uuid",
			expectedStatus: http.StatusBadRequest,
			expectedCode:   ErrorCodeInvalidID,
		},
		{
			name:           "missing analysis",
			analysisID:     uuid.NewString(),
			repositoryErr:  store.ErrNotFound,
			expectedStatus: http.StatusNotFound,
			expectedCode:   ErrorCodeAnalysisNotFound,
			expectedCalls:  1,
		},
		{
			name:           "analysis storage failure",
			analysisID:     uuid.NewString(),
			repositoryErr:  errors.New("private analysis storage detail"),
			expectedStatus: http.StatusInternalServerError,
			expectedCode:   ErrorCodeStorage,
			expectedCalls:  1,
			forbiddenText:  "private analysis storage detail",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			analyses := &fakeAnalysisRepository{
				getError: test.repositoryErr,
			}
			handler := newAnalysisTestRouter(
				&fakeScenarioRepository{},
				analyses,
				(&fakeAnalyzer{}).Analyze,
			)
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(
				http.MethodGet,
				"/api/v1/analyses/"+test.analysisID,
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

			if analyses.getCalls != test.expectedCalls {
				t.Fatalf(
					"expected %d repository calls, got %d",
					test.expectedCalls,
					analyses.getCalls,
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

func newAnalysisTestRouter(
	scenarios store.ScenarioRepository,
	analyses store.AnalysisRepository,
	analyze func(domain.Scenario) (domain.Analysis, error),
) http.Handler {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	return NewRouter(
		Dependencies{
			Scenarios: scenarios,
			Analyses:  analyses,
			Analyze:   analyze,
			DatabaseReady: func(context.Context) error {
				return nil
			},
			Logger: logger,
		},
		[]string{"http://localhost:5173"},
	)
}

func validAnalysisResult() domain.Analysis {
	return domain.Analysis{
		Safe:                true,
		EvaluatedBoundaries: 2,
		Boundaries:          []int{0, 90},
		Violations:          []domain.Violation{},
		EngineVersion:       "v1",
	}
}
