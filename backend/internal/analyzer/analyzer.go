package analyzer

import (
	"fmt"
	"sort"

	"github.com/kernelKain/timetrap/backend/internal/domain"
)

const engineVersion = "v1"

// Analyze evaluates a validated scenario over its bounded time horizon.
//
// Unsafe scenarios are returned as successful analysis results containing one
// or more violations. An error means the submitted model could not be
// evaluated and must never be interpreted as safe.
//
// Analyze treats the submitted scenario as immutable. All boundaries,
// violations, maps and result slices are allocated by the analyzer.
func Analyze(scenario domain.Scenario) (domain.Analysis, error) {
	if validationErrors := scenario.Validate(); !validationErrors.Empty() {
		return domain.Analysis{}, validationErrors
	}

	boundaries, err := CandidateBoundaries(scenario)
	if err != nil {
		return domain.Analysis{}, fmt.Errorf(
			"generate candidate boundaries: %w",
			err,
		)
	}

	if len(boundaries) < 2 {
		return domain.Analysis{}, fmt.Errorf(
			"analysis requires at least two boundaries, got %d",
			len(boundaries),
		)
	}

	if boundaries[0] != 0 {
		return domain.Analysis{}, fmt.Errorf(
			"first analysis boundary must be 0, got %d",
			boundaries[0],
		)
	}

	lastBoundary := boundaries[len(boundaries)-1]
	if lastBoundary != scenario.HorizonMinutes {
		return domain.Analysis{}, fmt.Errorf(
			"last analysis boundary must equal horizon %d, got %d",
			scenario.HorizonMinutes,
			lastBoundary,
		)
	}

	violations, err := buildViolationIntervals(
		scenario,
		boundaries,
	)
	if err != nil {
		return domain.Analysis{}, fmt.Errorf(
			"build violation intervals: %w",
			err,
		)
	}

	violations, err = attachRemediations(
		scenario,
		violations,
	)
	if err != nil {
		return domain.Analysis{}, fmt.Errorf(
			"attach deterministic remediations: %w",
			err,
		)
	}

	return analysisFromIntervals(boundaries, violations), nil
}

// analysisFromIntervals builds the deterministic public analysis result.
//
// It copies and sorts its inputs so callers retain ownership of their slices.
func analysisFromIntervals(
	boundaries []int,
	violations []domain.Violation,
) domain.Analysis {
	resultBoundaries := make([]int, len(boundaries))
	copy(resultBoundaries, boundaries)

	resultViolations := make([]domain.Violation, len(violations))
	copy(resultViolations, violations)

	sort.SliceStable(
		resultViolations,
		func(leftIndex, rightIndex int) bool {
			left := resultViolations[leftIndex]
			right := resultViolations[rightIndex]

			if left.StartMinute != right.StartMinute {
				return left.StartMinute < right.StartMinute
			}

			if left.InvariantIndex != right.InvariantIndex {
				return left.InvariantIndex < right.InvariantIndex
			}

			if left.EndMinute != right.EndMinute {
				return left.EndMinute < right.EndMinute
			}

			leftObjectKey := violationObjectKey(left)
			rightObjectKey := violationObjectKey(right)

			if leftObjectKey != rightObjectKey {
				return leftObjectKey < rightObjectKey
			}

			// This final comparison makes ordering deterministic even if two
			// otherwise identical findings have different invariant types.
			return left.InvariantType < right.InvariantType
		},
	)

	analysis := domain.Analysis{
		Safe:                len(resultViolations) == 0,
		EvaluatedBoundaries: len(resultBoundaries),
		Boundaries:          resultBoundaries,
		Violations:          resultViolations,
		EngineVersion:       engineVersion,
	}

	if len(resultViolations) > 0 {
		earliest := resultViolations[0]
		analysis.EarliestViolation = &earliest
	}

	return analysis
}

// violationObjectKey supplies a stable final ordering for findings involving
// different combinations of source, dependent and target objects.
func violationObjectKey(violation domain.Violation) string {
	return violation.SourceObjectID +
		"\x00" +
		violation.DependentObjectID +
		"\x00" +
		violation.TargetObjectID
}
