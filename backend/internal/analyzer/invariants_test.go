package analyzer

import (
	"testing"

	"github.com/kernelKain/timetrap/backend/internal/domain"
)

func TestEvaluateInvariantRevokedAccessGrace(t *testing.T) {
	t.Parallel()

	graceMinutes := 5

	invariant := domain.Invariant{
		Type:              domain.InvariantTypeRevokedAccessGrace,
		SourceObjectID:    "subscription",
		DependentObjectID: "premium-cache",
		GraceMinutes:      &graceMinutes,
	}

	testCases := []struct {
		name        string
		minute      int
		states      map[string]ObjectState
		wantFailure bool
	}{
		{
			name:        "before deadline",
			minute:      14,
			states:      revokedGraceStates(true),
			wantFailure: false,
		},
		{
			name:        "exactly at deadline",
			minute:      15,
			states:      revokedGraceStates(true),
			wantFailure: true,
		},
		{
			name:        "after deadline",
			minute:      30,
			states:      revokedGraceStates(true),
			wantFailure: true,
		},
		{
			name:        "dependent already invalid",
			minute:      15,
			states:      revokedGraceStates(false),
			wantFailure: false,
		},
		{
			name:   "source was reissued",
			minute: 25,
			states: map[string]ObjectState{
				"subscription": {
					ObjectID:        "subscription",
					Valid:           true,
					ValidSince:      20,
					HasValidSince:   true,
					LastEventType:   domain.EventTypeIssue,
					LastEventMinute: 20,
					HasLastEvent:    true,
				},
				"premium-cache": validState(
					"premium-cache",
					0,
					domain.EventTypeIssue,
					0,
				),
			},
			wantFailure: false,
		},
		{
			name:   "source expired rather than revoked",
			minute: 15,
			states: map[string]ObjectState{
				"subscription": {
					ObjectID:        "subscription",
					LastEventType:   domain.EventTypeExpire,
					LastEventMinute: 10,
					HasLastEvent:    true,
				},
				"premium-cache": validState(
					"premium-cache",
					0,
					domain.EventTypeIssue,
					0,
				),
			},
			wantFailure: false,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			failure, err := EvaluateInvariant(
				invariant,
				2,
				testCase.states,
				testCase.minute,
			)
			if err != nil {
				t.Fatalf("EvaluateInvariant() error = %v", err)
			}

			if testCase.wantFailure && failure == nil {
				t.Fatal("EvaluateInvariant() failure = nil, want failure")
			}

			if !testCase.wantFailure && failure != nil {
				t.Fatalf(
					"EvaluateInvariant() failure = %#v, want nil",
					failure,
				)
			}

			if failure != nil {
				if failure.InvariantIndex != 2 {
					t.Fatalf(
						"failure invariant index = %d, want 2",
						failure.InvariantIndex,
					)
				}

				if failure.SourceObjectID != "subscription" {
					t.Fatalf(
						"failure source = %q, want subscription",
						failure.SourceObjectID,
					)
				}

				if failure.DependentObjectID != "premium-cache" {
					t.Fatalf(
						"failure dependent = %q, want premium-cache",
						failure.DependentObjectID,
					)
				}
			}
		})
	}
}

func TestEvaluateInvariantRevokedAccessZeroGrace(t *testing.T) {
	t.Parallel()

	graceMinutes := 0

	invariant := domain.Invariant{
		Type:              domain.InvariantTypeRevokedAccessGrace,
		SourceObjectID:    "subscription",
		DependentObjectID: "premium-cache",
		GraceMinutes:      &graceMinutes,
	}

	failure, err := EvaluateInvariant(
		invariant,
		0,
		revokedGraceStates(true),
		10,
	)
	if err != nil {
		t.Fatalf("EvaluateInvariant() error = %v", err)
	}

	if failure == nil {
		t.Fatal("EvaluateInvariant() failure = nil, want failure at revoke minute")
	}
}

