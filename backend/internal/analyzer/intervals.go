package analyzer

import (
	"fmt"

	"github.com/kernelKain/timetrap/backend/internal/domain"
)

// activeViolation tracks a failure that is still active across adjacent
// boundary segments. It is not part of the public API.
type activeViolation struct {
	startMinute int
	failure     Failure
}

// buildViolationIntervals evaluates every half-open segment and merges
// consecutive failures of the same invariant into one violation interval.
//
// The horizon boundary is not evaluated because the horizon is the exclusive
// end of the analyzed interval.
func buildViolationIntervals(
	scenario domain.Scenario,
	boundaries []int,
) ([]domain.Violation, error) {
	if err := validateIntervalBoundaries(
		boundaries,
		scenario.HorizonMinutes,
	); err != nil {
		return nil, err
	}

	active := make(map[int]activeViolation)
	violations := make([]domain.Violation, 0)

	// Each pair of adjacent boundaries defines one non-empty segment.
	// Evaluate the invariant state at the inclusive segment start.
	for boundaryIndex := 0; boundaryIndex < len(boundaries)-1; boundaryIndex++ {
		segmentStart := boundaries[boundaryIndex]
		segmentEnd := boundaries[boundaryIndex+1]

		if segmentStart >= segmentEnd {
			return nil, fmt.Errorf(
				"invalid segment [%d,%d)",
				segmentStart,
				segmentEnd,
			)
		}

		states := statesAt(scenario.Objects, segmentStart)

		for invariantIndex, invariant := range scenario.Invariants {
			failure, err := EvaluateInvariant(
				invariant,
				invariantIndex,
				states,
				segmentStart,
			)
			if err != nil {
				return nil, fmt.Errorf(
					"evaluate invariant %d at minute %d: %w",
					invariantIndex,
					segmentStart,
					err,
				)
			}

			activeFinding, isActive := active[invariantIndex]

			switch {
			case failure != nil && !isActive:
				active[invariantIndex] = activeViolation{
					startMinute: segmentStart,
					failure:     *failure,
				}

			case failure != nil && isActive:
				// The same invariant continues to fail in an adjacent
				// segment. Keep the original opening minute and evidence.

			case failure == nil && isActive:
				violations = append(
					violations,
					closeViolation(
						activeFinding,
						segmentStart,
						false,
					),
				)
				delete(active, invariantIndex)

			case failure == nil && !isActive:
				// The invariant passes and has no open interval.
			}
		}
	}

	// Close findings that remain active at the exclusive horizon. Iterate by
	// invariant index rather than ranging over the map so output is stable.
	for invariantIndex := range scenario.Invariants {
		activeFinding, isActive := active[invariantIndex]
		if !isActive {
			continue
		}

		violations = append(
			violations,
			closeViolation(
				activeFinding,
				scenario.HorizonMinutes,
				true,
			),
		)
	}

	return violations, nil
}

func closeViolation(
	active activeViolation,
	endMinute int,
	ongoing bool,
) domain.Violation {
	return domain.Violation{
		InvariantIndex:    active.failure.InvariantIndex,
		InvariantType:     active.failure.InvariantType,
		StartMinute:       active.startMinute,
		EndMinute:         endMinute,
		DurationMinutes:   endMinute - active.startMinute,
		Ongoing:           ongoing,
		SourceObjectID:    active.failure.SourceObjectID,
		DependentObjectID: active.failure.DependentObjectID,
		TargetObjectID:    active.failure.TargetObjectID,
		Evidence: []domain.Evidence{
			active.failure.Evidence,
		},
		// Remediation is attached deterministically in the remediation step.
	}
}

func validateIntervalBoundaries(
	boundaries []int,
	horizon int,
) error {
	if len(boundaries) < 2 {
		return fmt.Errorf(
			"at least two boundaries are required, got %d",
			len(boundaries),
		)
	}

	if boundaries[0] != 0 {
		return fmt.Errorf(
			"first boundary must be 0, got %d",
			boundaries[0],
		)
	}

	if boundaries[len(boundaries)-1] != horizon {
		return fmt.Errorf(
			"last boundary must equal horizon %d, got %d",
			horizon,
			boundaries[len(boundaries)-1],
		)
	}

	for index := 1; index < len(boundaries); index++ {
		if boundaries[index] <= boundaries[index-1] {
			return fmt.Errorf(
				"boundaries must be strictly increasing: index %d contains %d after %d",
				index,
				boundaries[index],
				boundaries[index-1],
			)
		}
	}

	return nil
}
