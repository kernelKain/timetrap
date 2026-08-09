package analyzer

import (
	"fmt"

	"github.com/kernelKain/timetrap/backend/internal/domain"
)

const (
	remediationCodeInvalidateDependentOnSourceRevoke = "invalidate_dependent_on_source_revoke"
	remediationCodeAlignDependentWithSource          = "align_dependent_with_source"
	remediationCodeReduceMaximumValidity             = "reduce_maximum_validity"
)

// attachRemediations returns a copied violation slice with deterministic
// remediation attached. It never modifies the submitted scenario or the input
// violation slice.
func attachRemediations(
	scenario domain.Scenario,
	violations []domain.Violation,
) ([]domain.Violation, error) {
	result := make([]domain.Violation, len(violations))
	copy(result, violations)

	for violationIndex := range result {
		remediation, err := remediationForViolation(
			scenario,
			result[violationIndex],
		)
		if err != nil {
			return nil, fmt.Errorf(
				"build remediation for violation %d: %w",
				violationIndex,
				err,
			)
		}

		result[violationIndex].Remediation = remediation
		attachFindingDetails(
			scenario,
			&result[violationIndex],
		)
	}

	return result, nil
}

// remediationForViolation maps an analyzer finding to a stable remediation
// code and structured parameters. It does not apply the remediation.
func remediationForViolation(
	scenario domain.Scenario,
	violation domain.Violation,
) (domain.Remediation, error) {
	if violation.InvariantIndex < 0 ||
		violation.InvariantIndex >= len(scenario.Invariants) {
		return domain.Remediation{}, fmt.Errorf(
			"invariant index %d is outside scenario invariant range",
			violation.InvariantIndex,
		)
	}

	invariant := scenario.Invariants[violation.InvariantIndex]

	if invariant.Type != violation.InvariantType {
		return domain.Remediation{}, fmt.Errorf(
			"invariant %d type %q does not match violation type %q",
			violation.InvariantIndex,
			invariant.Type,
			violation.InvariantType,
		)
	}

	switch invariant.Type {
	case domain.InvariantTypeRevokedAccessGrace:
		return revocationGraceRemediation(
			scenario,
			invariant,
			violation,
		)

	case domain.InvariantTypeDependentNotOutliveSource:
		return dependentOutlivesSourceRemediation(
			scenario,
			invariant,
			violation,
		)

	case domain.InvariantTypeMaxValidity:
		return maximumValidityRemediation(
			invariant,
			violation,
		)

	default:
		return domain.Remediation{}, fmt.Errorf(
			"invariant %d has unsupported type %q",
			violation.InvariantIndex,
			invariant.Type,
		)
	}
}

func revocationGraceRemediation(
	scenario domain.Scenario,
	invariant domain.Invariant,
	violation domain.Violation,
) (domain.Remediation, error) {
	if invariant.GraceMinutes == nil {
		return domain.Remediation{}, fmt.Errorf(
			"invariant %d is missing graceMinutes",
			violation.InvariantIndex,
		)
	}

	sourceState, err := remediationSourceState(
		scenario,
		invariant,
		violation,
	)
	if err != nil {
		return domain.Remediation{}, err
	}

	if sourceState.Valid ||
		!sourceState.HasLastEvent ||
		sourceState.LastEventType != domain.EventTypeRevoke {
		return domain.Remediation{}, fmt.Errorf(
			"invariant %d has no source revocation at violation start %d",
			violation.InvariantIndex,
			violation.StartMinute,
		)
	}

	suggestedMinute := sourceState.LastEventMinute
	event := domain.Event{Type: domain.EventTypeRevoke, AtMinute: suggestedMinute}

	return domain.Remediation{
		Code:             remediationCodeInvalidateDependentOnSourceRevoke,
		Title:            "Invalidate the dependent authorization",
		Summary:          "Revoke the dependent when the source is revoked.",
		ObjectID:         invariant.DependentObjectID,
		SuggestedMinutes: &suggestedMinute,
		Operations: []domain.RemediationOperation{{
			Type:     domain.RemediationOperationAddEvent,
			ObjectID: invariant.DependentObjectID,
			Event:    &event,
		}},
	}, nil
}

