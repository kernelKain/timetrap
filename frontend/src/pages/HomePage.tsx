import { ArrowRight, Boxes, Code2, Database, GitFork, Layers3, Play, Route, ScanSearch } from 'lucide-react'
import { Link } from 'react-router'
import { ApiHealth } from '../components/ui/ApiHealth'

export function HomePage() {
  return <>
    <section className="hero-section"><div className="page-container hero-grid"><div className="hero-copy">
      <div className="kicker"><ScanSearch aria-hidden="true" /> Design-time authorization verifier</div>
      <h1>Find the minute <span>trust outlives truth.</span></h1>
      <p>Model subscriptions, caches, sessions, and tokens as bounded timelines. TimeTrap makes hidden authorization timing gaps visible before they become production behavior.</p>
      <div className="hero-actions"><Link className="button primary" to="/scenario/new?template=subscription-cancellation"><Play aria-hidden="true" /> Try subscription cancellation</Link><Link className="button secondary" to="/scenario/new">Start blank scenario <ArrowRight aria-hidden="true" /></Link></div>
      <ApiHealth />
    </div><MiniTimeline /></div></section>
    <section className="content-section"><div className="page-container"><div className="section-intro"><p className="eyebrow">One timing disagreement</p><h2>A cancellation at minute 10 can leave access alive until minute 60.</h2><p>TimeTrap separates total stale exposure from the actual policy violation: `[10,60)` is 50 minutes stale, while `[15,60)` is a 45-minute violation after the five-minute grace.</p></div>
      <div className="steps-grid"><Step icon={<Boxes />} number="01" title="Describe the model">Add timed objects and the events that change their validity.</Step><Step icon={<Layers3 />} number="02" title="Choose an invariant">State how dependent authorization must relate to its source.</Step><Step icon={<Route />} number="03" title="See the interval">Phase 6 will ask the deterministic Go analyzer for the counterexample.</Step></div>
    </div></section>
    <section className="content-section tinted"><div className="page-container"><div className="section-heading"><div><p className="eyebrow">Built-in starting points</p><h2>Load a model, then make it yours.</h2></div></div><div className="template-grid">
      <TemplateCard title="Subscription cancellation" description="A cached premium entitlement survives source revocation beyond a five-minute grace." to="/scenario/new?template=subscription-cancellation" featured />
      <TemplateCard title="Account suspension" description="An application session remains active after its account credential is suspended." to="/scenario/new?template=account-suspension" />
    </div></div></section>
    <section className="content-section"><div className="page-container architecture-card"><div><p className="eyebrow">Built on Zerops</p><h2>A focused React interface over a deterministic Go core.</h2><p>This Phase 5 experience is local-only. The existing health request remains, but scenarios are not sent to the API until Phase 6.</p><a className="text-link" href="https://github.com/kernelKain/timetrap" target="_blank" rel="noreferrer"><GitFork aria-hidden="true" /> Explore the repository</a></div><div className="architecture-flow"><span><Code2 /> React + TypeScript</span><ArrowRight /><span><Database /> Go API + PostgreSQL</span></div></div></section>
  </>
}
function MiniTimeline() { return <div className="hero-visual"><div className="visual-top"><span>Subscription cancellation</span><span className="unsafe-badge small">Illustrative</span></div><div className="mini-track"><div className="mini-window" /><Marker minute="10" label="Revoke" left="16.67%" tone="warning" /><Marker minute="15" label="Deadline" left="25%" tone="danger" /><Marker minute="60" label="Expiry" left="100%" tone="safe" /></div><div className="visual-callout"><strong>45-minute policy violation</strong><span>after the permitted grace</span></div><p>Static model preview. No backend analysis was run.</p></div> }
function Marker({ minute, label, left, tone }: { minute: string; label: string; left: string; tone: string }) { return <div className={`mini-marker ${tone}`} style={{ left }}><span /> <strong>{minute}m</strong><small>{label}</small></div> }
function Step({ icon, number, title, children }: { icon: React.ReactNode; number: string; title: string; children: React.ReactNode }) { return <article className="step-card"><div className="step-icon">{icon}</div><span>{number}</span><h3>{title}</h3><p>{children}</p></article> }
function TemplateCard({ title, description, to, featured = false }: { title: string; description: string; to: string; featured?: boolean }) { return <article className={`template-card ${featured ? 'featured' : ''}`}><div><p className="eyebrow">{featured ? 'Canonical demo' : 'Secondary example'}</p><h3>{title}</h3><p>{description}</p></div><Link className="button secondary" to={to}>Load template <ArrowRight aria-hidden="true" /></Link></article> }
