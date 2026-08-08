package domain

// Analysis is the deterministic result produced for one submitted scenario.
//
// A safe analysis has no earliest violation and an empty violations list.
type Analysis struct {
	Safe                bool        `json:"safe"`
	EvaluatedBoundaries int         `json:"evaluatedBoundaries"`
	EarliestViolation   *Violation  `json:"earliestViolation,omitempty"`
	Violations          []Violation `json:"violations"`
	EngineVersion       string      `json:"engineVersion"`
}

// Violation describes one continuous interval during which an invariant fails.
//
// EndMinute and DurationMinutes are nil when the violation remains active at
// the scenario horizon. The analyzer must not invent an ending beyond the
// bounded horizon.
type Violation struct {
	InvariantIndex  int           `json:"invariantIndex"`
	InvariantType   InvariantType `json:"invariantType"`
	StartMinute     int           `json:"startMinute"`
	EndMinute       *int          `json:"endMinute,omitempty"`
	DurationMinutes *int          `json:"durationMinutes,omitempty"`
	ObjectIDs       []string      `json:"objectIds"`
	Evidence        []Evidence    `json:"evidence"`
	Remediation     *Remediation  `json:"remediation,omitempty"`
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
	Code             string `json:"code"`
	Summary          string `json:"summary"`
	ObjectID         string `json:"objectId,omitempty"`
	SuggestedMinutes *int   `json:"suggestedMinutes,omitempty"`
}
