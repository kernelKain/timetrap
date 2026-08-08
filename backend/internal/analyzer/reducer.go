package analyzer

import (
	"sort"

	"github.com/kernelKain/timetrap/backend/internal/domain"
)

// ObjectState is one timed object's reduced state at a specific minute.
type ObjectState struct {
	ObjectID        string
	Valid           bool
	ValidSince      int
	HasValidSince   bool
	LastEventType   domain.EventType
	LastEventMinute int
	HasLastEvent    bool
}

// ReduceObjectAt returns an object's state after applying every event whose
// minute is less than or equal to the evaluation minute.
//
// Events are copied before sorting so the submitted object remains unchanged.
func ReduceObjectAt(
	object domain.TimedObject,
	minute int,
) ObjectState {
	events := append([]domain.Event(nil), object.Events...)

	sort.SliceStable(events, func(leftIndex, rightIndex int) bool {
		return events[leftIndex].AtMinute < events[rightIndex].AtMinute
	})

	state := ObjectState{
		ObjectID: object.ClientID,
	}

	for _, event := range events {
		if event.AtMinute > minute {
			break
		}

		switch event.Type {
		case domain.EventTypeIssue, domain.EventTypeRefresh:
			state.Valid = true
			state.ValidSince = event.AtMinute
			state.HasValidSince = true

		case domain.EventTypeRevoke, domain.EventTypeExpire:
			state.Valid = false
			state.ValidSince = 0
			state.HasValidSince = false
		}

		state.LastEventType = event.Type
		state.LastEventMinute = event.AtMinute
		state.HasLastEvent = true
	}

	return state
}

// statesAt returns a newly allocated object-state snapshot for one boundary
// minute.
//
// Recomputing from immutable submitted events keeps snapshots independent and
// deterministic. The bounded MVP input limits make this approach acceptable.
func statesAt(
	objects []domain.TimedObject,
	minute int,
) map[string]ObjectState {
	states := make(map[string]ObjectState, len(objects))

	for _, object := range objects {
		states[object.ClientID] = ReduceObjectAt(object, minute)
	}

	return states
}
