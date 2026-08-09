package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/kernelKain/timetrap/backend/internal/domain"
	"github.com/kernelKain/timetrap/backend/internal/store"
)

const saveAnalysisQuery = `
	INSERT INTO analyses (
		id,
		scenario_id,
		scenario_snapshot,
		safe,
		result,
		engine_version
	)
	VALUES (
		$1,
		$2,
		$3::jsonb,
		$4,
		$5::jsonb,
		$6
	)
	RETURNING created_at
`

const getAnalysisQuery = `
	SELECT
		id,
		scenario_id,
		scenario_snapshot,
		result,
		engine_version,
		created_at
	FROM analyses
	WHERE id = $1
`

// Compile-time verification that Store implements the analysis repository.
var _ store.AnalysisRepository = (*Store)(nil)

// SaveAnalysis persists an analysis with the exact scenario snapshot used to
// produce it.
func (s *Store) SaveAnalysis(
	ctx context.Context,
	scenarioID uuid.UUID,
	snapshot domain.Scenario,
	analysis domain.Analysis,
) (store.AnalysisRecord, error) {
	snapshotJSON, err := json.Marshal(snapshot)
	if err != nil {
		return store.AnalysisRecord{},
			fmt.Errorf("encode analysis scenario snapshot: %w", err)
	}

	analysisJSON, err := json.Marshal(analysis)
	if err != nil {
		return store.AnalysisRecord{},
			fmt.Errorf("encode analysis result: %w", err)
	}

	record := store.AnalysisRecord{
		ID:               s.newUUID(),
		ScenarioID:       scenarioID,
		ScenarioSnapshot: snapshot,
		Analysis:         analysis,
		EngineVersion:    analysis.EngineVersion,
	}

	queryContext, cancel := context.WithTimeout(ctx, s.queryTimeout)
	defer cancel()

	err = s.pool.QueryRow(
		queryContext,
		saveAnalysisQuery,
		record.ID,
		record.ScenarioID,
		string(snapshotJSON),
		analysis.Safe,
		string(analysisJSON),
		record.EngineVersion,
	).Scan(&record.CreatedAt)
	if err != nil {
		switch {
		case isUniqueViolation(err):
			return store.AnalysisRecord{},
				fmt.Errorf("save analysis: %w", store.ErrConflict)

		case isForeignKeyViolation(err):
			return store.AnalysisRecord{},
				fmt.Errorf("save analysis: %w", store.ErrNotFound)

		default:
			return store.AnalysisRecord{},
				fmt.Errorf("save analysis: %w", err)
		}
	}

	return record, nil
}

// GetAnalysis retrieves a persisted analysis and its immutable scenario
// snapshot.
func (s *Store) GetAnalysis(
	ctx context.Context,
	id uuid.UUID,
) (store.AnalysisRecord, error) {
	var (
		record       store.AnalysisRecord
		snapshotJSON []byte
		analysisJSON []byte
	)

	queryContext, cancel := context.WithTimeout(ctx, s.queryTimeout)
	defer cancel()

	err := s.pool.QueryRow(
		queryContext,
		getAnalysisQuery,
		id,
	).Scan(
		&record.ID,
		&record.ScenarioID,
		&snapshotJSON,
		&analysisJSON,
		&record.EngineVersion,
		&record.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return store.AnalysisRecord{},
			fmt.Errorf("get analysis: %w", store.ErrNotFound)
	}
	if err != nil {
		return store.AnalysisRecord{},
			fmt.Errorf("get analysis: %w", err)
	}

	if err := json.Unmarshal(
		snapshotJSON,
		&record.ScenarioSnapshot,
	); err != nil {
		return store.AnalysisRecord{},
			fmt.Errorf("decode stored analysis scenario snapshot: %w", err)
	}

	if err := json.Unmarshal(analysisJSON, &record.Analysis); err != nil {
		return store.AnalysisRecord{},
			fmt.Errorf("decode stored analysis result: %w", err)
	}

	return record, nil
}

func isForeignKeyViolation(err error) bool {
	var postgresError *pgconn.PgError

	return errors.As(err, &postgresError) &&
		postgresError.Code == "23503"
}
