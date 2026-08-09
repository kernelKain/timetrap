import type { EventType, Remediation, ScenarioRequest } from '../types/api'
import { fromScenarioRequest } from './scenario.ts'
import { validateScenario } from './validation.ts'

export type RemediationResult = { ok: true; scenario: ScenarioRequest } | { ok: false; reason: string }
const eventTypes = new Set<EventType>(['issue', 'refresh', 'revoke', 'expire'])
const correctedSuffix = ' — corrected'

export function correctedScenarioCopy(snapshot: ScenarioRequest): ScenarioRequest {
  const clone = structuredClone(snapshot)
  const available = 120 - correctedSuffix.length
  const base = Array.from(clone.name).slice(0, available).join('').trimEnd()
  clone.name = `${base}${correctedSuffix}`
  return clone
}

export function applyRemediation(snapshot: ScenarioRequest, remediation: Remediation): RemediationResult {
  if (!remediation.operations?.length) return { ok: false, reason: 'This stored recommendation has no supported automatic operation.' }
  const clone = correctedScenarioCopy(snapshot)

  for (const operation of remediation.operations) {
    switch (operation.type) {
      case 'add_event': { const object = clone.objects.find((item) => item.clientId === operation.objectId)
        if (!object) return { ok: false, reason: `The referenced object “${operation.objectId}” no longer exists in the snapshot.` }
        if (!eventTypes.has(operation.event.type) || !Number.isInteger(operation.event.atMinute) || operation.event.atMinute < 0 || operation.event.atMinute > clone.horizonMinutes) return { ok: false, reason: 'The proposed event is outside the supported scenario timeline.' }
        if (object.events.some((event) => event.atMinute === operation.event.atMinute)) return { ok: false, reason: `An event already exists on ${object.name} at minute ${operation.event.atMinute}.` }
        object.events.push(structuredClone(operation.event)); object.events.sort((left, right) => left.atMinute - right.atMinute); break }
      case 'replace_event': { const object = clone.objects.find((item) => item.clientId === operation.objectId)
        if (!object) return { ok: false, reason: `The referenced object “${operation.objectId}” no longer exists in the snapshot.` }
        if (!eventTypes.has(operation.event.type) || operation.event.atMinute < 0 || operation.event.atMinute > clone.horizonMinutes) return { ok: false, reason: 'The replacement event is outside the supported scenario timeline.' }
        const matches = object.events.map((event, index) => ({ event, index })).filter(({ event }) => event.type === operation.matchEvent.type && event.atMinute === operation.matchEvent.atMinute)
        if (matches.length !== 1) return { ok: false, reason: 'The event to replace is missing or ambiguous.' }
        if (object.events.some((event, index) => index !== matches[0].index && event.atMinute === operation.event.atMinute)) return { ok: false, reason: `Another event already exists on ${object.name} at minute ${operation.event.atMinute}.` }
        object.events[matches[0].index] = structuredClone(operation.event); break }
      case 'set_maximum_duration': { const invariant = clone.invariants[operation.invariantIndex]
        if (!invariant || invariant.type !== 'max_validity' || !Number.isInteger(operation.maximumMinutes) || operation.maximumMinutes <= 0 || operation.maximumMinutes > clone.horizonMinutes) return { ok: false, reason: 'The proposed maximum-duration change is not valid for this snapshot.' }
        invariant.maximumMinutes = operation.maximumMinutes; break }
      default: return { ok: false, reason: 'This remediation operation is not supported by this version of TimeTrap.' }
    }
  }

  const errors = validateScenario(fromScenarioRequest(clone))
  return errors.length ? { ok: false, reason: `The corrected scenario is not valid: ${errors[0].message}` } : { ok: true, scenario: clone }
}
