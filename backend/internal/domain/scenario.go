package domain

import (
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"
)

var clientIDRegexp = regexp.MustCompile(ClientIDPattern)

// ObjectKind identifies the role of an object in an authorization model.
type ObjectKind string

const (
	ObjectKindSubscription      ObjectKind = "subscription"
	ObjectKindCachedEntitlement ObjectKind = "cached_entitlement"
	ObjectKindSession           ObjectKind = "session"
	ObjectKindToken             ObjectKind = "token"
	ObjectKindInvitation        ObjectKind = "invitation"
	ObjectKindCredential        ObjectKind = "credential"
)

// IsValid reports whether the object kind is supported by the MVP.
func (kind ObjectKind) IsValid() bool {
	switch kind {
	case ObjectKindSubscription,
		ObjectKindCachedEntitlement,
		ObjectKindSession,
		ObjectKindToken,
		ObjectKindInvitation,
		ObjectKindCredential:
		return true
	default:
		return false
	}
}

// EventType identifies a state-changing event for a timed object.
type EventType string

const (
	EventTypeIssue   EventType = "issue"
	EventTypeRefresh EventType = "refresh"
	EventTypeRevoke  EventType = "revoke"
	EventTypeExpire  EventType = "expire"
)

// IsValid reports whether the event type is supported by the MVP.
func (eventType EventType) IsValid() bool {
	switch eventType {
	case EventTypeIssue,
		EventTypeRefresh,
		EventTypeRevoke,
		EventTypeExpire:
		return true
	default:
		return false
	}
}

// InvariantType identifies one of TimeTrap's supported rule templates.
type InvariantType string

const (
	InvariantTypeRevokedAccessGrace        InvariantType = "revoked_access_grace"
	InvariantTypeDependentNotOutliveSource InvariantType = "dependent_not_outlive_source"
	InvariantTypeMaxValidity               InvariantType = "max_validity"
)

// IsValid reports whether the invariant type is supported by the MVP.
func (invariantType InvariantType) IsValid() bool {
	switch invariantType {
	case InvariantTypeRevokedAccessGrace,
		InvariantTypeDependentNotOutliveSource,
		InvariantTypeMaxValidity:
		return true
	default:
		return false
	}
}

// Input limits keep submitted scenarios bounded and predictable.
const (
	MinScenarioNameLength = 1
	MaxScenarioNameLength = 120

	MinDescriptionLength = 0
	MaxDescriptionLength = 1000

	MinHorizonMinutes = 1
	MaxHorizonMinutes = 10080

	MinObjects = 1
	MaxObjects = 25

	MinObjectNameLength = 1
	MaxObjectNameLength = 80

	MinClientIDLength = 1
	MaxClientIDLength = 64

	MinEventsPerObject = 1
	MaxEventsPerObject = 50
	MaxTotalEvents     = 250

	MinInvariants = 1
	MaxInvariants = 10

	// Event timestamps may occur from minute zero through the scenario
	// horizon, inclusive.
	MinEventMinute = 0
)

// ClientIDPattern is the regular-expression pattern that client IDs must
// satisfy. Client IDs may contain only ASCII letters, numbers, hyphens and
// underscores.
const ClientIDPattern = `^[A-Za-z0-9_-]+$`

// Scenario is a complete bounded authorization model submitted by a user.
type Scenario struct {
	Name           string        `json:"name"`
	Description    string        `json:"description,omitempty"`
	HorizonMinutes int           `json:"horizonMinutes"`
	Objects        []TimedObject `json:"objects"`
	Invariants     []Invariant   `json:"invariants"`
}

// TimedObject is an authorization-related object whose validity changes over
// the scenario timeline.
type TimedObject struct {
	ClientID string     `json:"clientId"`
	Name     string     `json:"name"`
	Kind     ObjectKind `json:"kind"`
	Events   []Event    `json:"events"`
}

// Event changes an object's validity at a specific integer minute.
type Event struct {
	Type     EventType `json:"type"`
	AtMinute int       `json:"atMinute"`
}

