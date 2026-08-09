package testfixture

import "github.com/kernelKain/timetrap/backend/internal/domain"

// UnsafeSubscriptionScenario returns a fresh canonical cancellation model.
func UnsafeSubscriptionScenario() domain.Scenario {
	graceMinutes := 5

	return domain.Scenario{
		Name:           "Subscription cancellation leak",
		Description:    "A premium entitlement cache remains valid after its source subscription is revoked.",
		HorizonMinutes: 90,
		Objects: []domain.TimedObject{
			{
				ClientID: "subscription",
				Name:     "Premium subscription",
				Kind:     domain.ObjectKindSubscription,
				Events: []domain.Event{
					{Type: domain.EventTypeIssue, AtMinute: 0},
					{Type: domain.EventTypeRevoke, AtMinute: 10},
				},
			},
			{
				ClientID: "premium-cache",
				Name:     "Premium entitlement cache",
				Kind:     domain.ObjectKindCachedEntitlement,
				Events: []domain.Event{
					{Type: domain.EventTypeIssue, AtMinute: 0},
					{Type: domain.EventTypeExpire, AtMinute: 60},
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

// CorrectedSubscriptionScenario returns a fresh canonical model with active
// dependent revocation while retaining the later expiry event.
func CorrectedSubscriptionScenario() domain.Scenario {
	scenario := UnsafeSubscriptionScenario()
	scenario.Name = "Subscription cancellation leak — corrected"
	scenario.Objects[1].Events = []domain.Event{
		{Type: domain.EventTypeIssue, AtMinute: 0},
		{Type: domain.EventTypeRevoke, AtMinute: 10},
		{Type: domain.EventTypeExpire, AtMinute: 60},
	}

	return scenario
}

// SafeDependentScenario returns a fresh model whose dependent expires with
// its source.
func SafeDependentScenario() domain.Scenario {
	scenario := UnsafeSubscriptionScenario()
	scenario.Name = "Dependent ends with source"
	scenario.Objects[1].Events[1] = domain.Event{
		Type:     domain.EventTypeExpire,
		AtMinute: 10,
	}
	scenario.Invariants[0] = domain.Invariant{
		Type:              domain.InvariantTypeDependentNotOutliveSource,
		SourceObjectID:    "subscription",
		DependentObjectID: "premium-cache",
	}

	return scenario
}
