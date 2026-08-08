package analyzer

import (
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	"github.com/kernelKain/timetrap/backend/internal/domain"
)

func TestAnalysisFromIntervalsSelectsEarliestDeterministically(t *testing.T) {
	t.Parallel()

	boundaries := []int{0, 10, 20, 30, 40, 90}
	violations := []domain.Violation{
		{
			InvariantIndex:  0,
			InvariantType:   domain.InvariantTypeMaxValidity,
			StartMinute:     20,
			EndMinute:       30,
			DurationMinutes: 10,
			TargetObjectID:  "later",
		},
		{
			InvariantIndex:    2,
			InvariantType:     domain.InvariantTypeRevokedAccessGrace,
			StartMinute:       10,
			EndMinute:         30,
			DurationMinutes:   20,
			SourceObjectID:    "source-z",
			DependentObjectID: "dependent-z",
		},
		{
			InvariantIndex:  1,
			InvariantType:   domain.InvariantTypeMaxValidity,
			StartMinute:     10,
			EndMinute:       40,
			DurationMinutes: 30,
			TargetObjectID:  "target-b",
		},
		{
			InvariantIndex:  1,
			InvariantType:   domain.InvariantTypeMaxValidity,
			StartMinute:     10,
			EndMinute:       30,
			DurationMinutes: 20,
			TargetObjectID:  "target-z",
		},
		{
			InvariantIndex:  1,
			InvariantType:   domain.InvariantTypeMaxValidity,
			StartMinute:     10,
			EndMinute:       30,
			DurationMinutes: 20,
			TargetObjectID:  "target-a",
		},
	}

	originalBoundaries := append([]int(nil), boundaries...)
	originalViolations := append([]domain.Violation(nil), violations...)

	analysis := analysisFromIntervals(boundaries, violations)

	if analysis.Safe {
		t.Fatal("Safe = true, want false")
	}

	if analysis.EarliestViolation == nil {
		t.Fatal("EarliestViolation = nil, want violation")
	}

	earliest := analysis.EarliestViolation

	if earliest.StartMinute != 10 {
		t.Errorf("StartMinute = %d, want 10", earliest.StartMinute)
	}

	if earliest.InvariantIndex != 1 {
		t.Errorf("InvariantIndex = %d, want 1", earliest.InvariantIndex)
	}

	if earliest.EndMinute != 30 {
		t.Errorf("EndMinute = %d, want 30", earliest.EndMinute)
	}

	if earliest.TargetObjectID != "target-a" {
		t.Errorf(
			"TargetObjectID = %q, want %q",
			earliest.TargetObjectID,
			"target-a",
		)
	}

	wantTargetOrder := []string{
		"target-a",
		"target-z",
		"target-b",
		"",
		"later",
	}

	for index, wantTargetID := range wantTargetOrder {
		if analysis.Violations[index].TargetObjectID != wantTargetID {
			t.Errorf(
				"Violations[%d].TargetObjectID = %q, want %q",
				index,
				analysis.Violations[index].TargetObjectID,
				wantTargetID,
			)
		}
	}

	if !reflect.DeepEqual(boundaries, originalBoundaries) {
		t.Errorf(
			"input boundaries mutated: got %v, want %v",
			boundaries,
			originalBoundaries,
		)
	}

	if !reflect.DeepEqual(violations, originalViolations) {
		t.Error("input violations were mutated")
	}
}

func TestAnalysisFromIntervalsReturnsSafeResult(t *testing.T) {
	t.Parallel()

	boundaries := []int{0, 1, 10, 90}

	analysis := analysisFromIntervals(
		boundaries,
		[]domain.Violation{},
	)

	if !analysis.Safe {
		t.Fatal("Safe = false, want true")
	}

	if analysis.EarliestViolation != nil {
		t.Errorf(
			"EarliestViolation = %+v, want nil",
			analysis.EarliestViolation,
		)
	}

	if len(analysis.Violations) != 0 {
		t.Errorf(
			"len(Violations) = %d, want 0",
			len(analysis.Violations),
		)
	}

	if analysis.Violations == nil {
		t.Error("Violations = nil, want non-nil empty slice")
	}

	if analysis.EvaluatedBoundaries != 4 {
		t.Errorf(
			"EvaluatedBoundaries = %d, want 4",
			analysis.EvaluatedBoundaries,
		)
	}

	if analysis.EngineVersion == "" {
		t.Error("EngineVersion is empty")
	}
}