// Invariant configures one of TimeTrap's predefined temporal rules.
//
// GraceMinutes and MaximumMinutes are pointers because zero can be meaningful.
// A nil pointer means the JSON field was not provided. A non-nil pointer
// containing zero means the user explicitly provided zero.
type Invariant struct {
	Type              InvariantType `json:"type"`
	SourceObjectID    string        `json:"sourceObjectId,omitempty"`
	DependentObjectID string        `json:"dependentObjectId,omitempty"`
	TargetObjectID    string        `json:"targetObjectId,omitempty"`
	GraceMinutes      *int          `json:"graceMinutes,omitempty"`
	MaximumMinutes    *int          `json:"maximumMinutes,omitempty"`
}

// Validate checks whether the scenario satisfies the bounded MVP model.
//
// Validation reads the scenario but never trims, normalizes or otherwise
// modifies submitted values.
func (scenario Scenario) Validate() ValidationErrors {
	var validationErrors ValidationErrors

	validateScenarioBasics(scenario, &validationErrors)
	objectIDs := validateObjects(scenario, &validationErrors)
	validateInvariants(scenario, objectIDs, &validationErrors)

	return validationErrors
}

func validateScenarioBasics(
	scenario Scenario,
	validationErrors *ValidationErrors,
) {
	nameLength := utf8.RuneCountInString(scenario.Name)

	switch {
	case strings.TrimSpace(scenario.Name) == "":
		validationErrors.Add("name", "is required")
	case nameLength > MaxScenarioNameLength:
		validationErrors.Add(
			"name",
			fmt.Sprintf(
				"must contain at most %d characters",
				MaxScenarioNameLength,
			),
		)
	}

	descriptionLength := utf8.RuneCountInString(scenario.Description)
	if descriptionLength > MaxDescriptionLength {
		validationErrors.Add(
			"description",
			fmt.Sprintf(
				"must contain at most %d characters",
				MaxDescriptionLength,
			),
		)
	}

	if scenario.HorizonMinutes < MinHorizonMinutes ||
		scenario.HorizonMinutes > MaxHorizonMinutes {
		validationErrors.Add(
			"horizonMinutes",
			fmt.Sprintf(
				"must be between %d and %d",
				MinHorizonMinutes,
				MaxHorizonMinutes,
			),
		)
	}

	objectCount := len(scenario.Objects)
	if objectCount < MinObjects || objectCount > MaxObjects {
		validationErrors.Add(
			"objects",
			fmt.Sprintf(
				"must contain between %d and %d objects",
				MinObjects,
				MaxObjects,
			),
		)
	}
}

func validateObjects(
	scenario Scenario,
	validationErrors *ValidationErrors,
) map[string]struct{} {
	objectIDs := make(map[string]struct{}, len(scenario.Objects))
	totalEvents := 0

	for objectIndex, object := range scenario.Objects {
		fieldPrefix := fmt.Sprintf("objects[%d]", objectIndex)
		clientIDValid := true

		clientIDLength := utf8.RuneCountInString(object.ClientID)

		switch {
		case strings.TrimSpace(object.ClientID) == "":
			validationErrors.Add(
				fieldPrefix+".clientId",
				"is required",
			)
			clientIDValid = false
		case clientIDLength > MaxClientIDLength:
			validationErrors.Add(
				fieldPrefix+".clientId",
				fmt.Sprintf(
					"must contain at most %d characters",
					MaxClientIDLength,
				),
			)
			clientIDValid = false
		case !clientIDRegexp.MatchString(object.ClientID):
			validationErrors.Add(
				fieldPrefix+".clientId",
				"may contain only letters, numbers, hyphens and underscores",
			)
			clientIDValid = false
		}

		if clientIDValid {
			if _, exists := objectIDs[object.ClientID]; exists {
				validationErrors.Add(
					fieldPrefix+".clientId",
					"must be unique",
				)
			} else {
				objectIDs[object.ClientID] = struct{}{}
			}
		}

		nameLength := utf8.RuneCountInString(object.Name)

		switch {
		case strings.TrimSpace(object.Name) == "":
			validationErrors.Add(
				fieldPrefix+".name",
				"is required",
			)
		case nameLength > MaxObjectNameLength:
			validationErrors.Add(
				fieldPrefix+".name",
				fmt.Sprintf(
					"must contain at most %d characters",
					MaxObjectNameLength,
				),
			)
		}

		if !object.Kind.IsValid() {
			validationErrors.Add(
				fieldPrefix+".kind",
				"is not a supported object kind",
			)
		}

		eventCount := len(object.Events)
		totalEvents += eventCount

		if eventCount < MinEventsPerObject ||
			eventCount > MaxEventsPerObject {
			validationErrors.Add(
				fieldPrefix+".events",
				fmt.Sprintf(
					"must contain between %d and %d events",
					MinEventsPerObject,
					MaxEventsPerObject,
				),
			)
		}

		validateEvents(
			objectIndex,
			object.Events,
			scenario.HorizonMinutes,
			validationErrors,
		)
	}

	if totalEvents > MaxTotalEvents {
		validationErrors.Add(
			"objects",
			fmt.Sprintf(
				"must contain at most %d events in total",
				MaxTotalEvents,
			),
		)
	}

	return objectIDs
}

