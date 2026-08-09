import type { ScenarioDraft, ValidationIssue } from '../types/forms'

const issue = (field: string, message: string): ValidationIssue => ({ field, message })
export function validateScenario(draft: ScenarioDraft): ValidationIssue[] {
  const errors: ValidationIssue[] = []
  const len = (value: string) => [...value].length
  if (!draft.name.trim()) errors.push(issue('name', 'Scenario name is required.'))
  else if (len(draft.name) > 120) errors.push(issue('name', 'Must contain at most 120 characters.'))
  if (len(draft.description) > 1000) errors.push(issue('description', 'Must contain at most 1,000 characters.'))
  if (draft.horizonMinutes === '' || draft.horizonMinutes < 1 || draft.horizonMinutes > 10080) errors.push(issue('horizonMinutes', 'Must be between 1 and 10,080.'))
  if (draft.objects.length < 1 || draft.objects.length > 25) errors.push(issue('objects', 'Add between 1 and 25 objects.'))
  const ids = new Set<string>(); let totalEvents = 0
  draft.objects.forEach((object, oi) => {
    const prefix = `objects[${oi}]`; totalEvents += object.events.length
    if (!object.clientId.trim()) errors.push(issue(`${prefix}.clientId`, 'Client ID is required.'))
    else if (ids.has(object.clientId)) errors.push(issue(`${prefix}.clientId`, 'Client ID must be unique.'))
    else ids.add(object.clientId)
    if (!object.name.trim()) errors.push(issue(`${prefix}.name`, 'Object name is required.'))
    else if (len(object.name) > 80) errors.push(issue(`${prefix}.name`, 'Must contain at most 80 characters.'))
    if (object.events.length < 1 || object.events.length > 50) errors.push(issue(`${prefix}.events`, 'Add between 1 and 50 events.'))
    const minutes = new Map<number, number>()
    object.events.forEach((event, ei) => {
      const field = `${prefix}.events[${ei}].atMinute`
      if (event.atMinute === '' || !Number.isInteger(event.atMinute) || event.atMinute < 0 || (draft.horizonMinutes !== '' && event.atMinute > draft.horizonMinutes)) errors.push(issue(field, 'Must be a whole minute between 0 and the horizon.'))
      else if (minutes.has(event.atMinute)) errors.push(issue(field, `Conflicts with event ${minutes.get(event.atMinute)! + 1}; only one event is allowed at a minute.`))
      else minutes.set(event.atMinute, ei)
    })
  })
  if (totalEvents > 250) errors.push(issue('objects', 'Scenarios may contain at most 250 events total.'))
  if (draft.invariants.length < 1 || draft.invariants.length > 10) errors.push(issue('invariants', 'Add between 1 and 10 invariants.'))
  const objectIds = new Set(draft.objects.map((object) => object.clientId))
  draft.invariants.forEach((invariant, ii) => {
    const prefix = `invariants[${ii}]`
    const reference = (field: string, value: string) => { if (!value || !objectIds.has(value)) errors.push(issue(`${prefix}.${field}`, 'Select an existing object.')) }
    if (invariant.type === 'max_validity') {
      reference('targetObjectId', invariant.targetObjectId)
      if (invariant.maximumMinutes === '' || invariant.maximumMinutes <= 0 || (draft.horizonMinutes !== '' && invariant.maximumMinutes > draft.horizonMinutes)) errors.push(issue(`${prefix}.maximumMinutes`, 'Must be greater than zero and no more than the horizon.'))
    } else {
      reference('sourceObjectId', invariant.sourceObjectId); reference('dependentObjectId', invariant.dependentObjectId)
      if (invariant.sourceObjectId && invariant.sourceObjectId === invariant.dependentObjectId) errors.push(issue(`${prefix}.dependentObjectId`, 'Dependent must be different from source.'))
      if (invariant.type === 'revoked_access_grace' && (invariant.graceMinutes === '' || invariant.graceMinutes < 0 || (draft.horizonMinutes !== '' && invariant.graceMinutes > draft.horizonMinutes))) errors.push(issue(`${prefix}.graceMinutes`, 'Must be zero or greater and no more than the horizon.'))
    }
  })
  return errors
}

export const errorsFor = (issues: ValidationIssue[], field: string) => issues.filter((entry) => entry.field === field)
