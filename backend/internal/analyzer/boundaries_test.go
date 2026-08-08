package analyzer

import (
	"errors"
	"reflect"
	"testing"

	"github.com/kernelKain/timetrap/backend/internal/domain"
)

func TestCandidateBoundariesCanonicalScenario(t *testing.T) {
	t.Parallel()

	scenario := canonicalBoundaryScenario()

	got, err := CandidateBoundaries(scenario)
	if err != nil {
		t.Fatalf("CandidateBoundaries() error = %v", err)
	}

	want := []int{0, 1, 9, 10, 11, 15, 59, 60, 61, 90}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf(
			"CandidateBoundaries() = %v, want %v",
			got,
			want,
		)
	}
}

func TestCandidateBoundariesMaximumValidityDeadlines(t *testing.T) {
	t.Parallel()

	maximumMinutes := 30

	scenario := domain.Scenario{
		Name:           "Maximum validity boundaries",
		HorizonMinutes: 90,
		Objects: []domain.TimedObject{
			{
				ClientID: "session",
				Name:     "Application session",
				Kind:     domain.ObjectKindSession,
				Events: []domain.Event{
					{
						Type:     domain.EventTypeIssue,
						AtMinute: 0,
					},
					{
						Type:     domain.EventTypeRefresh,
						AtMinute: 20,
					},
					{
						Type:     domain.EventTypeExpire,
						AtMinute: 70,
					},
				},
			},
		},
		Invariants: []domain.Invariant{
			{
				Type:           domain.InvariantTypeMaxValidity,
				TargetObjectID: "session",
				MaximumMinutes: &maximumMinutes,
			},
		},
	}

	got, err := CandidateBoundaries(scenario)
	if err != nil {
		t.Fatalf("CandidateBoundaries() error = %v", err)
	}

	want := []int{0, 1, 19, 20, 21, 30, 50, 69, 70, 71, 90}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf(
			"CandidateBoundaries() = %v, want %v",
			got,
			want,
		)
	}
}

func TestCandidateBoundariesClipAdjacentMinutesToHorizon(t *testing.T) {
	t.Parallel()

	maximumMinutes := 90

	scenario := domain.Scenario{
		Name:           "Horizon boundaries",
		HorizonMinutes: 90,
		Objects: []domain.TimedObject{
			{
				ClientID: "token",
				Name:     "Access token",
				Kind:     domain.ObjectKindToken,
				Events: []domain.Event{
					{
						Type:     domain.EventTypeIssue,
						AtMinute: 0,
					},
					{
						Type:     domain.EventTypeExpire,
						AtMinute: 90,
					},
				},
			},
		},
		Invariants: []domain.Invariant{
			{
				Type:           domain.InvariantTypeMaxValidity,
				TargetObjectID: "token",
				MaximumMinutes: &maximumMinutes,
			},
		},
	}

	got, err := CandidateBoundaries(scenario)
	if err != nil {
		t.Fatalf("CandidateBoundaries() error = %v", err)
	}

	want := []int{0, 1, 89, 90}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf(
			"CandidateBoundaries() = %v, want %v",
			got,
			want,
		)
	}
}

func TestCandidateBoundariesRejectInvalidObjectReference(t *testing.T) {
	t.Parallel()

	scenario := canonicalBoundaryScenario()
	scenario.Invariants[0].SourceObjectID = "missing-source"

	_, err := CandidateBoundaries(scenario)
	if err == nil {
		t.Fatal("CandidateBoundaries() error = nil, want validation error")
	}

	var validationErrors domain.ValidationErrors
	if !errors.As(err, &validationErrors) {
		t.Fatalf(
			"CandidateBoundaries() error type = %T, want domain.ValidationErrors",
			err,
		)
	}

	if validationErrors.Empty() {
		t.Fatal("validation error contains no field errors")
	}
}

func TestCandidateBoundariesAreDeterministicAndDoNotMutateInput(
	t *testing.T,
) {
	t.Parallel()

	scenario := canonicalBoundaryScenario()
	original := canonicalBoundaryScenario()

	first, err := CandidateBoundaries(scenario)
	if err != nil {
		t.Fatalf("first CandidateBoundaries() error = %v", err)
	}

	for run := 0; run < 20; run++ {
		got, err := CandidateBoundaries(scenario)
		if err != nil {
			t.Fatalf(
				"run %d CandidateBoundaries() error = %v",
				run,
				err,
			)
		}

		if !reflect.DeepEqual(got, first) {
			t.Fatalf(
				"run %d boundaries = %v, want %v",
				run,
				got,
				first,
			)
		}
	}

	if !reflect.DeepEqual(scenario, original) {
		t.Fatalf(
			"CandidateBoundaries() mutated scenario:\ngot:  %#v\nwant: %#v",
			scenario,
			original,
		)
	}
}

func canonicalBoundaryScenario() domain.Scenario {
	graceMinutes := 5

	return domain.Scenario{
		Name:           "Subscription cancellation leak",
		Description:    "Premium cache survives subscription cancellation.",
		HorizonMinutes: 90,
		Objects: []domain.TimedObject{
			{
				ClientID: "subscription",
				Name:     "Premium subscription",
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
				ClientID: "premium-cache",
				Name:     "Premium entitlement cache",
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
				DependentObjectID: "premium-cache",
				GraceMinutes:      &graceMinutes,
			},
		},
	}
}

func TestCandidateBoundariesDeduplicatesSharedEventMinutes(t *testing.T) {
	t.Parallel()

	scenario := domain.Scenario{
		Name:           "shared event minutes",
		HorizonMinutes: 20,
		Objects: []domain.TimedObject{
			{
				ClientID: "source",
				Name:     "Source",
				Kind:     domain.ObjectKindSubscription,
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
				Name:     "Dependent",
				Kind:     domain.ObjectKindCachedEntitlement,
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
		},
		Invariants: []domain.Invariant{
			{
				Type:              domain.InvariantTypeDependentNotOutliveSource,
				SourceObjectID:    "source",
				DependentObjectID: "dependent",
			},
		},
	}

	got, err := CandidateBoundaries(scenario)
	if err != nil {
		t.Fatalf("CandidateBoundaries() error = %v", err)
	}

	want := []int{0, 1, 9, 10, 11, 20}

	if !reflect.DeepEqual(got, want) {
		t.Errorf(
			"CandidateBoundaries() = %v, want %v",
			got,
			want,
		)
	}

	seen := make(map[int]struct{}, len(got))
	for _, boundary := range got {
		if _, exists := seen[boundary]; exists {
			t.Errorf("duplicate boundary %d", boundary)
		}

		seen[boundary] = struct{}{}
	}
}
