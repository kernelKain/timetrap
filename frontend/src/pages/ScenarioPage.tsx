import { useEffect, useMemo, useRef, useState } from 'react'
import { ArrowRight, Clock3, LoaderCircle, Plus, RotateCcw, ShieldAlert, TriangleAlert } from 'lucide-react'
import { useLocation, useNavigate, useSearchParams } from 'react-router'
import { InvariantCard } from '../components/scenario/InvariantCard'
import { ObjectCard } from '../components/scenario/ObjectCard'
import { ScenarioStepNav } from '../components/scenario/ScenarioStepNav'
import { ValidationSummary } from '../components/scenario/ValidationSummary'
import { TimelinePreview } from '../components/timeline/TimelinePreview'
import { FieldError } from '../components/ui/FieldError'
import { loadTemplate } from '../data/templates'
import { ApiError, createAnalysis, createScenario, updateScenario } from '../lib/api'
import { blankScenario, cloneScenario, createInvariant, createObject, parseNumber, scenarioFingerprint, toScenarioRequest } from '../lib/scenario'
import { validateScenario } from '../lib/validation'
import type { ScenarioDraft } from '../types/forms'

type SubmissionStage = 'idle' | 'saving-scenario' | 'running-analysis' | 'opening-result' | 'failed'
interface SubmissionState { stage: SubmissionStage; scenarioId?: string; fingerprint?: string; analysisId?: string; message?: string; requestId?: string }

