import { Component, type ErrorInfo, type ReactNode } from 'react'
import { Link } from 'react-router'

export class AppErrorBoundary extends Component<{ children: ReactNode }, { failed: boolean }> {
  state = { failed: false }
  static getDerivedStateFromError() { return { failed: true } }
  componentDidCatch(error: Error, info: ErrorInfo) { if (import.meta.env.DEV) console.error('TimeTrap render failure', error, info) }
  render() { if (this.state.failed) return <div className="page-container not-found"><h1>Something went wrong while displaying TimeTrap.</h1><div className="state-actions"><button className="button primary" onClick={() => window.location.reload()}>Reload page</button><Link className="button secondary" to="/">Return home</Link></div></div>; return this.props.children }
}
