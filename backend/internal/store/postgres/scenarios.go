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

const createScenarioQuery = `
	INSERT INTO scenarios (
		id,
		name,
		description,
		definition
	)
	VALUES ($1, $2, $3, $4::jsonb)
	RETURNING created_at, updated_at
`

const getScenarioQuery = `
	SELECT
		id,
		definition,
		created_at,
		updated_at
	FROM scenarios
	WHERE id = $1
`

const updateScenarioQuery = `
	UPDATE scenarios
	SET
		name = $2,
		description = $3,
		definition = $4::jsonb,
		updated_at = NOW()
	WHERE id = $1
	RETURNING created_at, updated_at
`

// Compile-time verification that Store implements the scenario repository.
var _ store.ScenarioRepository = (*Store)(nil)

// CreateScenario persists a new scenario and assigns its public UUID.
func (s *Store) CreateScenario(
	ctx context.Context,
	scenario domain.Scenario,
) (store.ScenarioRecord, error) {
	definitionJSON, err := json.Marshal(scenario)
	if err != nil {
		return store.ScenarioRecord{},
			fmt.Errorf("encode scenario definition: %w", err)
	}

	record := store.ScenarioRecord{
		ID:       s.newUUID(),
		Scenario: scenario,
	}

	queryContext, cancel := context.WithTimeout(ctx, s.queryTimeout)
	defer cancel()

	err = s.pool.QueryRow(
		queryContext,
		createScenarioQuery,
		record.ID,
		scenario.Name,
		scenario.Description,
		string(definitionJSON),
	).Scan(
		&record.CreatedAt,
		&record.UpdatedAt,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return store.ScenarioRecord{},
				fmt.Errorf("create scenario: %w", store.ErrConflict)
		}

		return store.ScenarioRecord{},
			fmt.Errorf("create scenario: %w", err)
	}

	return record, nil
}

// GetScenario retrieves a scenario by its public UUID.
func (s *Store) GetScenario(
	ctx context.Context,
	id uuid.UUID,
) (store.ScenarioRecord, error) {
	var (
		record         store.ScenarioRecord
		definitionJSON []byte
	)

	queryContext, cancel := context.WithTimeout(ctx, s.queryTimeout)
	defer cancel()

	err := s.pool.QueryRow(
		queryContext,
		getScenarioQuery,
		id,
	).Scan(
		&record.ID,
		&definitionJSON,
		&record.CreatedAt,
		&record.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return store.ScenarioRecord{},
			fmt.Errorf("get scenario: %w", store.ErrNotFound)
	}
	if err != nil {
		return store.ScenarioRecord{},
			fmt.Errorf("get scenario: %w", err)
	}

	if err := json.Unmarshal(definitionJSON, &record.Scenario); err != nil {
		return store.ScenarioRecord{},
			fmt.Errorf("decode stored scenario definition: %w", err)
	}

	return record, nil
}

// UpdateScenario replaces a stored scenario definition.
func (s *Store) UpdateScenario(
	ctx context.Context,
	id uuid.UUID,
	scenario domain.Scenario,
) (store.ScenarioRecord, error) {
	definitionJSON, err := json.Marshal(scenario)
	if err != nil {
		return store.ScenarioRecord{},
			fmt.Errorf("encode scenario definition: %w", err)
	}

	record := store.ScenarioRecord{
		ID:       id,
		Scenario: scenario,
	}

	queryContext, cancel := context.WithTimeout(ctx, s.queryTimeout)
	defer cancel()

	err = s.pool.QueryRow(
		queryContext,
		updateScenarioQuery,
		id,
		scenario.Name,
		scenario.Description,
		string(definitionJSON),
	).Scan(
		&record.CreatedAt,
		&record.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return store.ScenarioRecord{},
			fmt.Errorf("update scenario: %w", store.ErrNotFound)
	}
	if err != nil {
		return store.ScenarioRecord{},
			fmt.Errorf("update scenario: %w", err)
	}

	return record, nil
}

func isUniqueViolation(err error) bool {
	var postgresError *pgconn.PgError

	return errors.As(err, &postgresError) &&
		postgresError.Code == "23505"
}
