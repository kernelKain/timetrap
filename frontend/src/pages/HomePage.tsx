import { ArrowRight, Boxes, Code2, Database, GitFork, Layers3, Play, Route, ScanSearch } from 'lucide-react'
import { Link } from 'react-router'
import { ApiHealth } from '../components/ui/ApiHealth'

export function HomePage() {
  return <>
    <section className="hero-section"><div className="page-container hero-grid"><div className="hero-copy">
      <div className="kicker"><ScanSearch aria-hidden="true" /> Authorization timing checker</div>
      <h1>Find the minute <span>trust outlives truth.</span></h1>
      <p>Describe when subscriptions, caches, sessions, and tokens start and stop. TimeTrap shows the exact window where access continues after it should have ended.</p>
      <div className="hero-actions"><Link className="button primary" to="/scenario/new?template=subscription-cancellation"><Play aria-hidden="true" /> Try subscription cancellation</Link><Link className="button secondary" to="/scenario/new">Start blank scenario <ArrowRight aria-hidden="true" /></Link></div>
      <ApiHealth />
    </div><MiniTimeline /></div></section>
    <section className="content-section"><div className="page-container"><div className="section-intro"><p className="eyebrow">A simple timing problem</p><h2>A cancellation can happen now while cached access remains active.</h2><p>TimeTrap saves your scenario, checks it with the deterministic Go analyzer, and gives you a clear result you can refresh or share.</p></div>
      <div className="steps-grid"><Step icon={<Boxes />} number="01" title="Describe the scenario">Add the access objects and the events that start, refresh, or end them.</Step><Step icon={<Layers3 />} number="02" title="Set the rule">Choose how one object should depend on another.</Step><Step icon={<Route />} number="03" title="See what went wrong">Get the exact time window, supporting evidence, and a suggested correction.</Step></div>
    </div></section>
    <section className="content-section tinted"><div className="page-container"><div className="section-heading"><div><p className="eyebrow">Built-in starting points</p><h2>Load a model, then make it yours.</h2></div></div><div className="template-grid">
      <TemplateCard title="Subscription cancellation" description="A cached premium entitlement survives source revocation beyond a five-minute grace." to="/scenario/new?template=subscription-cancellation" featured />
      <TemplateCard title="Account suspension" description="An application session remains active after its account credential is suspended." to="/scenario/new?template=account-suspension" />
    </div></div></section>
    <section className="content-section"><div className="page-container architecture-card"><div><p className="eyebrow">Built on Zerops</p><h2>A clear interface backed by a deterministic Go analyzer.</h2><p>Every scenario and result is stored by the API, so a result link still works after a refresh or when shared with someone else.</p><a className="text-link" href="https://github.com/kernelKain/timetrap" target="_blank" rel="noreferrer"><GitFork aria-hidden="true" /> Explore the repository</a></div><div className="architecture-flow"><span><Code2 /> React + TypeScript</span><ArrowRight /><span><Database /> Go API + PostgreSQL</span></div></div></section>
  </>
}
function MiniTimeline() { return <div className="hero-visual"><div className="visual-top"><span>How it works</span><span className="safe-badge small">Saved results</span></div><div className="workflow-stack"><span>Save scenario</span><ArrowRight /><span>Check timing</span><ArrowRight /><span>Share the result</span></div><div className="visual-callout"><strong>Results come from the Go analyzer</strong><span>clear and repeatable</span></div><p>Each result has its own permanent link and stored scenario snapshot.</p></div> }
function Step({ icon, number, title, children }: { icon: React.ReactNode; number: string; title: string; children: React.ReactNode }) { return <article className="step-card"><div className="step-icon">{icon}</div><span>{number}</span><h3>{title}</h3><p>{children}</p></article> }
function TemplateCard({ title, description, to, featured = false }: { title: string; description: string; to: string; featured?: boolean }) { return <article className={`template-card ${featured ? 'featured' : ''}`}><div><p className="eyebrow">{featured ? 'Canonical demo' : 'Secondary example'}</p><h3>{title}</h3><p>{description}</p></div><Link className="button secondary" to={to}>Load template <ArrowRight aria-hidden="true" /></Link></article> }
