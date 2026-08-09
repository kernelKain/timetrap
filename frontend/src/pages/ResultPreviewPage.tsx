import { ArrowLeft, FlaskConical } from 'lucide-react'
import { Link, useLocation } from 'react-router'
import { CanonicalResult } from '../components/result/CanonicalResult'
import { TimelinePreview } from '../components/timeline/TimelinePreview'
import { loadTemplate, isCanonicalSubscription } from '../data/templates'
import type { ScenarioDraft } from '../types/forms'

export function ResultPreviewPage() {
  const location = useLocation(); const supplied = (location.state as { scenario?: ScenarioDraft } | null)?.scenario
  const fallback = loadTemplate('subscription-cancellation')!; const scenario = supplied ?? fallback; const canonical = isCanonicalSubscription(scenario)
  return <div className="page-container result-page"><div className="preview-banner"><FlaskConical aria-hidden="true" /><div><strong>Static Phase 5 preview — no backend analysis was run.</strong><p>{supplied ? 'This layout reflects the scenario passed from the builder.' : 'Direct navigation loaded the safe canonical fallback fixture.'}</p></div></div>
    <header className="page-heading result-heading"><div><p className="eyebrow">Result preview</p><h1>{scenario.name || 'Untitled scenario'}</h1><p>{scenario.description || 'A static shell for the future deterministic analysis result.'}</p></div><Link className="button secondary" to="/scenario/new?template=subscription-cancellation"><ArrowLeft aria-hidden="true" /> Back to builder</Link></header>
    {canonical ? <CanonicalResult /> : <section className="panel custom-result"><p className="eyebrow">Custom scenario</p><h2>No result has been fabricated.</h2><p>Phase 6 will submit this model to the Go API and populate the verdict, evidence, counterexample, and remediation. Phase 5 intentionally shows only the layout shell.</p></section>}
    <TimelinePreview scenario={scenario} />
  </div>
}
