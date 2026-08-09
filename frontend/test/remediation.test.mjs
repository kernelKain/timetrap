import assert from 'node:assert/strict'
import test from 'node:test'
import { applyRemediation } from '../src/lib/remediation.ts'
import { fromScenarioRequest } from '../src/lib/scenario.ts'
import { validateScenario } from '../src/lib/validation.ts'

const snapshot = () => ({ name: 'Subscription cancellation leak', description: 'Canonical', horizonMinutes: 90, objects: [
  { clientId: 'subscription', name: 'Premium subscription', kind: 'subscription', events: [{ type: 'issue', atMinute: 0 }, { type: 'revoke', atMinute: 10 }] },
  { clientId: 'premium-cache', name: 'Premium entitlement cache', kind: 'cached_entitlement', events: [{ type: 'issue', atMinute: 0 }, { type: 'expire', atMinute: 60 }] },
], invariants: [{ type: 'revoked_access_grace', sourceObjectId: 'subscription', dependentObjectId: 'premium-cache', graceMinutes: 5 }] })
const canonical = { code: 'invalidate_dependent_on_source_revoke', summary: 'Revoke dependent.', operations: [{ type: 'add_event', objectId: 'premium-cache', event: { type: 'revoke', atMinute: 10 } }] }

test('applies canonical remediation immutably and preserves unrelated data', () => {
  const original = snapshot(); const before = structuredClone(original); const result = applyRemediation(original, canonical)
  assert.equal(result.ok, true); assert.deepEqual(original, before); assert.notEqual(result.scenario, original)
  assert.deepEqual(result.scenario.objects[1].events, [{ type: 'issue', atMinute: 0 }, { type: 'revoke', atMinute: 10 }, { type: 'expire', atMinute: 60 }])
  assert.deepEqual(result.scenario.objects[0], original.objects[0]); assert.equal(validateScenario(fromScenarioRequest(result.scenario)).length, 0)
})
test('keeps a corrected scenario name within the backend limit', () => {
  const original = snapshot(); original.name = 'A'.repeat(120); const result = applyRemediation(original, canonical)
  assert.equal(result.ok, true); assert.equal(Array.from(result.scenario.name).length, 120); assert.match(result.scenario.name, /corrected$/)
})
test('rejects missing objects and conflicting event minutes', () => {
  assert.equal(applyRemediation(snapshot(), { ...canonical, operations: [{ ...canonical.operations[0], objectId: 'missing' }] }).ok, false)
  assert.equal(applyRemediation(snapshot(), { ...canonical, operations: [{ ...canonical.operations[0], event: { type: 'revoke', atMinute: 60 } }] }).ok, false)
})
test('supports unambiguous replacement and maximum-duration operations', () => {
  const replaced = applyRemediation(snapshot(), { code: 'replace', summary: 'Replace.', operations: [{ type: 'replace_event', objectId: 'premium-cache', matchEvent: { type: 'expire', atMinute: 60 }, event: { type: 'expire', atMinute: 15 } }] })
  assert.equal(replaced.ok, true); assert.equal(replaced.scenario.objects[1].events[1].atMinute, 15)
  const maximumScenario = snapshot(); maximumScenario.invariants = [{ type: 'max_validity', targetObjectId: 'premium-cache', maximumMinutes: 60 }]
  const limited = applyRemediation(maximumScenario, { code: 'limit', summary: 'Limit.', operations: [{ type: 'set_maximum_duration', invariantIndex: 0, maximumMinutes: 15 }] })
  assert.equal(limited.ok, true); assert.equal(limited.scenario.invariants[0].maximumMinutes, 15)
})
test('rejects missing and unknown operations without guessing', () => {
  assert.equal(applyRemediation(snapshot(), { code: 'old', summary: 'Manual only.' }).ok, false)
  assert.equal(applyRemediation(snapshot(), { code: 'unknown', summary: 'Unknown.', operations: [{ type: 'parse_the_prose' }] }).ok, false)
})
