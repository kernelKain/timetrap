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
    <section className="content-section"><div className="page-container"><div className="section-intro"><p className="eyebrow">One timing disagreement</p><h2>A cancellation can leave dependent authorization alive beyond policy.</h2><p>TimeTrap persists the model, runs the deterministic Go analyzer, and returns the exact server-calculated counterexample at a shareable result URL.</p></div>
      <div className="steps-grid"><Step icon={<Boxes />} number="01" title="Describe the model">Add timed objects and the events that change their validity.</Step><Step icon={<Layers3 />} number="02" title="Choose an invariant">State how dependent authorization must relate to its source.</Step><Step icon={<Route />} number="03" title="See the interval">Phase 6 will ask the deterministic Go analyzer for the counterexample.</Step></div>
    </div></section>
    <section className="content-section tinted"><div className="page-container"><div className="section-heading"><div><p className="eyebrow">Built-in starting points</p><h2>Load a model, then make it yours.</h2></div></div><div className="template-grid">
      <TemplateCard title="Subscription cancellation" description="A cached premium entitlement survives source revocation beyond a five-minute grace." to="/scenario/new?template=subscription-cancellation" featured />
      <TemplateCard title="Account suspension" description="An application session remains active after its account credential is suspended." to="/scenario/new?template=account-suspension" />
    </div></div></section>
    <section className="content-section"><div className="page-container architecture-card"><div><p className="eyebrow">Built on Zerops</p><h2>A focused React interface over a deterministic Go core.</h2><p>Scenarios and immutable analysis snapshots are persisted by the Go API, so result links survive refresh and sharing.</p><a className="text-link" href="https://github.com/kernelKain/timetrap" target="_blank" rel="noreferrer"><GitFork aria-hidden="true" /> Explore the repository</a></div><div className="architecture-flow"><span><Code2 /> React + TypeScript</span><ArrowRight /><span><Database /> Go API + PostgreSQL</span></div></div></section>
  </>
}
function MiniTimeline() { return <div className="hero-visual"><div className="visual-top"><span>Persisted analysis workflow</span><span className="safe-badge small">Server-backed</span></div><div className="workflow-stack"><span>Save scenario</span><ArrowRight /><span>Run Go analyzer</span><ArrowRight /><span>Open shareable result</span></div><div className="visual-callout"><strong>Evidence comes from the API</strong><span>never recalculated in React</span></div><p>Every result page reloads its immutable analysis snapshot by UUID.</p></div> }
function Step({ icon, number, title, children }: { icon: React.ReactNode; number: string; title: string; children: React.ReactNode }) { return <article className="step-card"><div className="step-icon">{icon}</div><span>{number}</span><h3>{title}</h3><p>{children}</p></article> }
function TemplateCard({ title, description, to, featured = false }: { title: string; description: string; to: string; featured?: boolean }) { return <article className={`template-card ${featured ? 'featured' : ''}`}><div><p className="eyebrow">{featured ? 'Canonical demo' : 'Secondary example'}</p><h3>{title}</h3><p>{description}</p></div><Link className="button secondary" to={to}>Load template <ArrowRight aria-hidden="true" /></Link></article> }
