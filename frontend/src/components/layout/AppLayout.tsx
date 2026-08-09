import { GitFork as Github, ShieldCheck } from 'lucide-react'
import { Link, NavLink, Outlet } from 'react-router'
import { ApiHealth } from '../ui/ApiHealth'

const githubUrl = import.meta.env.VITE_GITHUB_URL || 'https://github.com/kernelKain/timetrap'
export function AppLayout() {
  return <div className="app-shell">
    <header className="site-header"><div className="page-container header-inner">
      <Link className="brand" to="/"><span className="brand-mark"><ShieldCheck aria-hidden="true" /></span><span><strong>TimeTrap</strong><small>Temporal Authorization Analyzer</small></span></Link>
      <nav aria-label="Primary navigation">
        <NavLink to="/" end>Overview</NavLink><NavLink to="/scenario/new">Builder</NavLink>
        <a className="repo-link" href={githubUrl} target="_blank" rel="noreferrer"><Github aria-hidden="true" /> Repository</a>
      </nav>
      <ApiHealth />
    </div></header>
    <main><Outlet /></main>
    <footer className="site-footer"><div className="page-container footer-inner">
      <div><strong>TimeTrap</strong><p>Check when access may remain valid for too long.</p></div>
      <p>TimeTrap checks the scenario you provide. It does not inspect a live application.</p>
    </div></footer>
  </div>
}
