import { useCallback, useEffect, useRef, useState } from 'react'
import { ArrowLeft, Clipboard, ClipboardCheck } from 'lucide-react'
import { Link, useNavigate, useParams } from 'react-router'
import { AnalysisTimeline } from '../components/timeline/AnalysisTimeline'
import { EvidenceCard } from '../components/result/EvidenceCard'
import { RemediationCard } from '../components/result/RemediationCard'
import { ResultHero } from '../components/result/ResultHero'
import { ResultMetadata } from '../components/result/ResultMetadata'
import { ResultMetrics } from '../components/result/ResultMetrics'
import { ErrorState } from '../components/ui/ErrorState'
import { ResultSkeleton } from '../components/ui/ResultSkeleton'
import { Toast } from '../components/ui/Toast'
import { ApiError, createAnalysis, createScenario, getAnalysis } from '../lib/api'
import { applyRemediation, correctedScenarioCopy } from '../lib/remediation'
import { fromScenarioRequest } from '../lib/scenario'
import type { AnalysisResponse, ScenarioRequest } from '../types/api'

type LoadState = { kind: 'loading' } | { kind: 'success'; response: AnalysisResponse } | { kind: 'not-found'; requestId?: string } | { kind: 'error'; message: string; requestId?: string }
type FixStage = 'idle' | 'applying-fix' | 'saving-corrected-scenario' | 'rerunning-analysis' | 'opening-corrected-result' | 'failed'
interface FixState { stage: FixStage; corrected?: ScenarioRequest; scenarioId?: string; analysisId?: string; error?: string; requestId?: string }
const uuidPattern = /^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i

