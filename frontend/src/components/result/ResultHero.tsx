import { AlertTriangle, ShieldCheck } from 'lucide-react'
import type { AnalysisResult, ScenarioRequest } from '../../types/api'
import { formatDuration, formatInterval } from '../../lib/timeline'

const nameFor = (scenario: ScenarioRequest, id?: string) => scenario.objects.find((object) => object.clientId === id)?.name ?? id ?? 'authorization'

export function ResultHero({ result, scenario }: { result: AnalysisResult; scenario: ScenarioRequest }) {
  const finding = result.earliestViolation
  if (result.safe) return <section className="result-hero safe-result"><div className="safe-badge"><ShieldCheck aria-hidden="true" /> Verified safe within horizon</div><div className="verdict-code"><code>HORIZON {scenario.horizonMinutes}m</code><strong>0 violations</strong></div><h1>All configured authorization rules passed.</h1><p>This result applies only to the submitted model and its {scenario.horizonMinutes}-minute horizon. It does not cover unmodelled application behavior.</p></section>
  const dependent = nameFor(scenario, finding?.dependentObjectId ?? finding?.targetObjectId)
  return <section className="result-hero unsafe-result"><div className="unsafe-badge"><AlertTriangle aria-hidden="true" /> Violation</div>{finding && <div className="verdict-code"><code>{formatInterval(finding.startMinute, finding.endMinute, finding.ongoing)}</code><strong>{formatDuration(finding.durationMinutes)}</strong></div>}<h1>{finding ? `${dependent} exceeded its policy deadline.` : 'The analyzer found a policy violation.'}</h1><p>{finding?.sourceInvalidatedMinute !== undefined && finding.policyDeadlineMinute !== undefined && finding.dependentInvalidatedMinute !== undefined ? `The source became invalid at minute ${finding.sourceInvalidatedMinute}. Access was required to end by minute ${finding.policyDeadlineMinute}, but ${dependent} remained valid until minute ${finding.dependentInvalidatedMinute}.` : 'The persisted analyzer report contains the counterexample below.'}</p></section>
}