func TestAnalysisFromIntervalsCopiesBoundaries(t *testing.T) {
	t.Parallel()

	boundaries := []int{0, 10, 90}

	analysis := analysisFromIntervals(
		boundaries,
		[]domain.Violation{},
	)

	boundaries[1] = 50

	if analysis.Boundaries[1] != 10 {
		t.Errorf(
			"analysis.Boundaries[1] = %d, want 10",
			analysis.Boundaries[1],
		)
	}
}

func TestBuildViolationIntervalsMergesAdjacentFailingSegments(t *testing.T) {
	t.Parallel()

	scenario := analyzerTestScenario(
		domain.EventTypeExpire,
		60,
	)

	boundaries := []int{
		0, 1, 9, 10, 11, 15, 59, 60, 61, 90,
	}

	violations, err := buildViolationIntervals(scenario, boundaries)
	if err != nil {
		t.Fatalf("buildViolationIntervals() error = %v", err)
	}

	if len(violations) != 1 {
		t.Fatalf("len(violations) = %d, want 1", len(violations))
	}

	violation := violations[0]

	if violation.InvariantIndex != 0 {
		t.Errorf(
			"InvariantIndex = %d, want 0",
			violation.InvariantIndex,
		)
	}

	if violation.InvariantType != domain.InvariantTypeRevokedAccessGrace {
		t.Errorf(
			"InvariantType = %q, want %q",
			violation.InvariantType,
			domain.InvariantTypeRevokedAccessGrace,
		)
	}

	if violation.StartMinute != 15 {
		t.Errorf("StartMinute = %d, want 15", violation.StartMinute)
	}

	if violation.EndMinute != 60 {
		t.Errorf("EndMinute = %d, want 60", violation.EndMinute)
	}

	if violation.DurationMinutes != 45 {
		t.Errorf(
			"DurationMinutes = %d, want 45",
			violation.DurationMinutes,
		)
	}

	if violation.Ongoing {
		t.Error("Ongoing = true, want false")
	}

	if violation.SourceObjectID != "subscription" {
		t.Errorf(
			"SourceObjectID = %q, want %q",
			violation.SourceObjectID,
			"subscription",
		)
	}

	if violation.DependentObjectID != "premium-cache" {
		t.Errorf(
			"DependentObjectID = %q, want %q",
			violation.DependentObjectID,
			"premium-cache",
		)
	}

	if len(violation.Evidence) != 1 {
		t.Fatalf(
			"len(Evidence) = %d, want 1",
			len(violation.Evidence),
		)
	}

	if violation.Evidence[0].AtMinute != 15 {
		t.Errorf(
			"Evidence[0].AtMinute = %d, want 15",
			violation.Evidence[0].AtMinute,
		)
	}
}

func TestBuildViolationIntervalsClosesActiveFindingAtHorizon(t *testing.T) {
	t.Parallel()

	scenario := analyzerTestScenario(
		domain.EventTypeExpire,
		60,
	)

	scenario.Objects[1].Events = []domain.Event{
		{
			Type:     domain.EventTypeIssue,
			AtMinute: 0,
		},
	}

	boundaries := []int{0, 1, 9, 10, 11, 15, 90}

	violations, err := buildViolationIntervals(scenario, boundaries)
	if err != nil {
		t.Fatalf("buildViolationIntervals() error = %v", err)
	}

	if len(violations) != 1 {
		t.Fatalf("len(violations) = %d, want 1", len(violations))
	}

	violation := violations[0]

	if violation.StartMinute != 15 {
		t.Errorf("StartMinute = %d, want 15", violation.StartMinute)
	}

	if violation.EndMinute != 90 {
		t.Errorf("EndMinute = %d, want 90", violation.EndMinute)
	}

	if violation.DurationMinutes != 75 {
		t.Errorf(
			"DurationMinutes = %d, want 75",
			violation.DurationMinutes,
		)
	}

	if !violation.Ongoing {
		t.Error("Ongoing = false, want true")
	}
}

