package analyzer

import (
	"reflect"
	"testing"

	"github.com/kernelKain/timetrap/backend/internal/domain"
)

func TestReduceObjectAtTimeline(t *testing.T) {
	t.Parallel()

	object := domain.TimedObject{
		ClientID: "session",
		Name:     "Application session",
		Kind:     domain.ObjectKindSession,
		Events: []domain.Event{
			{
				Type:     domain.EventTypeIssue,
				AtMinute: 5,
			},
			{
				Type:     domain.EventTypeRefresh,
				AtMinute: 8,
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
				Type:     domain.EventTypeExpire,
				AtMinute: 30,
			},
		},
	}

	testCases := []struct {
		name   string
		minute int
		want   ObjectState
	}{
		{
			name:   "before issuance",
			minute: 4,
			want: ObjectState{
				ObjectID: "session",
			},
		},
		{
			name:   "exactly at issuance",
			minute: 5,
			want: ObjectState{
				ObjectID:        "session",
				Valid:           true,
				ValidSince:      5,
				HasValidSince:   true,
				LastEventType:   domain.EventTypeIssue,
				LastEventMinute: 5,
				HasLastEvent:    true,
			},
		},
		{
			name:   "between issue and refresh",
			minute: 7,
			want: ObjectState{
				ObjectID:        "session",
				Valid:           true,
				ValidSince:      5,
				HasValidSince:   true,
				LastEventType:   domain.EventTypeIssue,
				LastEventMinute: 5,
				HasLastEvent:    true,
			},
		},
		{
			name:   "exactly at refresh",
			minute: 8,
			want: ObjectState{
				ObjectID:        "session",
				Valid:           true,
				ValidSince:      8,
				HasValidSince:   true,
				LastEventType:   domain.EventTypeRefresh,
				LastEventMinute: 8,
				HasLastEvent:    true,
			},
		},
		{
			name:   "immediately after refresh",
			minute: 9,
			want: ObjectState{
				ObjectID:        "session",
				Valid:           true,
				ValidSince:      8,
				HasValidSince:   true,
				LastEventType:   domain.EventTypeRefresh,
				LastEventMinute: 8,
				HasLastEvent:    true,
			},
		},
		{
			name:   "exactly at revoke",
			minute: 10,
			want: ObjectState{
				ObjectID:        "session",
				LastEventType:   domain.EventTypeRevoke,
				LastEventMinute: 10,
				HasLastEvent:    true,
			},
		},
		{
			name:   "after revoke",
			minute: 15,
			want: ObjectState{
				ObjectID:        "session",
				LastEventType:   domain.EventTypeRevoke,
				LastEventMinute: 10,
				HasLastEvent:    true,
			},
		},
		{
			name:   "issued again",
			minute: 20,
			want: ObjectState{
				ObjectID:        "session",
				Valid:           true,
				ValidSince:      20,
				HasValidSince:   true,
				LastEventType:   domain.EventTypeIssue,
				LastEventMinute: 20,
				HasLastEvent:    true,
			},
		},
		{
			name:   "exactly at expiration",
			minute: 30,
			want: ObjectState{
				ObjectID:        "session",
				LastEventType:   domain.EventTypeExpire,
				LastEventMinute: 30,
				HasLastEvent:    true,
			},
		},
		{
			name:   "after expiration",
			minute: 40,
			want: ObjectState{
				ObjectID:        "session",
				LastEventType:   domain.EventTypeExpire,
				LastEventMinute: 30,
				HasLastEvent:    true,
			},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got := ReduceObjectAt(object, testCase.minute)

			if !reflect.DeepEqual(got, testCase.want) {
				t.Fatalf(
					"ReduceObjectAt() = %#v, want %#v",
					got,
					testCase.want,
				)
			}
		})
	}
}

func TestReduceObjectAtMinuteZeroAndHorizon(t *testing.T) {
	t.Parallel()

	object := domain.TimedObject{
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
	}

	atZero := ReduceObjectAt(object, 0)
	wantAtZero := ObjectState{
		ObjectID:        "token",
		Valid:           true,
		ValidSince:      0,
		HasValidSince:   true,
		LastEventType:   domain.EventTypeIssue,
		LastEventMinute: 0,
		HasLastEvent:    true,
	}

	if !reflect.DeepEqual(atZero, wantAtZero) {
		t.Fatalf(
			"state at minute zero = %#v, want %#v",
			atZero,
			wantAtZero,
		)
	}

	atHorizon := ReduceObjectAt(object, 90)
	wantAtHorizon := ObjectState{
		ObjectID:        "token",
		LastEventType:   domain.EventTypeExpire,
		LastEventMinute: 90,
		HasLastEvent:    true,
	}

	if !reflect.DeepEqual(atHorizon, wantAtHorizon) {
		t.Fatalf(
			"state at horizon = %#v, want %#v",
			atHorizon,
			wantAtHorizon,
		)
	}
}

