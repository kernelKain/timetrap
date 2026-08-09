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
      <dl className="technical-facts"><div><dt>Source</dt><dd>{nameFor(scenario, finding.sourceObjectId)}</dd></div><div><dt>Dependent / target</dt><dd>{nameFor(scenario, finding.dependentObjectId ?? finding.targetObjectId)}</dd></div>{finding.graceMinutes !== undefined && <div><dt>Allowed grace</dt><dd>{formatDuration(finding.graceMinutes)}</dd></div>}<div><dt>Failed rule</dt><dd>{finding.invariantType.replaceAll('_', ' ')}</dd></div></dl>
      <ol className="evidence-list">{finding.evidence.map((item, index) => <li key={`${item.objectId}-${item.atMinute}-${index}`}><span>{item.atMinute}m</span><div><strong>{nameFor(scenario, item.objectId)} · {item.valid ? 'valid' : 'invalid'}</strong><p>{item.detail}</p>{item.eventType && <small>{item.eventType} at minute {item.eventMinute}</small>}</div></li>)}</ol>
    </div>
  </details>
}
