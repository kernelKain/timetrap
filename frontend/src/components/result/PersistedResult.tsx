import { AlertTriangle, CheckCircle2, Clock3, Lightbulb } from 'lucide-react'
import type { PersistedAnalysis, ScenarioRequest, Violation } from '../../types/api'

const objectName = (scenario: ScenarioRequest, id?: string) => scenario.objects.find((object) => object.clientId === id)?.name ?? id ?? 'Not applicable'
export function PersistedResult({ analysis, scenario }: { analysis: PersistedAnalysis; scenario: ScenarioRequest }) {
  const earliest = analysis.result.earliestViolation
  return <div className="result-grid">
    <section className={`panel verdict-card ${analysis.result.safe ? 'safe-result' : ''}`}>
      <div className={analysis.result.safe ? 'safe-badge' : 'unsafe-badge'}>{analysis.result.safe ? <CheckCircle2 aria-hidden="true" /> : <AlertTriangle aria-hidden="true" />}{analysis.result.safe ? 'Safe' : 'Unsafe'}</div>
      <h2>{analysis.result.safe ? 'No invariant failed inside this horizon.' : 'The analyzer found a policy violation.'}</h2>
      <p>This verdict and every interval below came from the persisted Go analysis.</p>
      <dl><div><dt>Horizon</dt><dd>{scenario.horizonMinutes} minutes</dd></div><div><dt>Boundaries evaluated</dt><dd>{analysis.result.evaluatedBoundaries}</dd></div><div><dt>Violations</dt><dd>{analysis.result.violations.length}</dd></div></dl>
    </section>
    <section className="panel evidence-card"><div className="section-heading"><div><p className="eyebrow">Persisted report</p><h2>{earliest ? 'Earliest counterexample' : 'Analysis details'}</h2></div><Clock3 aria-hidden="true" /></div>
      {earliest ? <ViolationDetails violation={earliest} scenario={scenario} /> : <p className="result-empty">The backend returned no violation evidence for this safe model.</p>}
    </section>
    {earliest?.remediation && <section className="panel remediation"><Lightbulb aria-hidden="true" /><div><p className="eyebrow">Deterministic remediation</p><h2>{earliest.remediation.summary}</h2><p>Code: <code>{earliest.remediation.code}</code>{earliest.remediation.objectId ? ` · Object: ${objectName(scenario, earliest.remediation.objectId)}` : ''}</p></div></section>}
    {analysis.result.violations.length > 1 && <section className="panel all-violations"><p className="eyebrow">All server findings</p><h2>{analysis.result.violations.length} violation intervals</h2><ol>{analysis.result.violations.map((violation, index) => <li key={`${violation.invariantIndex}-${violation.startMinute}-${index}`}><strong>[{violation.startMinute},{violation.endMinute})</strong> · {violation.durationMinutes} minutes · {violation.invariantType.replaceAll('_', ' ')}</li>)}</ol></section>}
  </div>
}

function ViolationDetails({ violation, scenario }: { violation: Violation; scenario: ScenarioRequest }) {
  return <div><dl className="violation-facts"><div><dt>Interval</dt><dd>[{violation.startMinute},{violation.endMinute})</dd></div><div><dt>Duration</dt><dd>{violation.durationMinutes} minutes</dd></div><div><dt>Invariant</dt><dd>{violation.invariantType.replaceAll('_', ' ')}</dd></div><div><dt>Source</dt><dd>{objectName(scenario, violation.sourceObjectId)}</dd></div><div><dt>Dependent / target</dt><dd>{objectName(scenario, violation.dependentObjectId ?? violation.targetObjectId)}</dd></div></dl>
    <h3>Evidence returned by the analyzer</h3><ol className="evidence-list">{violation.evidence.map((evidence, index) => <li key={`${evidence.objectId}-${evidence.atMinute}-${index}`}><span>{evidence.atMinute}m</span><div><strong>{objectName(scenario, evidence.objectId)} · {evidence.valid ? 'valid' : 'invalid'}</strong><p>{evidence.detail}</p>{evidence.eventType && <small>{evidence.eventType} event{evidence.eventMinute !== undefined ? ` at ${evidence.eventMinute}m` : ''}</small>}</div></li>)}</ol>
  </div>
}
