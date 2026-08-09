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
  return <div className="health-chip" aria-live="polite">
    {state.kind === 'checking' && <><LoaderCircle className="spin" aria-hidden="true" /> API check</>}
    {state.kind === 'healthy' && <><CircleCheck aria-hidden="true" /> API healthy · {state.health.database}</>}
    {state.kind === 'unavailable' && <><TriangleAlert aria-hidden="true" /> API unavailable · builder still works locally</>}
  </div>
}
