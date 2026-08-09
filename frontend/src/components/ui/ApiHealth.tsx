import { useEffect, useState } from 'react'
import { CircleCheck, LoaderCircle, TriangleAlert } from 'lucide-react'
import type { HealthResponse } from '../../types/api'

type State = { kind: 'checking' } | { kind: 'healthy'; health: HealthResponse } | { kind: 'unavailable' }
export function ApiHealth() {
  const [state, setState] = useState<State>({ kind: 'checking' })
  useEffect(() => {
    const controller = new AbortController()
    fetch('/api/v1/health', { headers: { Accept: 'application/json' }, signal: controller.signal })
      .then(async (response) => { if (!response.ok) throw new Error(); return response.json() as Promise<HealthResponse> })
      .then((health) => setState(health.status === 'ok' ? { kind: 'healthy', health } : { kind: 'unavailable' }))
      .catch((error: unknown) => { if (!(error instanceof DOMException && error.name === 'AbortError')) setState({ kind: 'unavailable' }) })
    return () => controller.abort()
  }, [])
  return <div className="health-chip" aria-live="polite">
    {state.kind === 'checking' && <><LoaderCircle className="spin" aria-hidden="true" /> API check</>}
    {state.kind === 'healthy' && <><CircleCheck aria-hidden="true" /> API healthy · {state.health.database}</>}
    {state.kind === 'unavailable' && <><TriangleAlert aria-hidden="true" /> API unavailable · builder still works locally</>}
  </div>
}
