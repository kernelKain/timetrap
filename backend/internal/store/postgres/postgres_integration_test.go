//go:build integration

package postgres

import (
	"context"
	"errors"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kernelKain/timetrap/backend/internal/domain"
	"github.com/kernelKain/timetrap/backend/internal/store"
)

func TestPostgresPersistenceIntegration(t *testing.T) {
	cfg := integrationTestConfig(t)
	database := openIntegrationStore(t, cfg)

	t.Cleanup(func() {
		if database != nil {
			database.Close()
		}
	})

	scenarioID := uuid.New()
	cleanupIntegrationScenario(t, cfg, scenarioID)

	originalUUIDGenerator := database.newUUID
	database.newUUID = func() uuid.UUID {
		return scenarioID
	}

	scenarioV1 := integrationScenario(
		"integration-" + scenarioID.String(),
	)

	createdScenario, err := database.CreateScenario(
		context.Background(),
		scenarioV1,
	)
	database.newUUID = originalUUIDGenerator

	if err != nil {
		t.Fatal("create scenario failed")
	}

	if createdScenario.ID != scenarioID {
		t.Fatalf(
			"expected scenario ID %s, got %s",
			scenarioID,
			createdScenario.ID,
		)
	}

	if createdScenario.CreatedAt.IsZero() ||
		createdScenario.UpdatedAt.IsZero() {
		t.Fatal("created scenario has zero timestamps")
	}

	retrievedScenario, err := database.GetScenario(
		context.Background(),
		scenarioID,
	)
	if err != nil {
		t.Fatal("retrieve created scenario failed")
	}

	if !reflect.DeepEqual(retrievedScenario.Scenario, scenarioV1) {
		t.Fatal("retrieved scenario does not match created scenario")
	}

	analysis := integrationAnalysis()

	savedAnalysis, err := database.SaveAnalysis(
		context.Background(),
		scenarioID,
		scenarioV1,
		analysis,
	)
	if err != nil {
		t.Fatal("save analysis failed")
	}

	if savedAnalysis.ID == uuid.Nil {
		t.Fatal("saved analysis has a nil UUID")
	}

	if savedAnalysis.ScenarioID != scenarioID {
		t.Fatalf(
			"expected analysis scenario ID %s, got %s",
			scenarioID,
			savedAnalysis.ScenarioID,
		)
	}

	if savedAnalysis.CreatedAt.IsZero() {
		t.Fatal("saved analysis has a zero creation timestamp")
	}

	scenarioV2 := scenarioV1
	scenarioV2.Name += " updated"
	scenarioV2.Description = "Updated after the original analysis."

	updatedScenario, err := database.UpdateScenario(
		context.Background(),
		scenarioID,
		scenarioV2,
	)
	if err != nil {
		t.Fatal("update scenario failed")
	}

	if updatedScenario.ID != scenarioID {
		t.Fatalf(
			"expected updated scenario ID %s, got %s",
			scenarioID,
			updatedScenario.ID,
		)
	}

	if !updatedScenario.CreatedAt.Equal(createdScenario.CreatedAt) {
		t.Fatal("scenario update changed created_at")
	}

	if updatedScenario.UpdatedAt.Before(createdScenario.UpdatedAt) {
		t.Fatal("scenario update moved updated_at backwards")
	}

	currentScenario, err := database.GetScenario(
		context.Background(),
		scenarioID,
	)
	if err != nil {
		t.Fatal("retrieve updated scenario failed")
	}

	if !reflect.DeepEqual(currentScenario.Scenario, scenarioV2) {
		t.Fatal("current scenario does not contain the update")
	}

	retrievedAnalysis, err := database.GetAnalysis(
		context.Background(),
		savedAnalysis.ID,
	)
	if err != nil {
		t.Fatal("retrieve analysis failed")
	}

	if !reflect.DeepEqual(
		retrievedAnalysis.ScenarioSnapshot,
		scenarioV1,
	) {
		t.Fatal("analysis snapshot changed after scenario update")
	}

	if reflect.DeepEqual(
		retrievedAnalysis.ScenarioSnapshot,
		scenarioV2,
	) {
		t.Fatal("analysis snapshot incorrectly uses current scenario")
	}

	if !reflect.DeepEqual(retrievedAnalysis.Analysis, analysis) {
		t.Fatal("retrieved analysis result does not match saved result")
	}

	if retrievedAnalysis.EngineVersion != analysis.EngineVersion {
		t.Fatalf(
			"expected engine version %q, got %q",
			analysis.EngineVersion,
			retrievedAnalysis.EngineVersion,
		)
	}

	_, err = database.GetScenario(
		context.Background(),
		uuid.New(),
	)
	if !errors.Is(err, store.ErrNotFound) {
		t.Fatal("unknown scenario did not return store.ErrNotFound")
	}

	_, err = database.GetAnalysis(
		context.Background(),
		uuid.New(),
	)
	if !errors.Is(err, store.ErrNotFound) {
		t.Fatal("unknown analysis did not return store.ErrNotFound")
	}

	_, err = database.SaveAnalysis(
		context.Background(),
		uuid.New(),
		scenarioV1,
		analysis,
	)
	if !errors.Is(err, store.ErrNotFound) {
		t.Fatal(
			"analysis with nonexistent scenario did not return store.ErrNotFound",
		)
	}

	analysisID := savedAnalysis.ID

	database.Close()
	database = nil

	database = openIntegrationStore(t, cfg)

	reopenedScenario, err := database.GetScenario(
		context.Background(),
		scenarioID,
	)
	if err != nil {
		t.Fatal("retrieve scenario after reopening pool failed")
	}

	if !reflect.DeepEqual(reopenedScenario.Scenario, scenarioV2) {
		t.Fatal("updated scenario was not preserved after reopening pool")
	}

	reopenedAnalysis, err := database.GetAnalysis(
		context.Background(),
		analysisID,
	)
	if err != nil {
		t.Fatal("retrieve analysis after reopening pool failed")
	}

	if !reflect.DeepEqual(
		reopenedAnalysis.ScenarioSnapshot,
		scenarioV1,
	) {
		t.Fatal("analysis snapshot was not preserved after reopening pool")
	}

	if !reflect.DeepEqual(reopenedAnalysis.Analysis, analysis) {
		t.Fatal("analysis result was not preserved after reopening pool")
	}
}

