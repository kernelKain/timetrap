import type { ScenarioDraft } from '../types/forms'
import { cloneScenario } from '../lib/scenario'

const subscriptionCancellation: ScenarioDraft = {
  name: 'Subscription cancellation leak',
  description: 'A premium entitlement cache remains valid after its source subscription is revoked.',
  horizonMinutes: 90,
  objects: [
    { clientId: 'subscription', name: 'Premium subscription', kind: 'subscription', events: [
      { formId: 'subscription-issue', type: 'issue', atMinute: 0 }, { formId: 'subscription-revoke', type: 'revoke', atMinute: 10 },
    ] },
    { clientId: 'cached_entitlement', name: 'Premium entitlement cache', kind: 'cached_entitlement', events: [
      { formId: 'cache-issue', type: 'issue', atMinute: 0 }, { formId: 'cache-expire', type: 'expire', atMinute: 60 },
    ] },
  ],
  invariants: [{ formId: 'revocation-grace', type: 'revoked_access_grace', sourceObjectId: 'subscription', dependentObjectId: 'cached_entitlement', targetObjectId: '', graceMinutes: 5, maximumMinutes: '' }],
}

const accountSuspension: ScenarioDraft = {
  name: 'Account suspension session', description: 'Check whether an active session outlives a suspended account.', horizonMinutes: 120,
  objects: [
    { clientId: 'account', name: 'Account credential', kind: 'credential', events: [{ formId: 'account-issue', type: 'issue', atMinute: 0 }, { formId: 'account-revoke', type: 'revoke', atMinute: 30 }] },
    { clientId: 'session', name: 'Application session', kind: 'session', events: [{ formId: 'session-issue', type: 'issue', atMinute: 5 }, { formId: 'session-expire', type: 'expire', atMinute: 90 }] },
  ],
  invariants: [{ formId: 'session-source', type: 'dependent_not_outlive_source', sourceObjectId: 'account', dependentObjectId: 'session', targetObjectId: '', graceMinutes: '', maximumMinutes: '' }],
}

export const templates = Object.freeze({ 'subscription-cancellation': subscriptionCancellation, 'account-suspension': accountSuspension })
export type TemplateId = keyof typeof templates
export const loadTemplate = (id: string): ScenarioDraft | null => id in templates ? cloneScenario(templates[id as TemplateId]) : null
export const isCanonicalSubscription = (draft: ScenarioDraft) => JSON.stringify(draft) === JSON.stringify(subscriptionCancellation)