func TestEvaluateInvariantDependentNotOutliveSource(t *testing.T) {
	t.Parallel()

	invariant := domain.Invariant{
		Type:              domain.InvariantTypeDependentNotOutliveSource,
		SourceObjectID:    "subscription",
		DependentObjectID: "premium-cache",
	}

	testCases := []struct {
		name        string
		states      map[string]ObjectState
		wantFailure bool
	}{
		{
			name: "dependent remains valid after source",
			states: map[string]ObjectState{
				"subscription": {
					ObjectID:        "subscription",
					LastEventType:   domain.EventTypeExpire,
					LastEventMinute: 60,
					HasLastEvent:    true,
				},
				"premium-cache": validState(
					"premium-cache",
					0,
					domain.EventTypeIssue,
					0,
				),
			},
			wantFailure: true,
		},
		{
			name: "source and dependent both invalid",
			states: map[string]ObjectState{
				"subscription": {
					ObjectID:        "subscription",
					LastEventType:   domain.EventTypeExpire,
					LastEventMinute: 60,
					HasLastEvent:    true,
				},
				"premium-cache": {
					ObjectID:        "premium-cache",
					LastEventType:   domain.EventTypeExpire,
					LastEventMinute: 60,
					HasLastEvent:    true,
				},
			},
			wantFailure: false,
		},
		{
			name: "both valid",
			states: map[string]ObjectState{
				"subscription": validState(
					"subscription",
					0,
					domain.EventTypeIssue,
					0,
				),
				"premium-cache": validState(
					"premium-cache",
					0,
					domain.EventTypeIssue,
					0,
				),
			},
			wantFailure: false,
		},
		{
			name: "source valid and dependent invalid",
			states: map[string]ObjectState{
				"subscription": validState(
					"subscription",
					0,
					domain.EventTypeIssue,
					0,
				),
				"premium-cache": {
					ObjectID: "premium-cache",
				},
			},
			wantFailure: false,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			failure, err := EvaluateInvariant(
				invariant,
				1,
				testCase.states,
				60,
			)
			if err != nil {
				t.Fatalf("EvaluateInvariant() error = %v", err)
			}

			if testCase.wantFailure && failure == nil {
				t.Fatal("EvaluateInvariant() failure = nil, want failure")
			}

			if !testCase.wantFailure && failure != nil {
				t.Fatalf(
					"EvaluateInvariant() failure = %#v, want nil",
					failure,
				)
			}
		})
	}
}

func TestEvaluateInvariantMaximumValidity(t *testing.T) {
	t.Parallel()

	maximumMinutes := 30

	invariant := domain.Invariant{
		Type:           domain.InvariantTypeMaxValidity,
		TargetObjectID: "session",
		MaximumMinutes: &maximumMinutes,
	}

	testCases := []struct {
		name        string
		minute      int
		target      ObjectState
		wantFailure bool
	}{
		{
			name:   "before initial deadline",
			minute: 29,
			target: validState(
				"session",
				0,
				domain.EventTypeIssue,
				0,
			),
			wantFailure: false,
		},
		{
			name:   "at initial deadline",
			minute: 30,
			target: validState(
				"session",
				0,
				domain.EventTypeIssue,
				0,
			),
			wantFailure: true,
		},
		{
			name:   "before refreshed deadline",
			minute: 49,
			target: validState(
				"session",
				20,
				domain.EventTypeRefresh,
				20,
			),
			wantFailure: false,
		},
		{
			name:   "at refreshed deadline",
			minute: 50,
			target: validState(
				"session",
				20,
				domain.EventTypeRefresh,
				20,
			),
			wantFailure: true,
		},
		{
			name:   "invalid target",
			minute: 60,
			target: ObjectState{
				ObjectID:        "session",
				LastEventType:   domain.EventTypeExpire,
				LastEventMinute: 60,
				HasLastEvent:    true,
			},
			wantFailure: false,
		},
		{
			name:   "valid target without known start",
			minute: 60,
			target: ObjectState{
				ObjectID: "session",
				Valid:    true,
			},
			wantFailure: false,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			failure, err := EvaluateInvariant(
				invariant,
				3,
				map[string]ObjectState{
					"session": testCase.target,
				},
				testCase.minute,
			)
			if err != nil {
				t.Fatalf("EvaluateInvariant() error = %v", err)
			}

			if testCase.wantFailure && failure == nil {
				t.Fatal("EvaluateInvariant() failure = nil, want failure")
			}

			if !testCase.wantFailure && failure != nil {
				t.Fatalf(
					"EvaluateInvariant() failure = %#v, want nil",
					failure,
				)
			}

			if failure != nil && failure.TargetObjectID != "session" {
				t.Fatalf(
					"failure target = %q, want session",
					failure.TargetObjectID,
				)
			}
		})
	}
}

