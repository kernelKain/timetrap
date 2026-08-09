import { AlertTriangle, ShieldCheck } from 'lucide-react'
import type { AnalysisResult, ScenarioRequest } from '../../types/api'
import { formatDuration } from '../../lib/timeline'

const nameFor = (scenario: ScenarioRequest, id?: string) => scenario.objects.find((object) => object.clientId === id)?.name ?? id ?? 'authorization'

export function ResultHero({ result, scenario }: { result: AnalysisResult; scenario: ScenarioRequest }) {
  const finding = result.earliestViolation
  if (result.safe) return <section className="result-hero safe-result"><div className="safe-badge"><ShieldCheck aria-hidden="true" /> Safe within this model</div><h1>All configured rules passed within the {scenario.horizonMinutes}-minute simulation horizon.</h1><p>This result applies only to the submitted model and horizon. It is not a guarantee about unmodelled application behavior.</p></section>
  const dependent = nameFor(scenario, finding?.dependentObjectId ?? finding?.targetObjectId)
  return <section className="result-hero unsafe-result"><div className="unsafe-badge"><AlertTriangle aria-hidden="true" /> Unsafe</div><h1>{finding ? `${dependent} exceeded its policy deadline by ${formatDuration(finding.durationMinutes)}.` : 'The analyzer found a policy violation.'}</h1><p>{finding?.sourceInvalidatedMinute !== undefined && finding.policyDeadlineMinute !== undefined && finding.dependentInvalidatedMinute !== undefined ? `The source became invalid at minute ${finding.sourceInvalidatedMinute}. Access was required to end by minute ${finding.policyDeadlineMinute}, but ${dependent} remained valid until minute ${finding.dependentInvalidatedMinute}.` : 'The persisted analyzer report contains the counterexample below.'}</p></section>
}
