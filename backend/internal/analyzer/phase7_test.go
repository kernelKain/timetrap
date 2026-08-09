package analyzer

import (
	"encoding/json"
	"testing"

	"github.com/kernelKain/timetrap/backend/internal/domain"
)

func TestCanonicalAnalysisIncludesTimelineAndStructuredRemediation(t *testing.T) {
	t.Parallel()
	scenario := analyzerTestScenario(domain.EventTypeExpire, 60)

	analysis, err := Analyze(scenario)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if analysis.Safe || analysis.EarliestViolation == nil {
		t.Fatal("canonical analysis did not return an unsafe finding")
	}
	if len(analysis.Timeline) != 2 || len(analysis.Timeline[1].Intervals) != 1 {
		t.Fatalf("Timeline = %#v, want two lanes and one dependent interval", analysis.Timeline)
	}
	interval := analysis.Timeline[1].Intervals[0]
	if interval.StartMinute != 0 || interval.EndMinute != 60 {
		t.Fatalf("dependent interval = [%d,%d), want [0,60)", interval.StartMinute, interval.EndMinute)
	}

	finding := analysis.EarliestViolation
	if finding.SourceInvalidatedMinute == nil || *finding.SourceInvalidatedMinute != 10 ||
		finding.PolicyDeadlineMinute == nil || *finding.PolicyDeadlineMinute != 15 ||
		finding.DependentInvalidatedMinute == nil || *finding.DependentInvalidatedMinute != 60 ||
		finding.TotalStaleExposureMinutes == nil || *finding.TotalStaleExposureMinutes != 50 {
		t.Fatalf("canonical finding details = %#v", finding)
	}
	if len(finding.Remediation.Operations) != 1 {
		t.Fatalf("operations = %#v, want one", finding.Remediation.Operations)
	}
	operation := finding.Remediation.Operations[0]
	if operation.Type != domain.RemediationOperationAddEvent || operation.ObjectID != "premium-cache" ||
		operation.Event == nil || operation.Event.Type != domain.EventTypeRevoke || operation.Event.AtMinute != 10 {
		t.Fatalf("operation = %#v, want canonical revoke", operation)
	}
}

func TestOldAnalysisJSONRemainsReadable(t *testing.T) {
	t.Parallel()
	var result domain.Analysis
	if err := json.Unmarshal([]byte(`{"safe":true,"evaluatedBoundaries":2,"boundaries":[0,90],"violations":[],"engineVersion":"v1"}`), &result); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if !result.Safe || result.Timeline != nil {
		t.Fatalf("old result decoded unexpectedly: %#v", result)
	}
}

func TestStructuredAnalysisJSONRoundTrip(t *testing.T) {
	t.Parallel()
	scenario := analyzerTestScenario(domain.EventTypeExpire, 60)
	analysis, err := Analyze(scenario)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	encoded, err := json.Marshal(analysis)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	var decoded domain.Analysis
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if decoded.EarliestViolation == nil || len(decoded.EarliestViolation.Remediation.Operations) != 1 || len(decoded.Timeline) != 2 {
		t.Fatalf("round-tripped result lost structured fields: %#v", decoded)
	}
}