export function ScenarioPage() {
  const [params] = useSearchParams(); const templateId = params.get('template'); const navigate = useNavigate(); const location = useLocation()
  const initialTemplate = templateId ? loadTemplate(templateId) : null
  const routedDraft = (location.state as { scenarioDraft?: ScenarioDraft } | null)?.scenarioDraft
  const [scenario, setScenario] = useState<ScenarioDraft>(() => routedDraft ? cloneScenario(routedDraft) : initialTemplate ?? blankScenario())
  const [touched, setTouched] = useState<Set<string>>(new Set()); const [attempted, setAttempted] = useState(false); const [unknownTemplate, setUnknownTemplate] = useState(Boolean(templateId && !initialTemplate)); const lastLocation = useRef(location.search)
  const [serverErrors, setServerErrors] = useState<Array<{ field: string; message: string }>>([])
  const [submission, setSubmission] = useState<SubmissionState>({ stage: 'idle' }); const inFlight = useRef(false); const activeController = useRef<AbortController | null>(null); const mounted = useRef(true)
  useEffect(() => { mounted.current = true; return () => { mounted.current = false; activeController.current?.abort() } }, [])
  useEffect(() => {
    if (lastLocation.current === location.search) return
    lastLocation.current = location.search; const nextId = new URLSearchParams(location.search).get('template'); const next = nextId ? loadTemplate(nextId) : blankScenario()
    setUnknownTemplate(Boolean(nextId && !next)); setScenario(next ?? blankScenario()); setTouched(new Set()); setAttempted(false); setServerErrors([]); setSubmission({ stage: 'idle' })
  }, [location.search])
  const errors = useMemo(() => validateScenario(scenario), [scenario])
  const clientVisible = attempted ? errors : errors.filter((error) => touched.has(error.field))
  const visibleErrors = [...clientVisible.filter((error) => !serverErrors.some((serverError) => serverError.field === error.field)), ...serverErrors]
  const touch = (field: string) => { setTouched((current) => new Set(current).add(field)); setServerErrors((current) => current.filter((error) => error.field !== field)) }
  const reset = () => { const clean = templateId ? loadTemplate(templateId) : blankScenario(); setScenario(clean ?? blankScenario()); setTouched(new Set()); setAttempted(false); setServerErrors([]); setSubmission({ stage: 'idle' }) }
  const submit = async () => {
    if (inFlight.current) return
    const currentErrors = validateScenario(scenario)
    setAttempted(true)
    if (currentErrors.length) {
      requestAnimationFrame(() => document.querySelector('.validation-summary')?.scrollIntoView({ behavior: 'smooth', block: 'center' }))
      return
    }
    const request = toScenarioRequest(scenario)
    if (!request) return
    const fingerprint = scenarioFingerprint(request)
    if (submission.analysisId) { setSubmission((current) => ({ ...current, stage: 'opening-result' })); navigate(`/results/${submission.analysisId}`); return }
    inFlight.current = true; const controller = new AbortController(); activeController.current = controller; setServerErrors([])
    let confirmedScenarioId = submission.scenarioId; let confirmedFingerprint = submission.fingerprint
    try {
      if (!confirmedScenarioId) {
        setSubmission({ stage: 'saving-scenario' })
        const saved = await createScenario(request, { signal: controller.signal })
        confirmedScenarioId = saved.scenario.id; confirmedFingerprint = fingerprint
        if (mounted.current) setSubmission({ stage: 'running-analysis', scenarioId: confirmedScenarioId, fingerprint: confirmedFingerprint })
      } else if (confirmedFingerprint !== fingerprint) {
        setSubmission((current) => ({ ...current, stage: 'saving-scenario', message: undefined, requestId: undefined }))
        await updateScenario(confirmedScenarioId, request, { signal: controller.signal })
        confirmedFingerprint = fingerprint
        if (mounted.current) setSubmission({ stage: 'running-analysis', scenarioId: confirmedScenarioId, fingerprint: confirmedFingerprint })
      } else setSubmission((current) => ({ ...current, stage: 'running-analysis', message: undefined, requestId: undefined }))
      const analyzed = await createAnalysis(confirmedScenarioId, { signal: controller.signal })
      const analysisId = analyzed.analysis.id
      if (mounted.current) setSubmission({ stage: 'opening-result', scenarioId: confirmedScenarioId, fingerprint: confirmedFingerprint, analysisId })
      navigate(`/results/${analysisId}`)
    } catch (error) {
      if (!mounted.current) return
      const apiError = error instanceof ApiError ? error : undefined
      setServerErrors(apiError?.fields ?? [])
      const aborted = error instanceof DOMException && error.name === 'AbortError'
      setSubmission({ stage: 'failed', scenarioId: confirmedScenarioId, fingerprint: confirmedFingerprint, message: aborted ? 'The request was cancelled. Your draft is unchanged.' : apiError?.message ?? 'The workflow failed. Your draft is unchanged.', requestId: apiError?.requestId })
    } finally { inFlight.current = false; activeController.current = null }
  }
  const active = submission.stage === 'saving-scenario' || submission.stage === 'running-analysis' || submission.stage === 'opening-result'
  const currentRequest = toScenarioRequest(scenario); const currentFingerprint = currentRequest ? scenarioFingerprint(currentRequest) : undefined
  const buttonLabel = submission.stage === 'saving-scenario' ? 'Saving scenario…' : submission.stage === 'running-analysis' ? 'Running analysis…' : submission.stage === 'opening-result' ? 'Opening result…' : submission.stage === 'failed' && submission.scenarioId ? (submission.fingerprint === currentFingerprint ? 'Retry analysis' : 'Save changes and retry') : submission.stage === 'failed' ? 'Retry save' : 'Save and analyze'
  const removeObject = (clientId: string) => setScenario((current) => ({ ...current, objects: current.objects.filter((object) => object.clientId !== clientId), invariants: current.invariants.map((invariant) => ({ ...invariant, sourceObjectId: invariant.sourceObjectId === clientId ? '' : invariant.sourceObjectId, dependentObjectId: invariant.dependentObjectId === clientId ? '' : invariant.dependentObjectId, targetObjectId: invariant.targetObjectId === clientId ? '' : invariant.targetObjectId })) }))
  return <div className="page-container builder-page">
    <header className="page-heading"><div><div className="kicker"><ShieldAlert aria-hidden="true" /> Scenario builder</div><h1>Model the lifetime of trust.</h1><p>Build a bounded design model locally. Client checks help with input quality; the backend remains authoritative.</p></div>{templateId && <button className="button secondary" type="button" onClick={reset}><RotateCcw aria-hidden="true" /> Reset template</button>}</header>
    <ScenarioStepNav objectCount={scenario.objects.length} invariantCount={scenario.invariants.length} errorCount={errors.length} />
    {unknownTemplate && <div className="notice warning" role="status">Unknown template ID. A blank scenario has been opened instead.</div>}
    {submission.stage === 'failed' && <div className="submission-error" role="alert"><TriangleAlert /><div><strong>{submission.scenarioId ? 'Scenario saved; analysis was not completed.' : 'Scenario was not saved.'}</strong><p>{submission.message}</p>{submission.requestId && <small>Request ID: <code>{submission.requestId}</code></small>}</div></div>}
    {attempted && <ValidationSummary errors={visibleErrors} />}
    <fieldset className="builder-fieldset" disabled={active}><div className="builder-grid"><div className="builder-main">
      <section id="scenario-config" className="panel scenario-details"><div className="section-heading"><div><p className="eyebrow">Configuration</p><h2>Scenario</h2></div><span className="config-hint"><Clock3 aria-hidden="true" /> bounded simulation</span></div><div className="form-grid two">
        <label><span>Scenario name</span><input value={scenario.name} maxLength={120} aria-invalid={visibleErrors.some((e) => e.field === 'name')} aria-describedby={visibleErrors.some((e) => e.field === 'name') ? 'name-error' : undefined} onBlur={() => touch('name')} onChange={(e) => { touch('name'); setScenario({ ...scenario, name: e.target.value }) }} placeholder="e.g. Subscription cancellation leak" /><FieldError id="name-error" errors={visibleErrors.filter((e) => e.field === 'name')} /></label>
        <label><span>Simulation horizon <small>minutes</small></span><div className="input-with-suffix"><input type="number" min="1" max="10080" step="1" value={scenario.horizonMinutes} aria-invalid={visibleErrors.some((e) => e.field === 'horizonMinutes')} aria-describedby={visibleErrors.some((e) => e.field === 'horizonMinutes') ? 'horizon-error' : undefined} onBlur={() => touch('horizonMinutes')} onChange={(e) => { touch('horizonMinutes'); setScenario({ ...scenario, horizonMinutes: parseNumber(e.target.value) }) }} /><code>min</code></div><FieldError id="horizon-error" errors={visibleErrors.filter((e) => e.field === 'horizonMinutes')} /></label>
      </div><label><span>Description <small>Optional</small></span><textarea value={scenario.description} maxLength={1000} rows={3} aria-invalid={visibleErrors.some((e) => e.field === 'description')} aria-describedby={visibleErrors.some((e) => e.field === 'description') ? 'description-error' : undefined} onBlur={() => touch('description')} onChange={(e) => { touch('description'); setScenario({ ...scenario, description: e.target.value }) }} /><div className="field-meta"><FieldError id="description-error" errors={visibleErrors.filter((e) => e.field === 'description')} /><span>{scenario.description.length}/1,000</span></div></label></section>
      <BuilderSection id="objects" step="02" title="Objects & events" description="Model each authorization object and the events that change its validity." action={<button className="button secondary small" type="button" disabled={scenario.objects.length >= 25} onClick={() => setScenario({ ...scenario, objects: [...scenario.objects, createObject()] })}><Plus /> Add object</button>}>
        {scenario.objects.length === 0 ? <div className="empty-state large"><h3>No objects yet</h3><p>Add a subscription, cache, session, token, invitation, or credential.</p></div> : scenario.objects.map((object, index) => <ObjectCard key={object.clientId} object={object} index={index} errors={visibleErrors} onTouch={touch} onChange={(value) => setScenario({ ...scenario, objects: scenario.objects.map((item) => item.clientId === object.clientId ? value : item) })} onRemove={() => removeObject(object.clientId)} />)}
        <FieldError id="objects-error" errors={visibleErrors.filter((e) => e.field === 'objects')} />
      </BuilderSection>
      <BuilderSection id="invariants" step="03" title="Policy invariants" description="Define the relationship the analyzer must enforce." action={<button className="button secondary small" type="button" disabled={scenario.invariants.length >= 10} onClick={() => setScenario({ ...scenario, invariants: [...scenario.invariants, createInvariant()] })}><Plus /> Add invariant</button>}>
        {scenario.invariants.length === 0 ? <div className="empty-state large"><h3>No invariant configured</h3><p>Add a rule before saving and analyzing the scenario.</p></div> : scenario.invariants.map((invariant, index) => <InvariantCard key={invariant.formId} invariant={invariant} index={index} objects={scenario.objects} errors={visibleErrors} onTouch={touch} onChange={(value) => setScenario({ ...scenario, invariants: scenario.invariants.map((item) => item.formId === invariant.formId ? value : item) })} onRemove={() => setScenario({ ...scenario, invariants: scenario.invariants.filter((item) => item.formId !== invariant.formId) })} />)}
        <FieldError id="invariants-error" errors={visibleErrors.filter((e) => e.field === 'invariants')} />
      </BuilderSection>
    </div><aside className="builder-aside"><TimelinePreview scenario={scenario} compact /></aside></div></fieldset>
    <div id="analyze" className="sticky-actions"><div className={`readiness-indicator ${errors.length ? 'invalid' : 'ready'}`} aria-hidden="true"><span /></div><div><strong>{active ? buttonLabel : errors.length === 0 ? 'Ready to analyze' : `${errors.length} ${errors.length === 1 ? 'issue' : 'issues'} to resolve`}</strong><span>{active ? 'Keep this page open while the persisted workflow completes.' : errors.length ? 'Review the highlighted fields before running the analyzer.' : 'The scenario will be saved before the Go analyzer runs.'}</span></div><button className="button primary" type="button" disabled={active} aria-disabled={active || errors.length > 0} onClick={() => void submit()}>{active && <LoaderCircle className="spin" aria-hidden="true" />}{buttonLabel} {!active && <ArrowRight aria-hidden="true" />}</button></div>
  </div>
}
function BuilderSection({ id, step, title, description, action, children }: { id: string; step: string; title: string; description: string; action: React.ReactNode; children: React.ReactNode }) { return <section id={id} className="builder-section"><div className="section-heading"><div><p className="eyebrow">Step {step}</p><h2>{title}</h2><p>{description}</p></div>{action}</div><div className="card-stack">{children}</div></section> }
