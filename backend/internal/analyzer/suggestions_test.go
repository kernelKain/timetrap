package analyzer

import (
	"encoding/json"
	"testing"

	"github.com/kernelKain/timetrap/backend/internal/domain"
)

func TestRemediationForViolation(t *testing.T) {
	t.Parallel()

	graceMinutes := 5
	maximumMinutes := 30

	tests := []struct {
		name         string
		scenario     domain.Scenario
		violation    domain.Violation
		wantCode     string
		wantObjectID string
		wantMinute   int
	}{
		{
			name: "revocation grace",
			scenario: domain.Scenario{
				HorizonMinutes: 90,
				Objects: []domain.TimedObject{
					{
						ClientID: "subscription",
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
			},
			violation: domain.Violation{
				InvariantIndex:    0,
				InvariantType:     domain.InvariantTypeRevokedAccessGrace,
				StartMinute:       15,
				EndMinute:         60,
				DurationMinutes:   45,
				SourceObjectID:    "subscription",
				DependentObjectID: "premium-cache",
			},
			wantCode:     remediationCodeInvalidateDependentOnSourceRevoke,
			wantObjectID: "premium-cache",
			wantMinute:   10,
		},
		{
			name: "dependent outlives source",
			scenario: domain.Scenario{
				HorizonMinutes: 90,
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
								AtMinute: 20,
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
						Type:              domain.InvariantTypeDependentNotOutliveSource,
						SourceObjectID:    "source",
						DependentObjectID: "dependent",
					},
				},
			},
			violation: domain.Violation{
				InvariantIndex:    0,
				InvariantType:     domain.InvariantTypeDependentNotOutliveSource,
				StartMinute:       20,
				EndMinute:         90,
				DurationMinutes:   70,
				Ongoing:           true,
				SourceObjectID:    "source",
				DependentObjectID: "dependent",
			},
			wantCode:     remediationCodeAlignDependentWithSource,
			wantObjectID: "dependent",
			wantMinute:   20,
		},
		{
			name: "maximum validity",
			scenario: domain.Scenario{
				HorizonMinutes: 90,
				Objects: []domain.TimedObject{
					{
						ClientID: "session",
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
						Type:           domain.InvariantTypeMaxValidity,
						TargetObjectID: "session",
						MaximumMinutes: &maximumMinutes,
					},
				},
			},
			violation: domain.Violation{
				InvariantIndex:  0,
				InvariantType:   domain.InvariantTypeMaxValidity,
				StartMinute:     30,
				EndMinute:       90,
				DurationMinutes: 60,
				Ongoing:         true,
				TargetObjectID:  "session",
			},
			wantCode:     remediationCodeReduceMaximumValidity,
			wantObjectID: "session",
			wantMinute:   30,
		},
	}

	for _, test := range tests {
		test := test

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			remediation, err := remediationForViolation(
				test.scenario,
				test.violation,
			)
			if err != nil {
				t.Fatalf("remediationForViolation() error = %v", err)
			}

			if remediation.Code != test.wantCode {
				t.Errorf(
					"Code = %q, want %q",
					remediation.Code,
					test.wantCode,
				)
			}

			if remediation.ObjectID != test.wantObjectID {
				t.Errorf(
					"ObjectID = %q, want %q",
					remediation.ObjectID,
					test.wantObjectID,
				)
			}

			if remediation.SuggestedMinutes == nil {
				t.Fatal("SuggestedMinutes = nil, want value")
			}

			if *remediation.SuggestedMinutes != test.wantMinute {
				t.Errorf(
					"SuggestedMinutes = %d, want %d",
					*remediation.SuggestedMinutes,
					test.wantMinute,
				)
			}

			if remediation.Summary == "" {
				t.Error("Summary is empty")
			}
		})
	}
}

