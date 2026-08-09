import { ChevronDown } from 'lucide-react'
import type { PersistedAnalysis, ScenarioRequest } from '../../types/api'
import { formatDuration, formatInterval } from '../../lib/timeline'

const nameFor = (scenario: ScenarioRequest, id?: string) => scenario.objects.find((object) => object.clientId === id)?.name ?? id ?? 'Not provided'

export function EvidenceCard({ analysis, scenario }: { analysis: PersistedAnalysis; scenario: ScenarioRequest }) {
  const finding = analysis.result.earliestViolation
  if (!finding) return <section className="panel result-section evidence-panel"><p className="eyebrow">Evidence</p><h2>No counterexample was returned</h2><p>The server evaluated {analysis.result.evaluatedBoundaries} boundaries and found no configured invariant violation.</p></section>

  return <details className="panel result-section evidence-panel" open>
    <summary><span><span className="eyebrow">Evidence</span><h2>What failed, when, and for how long</h2></span><ChevronDown aria-hidden="true" /></summary>
    <div className="evidence-body"><p><strong>{nameFor(scenario, finding.dependentObjectId ?? finding.targetObjectId)}</strong> violated <code>{finding.invariantType}</code> during {formatInterval(finding.startMinute, finding.endMinute, finding.ongoing)} for {formatDuration(finding.durationMinutes)}.</p>
      <dl className="diagnostic-output"><div><dt>source.object</dt><dd>{nameFor(scenario, finding.sourceObjectId)}</dd></div><div><dt>dependent.object</dt><dd>{nameFor(scenario, finding.dependentObjectId ?? finding.targetObjectId)}</dd></div>{finding.sourceInvalidatedMinute !== undefined && <div><dt>source.revoked_at</dt><dd>{finding.sourceInvalidatedMinute}m</dd></div>}{finding.policyDeadlineMinute !== undefined && <div><dt>grace.deadline</dt><dd>{finding.policyDeadlineMinute}m</dd></div>}{finding.dependentInvalidatedMinute !== undefined && <div><dt>dependent.expires_at</dt><dd>{finding.dependentInvalidatedMinute}m</dd></div>}<div><dt>violation.interval</dt><dd>{formatInterval(finding.startMinute, finding.endMinute, finding.ongoing)}</dd></div><div><dt>violation.duration</dt><dd>{formatDuration(finding.durationMinutes)}</dd></div><div><dt>invariant</dt><dd>{finding.invariantType}</dd></div></dl>
      <ol className="evidence-list">{finding.evidence.map((item, index) => <li key={`${item.objectId}-${item.atMinute}-${index}`}><span>{item.atMinute}m</span><div><strong>{nameFor(scenario, item.objectId)} · {item.valid ? 'valid' : 'invalid'}</strong><p>{item.detail}</p>{item.eventType && <small>{item.eventType} at minute {item.eventMinute}</small>}</div></li>)}</ol>
    </div>
  </details>
}