export function ResultPage() {
  const { analysisId } = useParams(); const navigate = useNavigate(); const validId = Boolean(analysisId && uuidPattern.test(analysisId))
  const [load, setLoad] = useState<LoadState>(validId ? { kind: 'loading' } : { kind: 'error', message: 'This result link does not contain a valid analysis UUID.' }); const [attempt, setAttempt] = useState(0)
  const [fix, setFix] = useState<FixState>({ stage: 'idle' }); const fixLock = useRef(false); const fixController = useRef<AbortController | null>(null); const mounted = useRef(true)
  useEffect(() => { mounted.current = true; return () => { mounted.current = false; fixController.current?.abort() } }, [])
  useEffect(() => {
    if (!analysisId || !validId) return
    const controller = new AbortController(); let current = true
    getAnalysis(analysisId, { signal: controller.signal }).then((response) => { if (current) { setLoad({ kind: 'success', response }); setFix({ stage: 'idle' }) } }).catch((error: unknown) => {
      if (!current || (error instanceof DOMException && error.name === 'AbortError')) return
      if (error instanceof ApiError && error.status === 404) setLoad({ kind: 'not-found', requestId: error.requestId })
      else setLoad({ kind: 'error', message: error instanceof ApiError ? error.message : 'The result could not be loaded.', requestId: error instanceof ApiError ? error.requestId : undefined })
    })
    return () => { current = false; controller.abort() }
  }, [analysisId, attempt, validId])
  const retryLoad = useCallback(() => { setLoad({ kind: 'loading' }); setAttempt((value) => value + 1) }, [])
  if (load.kind === 'loading') return <ResultSkeleton />
  if (load.kind === 'not-found') return <ErrorState title="Analysis not found" message="No persisted analysis exists for this UUID." requestId={load.requestId} />
  if (load.kind === 'error') return <ErrorState title="Unable to open this result" message={load.message} requestId={load.requestId} retry={validId ? retryLoad : undefined} />
  if (load.response.analysis.id !== analysisId) return <ResultSkeleton />
  const { analysis } = load.response; const scenario = analysis.scenarioSnapshot
  if (!scenario) return <ErrorState title="Incomplete persisted result" message="This older analysis does not include the scenario snapshot required to display the report." requestId={load.response.requestId} retry={retryLoad} />

  const applyAndRerun = async () => {
    const remediation = analysis.result.earliestViolation?.remediation
    if (!remediation || fixLock.current) return
    fixLock.current = true; const controller = new AbortController(); fixController.current = controller
    let corrected = fix.corrected; let scenarioId = fix.scenarioId; let correctedAnalysisId = fix.analysisId
    try {
      if (correctedAnalysisId) { setFix((current) => ({ ...current, stage: 'opening-corrected-result' })); navigate(`/results/${correctedAnalysisId}`); return }
      if (!corrected) {
        setFix({ stage: 'applying-fix' })
        const applied = applyRemediation(scenario, remediation)
        if (!applied.ok) throw new Error(applied.reason)
        corrected = applied.scenario
      }
      if (!scenarioId) {
        setFix({ stage: 'saving-corrected-scenario', corrected })
        const saved = await createScenario(corrected, { signal: controller.signal })
        scenarioId = saved.scenario.id
        if (mounted.current) setFix({ stage: 'rerunning-analysis', corrected, scenarioId })
      } else setFix((current) => ({ ...current, stage: 'rerunning-analysis', error: undefined, requestId: undefined }))
      const rerun = await createAnalysis(scenarioId, { signal: controller.signal })
      correctedAnalysisId = rerun.analysis.id
      if (mounted.current) setFix({ stage: 'opening-corrected-result', corrected, scenarioId, analysisId: correctedAnalysisId })
      navigate(`/results/${correctedAnalysisId}`)
    } catch (error) {
      if (!mounted.current) return
      const apiError = error instanceof ApiError ? error : undefined
      setFix({ stage: 'failed', corrected, scenarioId, analysisId: correctedAnalysisId, error: apiError?.message ?? (error instanceof Error ? error.message : 'The correction could not be completed.'), requestId: apiError?.requestId })
    } finally { fixLock.current = false; fixController.current = null }
  }
  const editManually = () => navigate('/scenario/new', { state: { scenarioDraft: fromScenarioRequest(correctedScenarioCopy(scenario)) } })
  const finding = analysis.result.earliestViolation
  return <div className="page-container polished-result"><header className="result-toolbar"><div><p className="eyebrow">Persisted analysis</p><p className="result-name">{scenario.name}</p></div><div className="result-actions"><CopyResultLink /><Link className="button secondary" to="/scenario/new"><ArrowLeft aria-hidden="true" /> Builder</Link></div></header><ResultHero result={analysis.result} scenario={scenario} /><ResultMetrics result={analysis.result} scenario={scenario} /><AnalysisTimeline scenario={scenario} lanes={analysis.result.timeline} violation={finding} /><EvidenceCard analysis={analysis} scenario={scenario} />{finding?.remediation && <RemediationCard remediation={finding.remediation} scenario={scenario} stage={fix.stage} error={fix.error && `${fix.error}${fix.requestId ? ` Request ID: ${fix.requestId}` : ''}`} onApply={() => void applyAndRerun()} onEdit={editManually} />}<ResultMetadata analysis={analysis} /></div>
}

function CopyResultLink() {
  const [status, setStatus] = useState<'idle' | 'copied' | 'failed'>('idle'); const timer = useRef<ReturnType<typeof setTimeout> | null>(null); const url = window.location.href
  useEffect(() => () => { if (timer.current) clearTimeout(timer.current) }, [])
  const copy = async () => { if (timer.current) clearTimeout(timer.current); try { if (!navigator.clipboard?.writeText) throw new Error(); await navigator.clipboard.writeText(url); setStatus('copied'); timer.current = setTimeout(() => setStatus('idle'), 2400) } catch { setStatus('failed') } }
  return <div className="copy-link"><button className="button secondary" type="button" onClick={() => void copy()}>{status === 'copied' ? <ClipboardCheck aria-hidden="true" /> : <Clipboard aria-hidden="true" />}{status === 'copied' ? 'Copied' : 'Copy link'}</button>{status === 'copied' && <Toast kind="success" message="Copied analysis link" />}{status === 'failed' && <><Toast kind="error" message="Could not copy link" /><label className="manual-link"><span>Select the result URL</span><input value={url} readOnly onFocus={(event) => event.currentTarget.select()} /></label></>}</div>
}
