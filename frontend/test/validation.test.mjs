import assert from 'node:assert/strict'
import test from 'node:test'
import { validateScenario } from '../src/lib/validation.ts'

const canonical = () => ({
  name: 'Subscription cancellation leak', description: '', horizonMinutes: 90,
  objects: [
    { clientId: 'subscription', name: 'Subscription', kind: 'subscription', events: [{ formId: 's0', type: 'issue', atMinute: 0 }, { formId: 's10', type: 'revoke', atMinute: 10 }] },
    { clientId: 'premium-cache', name: 'Premium cache', kind: 'cached_entitlement', events: [{ formId: 'c0', type: 'issue', atMinute: 0 }, { formId: 'c60', type: 'expire', atMinute: 60 }] },
  ],
  invariants: [{ formId: 'rule', type: 'revoked_access_grace', sourceObjectId: 'subscription', dependentObjectId: 'premium-cache', targetObjectId: '', graceMinutes: 5, maximumMinutes: '' }],
})

const fields = (draft) => validateScenario(draft).map((entry) => entry.field)

test('accepts the canonical draft including minute zero', () => assert.deepEqual(validateScenario(canonical()), []))
test('reports empty numeric inputs with stable paths', () => {
  const draft = canonical(); draft.horizonMinutes = ''; draft.objects[0].events[0].atMinute = ''; draft.invariants[0].graceMinutes = ''
  assert.deepEqual(fields(draft), ['horizonMinutes', 'objects[0].events[0].atMinute', 'invariants[0].graceMinutes'])
})
test('reports horizon overflow and a duplicate event minute at the edited row', () => {
  const outside = canonical(); outside.objects[1].events[1].atMinute = 91
  assert.deepEqual(fields(outside), ['objects[1].events[1].atMinute'])
  const duplicate = canonical(); duplicate.objects[1].events[1].atMinute = 0
  assert.deepEqual(fields(duplicate), ['objects[1].events[1].atMinute'])
})
test('reports removed and identical invariant references without discarding the rule', () => {
  const removed = canonical(); removed.objects.splice(1, 1)
  assert.deepEqual(fields(removed), ['invariants[0].dependentObjectId'])
  const same = canonical(); same.invariants[0].dependentObjectId = 'subscription'
  assert.deepEqual(fields(same), ['invariants[0].dependentObjectId'])
})
