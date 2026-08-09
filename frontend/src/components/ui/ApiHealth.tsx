import { useEffect, useState } from 'react'
import { CircleCheck, LoaderCircle, TriangleAlert } from 'lucide-react'
import type { HealthResponse } from '../../types/api'
import { getHealth } from '../../lib/api'

type State = { kind: 'checking' } | { kind: 'healthy'; health: HealthResponse } | { kind: 'unavailable' }
export function ApiHealth() {
  const [state, setState] = useState<State>({ kind: 'checking' })
  useEffect(() => {
    const controller = new AbortController()
    getHealth({ signal: controller.signal })
      .then((health) => setState(health.status === 'ok' ? { kind: 'healthy', health } : { kind: 'unavailable' }))
      .catch((error: unknown) => { if (!(error instanceof DOMException && error.name === 'AbortError')) setState({ kind: 'unavailable' }) })
    return () => controller.abort()
  }, [])
  return <div className={`health-chip ${state.kind}`} aria-live="polite" title={state.kind === 'healthy' ? `Database: ${state.health.database}` : undefined}>
    {state.kind === 'checking' && <><LoaderCircle className="spin" aria-hidden="true" /> Connecting</>}
    {state.kind === 'healthy' && <><CircleCheck aria-hidden="true" /> API online</>}
    {state.kind === 'unavailable' && <><TriangleAlert aria-hidden="true" /> API offline</>}
  </div>
}
