package domain

import (
	"fmt"
	"testing"
)

func TestScenarioValidate(t *testing.T) {
	tests := []struct {
		name          string
		modify        func(*Scenario)
		expectedField string
	}{
		{
			name:          "valid subscription scenario",
			expectedField: "",
		},
		{
			name: "empty objects",
			modify: func(scenario *Scenario) {
				scenario.Objects = nil
			},
			expectedField: "objects",
		},
		{
			name: "event outside horizon",
			modify: func(scenario *Scenario) {
				scenario.Objects[0].Events[1].AtMinute =
					scenario.HorizonMinutes + 1
			},
			expectedField: "objects[0].events[1].atMinute",
		},
		{
			name: "unknown object kind",
			modify: func(scenario *Scenario) {
				scenario.Objects[0].Kind = ObjectKind("unknown_kind")
			},
			expectedField: "objects[0].kind",
		},
		{
			name: "unknown event type",
			modify: func(scenario *Scenario) {
				scenario.Objects[0].Events[0].Type =
					EventType("unknown_event")
			},
			expectedField: "objects[0].events[0].type",
		},
		{
			name: "missing invariant source",
			modify: func(scenario *Scenario) {
				scenario.Invariants[0].SourceObjectID = ""
			},
			expectedField: "invariants[0].sourceObjectId",
		},
		{
			name: "missing dependent reference",
			modify: func(scenario *Scenario) {
				scenario.Invariants[0].DependentObjectID = ""
			},
			expectedField: "invariants[0].dependentObjectId",
		},
		{
			name: "unknown referenced object",
			modify: func(scenario *Scenario) {
				scenario.Invariants[0].SourceObjectID = "missing-object"
			},
			expectedField: "invariants[0].sourceObjectId",
		},
		{
			name: "same source and dependent",
			modify: func(scenario *Scenario) {
				scenario.Invariants[0].DependentObjectID =
					scenario.Invariants[0].SourceObjectID
			},
			expectedField: "invariants[0].dependentObjectId",
		},
		{
			name: "two events on same object at same minute",
			modify: func(scenario *Scenario) {
				scenario.Objects[0].Events[1].AtMinute =
					scenario.Objects[0].Events[0].AtMinute
			},
			expectedField: "objects[0].events[1].atMinute",
		},
		{
			name: "duplicate client IDs",
			modify: func(scenario *Scenario) {
				scenario.Objects[1].ClientID =
					scenario.Objects[0].ClientID
			},
			expectedField: "objects[1].clientId",
		},
		{
			name: "empty invariants",
			modify: func(scenario *Scenario) {
				scenario.Invariants = nil
			},
			expectedField: "invariants",
		},
		{
			name: "zero minute grace accepted",
			modify: func(scenario *Scenario) {
				scenario.Invariants[0].GraceMinutes = intPointer(0)
			},
			expectedField: "",
		},
		{
			name: "excessive object count",
			modify: func(scenario *Scenario) {
				for index := len(scenario.Objects); index <= MaxObjects; index++ {
					scenario.Objects = append(
						scenario.Objects,
						TimedObject{
							ClientID: fmt.Sprintf("object-%d", index),
							Name:     fmt.Sprintf("Object %d", index),
							Kind:     ObjectKindToken,
							Events: []Event{
								{
									Type:     EventTypeIssue,
									AtMinute: 0,
								},
							},
						},
					)
				}
			},
			expectedField: "objects",
		},
		{
			name: "excessive events on one object",
			modify: func(scenario *Scenario) {
				scenario.Objects[0].Events =
					makeEvents(MaxEventsPerObject + 1)
			},
			expectedField: "objects[0].events",
		},
		{
			name: "excessive total event count",
			modify: func(scenario *Scenario) {
				scenario.Objects = make([]TimedObject, 6)

				for index := range scenario.Objects {
					clientID := fmt.Sprintf("object-%d", index)

					if index == 0 {
						clientID = "subscription"
					}

					if index == 1 {
						clientID = "premium-cache"
					}

					scenario.Objects[index] = TimedObject{
						ClientID: clientID,
						Name:     fmt.Sprintf("Object %d", index),
						Kind:     ObjectKindToken,
						Events:   makeEvents(MaxEventsPerObject),
					}
				}
			},
			expectedField: "objects",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			scenario := validScenario()

			if test.modify != nil {
				test.modify(&scenario)
			}

			validationErrors := scenario.Validate()

			if test.expectedField == "" {
				if !validationErrors.Empty() {
					t.Fatalf(
						"expected valid scenario, got errors: %#v",
						validationErrors,
					)
				}

				return
			}

			if validationErrors.Empty() {
				t.Fatalf(
					"expected validation error for field %q",
					test.expectedField,
				)
			}

			if !containsField(validationErrors, test.expectedField) {
				t.Fatalf(
					"expected error for field %q, got %#v",
					test.expectedField,
					validationErrors,
				)
			}
		})
	}
}

func validScenario() Scenario {
	return Scenario{
		Name:           "Subscription cancellation leak",
		Description:    "Premium cache survives cancellation.",
		HorizonMinutes: 90,
		Objects: []TimedObject{
			{
				ClientID: "subscription",
				Name:     "Premium subscription",
				Kind:     ObjectKindSubscription,
				Events: []Event{
					{
						Type:     EventTypeIssue,
						AtMinute: 0,
					},
					{
						Type:     EventTypeRevoke,
						AtMinute: 10,
					},
				},
			},
			{
				ClientID: "premium-cache",
				Name:     "Premium entitlement cache",
				Kind:     ObjectKindCachedEntitlement,
				Events: []Event{
					{
						Type:     EventTypeIssue,
						AtMinute: 0,
					},
					{
						Type:     EventTypeExpire,
						AtMinute: 60,
					},
				},
			},
		},
		Invariants: []Invariant{
			{
				Type:              InvariantTypeRevokedAccessGrace,
				SourceObjectID:    "subscription",
				DependentObjectID: "premium-cache",
				GraceMinutes:      intPointer(5),
			},
		},
	}
}

func makeEvents(count int) []Event {
	events := make([]Event, count)

	for index := range events {
		events[index] = Event{
			Type:     EventTypeIssue,
			AtMinute: index,
		}
	}

	return events
}

func intPointer(value int) *int {
	return &value
}

func containsField(
	validationErrors ValidationErrors,
	expectedField string,
) bool {
	for _, fieldError := range validationErrors {
		if fieldError.Field == expectedField {
			return true
		}
	}

	return false
}
