import type { ScenarioRequest } from '../types/api'
import type { EventDraft, InvariantDraft, ScenarioDraft, TimedObjectDraft } from '../types/forms'

export const newId = (prefix: string) => `${prefix}_${crypto.randomUUID().replaceAll('-', '')}`
export const cloneScenario = (draft: ScenarioDraft): ScenarioDraft => structuredClone(draft)

export function createEvent(type: EventDraft['type'] = 'issue'): EventDraft {
  return { formId: newId('event'), type, atMinute: 0 }
}
export function createObject(): TimedObjectDraft {
  return { clientId: newId('object'), name: 'New object', kind: 'session', events: [createEvent()] }
}
export function createInvariant(type: InvariantDraft['type'] = 'revoked_access_grace'): InvariantDraft {
  return { formId: newId('invariant'), type, sourceObjectId: '', dependentObjectId: '', targetObjectId: '', graceMinutes: 0, maximumMinutes: 60 }
}
export const blankScenario = (): ScenarioDraft => ({ name: '', description: '', horizonMinutes: 90, objects: [], invariants: [] })

export function toScenarioRequest(draft: ScenarioDraft): ScenarioRequest | null {
  if (draft.horizonMinutes === '') return null
  return {
    name: draft.name,
    ...(draft.description ? { description: draft.description } : {}),
    horizonMinutes: draft.horizonMinutes,
    objects: draft.objects.map((object) => ({
      clientId: object.clientId, name: object.name, kind: object.kind,
      events: object.events.filter((event) => event.atMinute !== '').map(({ type, atMinute }) => ({ type, atMinute: atMinute as number })),
    })),
    invariants: draft.invariants.map((invariant) => {
      if (invariant.type === 'max_validity') return { type: invariant.type, targetObjectId: invariant.targetObjectId, ...(invariant.maximumMinutes !== '' ? { maximumMinutes: invariant.maximumMinutes } : {}) }
      return {
        type: invariant.type, sourceObjectId: invariant.sourceObjectId, dependentObjectId: invariant.dependentObjectId,
        ...(invariant.type === 'revoked_access_grace' && invariant.graceMinutes !== '' ? { graceMinutes: invariant.graceMinutes } : {}),
      }
    }),
  }
}

export const parseNumber = (value: string): number | '' => value === '' ? '' : Number(value)

export const scenarioFingerprint = (request: ScenarioRequest): string => JSON.stringify(request)

export function fromScenarioRequest(request: ScenarioRequest): ScenarioDraft {
  return {
    name: request.name, description: request.description ?? '', horizonMinutes: request.horizonMinutes,
    objects: request.objects.map((object, objectIndex) => ({ ...object, events: object.events.map((event, eventIndex) => ({ ...event, formId: `snapshot-event-${objectIndex}-${eventIndex}` })) })),
    invariants: request.invariants.map((invariant, index) => ({
      formId: `snapshot-invariant-${index}`, type: invariant.type, sourceObjectId: invariant.sourceObjectId ?? '', dependentObjectId: invariant.dependentObjectId ?? '', targetObjectId: invariant.targetObjectId ?? '', graceMinutes: invariant.graceMinutes ?? '', maximumMinutes: invariant.maximumMinutes ?? '',
    })),
  }
}
