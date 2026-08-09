import { AlertTriangle, Clock3, Lightbulb } from 'lucide-react'
import { canonicalMockAnalysis as result } from '../../data/mockAnalysis'
export function CanonicalResult() {
  return <div className="result-grid">
    <section className="panel verdict-card"><div className="unsafe-badge"><AlertTriangle aria-hidden="true" /> {result.verdict}</div><h2>Cached access violates the configured policy.</h2><p>The five-minute grace is permitted. The violation begins at the deadline, not at revocation.</p><dl><div><dt>Policy violation</dt><dd>{result.violationInterval}</dd></div><div><dt>Violation duration</dt><dd>{result.violationDuration} minutes</dd></div><div><dt>Total stale exposure</dt><dd>{result.staleDuration} minutes</dd></div></dl></section>
    <section className="panel evidence-card"><div className="section-heading"><div><p className="eyebrow">Counterexample</p><h2>Evidence timeline</h2></div><Clock3 aria-hidden="true" /></div><ol className="result-steps"><li><span>10</span><div><strong>Source revoked</strong><p>The subscription stops authorizing access.</p></div></li><li><span>15</span><div><strong>Policy deadline</strong><p>Violation begins after the permitted grace.</p></div></li><li><span>60</span><div><strong>Cache expires</strong><p>The dependent finally becomes invalid.</p></div></li></ol></section>
    <section className="panel remediation"><Lightbulb aria-hidden="true" /><div><p className="eyebrow">Deterministic remediation</p><h2>{result.remediation}</h2><p>Phase 6 will return remediation from the Go analyzer. This fixture is intentionally static.</p></div></section>
  </div>
}