func TestBuildViolationIntervalsReopensAfterPassingSegment(t *testing.T) {
	t.Parallel()

	graceMinutes := 0
	scenario := domain.Scenario{
		Name:           "reopened interval scenario",
		HorizonMinutes: 40,
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
						Type:     domain.EventTypeRevoke,
						AtMinute: 10,
					},
					{
						Type:     domain.EventTypeIssue,
						AtMinute: 20,
					},
					{
						Type:     domain.EventTypeRevoke,
						AtMinute: 30,
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

	boundaries := []int{0, 10, 20, 30, 40}

	violations, err := buildViolationIntervals(scenario, boundaries)
	if err != nil {
		t.Fatalf("buildViolationIntervals() error = %v", err)
	}

	if len(violations) != 2 {
		t.Fatalf("len(violations) = %d, want 2", len(violations))
	}

	first := violations[0]
	if first.StartMinute != 10 ||
		first.EndMinute != 20 ||
		first.DurationMinutes != 10 ||
		first.Ongoing {
		t.Errorf("first violation = %+v, want [10,20) closed", first)
	}

	second := violations[1]
	if second.StartMinute != 30 ||
		second.EndMinute != 40 ||
		second.DurationMinutes != 10 ||
		!second.Ongoing {
		t.Errorf("second violation = %+v, want [30,40) ongoing", second)
	}
}

func TestBuildViolationIntervalsRejectsInvalidBoundaries(t *testing.T) {
	t.Parallel()

	scenario := domain.Scenario{
		HorizonMinutes: 90,
	}

	tests := []struct {
		name       string
		boundaries []int
	}{
		{
			name:       "fewer than two",
			boundaries: []int{0},
		},
		{
			name:       "does not begin at zero",
			boundaries: []int{1, 90},
		},
		{
			name:       "does not end at horizon",
			boundaries: []int{0, 89},
		},
		{
			name:       "duplicate boundary",
			boundaries: []int{0, 10, 10, 90},
		},
		{
			name:       "descending boundary",
			boundaries: []int{0, 20, 10, 90},
		},
	}

	for _, testCase := range tests {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			_, err := buildViolationIntervals(
				scenario,
				testCase.boundaries,
			)
			if err == nil {
				t.Fatal(
					"buildViolationIntervals() error = nil, want error",
				)
			}
		})
	}
}

func TestAnalyzeCorrectedScenarioIsSafe(t *testing.T) {
	t.Parallel()

	scenario := requiredDemoScenario(
		domain.EventTypeRevoke,
		10,
	)

	analysis, err := Analyze(scenario)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}

	if !analysis.Safe {
		t.Fatal("Safe = false, want true")
	}

	if analysis.EarliestViolation != nil {
		t.Errorf(
			"EarliestViolation = %+v, want nil",
			analysis.EarliestViolation,
		)
	}

	if len(analysis.Violations) != 0 {
		t.Errorf(
			"len(Violations) = %d, want 0",
			len(analysis.Violations),
		)
	}

	if analysis.Violations == nil {
		t.Error("Violations = nil, want non-nil empty slice")
	}

	if analysis.Boundaries == nil {
		t.Error("Boundaries = nil, want allocated slice")
	}
}

func TestAnalyzeEventAtHorizonCannotCreateViolation(t *testing.T) {
	t.Parallel()

	graceMinutes := 0
	scenario := domain.Scenario{
		Name:           "event at horizon",
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
						AtMinute: 90,
					},
				},
			},
			{
				ClientID: "premium-cache",
				Name:     "Premium cache",
				Kind:     domain.ObjectKindCachedEntitlement,
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
				SourceObjectID:    "subscription",
				DependentObjectID: "premium-cache",
				GraceMinutes:      &graceMinutes,
			},
		},
	}

	analysis, err := Analyze(scenario)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}

	if !analysis.Safe {
		t.Fatalf(
			"Safe = false, violations = %+v",
			analysis.Violations,
		)
	}

	if len(analysis.Violations) != 0 {
		t.Errorf(
			"len(Violations) = %d, want 0",
			len(analysis.Violations),
		)
	}
}

