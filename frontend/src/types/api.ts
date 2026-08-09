export type ObjectKind = 'subscription' | 'cached_entitlement' | 'session' | 'token' | 'invitation' | 'credential'
export type EventType = 'issue' | 'refresh' | 'revoke' | 'expire'
export type InvariantType = 'revoked_access_grace' | 'dependent_not_outlive_source' | 'max_validity'

export interface ScenarioRequest {
  name: string
  description?: string
  horizonMinutes: number
  objects: TimedObjectRequest[]
  invariants: InvariantRequest[]
}

export interface TimedObjectRequest { clientId: string; name: string; kind: ObjectKind; events: EventRequest[] }
export interface EventRequest { type: EventType; atMinute: number }
export interface InvariantRequest {
  type: InvariantType
  sourceObjectId?: string
  dependentObjectId?: string
  targetObjectId?: string
  graceMinutes?: number
  maximumMinutes?: number
}

export interface HealthResponse { status: string; service: string; version: string; database: string; timestamp: string }
