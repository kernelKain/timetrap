import type { EventType, InvariantType, ObjectKind } from './api'

export type NumericDraft = number | ''
export interface EventDraft { formId: string; type: EventType; atMinute: NumericDraft }
export interface TimedObjectDraft { clientId: string; name: string; kind: ObjectKind; events: EventDraft[] }
export interface InvariantDraft {
  formId: string; type: InvariantType; sourceObjectId: string; dependentObjectId: string; targetObjectId: string
  graceMinutes: NumericDraft; maximumMinutes: NumericDraft
}
export interface ScenarioDraft {
  name: string; description: string; horizonMinutes: NumericDraft; objects: TimedObjectDraft[]; invariants: InvariantDraft[]
}
export interface ValidationIssue { field: string; message: string }
