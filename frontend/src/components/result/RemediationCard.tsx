import { Lightbulb, LoaderCircle, Pencil, WandSparkles } from 'lucide-react'
import type { Remediation, ScenarioRequest } from '../../types/api'

const objectName = (scenario: ScenarioRequest, id?: string) => scenario.objects.find((object) => object.clientId === id)?.name ?? id ?? 'Unknown object'
const operationText = (remediation: Remediation, scenario: ScenarioRequest) => (remediation.operations ?? []).map((operation) => {
  if (operation.type === 'add_event') return `Add ${operation.event.type} to ${objectName(scenario, operation.objectId)} at minute ${operation.event.atMinute}`
  if (operation.type === 'replace_event') return `Replace ${operation.matchEvent.type} at minute ${operation.matchEvent.atMinute} with ${operation.event.type} at minute ${operation.event.atMinute} on ${objectName(scenario, operation.objectId)}`
  return `Set invariant ${operation.invariantIndex + 1} maximum duration to ${operation.maximumMinutes} minutes`
})

export function RemediationCard({ remediation, scenario, stage, error, onApply, onEdit }: { remediation: Remediation; scenario: ScenarioRequest; stage: string; error?: string; onApply: () => void; onEdit: () => void }) {
  const operations = operationText(remediation, scenario); const supported = operations.length > 0; const active = !['idle', 'failed'].includes(stage)
  const label = stage === 'applying-fix' ? 'Applying fix…' : stage === 'saving-corrected-scenario' ? 'Saving corrected scenario…' : stage === 'rerunning-analysis' ? 'Rerunning analysis…' : stage === 'opening-corrected-result' ? 'Opening corrected result…' : stage === 'failed' ? 'Retry fix and rerun' : 'Apply fix and rerun'
  return <section className="panel result-section remediation-panel"><div className="remediation-heading"><Lightbulb aria-hidden="true" /><div><p className="eyebrow">Suggested correction</p><h2>{remediation.title || remediation.summary}</h2><p>{remediation.summary}</p></div></div><div className="operation-preview"><strong>Proposed operation</strong>{operations.length ? <ul>{operations.map((operation) => <li key={operation}>{operation}</li>)}</ul> : <p>Automatic application is unavailable for this older or unsupported recommendation.</p>}</div>{error && <div className="apply-error" role="alert">{error}</div>}<div className="remediation-actions"><button className="button primary" disabled={!supported || active} onClick={onApply}>{active ? <LoaderCircle className="spin" aria-hidden="true" /> : <WandSparkles aria-hidden="true" />}{label}</button><button className="button secondary" disabled={active} onClick={onEdit}><Pencil aria-hidden="true" /> Edit manually</button></div></section>
}
