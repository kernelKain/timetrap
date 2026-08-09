package domain

// Analysis is the deterministic result produced for one submitted scenario.
//
// A safe analysis has no earliest violation and an empty violations list.
// Boundaries contains the sorted, deduplicated boundary minutes evaluated by
// the analyzer.
type Analysis struct {
	Safe                bool           `json:"safe"`
	EvaluatedBoundaries int            `json:"evaluatedBoundaries"`
	Boundaries          []int          `json:"boundaries"`
	Timeline            []TimelineLane `json:"timeline,omitempty"`
	EarliestViolation   *Violation     `json:"earliestViolation,omitempty"`
	Violations          []Violation    `json:"violations"`
	EngineVersion       string         `json:"engineVersion"`
}

// Violation describes one continuous half-open interval during which an
// invariant fails.
//
// EndMinute is always the exclusive end of the reported interval. When Ongoing
// is true, EndMinute equals the scenario horizon because the analyzer must not
// invent an ending beyond the bounded model.
//
// DurationMinutes is always EndMinute - StartMinute.
type Violation struct {
	InvariantIndex             int           `json:"invariantIndex"`
	InvariantType              InvariantType `json:"invariantType"`
	StartMinute                int           `json:"startMinute"`
	EndMinute                  int           `json:"endMinute"`
	DurationMinutes            int           `json:"durationMinutes"`
	Ongoing                    bool          `json:"ongoing"`
	SourceObjectID             string        `json:"sourceObjectId,omitempty"`
	DependentObjectID          string        `json:"dependentObjectId,omitempty"`
	TargetObjectID             string        `json:"targetObjectId,omitempty"`
	Evidence                   []Evidence    `json:"evidence"`
	Remediation                Remediation   `json:"remediation"`
	SourceInvalidatedMinute    *int          `json:"sourceInvalidatedMinute,omitempty"`
	PolicyDeadlineMinute       *int          `json:"policyDeadlineMinute,omitempty"`
	DependentInvalidatedMinute *int          `json:"dependentInvalidatedMinute,omitempty"`
	GraceMinutes               *int          `json:"graceMinutes,omitempty"`
	TotalStaleExposureMinutes  *int          `json:"totalStaleExposureMinutes,omitempty"`
}

// TimelineLane is presentation-ready analyzer output for one timed object.
// Intervals are normalized half-open validity ranges on the scenario horizon.
type TimelineLane struct {
	ObjectID  string             `json:"objectId"`
	Intervals []ValidityInterval `json:"intervals"`
}

type ValidityInterval struct {
	StartMinute int `json:"startMinute"`
	EndMinute   int `json:"endMinute"`
}

// Evidence records one object state or event that helps explain a violation.
//
// EventType and EventMinute are optional because some evidence may describe a
// calculated rule boundary rather than one specific submitted event.
type Evidence struct {
	ObjectID    string     `json:"objectId"`
	AtMinute    int        `json:"atMinute"`
	Valid       bool       `json:"valid"`
	EventType   *EventType `json:"eventType,omitempty"`
	EventMinute *int       `json:"eventMinute,omitempty"`
	Detail      string     `json:"detail"`
}

// Remediation describes a deterministic configuration-level correction.
//
// SuggestedMinutes is optional because some remediations, such as adding active
// invalidation, may not require a replacement duration.
type Remediation struct {
	Code             string                 `json:"code"`
	Title            string                 `json:"title,omitempty"`
	Summary          string                 `json:"summary"`
	ObjectID         string                 `json:"objectId,omitempty"`
	SuggestedMinutes *int                   `json:"suggestedMinutes,omitempty"`
	Operations       []RemediationOperation `json:"operations,omitempty"`
}

type RemediationOperationType string

const (
	RemediationOperationAddEvent           RemediationOperationType = "add_event"
	RemediationOperationReplaceEvent       RemediationOperationType = "replace_event"
	RemediationOperationSetMaximumDuration RemediationOperationType = "set_maximum_duration"
)

// RemediationOperation is a discriminated machine-actionable correction.
// Fields not used by an operation type are omitted from JSON.
type RemediationOperation struct {
	Type           RemediationOperationType `json:"type"`
	ObjectID       string                   `json:"objectId,omitempty"`
	Event          *Event                   `json:"event,omitempty"`
	MatchEvent     *Event                   `json:"matchEvent,omitempty"`
	InvariantIndex *int                     `json:"invariantIndex,omitempty"`
	MaximumMinutes *int                     `json:"maximumMinutes,omitempty"`
}
