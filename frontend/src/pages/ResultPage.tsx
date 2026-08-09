import { useCallback, useEffect, useState } from 'react'
import { ArrowLeft, Clipboard, ClipboardCheck, Link2Off, LoaderCircle, RefreshCw, TriangleAlert } from 'lucide-react'
import { Link, useParams } from 'react-router'
import { PersistedResult } from '../components/result/PersistedResult'
import { TimelinePreview } from '../components/timeline/TimelinePreview'
import { ApiError, getAnalysis } from '../lib/api'
import { fromScenarioRequest } from '../lib/scenario'
import type { AnalysisResponse } from '../types/api'

type LoadState = { kind: 'loading' } | { kind: 'success'; response: AnalysisResponse } | { kind: 'not-found'; requestId?: string } | { kind: 'error'; message: string; requestId?: string }
const uuidPattern = /^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i

export function ResultPage() {
  const { analysisId } = useParams(); const validId = Boolean(analysisId && uuidPattern.test(analysisId)); const [state, setState] = useState<LoadState>(validId ? { kind: 'loading' } : { kind: 'error', message: 'This result link does not contain a valid analysis UUID.' }); const [attempt, setAttempt] = useState(0)
  useEffect(() => {
    if (!analysisId || !validId) return
    const controller = new AbortController(); let current = true
    getAnalysis(analysisId, { signal: controller.signal }).then((response) => { if (current) setState({ kind: 'success', response }) }).catch((error: unknown) => {
      if (!current || (error instanceof DOMException && error.name === 'AbortError')) return
      if (error instanceof ApiError && error.status === 404) setState({ kind: 'not-found', requestId: error.requestId })
      else setState({ kind: 'error', message: error instanceof ApiError ? error.message : 'The result could not be loaded.', requestId: error instanceof ApiError ? error.requestId : undefined })
    })
    return () => { current = false; controller.abort() }
  }, [analysisId, attempt, validId])
  const retry = useCallback(() => { setState({ kind: 'loading' }); setAttempt((value) => value + 1) }, [])
  if (state.kind === 'loading') return <ResultState icon={<LoaderCircle className="spin" />} title="Loading persisted analysis…" message="Fetching the report identified by this URL." />
  if (state.kind === 'not-found') return <ResultState icon={<Link2Off />} title="Analysis not found" message="No persisted analysis exists for this UUID." requestId={state.requestId} />
  if (state.kind === 'error') return <ResultState icon={<TriangleAlert />} title="Unable to open this result" message={state.message} requestId={state.requestId} retry={validId ? retry : undefined} />
  const { analysis } = state.response; const scenario = analysis.scenarioSnapshot
  if (!scenario) return <ResultState icon={<TriangleAlert />} title="Incomplete persisted result" message="The API response did not include the scenario snapshot required to display this report." requestId={state.response.requestId} retry={retry} />
  return <div className="page-container result-page"><header className="page-heading result-heading"><div><p className="eyebrow">Persisted analysis</p><h1>{scenario.name}</h1><p>{scenario.description || 'A server-backed deterministic TimeTrap report.'}</p></div><div className="result-actions"><CopyResultLink /><Link className="button secondary" to="/scenario/new?template=subscription-cancellation"><ArrowLeft aria-hidden="true" /> Back to builder</Link></div></header>
    <div className="analysis-meta"><span>Analysis <code>{analysis.id}</code></span><span>Scenario <code>{analysis.scenarioId}</code></span><span>Created {new Date(analysis.createdAt).toLocaleString()}</span><span>Engine {analysis.engineVersion}</span></div>
    <PersistedResult analysis={analysis} scenario={scenario} /><TimelinePreview scenario={fromScenarioRequest(scenario)} />
  </div>
}

function ResultState({ icon, title, message, requestId, retry }: { icon: React.ReactNode; title: string; message: string; requestId?: string; retry?: () => void }) { return <div className="page-container not-found result-state">{icon}<p className="eyebrow">Persisted result</p><h1>{title}</h1><p>{message}</p>{requestId && <small>Request ID: <code>{requestId}</code></small>}<div className="state-actions">{retry && <button className="button primary" onClick={retry}><RefreshCw /> Retry</button>}<Link className="button secondary" to="/scenario/new"><ArrowLeft /> Return to builder</Link></div></div> }
function CopyResultLink() {
  const [status, setStatus] = useState<'idle' | 'copied' | 'manual'>('idle'); const url = window.location.href
  const copy = async () => { try { if (!navigator.clipboard?.writeText) throw new Error(); await navigator.clipboard.writeText(url); setStatus('copied'); setTimeout(() => setStatus('idle'), 2200) } catch { setStatus('manual') } }
  return <div className="copy-link"><button className="button secondary" type="button" onClick={() => void copy()}>{status === 'copied' ? <ClipboardCheck /> : <Clipboard />} {status === 'copied' ? 'Link copied' : 'Copy result link'}</button><span className="sr-only" aria-live="polite">{status === 'copied' ? 'Result link copied.' : status === 'manual' ? 'Automatic copying unavailable. Select the URL shown.' : ''}</span>{status === 'manual' && <input aria-label="Result URL for manual copying" value={url} readOnly onFocus={(event) => event.currentTarget.select()} />}</div>
}