func TestAnalyzeRejectsInvalidScenario(t *testing.T) {
	t.Parallel()

	analysis, err := Analyze(domain.Scenario{})
	if err == nil {
		t.Fatal("Analyze() error = nil, want validation error")
	}

	if analysis.Safe {
		t.Error("invalid scenario returned Safe = true")
	}

	var validationErrors domain.ValidationErrors
	if !errors.As(err, &validationErrors) {
		t.Fatalf(
			"Analyze() error type = %T, want domain.ValidationErrors",
			err,
		)
	}

	if validationErrors.Empty() {
		t.Error("validation error collection is empty")
	}
}

func TestAnalyzeIsDeterministicAndDoesNotMutateScenario(t *testing.T) {
	t.Parallel()

	scenario := analyzerTestScenario(
		domain.EventTypeExpire,
		60,
	)

	scenarioBefore, err := json.Marshal(scenario)
	if err != nil {
		t.Fatalf("json.Marshal(scenario) error = %v", err)
	}

	first, err := Analyze(scenario)
	if err != nil {
		t.Fatalf("first Analyze() error = %v", err)
	}

	second, err := Analyze(scenario)
	if err != nil {
		t.Fatalf("second Analyze() error = %v", err)
	}

	if !reflect.DeepEqual(first, second) {
		t.Errorf(
			"repeated analyses differ:\nfirst:  %+v\nsecond: %+v",
			first,
			second,
		)
	}

	scenarioAfter, err := json.Marshal(scenario)
	if err != nil {
		t.Fatalf(
			"json.Marshal(scenario) after analysis error = %v",
			err,
		)
	}

	if string(scenarioAfter) != string(scenarioBefore) {
		t.Error("Analyze() mutated the submitted scenario")
	}

	first.Boundaries[0] = 999
	first.Violations[0].StartMinute = 999

	if second.Boundaries[0] != 0 {
		t.Error("separate Analyze() calls share boundary storage")
	}

	if second.Violations[0].StartMinute != 15 {
		t.Errorf(
			"second.Violations[0].StartMinute = %d, want 15",
			second.Violations[0].StartMinute,
		)
	}
}

func requiredDemoScenario(
	dependentEndType domain.EventType,
	dependentEndMinute int,
) domain.Scenario {
	scenario := analyzerTestScenario(
		dependentEndType,
		dependentEndMinute,
	)

	graceInvariant := scenario.Invariants[0]

	scenario.Invariants = []domain.Invariant{
		{
			Type:              domain.InvariantTypeDependentNotOutliveSource,
			SourceObjectID:    "subscription",
			DependentObjectID: "premium-cache",
		},
		graceInvariant,
	}

	return scenario
}

