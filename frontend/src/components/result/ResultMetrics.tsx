import type { AnalysisResult, ScenarioRequest } from '../../types/api'
import { formatDuration } from '../../lib/timeline'

export function ResultMetrics({ result, scenario }: { result: AnalysisResult; scenario: ScenarioRequest }) {
  const finding = result.earliestViolation
  const metrics = result.safe ? [
    ['Simulation horizon', formatDuration(scenario.horizonMinutes)], ['Boundaries evaluated', String(result.evaluatedBoundaries)], ['Violations', '0'],
  ] : [
    ['Source invalidated', finding?.sourceInvalidatedMinute === undefined ? 'Not provided' : `Minute ${finding.sourceInvalidatedMinute}`],
    ['Policy deadline', finding?.policyDeadlineMinute === undefined ? 'Not provided' : `Minute ${finding.policyDeadlineMinute}`],
    ['Dependent ended', finding?.dependentInvalidatedMinute === undefined ? (finding?.ongoing ? 'Continues to horizon' : 'Not provided') : `Minute ${finding.dependentInvalidatedMinute}`],
    ['Policy violation', finding ? formatDuration(finding.durationMinutes) : 'Not provided'],
    ['Total stale exposure', finding?.totalStaleExposureMinutes === undefined ? 'Not provided' : formatDuration(finding.totalStaleExposureMinutes)],
  ]
  return <section className="result-metrics" aria-label="Analysis summary">{metrics.map(([label, value]) => <div className="metric-card" key={label}><span>{label}</span><strong>{value}</strong></div>)}</section>
}