func validateEvents(
	objectIndex int,
	events []Event,
	horizonMinutes int,
	validationErrors *ValidationErrors,
) {
	eventsByMinute := make(map[int]int, len(events))

	for eventIndex, event := range events {
		fieldPrefix := fmt.Sprintf(
			"objects[%d].events[%d]",
			objectIndex,
			eventIndex,
		)

		if !event.Type.IsValid() {
			validationErrors.Add(
				fieldPrefix+".type",
				"is not a supported event type",
			)
		}

		if event.AtMinute < MinEventMinute ||
			event.AtMinute > horizonMinutes {
			validationErrors.Add(
				fieldPrefix+".atMinute",
				fmt.Sprintf(
					"must be between %d and horizonMinutes",
					MinEventMinute,
				),
			)
		}

		if previousEventIndex, exists := eventsByMinute[event.AtMinute]; exists {
			validationErrors.Add(
				fieldPrefix+".atMinute",
				fmt.Sprintf(
					"conflicts with objects[%d].events[%d]; only one event per object is allowed at the same minute",
					objectIndex,
					previousEventIndex,
				),
			)
		} else {
			eventsByMinute[event.AtMinute] = eventIndex
		}
	}
}

func validateInvariants(
	scenario Scenario,
	objectIDs map[string]struct{},
	validationErrors *ValidationErrors,
) {
	invariantCount := len(scenario.Invariants)

	if invariantCount < MinInvariants ||
		invariantCount > MaxInvariants {
		validationErrors.Add(
			"invariants",
			fmt.Sprintf(
				"must contain between %d and %d invariants",
				MinInvariants,
				MaxInvariants,
			),
		)
	}

	for invariantIndex, invariant := range scenario.Invariants {
		fieldPrefix := fmt.Sprintf("invariants[%d]", invariantIndex)

		if !invariant.Type.IsValid() {
			validationErrors.Add(
				fieldPrefix+".type",
				"is not a supported invariant type",
			)
			continue
		}

		switch invariant.Type {
		case InvariantTypeRevokedAccessGrace:
			validateRevokedAccessGraceInvariant(
				scenario,
				invariant,
				fieldPrefix,
				objectIDs,
				validationErrors,
			)

		case InvariantTypeDependentNotOutliveSource:
			validateDependentNotOutliveSourceInvariant(
				invariant,
				fieldPrefix,
				objectIDs,
				validationErrors,
			)

		case InvariantTypeMaxValidity:
			validateMaxValidityInvariant(
				scenario,
				invariant,
				fieldPrefix,
				objectIDs,
				validationErrors,
			)
		}
	}
}

