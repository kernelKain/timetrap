package analyzer

import (
	"fmt"
	"sort"

	"github.com/kernelKain/timetrap/backend/internal/domain"
)

// CandidateBoundaries returns the sorted, deduplicated minutes at which the
// analyzer must reconsider object state or invariant behavior.
//
// The function treats the submitted scenario as immutable.
func CandidateBoundaries(scenario domain.Scenario) ([]int, error) {
	if validationErrors := scenario.Validate(); !validationErrors.Empty() {
		return nil, validationErrors
	}

	boundarySet := make(map[int]struct{})

	addBoundary := func(minute int) {
		if minute < 0 || minute > scenario.HorizonMinutes {
			return
		}

		boundarySet[minute] = struct{}{}
	}

	addBoundary(0)
	addBoundary(scenario.HorizonMinutes)

	objectsByID := make(map[string]domain.TimedObject, len(scenario.Objects))

	for _, object := range scenario.Objects {
		objectsByID[object.ClientID] = object

		for _, event := range object.Events {
			addBoundary(event.AtMinute)
			addBoundary(event.AtMinute - 1)
			addBoundary(event.AtMinute + 1)
		}
	}

	for invariantIndex, invariant := range scenario.Invariants {
		switch invariant.Type {
		case domain.InvariantTypeRevokedAccessGrace:
			source, exists := objectsByID[invariant.SourceObjectID]
			if !exists {
				return nil, fmt.Errorf(
					"invariant %d references missing source object %q",
					invariantIndex,
					invariant.SourceObjectID,
				)
			}

			if invariant.GraceMinutes == nil {
				return nil, fmt.Errorf(
					"invariant %d is missing graceMinutes",
					invariantIndex,
				)
			}

			for _, event := range source.Events {
				if event.Type != domain.EventTypeRevoke {
					continue
				}

				addBoundary(event.AtMinute + *invariant.GraceMinutes)
			}

		case domain.InvariantTypeMaxValidity:
			target, exists := objectsByID[invariant.TargetObjectID]
			if !exists {
				return nil, fmt.Errorf(
					"invariant %d references missing target object %q",
					invariantIndex,
					invariant.TargetObjectID,
				)
			}

			if invariant.MaximumMinutes == nil {
				return nil, fmt.Errorf(
					"invariant %d is missing maximumMinutes",
					invariantIndex,
				)
			}

			for _, event := range target.Events {
				switch event.Type {
				case domain.EventTypeIssue, domain.EventTypeRefresh:
					addBoundary(event.AtMinute + *invariant.MaximumMinutes)
				}
			}

		case domain.InvariantTypeDependentNotOutliveSource:
			// This invariant changes only when one of its objects changes
			// state, so the submitted event boundaries are sufficient.

		default:
			return nil, fmt.Errorf(
				"invariant %d has unsupported type %q",
				invariantIndex,
				invariant.Type,
			)
		}
	}

	boundaries := make([]int, 0, len(boundarySet))
	for minute := range boundarySet {
		boundaries = append(boundaries, minute)
	}

	sort.Ints(boundaries)

	return boundaries, nil
}