func dependentOutlivesSourceRemediation(
	scenario domain.Scenario,
	invariant domain.Invariant,
	violation domain.Violation,
) (domain.Remediation, error) {
	sourceState, err := remediationSourceState(
		scenario,
		invariant,
		violation,
	)
	if err != nil {
		return domain.Remediation{}, err
	}

	// A source can be invalid before its first issue event. In that case there
	// is no submitted invalidation event, so the bounded violation start is the
	// deterministic alignment minute.
	suggestedMinute := violation.StartMinute
	eventType := domain.EventTypeExpire

	if !sourceState.Valid && sourceState.HasLastEvent {
		switch sourceState.LastEventType {
		case domain.EventTypeRevoke, domain.EventTypeExpire:
			suggestedMinute = sourceState.LastEventMinute
			eventType = sourceState.LastEventType
		}
	}
	event := domain.Event{Type: eventType, AtMinute: suggestedMinute}

	return domain.Remediation{
		Code:             remediationCodeAlignDependentWithSource,
		Title:            "Align dependent authorization with its source",
		Summary:          "Invalidate the dependent when its source becomes invalid.",
		ObjectID:         invariant.DependentObjectID,
		SuggestedMinutes: &suggestedMinute,
		Operations: []domain.RemediationOperation{{
			Type:     domain.RemediationOperationAddEvent,
			ObjectID: invariant.DependentObjectID,
			Event:    &event,
		}},
	}, nil
}

func maximumValidityRemediation(
	invariant domain.Invariant,
	violation domain.Violation,
) (domain.Remediation, error) {
	invariantIndex := violation.InvariantIndex
	if invariant.MaximumMinutes == nil {
		return domain.Remediation{}, fmt.Errorf(
			"invariant %d is missing maximumMinutes",
			invariantIndex,
		)
	}

	if *invariant.MaximumMinutes <= 0 {
		return domain.Remediation{}, fmt.Errorf(
			"invariant %d has non-positive maximumMinutes",
			invariantIndex,
		)
	}

	suggestedMaximum := *invariant.MaximumMinutes
	event := domain.Event{Type: domain.EventTypeExpire, AtMinute: violation.StartMinute}

	return domain.Remediation{
		Code:             remediationCodeReduceMaximumValidity,
		Title:            "End validity at the configured maximum",
		Summary:          "Expire the target when its maximum validity is reached.",
		ObjectID:         invariant.TargetObjectID,
		SuggestedMinutes: &suggestedMaximum,
		Operations: []domain.RemediationOperation{{
			Type:     domain.RemediationOperationAddEvent,
			ObjectID: invariant.TargetObjectID,
			Event:    &event,
		}},
	}, nil
}

func attachFindingDetails(scenario domain.Scenario, violation *domain.Violation) {
	if violation.InvariantIndex < 0 || violation.InvariantIndex >= len(scenario.Invariants) {
		return
	}

	invariant := scenario.Invariants[violation.InvariantIndex]
	if invariant.Type != domain.InvariantTypeRevokedAccessGrace || invariant.GraceMinutes == nil {
		return
	}

	states := statesAt(scenario.Objects, violation.StartMinute)
	source, exists := states[invariant.SourceObjectID]
	if !exists || source.Valid || !source.HasLastEvent || source.LastEventType != domain.EventTypeRevoke {
		return
	}

	sourceMinute := source.LastEventMinute
	deadline := sourceMinute + *invariant.GraceMinutes
	grace := *invariant.GraceMinutes
	violation.SourceInvalidatedMinute = &sourceMinute
	violation.PolicyDeadlineMinute = &deadline
	violation.GraceMinutes = &grace

	if !violation.Ongoing {
		dependentMinute := violation.EndMinute
		staleExposure := dependentMinute - sourceMinute
		violation.DependentInvalidatedMinute = &dependentMinute
		violation.TotalStaleExposureMinutes = &staleExposure
	}
}

func remediationSourceState(
	scenario domain.Scenario,
	invariant domain.Invariant,
	violation domain.Violation,
) (ObjectState, error) {
	states := statesAt(
		scenario.Objects,
		violation.StartMinute,
	)

	return requiredState(
		states,
		invariant.SourceObjectID,
		violation.InvariantIndex,
		"source",
	)
}
