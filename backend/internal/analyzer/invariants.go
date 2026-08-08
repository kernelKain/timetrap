package analyzer

import (
	"fmt"

	"github.com/kernelKain/timetrap/backend/internal/domain"
)

// Failure describes one invariant failing at one evaluated boundary.
//
// It is an analyzer-internal observation, not a completed violation interval.
type Failure struct {
	InvariantIndex    int
	InvariantType     domain.InvariantType
	SourceObjectID    string
	DependentObjectID string
	TargetObjectID    string
	Evidence          domain.Evidence
}

// EvaluateInvariant evaluates one invariant against one object-state snapshot.
//
// A nil failure and nil error means the invariant passed. A non-nil error means
// the model cannot be evaluated reliably and must never be treated as safe.
func EvaluateInvariant(
	invariant domain.Invariant,
	invariantIndex int,
	states map[string]ObjectState,
	minute int,
) (*Failure, error) {
	if minute < 0 {
		return nil, fmt.Errorf(
			"invariant %d cannot be evaluated at negative minute %d",
			invariantIndex,
			minute,
		)
	}

	switch invariant.Type {
	case domain.InvariantTypeRevokedAccessGrace:
		return evaluateRevokedAccessGrace(
			invariant,
			invariantIndex,
			states,
			minute,
		)

	case domain.InvariantTypeDependentNotOutliveSource:
		return evaluateDependentNotOutliveSource(
			invariant,
			invariantIndex,
			states,
			minute,
		)

	case domain.InvariantTypeMaxValidity:
		return evaluateMaximumValidity(
			invariant,
			invariantIndex,
			states,
			minute,
		)

	default:
		return nil, fmt.Errorf(
			"invariant %d has unsupported type %q",
			invariantIndex,
			invariant.Type,
		)
	}
}

func evaluateRevokedAccessGrace(
	invariant domain.Invariant,
	invariantIndex int,
	states map[string]ObjectState,
	minute int,
) (*Failure, error) {
	source, err := requiredState(
		states,
		invariant.SourceObjectID,
		invariantIndex,
		"source",
	)
	if err != nil {
		return nil, err
	}

	dependent, err := requiredState(
		states,
		invariant.DependentObjectID,
		invariantIndex,
		"dependent",
	)
	if err != nil {
		return nil, err
	}

	if invariant.GraceMinutes == nil {
		return nil, fmt.Errorf(
			"invariant %d is missing graceMinutes",
			invariantIndex,
		)
	}

	if *invariant.GraceMinutes < 0 {
		return nil, fmt.Errorf(
			"invariant %d has negative graceMinutes",
			invariantIndex,
		)
	}

	sourceWasRevoked := !source.Valid &&
		source.HasLastEvent &&
		source.LastEventType == domain.EventTypeRevoke

	if !sourceWasRevoked || !dependent.Valid {
		return nil, nil
	}

	deadline := source.LastEventMinute + *invariant.GraceMinutes
	if minute < deadline {
		return nil, nil
	}

	return &Failure{
		InvariantIndex:    invariantIndex,
		InvariantType:     invariant.Type,
		SourceObjectID:    invariant.SourceObjectID,
		DependentObjectID: invariant.DependentObjectID,
		Evidence: evidenceFromState(
			dependent,
			minute,
			fmt.Sprintf(
				"dependent remains valid at or after source revocation grace deadline %d",
				deadline,
			),
		),
	}, nil
}

func evaluateDependentNotOutliveSource(
	invariant domain.Invariant,
	invariantIndex int,
	states map[string]ObjectState,
	minute int,
) (*Failure, error) {
	source, err := requiredState(
		states,
		invariant.SourceObjectID,
		invariantIndex,
		"source",
	)
	if err != nil {
		return nil, err
	}

	dependent, err := requiredState(
		states,
		invariant.DependentObjectID,
		invariantIndex,
		"dependent",
	)
	if err != nil {
		return nil, err
	}

	if !dependent.Valid || source.Valid {
		return nil, nil
	}

	return &Failure{
		InvariantIndex:    invariantIndex,
		InvariantType:     invariant.Type,
		SourceObjectID:    invariant.SourceObjectID,
		DependentObjectID: invariant.DependentObjectID,
		Evidence: evidenceFromState(
			dependent,
			minute,
			"dependent remains valid while source is invalid",
		),
	}, nil
}

func evaluateMaximumValidity(
	invariant domain.Invariant,
	invariantIndex int,
	states map[string]ObjectState,
	minute int,
) (*Failure, error) {
	target, err := requiredState(
		states,
		invariant.TargetObjectID,
		invariantIndex,
		"target",
	)
	if err != nil {
		return nil, err
	}

	if invariant.MaximumMinutes == nil {
		return nil, fmt.Errorf(
			"invariant %d is missing maximumMinutes",
			invariantIndex,
		)
	}

	if *invariant.MaximumMinutes <= 0 {
		return nil, fmt.Errorf(
			"invariant %d has non-positive maximumMinutes",
			invariantIndex,
		)
	}

	if !target.Valid || !target.HasValidSince {
		return nil, nil
	}

	deadline := target.ValidSince + *invariant.MaximumMinutes
	if minute < deadline {
		return nil, nil
	}

	return &Failure{
		InvariantIndex: invariantIndex,
		InvariantType:  invariant.Type,
		TargetObjectID: invariant.TargetObjectID,
		Evidence: evidenceFromState(
			target,
			minute,
			fmt.Sprintf(
				"target remains valid at or after maximum-validity deadline %d",
				deadline,
			),
		),
	}, nil
}

// evidenceFromState creates a stable evidence value for one failed boundary.
//
// Event values are copied before their addresses are stored so the evidence
// does not reference mutable submitted scenario data.
func evidenceFromState(
	state ObjectState,
	minute int,
	detail string,
) domain.Evidence {
	evidence := domain.Evidence{
		ObjectID: state.ObjectID,
		AtMinute: minute,
		Valid:    state.Valid,
		Detail:   detail,
	}

	if state.HasLastEvent {
		eventType := state.LastEventType
		eventMinute := state.LastEventMinute
		evidence.EventType = &eventType
		evidence.EventMinute = &eventMinute
	}

	return evidence
}

func requiredState(
	states map[string]ObjectState,
	objectID string,
	invariantIndex int,
	role string,
) (ObjectState, error) {
	state, exists := states[objectID]
	if !exists {
		return ObjectState{}, fmt.Errorf(
			"invariant %d references missing %s object %q",
			invariantIndex,
			role,
			objectID,
		)
	}

	return state, nil
}