func analyzerTestScenario(
	dependentEndType domain.EventType,
	dependentEndMinute int,
) domain.Scenario {
	graceMinutes := 5

	return domain.Scenario{
		Name:           "canonical subscription scenario",
		Description:    "Subscription revocation and cached entitlement",
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
				ClientID: "premium-cache",
				Name:     "Premium cache",
				Kind:     domain.ObjectKindCachedEntitlement,
				Events: []domain.Event{
					{
						Type:     domain.EventTypeIssue,
						AtMinute: 0,
					},
					{
						Type:     dependentEndType,
						AtMinute: dependentEndMinute,
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

func TestAnalyzeCanonicalScenario(t *testing.T) {
	t.Parallel()

	scenario := requiredDemoScenario(
		domain.EventTypeExpire,
		60,
	)

	analysis, err := Analyze(scenario)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}

	if analysis.Safe {
		t.Fatal("Safe = true, want false")
	}

	wantBoundaries := []int{
		0, 1, 9, 10, 11, 15, 59, 60, 61, 90,
	}

	if !reflect.DeepEqual(analysis.Boundaries, wantBoundaries) {
		t.Errorf(
			"Boundaries = %v, want %v",
			analysis.Boundaries,
			wantBoundaries,
		)
	}

	if analysis.EvaluatedBoundaries != len(wantBoundaries) {
		t.Errorf(
			"EvaluatedBoundaries = %d, want %d",
			analysis.EvaluatedBoundaries,
			len(wantBoundaries),
		)
	}

	if len(analysis.Violations) != 2 {
		t.Fatalf(
			"len(Violations) = %d, want 2",
			len(analysis.Violations),
		)
	}

	staleExposure := analysis.Violations[0]

	if staleExposure.InvariantIndex != 0 {
		t.Errorf(
			"stale exposure InvariantIndex = %d, want 0",
			staleExposure.InvariantIndex,
		)
	}

	if staleExposure.InvariantType !=
		domain.InvariantTypeDependentNotOutliveSource {
		t.Errorf(
			"stale exposure InvariantType = %q, want %q",
			staleExposure.InvariantType,
			domain.InvariantTypeDependentNotOutliveSource,
		)
	}

	if staleExposure.StartMinute != 10 {
		t.Errorf(
			"stale exposure StartMinute = %d, want 10",
			staleExposure.StartMinute,
		)
	}

	if staleExposure.EndMinute != 60 {
		t.Errorf(
			"stale exposure EndMinute = %d, want 60",
			staleExposure.EndMinute,
		)
	}

	if staleExposure.DurationMinutes != 50 {
		t.Errorf(
			"stale exposure DurationMinutes = %d, want 50",
			staleExposure.DurationMinutes,
		)
	}

	if staleExposure.Ongoing {
		t.Error("stale exposure Ongoing = true, want false")
	}

	if staleExposure.Remediation.Code !=
		remediationCodeAlignDependentWithSource {
		t.Errorf(
			"stale exposure remediation Code = %q, want %q",
			staleExposure.Remediation.Code,
			remediationCodeAlignDependentWithSource,
		)
	}

	if staleExposure.Remediation.SuggestedMinutes == nil {
		t.Fatal(
			"stale exposure SuggestedMinutes = nil, want 10",
		)
	}

	if *staleExposure.Remediation.SuggestedMinutes != 10 {
		t.Errorf(
			"stale exposure SuggestedMinutes = %d, want 10",
			*staleExposure.Remediation.SuggestedMinutes,
		)
	}

	policyViolation := analysis.Violations[1]

	if policyViolation.InvariantIndex != 1 {
		t.Errorf(
			"policy violation InvariantIndex = %d, want 1",
			policyViolation.InvariantIndex,
		)
	}

	if policyViolation.InvariantType !=
		domain.InvariantTypeRevokedAccessGrace {
		t.Errorf(
			"policy violation InvariantType = %q, want %q",
			policyViolation.InvariantType,
			domain.InvariantTypeRevokedAccessGrace,
		)
	}

	if policyViolation.StartMinute != 15 {
		t.Errorf(
			"policy violation StartMinute = %d, want 15",
			policyViolation.StartMinute,
		)
	}

	if policyViolation.EndMinute != 60 {
		t.Errorf(
			"policy violation EndMinute = %d, want 60",
			policyViolation.EndMinute,
		)
	}

	if policyViolation.DurationMinutes != 45 {
		t.Errorf(
			"policy violation DurationMinutes = %d, want 45",
			policyViolation.DurationMinutes,
		)
	}

	if policyViolation.Ongoing {
		t.Error("policy violation Ongoing = true, want false")
	}

	if policyViolation.Remediation.Code !=
		remediationCodeInvalidateDependentOnSourceRevoke {
		t.Errorf(
			"policy violation remediation Code = %q, want %q",
			policyViolation.Remediation.Code,
			remediationCodeInvalidateDependentOnSourceRevoke,
		)
	}

	if policyViolation.Remediation.SuggestedMinutes == nil {
		t.Fatal(
			"policy violation SuggestedMinutes = nil, want 10",
		)
	}

	if *policyViolation.Remediation.SuggestedMinutes != 10 {
		t.Errorf(
			"policy violation SuggestedMinutes = %d, want 10",
			*policyViolation.Remediation.SuggestedMinutes,
		)
	}

	if analysis.EarliestViolation == nil {
		t.Fatal("EarliestViolation = nil, want stale exposure")
	}

	if !reflect.DeepEqual(
		*analysis.EarliestViolation,
		staleExposure,
	) {
		t.Errorf(
			"EarliestViolation = %+v, want %+v",
			*analysis.EarliestViolation,
			staleExposure,
		)
	}

	if analysis.EngineVersion != engineVersion {
		t.Errorf(
			"EngineVersion = %q, want %q",
			analysis.EngineVersion,
			engineVersion,
		)
	}
}

func TestAnalyzeRequiredEdgeCases(t *testing.T) {
	t.Parallel()

	eventAtZero := 0

	tests := []struct {
		name                    string
		scenario                domain.Scenario
		wantSafe                bool
		wantViolationCount      int
		wantInvariantType       domain.InvariantType
		wantStart               int
		wantEnd                 int
		wantDuration            int
		wantOngoing             bool
		wantEvidenceEventMinute *int
	}{
		{
			name: "safe scenario",
			scenario: requiredDependentScenario(
				30,
				20,
				40,
			),
			wantSafe:           true,
			wantViolationCount: 0,
		},
		{
			name: "dependent expires too late",
			scenario: requiredDependentScenario(
				20,
				30,
				40,
			),
			wantSafe:           false,
			wantViolationCount: 1,
			wantInvariantType:  domain.InvariantTypeDependentNotOutliveSource,
			wantStart:          20,
			wantEnd:            30,
			wantDuration:       10,
			wantOngoing:        false,
		},
		{
			name: "source and dependent expire together",
			scenario: requiredDependentScenario(
				20,
				20,
				40,
			),
			wantSafe:           true,
			wantViolationCount: 0,
		},
		{
			name: "violation continues to horizon",
			scenario: requiredDependentScenario(
				20,
				-1,
				40,
			),
			wantSafe:           false,
			wantViolationCount: 1,
			wantInvariantType:  domain.InvariantTypeDependentNotOutliveSource,
			wantStart:          20,
			wantEnd:            40,
			wantDuration:       20,
			wantOngoing:        true,
		},
		{
			name: "zero minute grace",
			scenario: requiredGraceScenario(
				10,
				20,
				0,
				30,
			),
			wantSafe:           false,
			wantViolationCount: 1,
			wantInvariantType:  domain.InvariantTypeRevokedAccessGrace,
			wantStart:          10,
			wantEnd:            20,
			wantDuration:       10,
			wantOngoing:        false,
		},
		{
			name: "event at zero takes effect immediately",
			scenario: requiredMaximumScenario(
				[]domain.Event{
					{
						Type:     domain.EventTypeIssue,
						AtMinute: 0,
					},
					{
						Type:     domain.EventTypeExpire,
						AtMinute: 20,
					},
				},
				10,
				30,
			),
			wantSafe:                false,
			wantViolationCount:      1,
			wantInvariantType:       domain.InvariantTypeMaxValidity,
			wantStart:               10,
			wantEnd:                 20,
			wantDuration:            10,
			wantOngoing:             false,
			wantEvidenceEventMinute: &eventAtZero,
		},
		{
			name: "refresh resets maximum validity clock",
			scenario: requiredMaximumScenario(
				[]domain.Event{
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
						AtMinute: 60,
					},
				},
				30,
				70,
			),
			wantSafe:           false,
			wantViolationCount: 1,
			wantInvariantType:  domain.InvariantTypeMaxValidity,
			wantStart:          50,
			wantEnd:            60,
			wantDuration:       10,
			wantOngoing:        false,
		},
	}

	for _, testCase := range tests {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			analysis, err := Analyze(testCase.scenario)
			if err != nil {
				t.Fatalf("Analyze() error = %v", err)
			}

			if analysis.Safe != testCase.wantSafe {
				t.Errorf(
					"Safe = %t, want %t",
					analysis.Safe,
					testCase.wantSafe,
				)
			}

			if len(analysis.Violations) !=
				testCase.wantViolationCount {
				t.Fatalf(
					"len(Violations) = %d, want %d",
					len(analysis.Violations),
					testCase.wantViolationCount,
				)
			}

			if testCase.wantViolationCount == 0 {
				if analysis.EarliestViolation != nil {
					t.Errorf(
						"EarliestViolation = %+v, want nil",
						analysis.EarliestViolation,
					)
				}

				return
			}

			violation := analysis.Violations[0]

			if violation.InvariantType != testCase.wantInvariantType {
				t.Errorf(
					"InvariantType = %q, want %q",
					violation.InvariantType,
					testCase.wantInvariantType,
				)
			}

			if violation.StartMinute != testCase.wantStart {
				t.Errorf(
					"StartMinute = %d, want %d",
					violation.StartMinute,
					testCase.wantStart,
				)
			}

			if violation.EndMinute != testCase.wantEnd {
				t.Errorf(
					"EndMinute = %d, want %d",
					violation.EndMinute,
					testCase.wantEnd,
				)
			}

			if violation.DurationMinutes != testCase.wantDuration {
				t.Errorf(
					"DurationMinutes = %d, want %d",
					violation.DurationMinutes,
					testCase.wantDuration,
				)
			}

			if violation.Ongoing != testCase.wantOngoing {
				t.Errorf(
					"Ongoing = %t, want %t",
					violation.Ongoing,
					testCase.wantOngoing,
				)
			}

			if analysis.EarliestViolation == nil {
				t.Fatal(
					"EarliestViolation = nil, want violation",
				)
			}

			if !reflect.DeepEqual(
				*analysis.EarliestViolation,
				violation,
			) {
				t.Errorf(
					"EarliestViolation = %+v, want %+v",
					*analysis.EarliestViolation,
					violation,
				)
			}

			if testCase.wantEvidenceEventMinute != nil {
				if len(violation.Evidence) == 0 {
					t.Fatal("Evidence is empty")
				}

				eventMinute :=
					violation.Evidence[0].EventMinute

				if eventMinute == nil {
					t.Fatal(
						"Evidence.EventMinute = nil",
					)
				}

				if *eventMinute !=
					*testCase.wantEvidenceEventMinute {
					t.Errorf(
						"Evidence.EventMinute = %d, want %d",
						*eventMinute,
						*testCase.wantEvidenceEventMinute,
					)
				}
			}
		})
	}
}

