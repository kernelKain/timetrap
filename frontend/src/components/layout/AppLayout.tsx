import { GitFork, ShieldCheck } from 'lucide-react'
import { Link, NavLink, Outlet } from 'react-router'

const githubUrl = import.meta.env.VITE_GITHUB_URL || 'https://github.com/kernelKain/timetrap'
export function AppLayout() {
  return <div className="app-shell">
    <header className="site-header"><div className="page-container header-inner">
      <Link className="brand" to="/"><ShieldCheck aria-hidden="true" /> <span>TimeTrap</span></Link>
      <nav aria-label="Primary navigation">
        <NavLink to="/" end>Overview</NavLink><NavLink to="/scenario/new">Builder</NavLink>
        <a href={githubUrl} target="_blank" rel="noreferrer"><GitFork aria-hidden="true" /> Repository</a>
      </nav>
    </div></header>
    <main><Outlet /></main>
    <footer className="site-footer"><div className="page-container footer-inner">
      <div><strong>TimeTrap</strong><p>A design-time verifier for bounded authorization models.</p></div>
      <p>It does not inspect live systems or prove an application is secure.</p>
    </div></footer>
  </div>
}
