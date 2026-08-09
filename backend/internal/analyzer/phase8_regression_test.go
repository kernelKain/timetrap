package analyzer

import (
	"reflect"
	"testing"

	"github.com/kernelKain/timetrap/backend/internal/domain"
	"github.com/kernelKain/timetrap/backend/internal/testfixture"
)

func TestAnalyze_CanonicalSubscriptionCancellation(t *testing.T) {
	scenario := testfixture.UnsafeSubscriptionScenario()
	before := testfixture.UnsafeSubscriptionScenario()

	analysis, err := Analyze(scenario)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if analysis.Safe {
		t.Fatal("Safe = true, want false")
	}
	if !reflect.DeepEqual(analysis.Boundaries, []int{0, 1, 9, 10, 11, 15, 59, 60, 61, 90}) {
		t.Fatalf("Boundaries = %v, want exact canonical boundaries", analysis.Boundaries)
	}
	if analysis.EvaluatedBoundaries != 10 {
		t.Fatalf("EvaluatedBoundaries = %d, want 10", analysis.EvaluatedBoundaries)
	}
	if len(analysis.Violations) != 1 || analysis.EarliestViolation == nil {
		t.Fatalf("Violations = %#v, EarliestViolation = %#v, want one earliest violation", analysis.Violations, analysis.EarliestViolation)
	}

	violation := analysis.Violations[0]
	if violation.InvariantType != domain.InvariantTypeRevokedAccessGrace ||
		violation.SourceObjectID != "subscription" || violation.DependentObjectID != "premium-cache" ||
		violation.StartMinute != 15 || violation.EndMinute != 60 || violation.DurationMinutes != 45 || violation.Ongoing {
		t.Fatalf("canonical violation = %#v, want revoked-access grace [15,60) lasting 45 minutes", violation)
	}
	assertMinutePointer(t, "SourceInvalidatedMinute", violation.SourceInvalidatedMinute, 10)
	assertMinutePointer(t, "PolicyDeadlineMinute", violation.PolicyDeadlineMinute, 15)
	assertMinutePointer(t, "DependentInvalidatedMinute", violation.DependentInvalidatedMinute, 60)
	assertMinutePointer(t, "GraceMinutes", violation.GraceMinutes, 5)
	assertMinutePointer(t, "TotalStaleExposureMinutes", violation.TotalStaleExposureMinutes, 50)
	if !reflect.DeepEqual(*analysis.EarliestViolation, violation) {
		t.Fatalf("EarliestViolation = %#v, want %#v", *analysis.EarliestViolation, violation)
	}
	if !reflect.DeepEqual(scenario, before) {
		t.Fatal("Analyze() mutated the canonical input")
	}
}

func TestAnalyze_CanonicalCorrectionIsSafe(t *testing.T) {
	var first domain.Analysis
	for iteration := 0; iteration < 20; iteration++ {
		scenario := testfixture.CorrectedSubscriptionScenario()
		before := testfixture.CorrectedSubscriptionScenario()
		analysis, err := Analyze(scenario)
		if err != nil {
			t.Fatalf("iteration %d: Analyze() error = %v", iteration, err)
		}
		if !analysis.Safe || len(analysis.Violations) != 0 || analysis.EarliestViolation != nil {
			t.Fatalf("iteration %d: corrected analysis = %#v, want safe with no violations", iteration, analysis)
		}
		if !reflect.DeepEqual(scenario, before) {
			t.Fatalf("iteration %d: Analyze() mutated the corrected input", iteration)
		}
		if iteration == 0 {
			first = analysis
		} else if !reflect.DeepEqual(analysis, first) {
			t.Fatalf("iteration %d: result differs from first result", iteration)
		}
	}
}

func TestScenarioFixturesReturnIndependentValues(t *testing.T) {
	first := testfixture.UnsafeSubscriptionScenario()
	second := testfixture.UnsafeSubscriptionScenario()
	first.Objects[0].Events[0].AtMinute = 7
	*first.Invariants[0].GraceMinutes = 9
	if second.Objects[0].Events[0].AtMinute != 0 || *second.Invariants[0].GraceMinutes != 5 {
		t.Fatal("scenario fixtures share mutable storage")
	}
}

func assertMinutePointer(t *testing.T, name string, got *int, want int) {
	t.Helper()
	if got == nil || *got != want {
		t.Fatalf("%s = %v, want %d", name, got, want)
	}
}