func TestEvaluateInvariantErrors(t *testing.T) {
	t.Parallel()

	graceMinutes := 5
	maximumMinutes := 30
	zeroMinutes := 0

	validStates := map[string]ObjectState{
		"subscription": validState(
			"subscription",
			0,
			domain.EventTypeIssue,
			0,
		),
		"premium-cache": validState(
			"premium-cache",
			0,
			domain.EventTypeIssue,
			0,
		),
		"session": validState(
			"session",
			0,
			domain.EventTypeIssue,
			0,
		),
	}

	testCases := []struct {
		name      string
		invariant domain.Invariant
		states    map[string]ObjectState
		minute    int
	}{
		{
			name: "missing source",
			invariant: domain.Invariant{
				Type:              domain.InvariantTypeRevokedAccessGrace,
				SourceObjectID:    "missing",
				DependentObjectID: "premium-cache",
				GraceMinutes:      &graceMinutes,
			},
			states: validStates,
			minute: 15,
		},
		{
			name: "missing dependent",
			invariant: domain.Invariant{
				Type:              domain.InvariantTypeDependentNotOutliveSource,
				SourceObjectID:    "subscription",
				DependentObjectID: "missing",
			},
			states: validStates,
			minute: 15,
		},
		{
			name: "missing target",
			invariant: domain.Invariant{
				Type:           domain.InvariantTypeMaxValidity,
				TargetObjectID: "missing",
				MaximumMinutes: &maximumMinutes,
			},
			states: validStates,
			minute: 30,
		},
		{
			name: "missing grace",
			invariant: domain.Invariant{
				Type:              domain.InvariantTypeRevokedAccessGrace,
				SourceObjectID:    "subscription",
				DependentObjectID: "premium-cache",
			},
			states: validStates,
			minute: 15,
		},
		{
			name: "negative grace",
			invariant: domain.Invariant{
				Type:              domain.InvariantTypeRevokedAccessGrace,
				SourceObjectID:    "subscription",
				DependentObjectID: "premium-cache",
				GraceMinutes:      intPointer(-1),
			},
			states: validStates,
			minute: 15,
		},
		{
			name: "missing maximum",
			invariant: domain.Invariant{
				Type:           domain.InvariantTypeMaxValidity,
				TargetObjectID: "session",
			},
			states: validStates,
			minute: 30,
		},
		{
			name: "zero maximum",
			invariant: domain.Invariant{
				Type:           domain.InvariantTypeMaxValidity,
				TargetObjectID: "session",
				MaximumMinutes: &zeroMinutes,
			},
			states: validStates,
			minute: 30,
		},
		{
			name: "unsupported invariant",
			invariant: domain.Invariant{
				Type: "unsupported",
			},
			states: validStates,
			minute: 15,
		},
		{
			name: "negative minute",
			invariant: domain.Invariant{
				Type:              domain.InvariantTypeRevokedAccessGrace,
				SourceObjectID:    "subscription",
				DependentObjectID: "premium-cache",
				GraceMinutes:      &graceMinutes,
			},
			states: validStates,
			minute: -1,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			failure, err := EvaluateInvariant(
				testCase.invariant,
				0,
				testCase.states,
				testCase.minute,
			)

			if err == nil {
				t.Fatalf(
					"EvaluateInvariant() failure = %#v, error = nil; want error",
					failure,
				)
			}

			if failure != nil {
				t.Fatalf(
					"EvaluateInvariant() failure = %#v, want nil on error",
					failure,
				)
			}
		})
	}
}

func revokedGraceStates(dependentValid bool) map[string]ObjectState {
	dependent := ObjectState{
		ObjectID: "premium-cache",
	}

	if dependentValid {
		dependent = validState(
			"premium-cache",
			0,
			domain.EventTypeIssue,
			0,
		)
	}

	return map[string]ObjectState{
		"subscription": {
			ObjectID:        "subscription",
			LastEventType:   domain.EventTypeRevoke,
			LastEventMinute: 10,
			HasLastEvent:    true,
		},
		"premium-cache": dependent,
	}
}

func validState(
	objectID string,
	validSince int,
	lastEventType domain.EventType,
	lastEventMinute int,
) ObjectState {
	return ObjectState{
		ObjectID:        objectID,
		Valid:           true,
		ValidSince:      validSince,
		HasValidSince:   true,
		LastEventType:   lastEventType,
		LastEventMinute: lastEventMinute,
		HasLastEvent:    true,
	}
}

