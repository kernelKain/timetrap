import { Check, CircleAlert } from 'lucide-react'

export function ScenarioStepNav({ objectCount, invariantCount, errorCount }: { objectCount: number; invariantCount: number; errorCount: number }) {
  const steps = [
    { label: 'Scenario', complete: true },
    { label: 'Objects & events', complete: objectCount > 0 },
    { label: 'Invariants', complete: invariantCount > 0 },
    { label: 'Analyze', complete: errorCount === 0, invalid: errorCount > 0 },
  ]
  return <nav className="scenario-step-nav" aria-label="Scenario progress">{steps.map((step, index) => <a key={step.label} href={index === 0 ? '#scenario-config' : index === 1 ? '#objects' : index === 2 ? '#invariants' : '#analyze'} className={step.invalid ? 'invalid' : step.complete ? 'complete' : ''}><span>{step.invalid ? <CircleAlert aria-hidden="true" /> : step.complete ? <Check aria-hidden="true" /> : index + 1}</span>{step.label}</a>)}</nav>
}
