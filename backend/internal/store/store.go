package store

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/kernelKain/timetrap/backend/internal/domain"
)

// Sentinel repository errors allow callers to select an appropriate response
// without depending on PostgreSQL-specific errors.
var (
	ErrNotFound = errors.New("record not found")
	ErrConflict = errors.New("record conflict")
)

// ScenarioRecord combines a persisted scenario with its storage metadata.
type ScenarioRecord struct {
	ID        uuid.UUID
	Scenario  domain.Scenario
	CreatedAt time.Time
	UpdatedAt time.Time
}

// AnalysisRecord combines a persisted analysis with the exact scenario
// snapshot that was analyzed and its storage metadata.
type AnalysisRecord struct {
	ID               uuid.UUID
	ScenarioID       uuid.UUID
	ScenarioSnapshot domain.Scenario
	Analysis         domain.Analysis
	EngineVersion    string
	CreatedAt        time.Time
}

// ScenarioRepository defines persistence operations for scenarios.
type ScenarioRepository interface {
	CreateScenario(
		ctx context.Context,
		scenario domain.Scenario,
	) (ScenarioRecord, error)

	GetScenario(
		ctx context.Context,
		id uuid.UUID,
	) (ScenarioRecord, error)

	UpdateScenario(
		ctx context.Context,
		id uuid.UUID,
		scenario domain.Scenario,
	) (ScenarioRecord, error)
}

// AnalysisRepository defines persistence operations for analysis reports.
type AnalysisRepository interface {
	SaveAnalysis(
		ctx context.Context,
		scenarioID uuid.UUID,
		snapshot domain.Scenario,
		analysis domain.Analysis,
	) (AnalysisRecord, error)

	GetAnalysis(
		ctx context.Context,
		id uuid.UUID,
	) (AnalysisRecord, error)
}