func integrationTestConfig(t *testing.T) Config {
	t.Helper()

	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}

	parsedConfig, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		t.Fatal("parse TEST_DATABASE_URL failed")
	}

	const requiredDatabase = "timetrap_test"
	if parsedConfig.ConnConfig.Database != requiredDatabase {
		t.Fatalf(
			"refusing integration test against database %q; require %q",
			parsedConfig.ConnConfig.Database,
			requiredDatabase,
		)
	}

	return Config{
		DatabaseURL:    databaseURL,
		MaxConns:       3,
		MinConns:       0,
		ConnectTimeout: 5 * time.Second,
		QueryTimeout:   3 * time.Second,
	}
}

func openIntegrationStore(t *testing.T, cfg Config) *Store {
	t.Helper()

	startupContext, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	database, err := Open(startupContext, cfg)
	if err != nil {
		// Do not include the connection error because it may contain private
		// database addressing or identity details.
		t.Fatal("open integration-test database failed")
	}

	return database
}

func cleanupIntegrationScenario(
	t *testing.T,
	cfg Config,
	scenarioID uuid.UUID,
) {
	t.Helper()

	t.Cleanup(func() {
		database := openIntegrationStoreForCleanup(t, cfg)
		if database == nil {
			return
		}
		defer database.Close()

		cleanupContext, cancel := context.WithTimeout(
			context.Background(),
			3*time.Second,
		)
		defer cancel()

		commandTag, err := database.pool.Exec(
			cleanupContext,
			`DELETE FROM scenarios WHERE id = $1`,
			scenarioID,
		)
		if err != nil {
			t.Error("clean up integration scenario failed")
			return
		}

		if commandTag.RowsAffected() > 1 {
			t.Errorf(
				"cleanup removed %d scenarios; expected at most one",
				commandTag.RowsAffected(),
			)
		}
	})
}

func openIntegrationStoreForCleanup(
	t *testing.T,
	cfg Config,
) *Store {
	t.Helper()

	startupContext, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	database, err := Open(startupContext, cfg)
	if err != nil {
		t.Error("open database for integration cleanup failed")
		return nil
	}

	return database
}

func integrationScenario(name string) domain.Scenario {
	graceMinutes := 5

	return domain.Scenario{
		Name:           name,
		Description:    "PostgreSQL integration test scenario.",
		HorizonMinutes: 90,
		Objects: []domain.TimedObject{
			{
				ClientID: "subscription",
				Name:     "Subscription",
				Kind:     domain.ObjectKindSubscription,
				Events: []domain.Event{
					{
						Type:     domain.EventTypeIssue,
						AtMinute: 0,
					},
					{
						Type:     domain.EventTypeRevoke,
						AtMinute: 10,
					},
				},
			},
			{
				ClientID: "cached-access",
				Name:     "Cached access",
				Kind:     domain.ObjectKindCachedEntitlement,
				Events: []domain.Event{
					{
						Type:     domain.EventTypeIssue,
						AtMinute: 0,
					},
					{
						Type:     domain.EventTypeExpire,
						AtMinute: 60,
					},
				},
			},
		},
		Invariants: []domain.Invariant{
			{
				Type:              domain.InvariantTypeRevokedAccessGrace,
				SourceObjectID:    "subscription",
				DependentObjectID: "cached-access",
				GraceMinutes:      &graceMinutes,
			},
		},
	}
}

func integrationAnalysis() domain.Analysis {
	return domain.Analysis{
		Safe:                true,
		EvaluatedBoundaries: 2,
		Boundaries:          []int{0, 90},
		Violations:          []domain.Violation{},
		EngineVersion:       "integration-test",
	}
}