func validateRevokedAccessGraceInvariant(
	scenario Scenario,
	invariant Invariant,
	fieldPrefix string,
	objectIDs map[string]struct{},
	validationErrors *ValidationErrors,
) {
	sourceValid := validateObjectReference(
		fieldPrefix+".sourceObjectId",
		invariant.SourceObjectID,
		objectIDs,
		validationErrors,
	)

	dependentValid := validateObjectReference(
		fieldPrefix+".dependentObjectId",
		invariant.DependentObjectID,
		objectIDs,
		validationErrors,
	)

	if sourceValid &&
		dependentValid &&
		invariant.SourceObjectID == invariant.DependentObjectID {
		validationErrors.Add(
			fieldPrefix+".dependentObjectId",
			"must be different from sourceObjectId",
		)
	}

	switch {
	case invariant.GraceMinutes == nil:
		validationErrors.Add(
			fieldPrefix+".graceMinutes",
			"is required",
		)
	case *invariant.GraceMinutes < 0:
		validationErrors.Add(
			fieldPrefix+".graceMinutes",
			"must be zero or greater",
		)
	case *invariant.GraceMinutes > scenario.HorizonMinutes:
		validationErrors.Add(
			fieldPrefix+".graceMinutes",
			"must not exceed horizonMinutes",
		)
	}

	if strings.TrimSpace(invariant.TargetObjectID) != "" {
		validationErrors.Add(
			fieldPrefix+".targetObjectId",
			"must not be provided for revoked_access_grace",
		)
	}

	if invariant.MaximumMinutes != nil {
		validationErrors.Add(
			fieldPrefix+".maximumMinutes",
			"must not be provided for revoked_access_grace",
		)
	}
}

func validateDependentNotOutliveSourceInvariant(
	invariant Invariant,
	fieldPrefix string,
	objectIDs map[string]struct{},
	validationErrors *ValidationErrors,
) {
	sourceValid := validateObjectReference(
		fieldPrefix+".sourceObjectId",
		invariant.SourceObjectID,
		objectIDs,
		validationErrors,
	)

	dependentValid := validateObjectReference(
		fieldPrefix+".dependentObjectId",
		invariant.DependentObjectID,
		objectIDs,
		validationErrors,
	)

	if sourceValid &&
		dependentValid &&
		invariant.SourceObjectID == invariant.DependentObjectID {
		validationErrors.Add(
			fieldPrefix+".dependentObjectId",
			"must be different from sourceObjectId",
		)
	}

	if strings.TrimSpace(invariant.TargetObjectID) != "" {
		validationErrors.Add(
			fieldPrefix+".targetObjectId",
			"must not be provided for dependent_not_outlive_source",
		)
	}

	if invariant.GraceMinutes != nil {
		validationErrors.Add(
			fieldPrefix+".graceMinutes",
			"must not be provided for dependent_not_outlive_source",
		)
	}

	if invariant.MaximumMinutes != nil {
		validationErrors.Add(
			fieldPrefix+".maximumMinutes",
			"must not be provided for dependent_not_outlive_source",
		)
	}
}

func validateMaxValidityInvariant(
	scenario Scenario,
	invariant Invariant,
	fieldPrefix string,
	objectIDs map[string]struct{},
	validationErrors *ValidationErrors,
) {
	validateObjectReference(
		fieldPrefix+".targetObjectId",
		invariant.TargetObjectID,
		objectIDs,
		validationErrors,
	)

	switch {
	case invariant.MaximumMinutes == nil:
		validationErrors.Add(
			fieldPrefix+".maximumMinutes",
			"is required",
		)
	case *invariant.MaximumMinutes <= 0:
		validationErrors.Add(
			fieldPrefix+".maximumMinutes",
			"must be greater than zero",
		)
	case *invariant.MaximumMinutes > scenario.HorizonMinutes:
		validationErrors.Add(
			fieldPrefix+".maximumMinutes",
			"must not exceed horizonMinutes",
		)
	}

	if strings.TrimSpace(invariant.SourceObjectID) != "" {
		validationErrors.Add(
			fieldPrefix+".sourceObjectId",
			"must not be provided for max_validity",
		)
	}

	if strings.TrimSpace(invariant.DependentObjectID) != "" {
		validationErrors.Add(
			fieldPrefix+".dependentObjectId",
			"must not be provided for max_validity",
		)
	}

	if invariant.GraceMinutes != nil {
		validationErrors.Add(
			fieldPrefix+".graceMinutes",
			"must not be provided for max_validity",
		)
	}
}

func validateObjectReference(
	field string,
	objectID string,
	objectIDs map[string]struct{},
	validationErrors *ValidationErrors,
) bool {
	if strings.TrimSpace(objectID) == "" {
		validationErrors.Add(field, "is required")
		return false
	}

	if _, exists := objectIDs[objectID]; !exists {
		validationErrors.Add(field, "must reference an existing object")
		return false
	}

	return true
}
