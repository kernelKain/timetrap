export type ObjectKind = 'subscription' | 'cached_entitlement' | 'session' | 'token' | 'invitation' | 'credential'
export type EventType = 'issue' | 'refresh' | 'revoke' | 'expire'
export type InvariantType = 'revoked_access_grace' | 'dependent_not_outlive_source' | 'max_validity'

export interface ScenarioRequest { name: string; description?: string; horizonMinutes: number; objects: TimedObjectRequest[]; invariants: InvariantRequest[] }
export interface TimedObjectRequest { clientId: string; name: string; kind: ObjectKind; events: EventRequest[] }
export interface EventRequest { type: EventType; atMinute: number }
export interface InvariantRequest { type: InvariantType; sourceObjectId?: string; dependentObjectId?: string; targetObjectId?: string; graceMinutes?: number; maximumMinutes?: number }

export interface PersistedScenario { id: string; definition: ScenarioRequest; createdAt: string; updatedAt: string }
export interface ScenarioResponse { scenario: PersistedScenario; requestId: string }

export interface Evidence { objectId: string; atMinute: number; valid: boolean; eventType?: EventType; eventMinute?: number; detail: string }
export interface AddEventOperation { type: 'add_event'; objectId: string; event: EventRequest }
export interface ReplaceEventOperation { type: 'replace_event'; objectId: string; matchEvent: EventRequest; event: EventRequest }
export interface SetMaximumDurationOperation { type: 'set_maximum_duration'; invariantIndex: number; maximumMinutes: number }
export type RemediationOperation = AddEventOperation | ReplaceEventOperation | SetMaximumDurationOperation
export interface Remediation { code: string; title?: string; summary: string; objectId?: string; suggestedMinutes?: number; operations?: RemediationOperation[] }
export interface ValidityInterval { startMinute: number; endMinute: number }
export interface TimelineLane { objectId: string; intervals: ValidityInterval[] }
export interface Violation {
  invariantIndex: number; invariantType: InvariantType; startMinute: number; endMinute: number; durationMinutes: number; ongoing: boolean
  sourceObjectId?: string; dependentObjectId?: string; targetObjectId?: string; evidence: Evidence[]; remediation: Remediation
  sourceInvalidatedMinute?: number; policyDeadlineMinute?: number; dependentInvalidatedMinute?: number; graceMinutes?: number; totalStaleExposureMinutes?: number
}
export interface AnalysisResult { safe: boolean; evaluatedBoundaries: number; boundaries: number[]; timeline?: TimelineLane[]; earliestViolation?: Violation; violations: Violation[]; engineVersion: string }
export interface PersistedAnalysis { id: string; scenarioId: string; scenarioSnapshot?: ScenarioRequest; result: AnalysisResult; engineVersion: string; createdAt: string }
export interface AnalysisResponse { analysis: PersistedAnalysis; requestId: string }

export interface ApiFieldError { field: string; message: string }
export interface ApiErrorEnvelope { error: { code: string; message: string; fields?: ApiFieldError[] }; requestId: string }
export interface HealthResponse { status: string; service: string; version?: string; database: string; timestamp: string }
