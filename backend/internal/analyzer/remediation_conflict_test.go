package analyzer

import (
	"testing"

	"github.com/kernelKain/timetrap/backend/internal/domain"
)

func TestRemediationAvoidsOccupiedDependentMinute(t *testing.T) {
	grace := 5
	scenario := domain.Scenario{
		Name:           "repro",
		HorizonMinutes: 90,
		Objects: []domain.TimedObject{
			{ClientID: "src", Name: "Source", Kind: domain.ObjectKindSubscription, Events: []domain.Event{{Type: domain.EventTypeIssue, AtMinute: 0}, {Type: domain.EventTypeRevoke, AtMinute: 10}}},
			{ClientID: "dep", Name: "Dependent", Kind: domain.ObjectKindCachedEntitlement, Events: []domain.Event{{Type: domain.EventTypeIssue, AtMinute: 10}}},
		},
		Invariants: []domain.Invariant{{Type: domain.InvariantTypeRevokedAccessGrace, SourceObjectID: "src", DependentObjectID: "dep", GraceMinutes: &grace}},
	}
	analysis, err := Analyze(scenario)
	if err != nil {
		t.Fatalf("Analyze error: %v", err)
	}
	if analysis.Safe {
		t.Fatalf("expected unsafe")
	}
	rem := analysis.Violations[0].Remediation
	if len(rem.Operations) == 0 {
		t.Fatalf("expected an applicable remediation")
	}
	op := rem.Operations[0]
	for _, o := range scenario.Objects {
		if o.ClientID != op.ObjectID {
			continue
		}
		for _, ev := range o.Events {
			if ev.AtMinute == op.Event.AtMinute {
				t.Fatalf("suggested minute %d conflicts with existing event", op.Event.AtMinute)
			}
		}
	}
	// Apply the correction and re-analyze: must validate and be safe.
	fixed := scenario
	for i, o := range fixed.Objects {
		if o.ClientID == op.ObjectID {
			fixed.Objects[i].Events = append(append([]domain.Event(nil), o.Events...), *op.Event)
		}
	}
	if errs := fixed.Validate(); !errs.Empty() {
		t.Fatalf("corrected scenario invalid: %v", errs)
	}
	fixedAnalysis, err := Analyze(fixed)
	if err != nil {
		t.Fatalf("re-Analyze error: %v", err)
	}
	if !fixedAnalysis.Safe {
		t.Fatalf("corrected scenario still unsafe: %#v", fixedAnalysis.Violations)
	}
}