func TestAnalyzeInvalidObjectReferenceReturnsError(t *testing.T) {
	t.Parallel()

	scenario := requiredDependentScenario(
		20,
		30,
		40,
	)
	scenario.Invariants[0].SourceObjectID = "missing-source"

	analysis, err := Analyze(scenario)
	if err == nil {
		t.Fatal("Analyze() error = nil, want error")
	}

	if analysis.Safe {
		t.Error("invalid object reference returned Safe = true")
	}

	if len(analysis.Violations) != 0 {
		t.Errorf(
			"len(Violations) = %d, want 0",
			len(analysis.Violations),
		)
	}

	var validationErrors domain.ValidationErrors
	if !errors.As(err, &validationErrors) {
		t.Fatalf(
			"Analyze() error type = %T, want domain.ValidationErrors",
			err,
		)
	}

	if validationErrors.Empty() {
		t.Error("validation error collection is empty")
	}
}

func requiredDependentScenario(
	sourceExpiration int,
	dependentExpiration int,
	horizon int,
) domain.Scenario {
	dependentEvents := []domain.Event{
		{
			Type:     domain.EventTypeIssue,
			AtMinute: 0,
		},
	}

	if dependentExpiration >= 0 {
		dependentEvents = append(
			dependentEvents,
			domain.Event{
				Type:     domain.EventTypeExpire,
				AtMinute: dependentExpiration,
			},
		)
	}

	return domain.Scenario{
		Name:           "dependent lifetime scenario",
		HorizonMinutes: horizon,
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
						AtMinute: sourceExpiration,
					},
				},
			},
			{
				ClientID: "dependent",
				Name:     "Dependent",
				Kind:     domain.ObjectKindCachedEntitlement,
				Events:   dependentEvents,
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
}

func requiredGraceScenario(
	sourceRevocation int,
	dependentExpiration int,
	graceMinutes int,
	horizon int,
) domain.Scenario {
	return domain.Scenario{
		Name:           "revocation grace scenario",
		HorizonMinutes: horizon,
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
						Type:     domain.EventTypeRevoke,
						AtMinute: sourceRevocation,
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
						AtMinute: dependentExpiration,
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
}

func requiredMaximumScenario(
	events []domain.Event,
	maximumMinutes int,
	horizon int,
) domain.Scenario {
	resultEvents := make([]domain.Event, len(events))
	copy(resultEvents, events)

	return domain.Scenario{
		Name:           "maximum validity scenario",
		HorizonMinutes: horizon,
		Objects: []domain.TimedObject{
			{
				ClientID: "target",
				Name:     "Target",
				Kind:     domain.ObjectKindSession,
				Events:   resultEvents,
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
}