func intPointer(value int) *int {
	return &value
}

func TestEvaluateInvariantScenarioTable(t *testing.T) {
	t.Parallel()

	graceMinutes := 5
	maximumMinutes := 30

	graceScenario := domain.Scenario{
		Objects: []domain.TimedObject{
			{
				ClientID: "source",
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
				ClientID: "dependent",
				Events: []domain.Event{
					{
						Type:     domain.EventTypeIssue,
						AtMinute: 0,
					},
				},
			},
		},
		Invariants: []domain.Invariant{
			{
				Type:              domain.InvariantTypeRevokedAccessGrace,
				SourceObjectID:    "source",
				DependentObjectID: "dependent",
				GraceMinutes:      &graceMinutes,
			},
		},
	}

	outlivesScenario := domain.Scenario{
		Objects: []domain.TimedObject{
			{
				ClientID: "source",
				Events: []domain.Event{
					{
						Type:     domain.EventTypeIssue,
						AtMinute: 0,
					},
					{
						Type:     domain.EventTypeExpire,
						AtMinute: 10,
					},
				},
			},
			{
				ClientID: "dependent",
				Events: []domain.Event{
					{
						Type:     domain.EventTypeIssue,
						AtMinute: 0,
					},
					{
						Type:     domain.EventTypeExpire,
						AtMinute: 20,
					},
				},
			},
		},
		Invariants: []domain.Invariant{
			{
				Type:              domain.InvariantTypeDependentNotOutliveSource,
				SourceObjectID:    "source",
				DependentObjectID: "dependent",
			},
		},
	}

	maximumScenario := domain.Scenario{
		Objects: []domain.TimedObject{
			{
				ClientID: "target",
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
				Type:           domain.InvariantTypeMaxValidity,
				TargetObjectID: "target",
				MaximumMinutes: &maximumMinutes,
			},
		},
	}

	missingReferenceScenario := domain.Scenario{
		Objects: []domain.TimedObject{
			{
				ClientID: "dependent",
				Events: []domain.Event{
					{
						Type:     domain.EventTypeIssue,
						AtMinute: 0,
					},
				},
			},
		},
		Invariants: []domain.Invariant{
			{
				Type:              domain.InvariantTypeDependentNotOutliveSource,
				SourceObjectID:    "missing-source",
				DependentObjectID: "dependent",
			},
		},
	}

	tests := []struct {
		name        string
		scenario    domain.Scenario
		minute      int
		wantFailure bool
		wantErr     bool
	}{
		{
			name:        "revocation grace passes before deadline",
			scenario:    graceScenario,
			minute:      14,
			wantFailure: false,
		},
		{
			name:        "revocation grace fails at deadline",
			scenario:    graceScenario,
			minute:      15,
			wantFailure: true,
		},
		{
			name:        "dependent passes before source expiration",
			scenario:    outlivesScenario,
			minute:      9,
			wantFailure: false,
		},
		{
			name:        "dependent fails at source expiration",
			scenario:    outlivesScenario,
			minute:      10,
			wantFailure: true,
		},
		{
			name:        "maximum validity passes before deadline",
			scenario:    maximumScenario,
			minute:      29,
			wantFailure: false,
		},
		{
			name:        "maximum validity fails at deadline",
			scenario:    maximumScenario,
			minute:      30,
			wantFailure: true,
		},
		{
			name:        "missing object reference returns error",
			scenario:    missingReferenceScenario,
			minute:      10,
			wantFailure: false,
			wantErr:     true,
		},
	}

	for _, testCase := range tests {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			states := statesAt(
				testCase.scenario.Objects,
				testCase.minute,
			)

			failure, err := EvaluateInvariant(
				testCase.scenario.Invariants[0],
				0,
				states,
				testCase.minute,
			)

			if testCase.wantErr {
				if err == nil {
					t.Fatal("EvaluateInvariant() error = nil, want error")
				}

				return
			}

			if err != nil {
				t.Fatalf("EvaluateInvariant() error = %v", err)
			}

			gotFailure := failure != nil
			if gotFailure != testCase.wantFailure {
				t.Errorf(
					"failure present = %t, want %t",
					gotFailure,
					testCase.wantFailure,
				)
			}
		})
	}
}