func TestReduceObjectAtSortsACopyWithoutMutatingInput(t *testing.T) {
	t.Parallel()

	object := domain.TimedObject{
		ClientID: "premium-cache",
		Name:     "Premium entitlement cache",
		Kind:     domain.ObjectKindCachedEntitlement,
		Events: []domain.Event{
			{
				Type:     domain.EventTypeExpire,
				AtMinute: 60,
			},
			{
				Type:     domain.EventTypeRevoke,
				AtMinute: 10,
			},
			{
				Type:     domain.EventTypeIssue,
				AtMinute: 0,
			},
		},
	}

	originalEvents := append([]domain.Event(nil), object.Events...)

	got := ReduceObjectAt(object, 15)
	want := ObjectState{
		ObjectID:        "premium-cache",
		LastEventType:   domain.EventTypeRevoke,
		LastEventMinute: 10,
		HasLastEvent:    true,
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf(
			"ReduceObjectAt() = %#v, want %#v",
			got,
			want,
		)
	}

	if !reflect.DeepEqual(object.Events, originalEvents) {
		t.Fatalf(
			"ReduceObjectAt() mutated events:\ngot:  %#v\nwant: %#v",
			object.Events,
			originalEvents,
		)
	}
}

func TestStatesAtReturnsExpectedBoundarySnapshots(t *testing.T) {
	t.Parallel()

	objects := snapshotObjects()

	atMinuteFive := statesAt(objects, 5)
	wantAtMinuteFive := map[string]ObjectState{
		"subscription": {
			ObjectID:        "subscription",
			Valid:           true,
			ValidSince:      0,
			HasValidSince:   true,
			LastEventType:   domain.EventTypeIssue,
			LastEventMinute: 0,
			HasLastEvent:    true,
		},
		"premium-cache": {
			ObjectID:        "premium-cache",
			Valid:           true,
			ValidSince:      0,
			HasValidSince:   true,
			LastEventType:   domain.EventTypeIssue,
			LastEventMinute: 0,
			HasLastEvent:    true,
		},
	}

	if !reflect.DeepEqual(atMinuteFive, wantAtMinuteFive) {
		t.Fatalf(
			"statesAt(objects, 5) = %#v, want %#v",
			atMinuteFive,
			wantAtMinuteFive,
		)
	}

	atMinuteFifteen := statesAt(objects, 15)
	wantAtMinuteFifteen := map[string]ObjectState{
		"subscription": {
			ObjectID:        "subscription",
			LastEventType:   domain.EventTypeRevoke,
			LastEventMinute: 10,
			HasLastEvent:    true,
		},
		"premium-cache": {
			ObjectID:        "premium-cache",
			Valid:           true,
			ValidSince:      0,
			HasValidSince:   true,
			LastEventType:   domain.EventTypeIssue,
			LastEventMinute: 0,
			HasLastEvent:    true,
		},
	}

	if !reflect.DeepEqual(atMinuteFifteen, wantAtMinuteFifteen) {
		t.Fatalf(
			"statesAt(objects, 15) = %#v, want %#v",
			atMinuteFifteen,
			wantAtMinuteFifteen,
		)
	}
}

func TestStatesAtReturnsIndependentMaps(t *testing.T) {
	t.Parallel()

	objects := snapshotObjects()

	first := statesAt(objects, 15)
	second := statesAt(objects, 15)

	delete(first, "subscription")
	first["premium-cache"] = ObjectState{
		ObjectID: "changed",
	}

	wantSecond := map[string]ObjectState{
		"subscription": {
			ObjectID:        "subscription",
			LastEventType:   domain.EventTypeRevoke,
			LastEventMinute: 10,
			HasLastEvent:    true,
		},
		"premium-cache": {
			ObjectID:        "premium-cache",
			Valid:           true,
			ValidSince:      0,
			HasValidSince:   true,
			LastEventType:   domain.EventTypeIssue,
			LastEventMinute: 0,
			HasLastEvent:    true,
		},
	}

	if !reflect.DeepEqual(second, wantSecond) {
		t.Fatalf(
			"mutating first snapshot changed second:\ngot:  %#v\nwant: %#v",
			second,
			wantSecond,
		)
	}
}

func TestStatesAtIsIndependentOfCallOrder(t *testing.T) {
	t.Parallel()

	objects := snapshotObjects()

	lateFirst := statesAt(objects, 15)
	_ = statesAt(objects, 5)
	lateSecond := statesAt(objects, 15)

	if !reflect.DeepEqual(lateFirst, lateSecond) {
		t.Fatalf(
			"same boundary produced different snapshots:\nfirst:  %#v\nsecond: %#v",
			lateFirst,
			lateSecond,
		)
	}
}

func snapshotObjects() []domain.TimedObject {
	return []domain.TimedObject{
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
	}
}