func TestAttachRemediationsDoesNotMutateInputs(t *testing.T) {
	t.Parallel()

	maximumMinutes := 30
	scenario := domain.Scenario{
		HorizonMinutes: 90,
		Objects: []domain.TimedObject{
			{
				ClientID: "session",
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
				Type:           domain.InvariantTypeMaxValidity,
				TargetObjectID: "session",
				MaximumMinutes: &maximumMinutes,
			},
		},
	}

	violations := []domain.Violation{
		{
			InvariantIndex:  0,
			InvariantType:   domain.InvariantTypeMaxValidity,
			StartMinute:     30,
			EndMinute:       90,
			DurationMinutes: 60,
			Ongoing:         true,
			TargetObjectID:  "session",
		},
	}

	scenarioBefore, err := json.Marshal(scenario)
	if err != nil {
		t.Fatalf("json.Marshal(scenario) error = %v", err)
	}

	result, err := attachRemediations(scenario, violations)
	if err != nil {
		t.Fatalf("attachRemediations() error = %v", err)
	}

	scenarioAfter, err := json.Marshal(scenario)
	if err != nil {
		t.Fatalf("json.Marshal(scenario) after call error = %v", err)
	}

	if string(scenarioAfter) != string(scenarioBefore) {
		t.Error("attachRemediations() mutated scenario")
	}

	if violations[0].Remediation.Code != "" {
		t.Errorf(
			"input remediation Code = %q, want empty",
			violations[0].Remediation.Code,
		)
	}

	if violations[0].Remediation.SuggestedMinutes != nil {
		t.Error("input SuggestedMinutes was modified")
	}

	if result[0].Remediation.Code !=
		remediationCodeReduceMaximumValidity {
		t.Errorf(
			"result remediation Code = %q, want %q",
			result[0].Remediation.Code,
			remediationCodeReduceMaximumValidity,
		)
	}

	result[0].StartMinute = 99
	if violations[0].StartMinute != 30 {
		t.Error("result shares violation slice backing storage with input")
	}
}

func TestRemediationForViolationErrors(t *testing.T) {
	t.Parallel()

	maximumMinutes := 30

	tests := []struct {
		name      string
		scenario  domain.Scenario
		violation domain.Violation
	}{
		{
			name:     "invariant index outside range",
			scenario: domain.Scenario{},
			violation: domain.Violation{
				InvariantIndex: 0,
				InvariantType:  domain.InvariantTypeMaxValidity,
			},
		},
		{
			name: "invariant type mismatch",
			scenario: domain.Scenario{
				Invariants: []domain.Invariant{
					{
						Type:           domain.InvariantTypeMaxValidity,
						TargetObjectID: "target",
						MaximumMinutes: &maximumMinutes,
					},
				},
			},
			violation: domain.Violation{
				InvariantIndex: 0,
				InvariantType:  domain.InvariantTypeRevokedAccessGrace,
			},
		},
		{
			name: "maximum missing duration",
			scenario: domain.Scenario{
				Invariants: []domain.Invariant{
					{
						Type:           domain.InvariantTypeMaxValidity,
						TargetObjectID: "target",
					},
				},
			},
			violation: domain.Violation{
				InvariantIndex: 0,
				InvariantType:  domain.InvariantTypeMaxValidity,
			},
		},
		{
			name: "missing source object",
			scenario: domain.Scenario{
				Invariants: []domain.Invariant{
					{
						Type:              domain.InvariantTypeDependentNotOutliveSource,
						SourceObjectID:    "missing",
						DependentObjectID: "dependent",
					},
				},
			},
			violation: domain.Violation{
				InvariantIndex: 0,
				InvariantType:  domain.InvariantTypeDependentNotOutliveSource,
				StartMinute:    10,
			},
		},
	}

	for _, test := range tests {
		test := test

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			_, err := remediationForViolation(
				test.scenario,
				test.violation,
			)
			if err == nil {
				t.Fatal("remediationForViolation() error = nil, want error")
			}
		})
	}
}
