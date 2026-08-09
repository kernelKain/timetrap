import { useEffect, useMemo, useRef, useState } from 'react'
import { ArrowRight, Plus, RotateCcw, ShieldAlert } from 'lucide-react'
import { useLocation, useNavigate, useSearchParams } from 'react-router'
import { InvariantCard } from '../components/scenario/InvariantCard'
import { ObjectCard } from '../components/scenario/ObjectCard'
import { ValidationSummary } from '../components/scenario/ValidationSummary'
import { TimelinePreview } from '../components/timeline/TimelinePreview'
import { FieldError } from '../components/ui/FieldError'
import { loadTemplate } from '../data/templates'
import { blankScenario, cloneScenario, createInvariant, createObject, parseNumber } from '../lib/scenario'
import { validateScenario } from '../lib/validation'
import type { ScenarioDraft } from '../types/forms'

export function ScenarioPage() {
  const [params] = useSearchParams(); const templateId = params.get('template'); const navigate = useNavigate(); const location = useLocation()
  const initialTemplate = templateId ? loadTemplate(templateId) : null
  const [scenario, setScenario] = useState<ScenarioDraft>(() => initialTemplate ?? blankScenario())
  const [touched, setTouched] = useState<Set<string>>(new Set()); const [attempted, setAttempted] = useState(false); const [unknownTemplate, setUnknownTemplate] = useState(Boolean(templateId && !initialTemplate)); const lastLocation = useRef(location.search)
  useEffect(() => {
    if (lastLocation.current === location.search) return
    lastLocation.current = location.search; const nextId = new URLSearchParams(location.search).get('template'); const next = nextId ? loadTemplate(nextId) : blankScenario()
    setUnknownTemplate(Boolean(nextId && !next)); setScenario(next ?? blankScenario()); setTouched(new Set()); setAttempted(false)
  }, [location.search])
  const errors = useMemo(() => validateScenario(scenario), [scenario])
  const visibleErrors = attempted ? errors : errors.filter((error) => touched.has(error.field))
  const touch = (field: string) => setTouched((current) => new Set(current).add(field))
  const reset = () => { const clean = templateId ? loadTemplate(templateId) : blankScenario(); setScenario(clean ?? blankScenario()); setTouched(new Set()); setAttempted(false) }
  const preview = () => {
    const currentErrors = validateScenario(scenario)
    setAttempted(true)
    if (currentErrors.length) {
      requestAnimationFrame(() => document.querySelector('.validation-summary')?.scrollIntoView({ behavior: 'smooth', block: 'center' }))
      return
    }
    navigate('/result/preview', { state: { scenario: cloneScenario(scenario) } })
  }
  const removeObject = (clientId: string) => setScenario((current) => ({ ...current, objects: current.objects.filter((object) => object.clientId !== clientId), invariants: current.invariants.map((invariant) => ({ ...invariant, sourceObjectId: invariant.sourceObjectId === clientId ? '' : invariant.sourceObjectId, dependentObjectId: invariant.dependentObjectId === clientId ? '' : invariant.dependentObjectId, targetObjectId: invariant.targetObjectId === clientId ? '' : invariant.targetObjectId })) }))
  return <div className="page-container builder-page">
    <header className="page-heading"><div><div className="kicker"><ShieldAlert aria-hidden="true" /> Scenario builder</div><h1>Model the lifetime of trust.</h1><p>Build a bounded design model locally. Client checks help with input quality; the backend remains authoritative.</p></div>{templateId && <button className="button secondary" type="button" onClick={reset}><RotateCcw aria-hidden="true" /> Reset template</button>}</header>
    {unknownTemplate && <div className="notice warning" role="status">Unknown template ID. A blank scenario has been opened instead.</div>}
    {attempted && <ValidationSummary errors={errors} />}
    <div className="builder-grid"><div className="builder-main">
      <section className="panel scenario-details"><div className="section-heading"><div><p className="eyebrow">Step 1</p><h2>Scenario information</h2></div></div><div className="form-grid two">
        <label><span>Scenario name</span><input value={scenario.name} maxLength={180} aria-invalid={visibleErrors.some((e) => e.field === 'name')} aria-describedby={visibleErrors.some((e) => e.field === 'name') ? 'name-error' : undefined} onBlur={() => touch('name')} onChange={(e) => setScenario({ ...scenario, name: e.target.value })} placeholder="e.g. Subscription cancellation leak" /><FieldError id="name-error" errors={visibleErrors.filter((e) => e.field === 'name')} /></label>
        <label><span>Horizon minutes</span><input type="number" min="1" max="10080" step="1" value={scenario.horizonMinutes} aria-invalid={visibleErrors.some((e) => e.field === 'horizonMinutes')} aria-describedby={visibleErrors.some((e) => e.field === 'horizonMinutes') ? 'horizon-error' : undefined} onBlur={() => touch('horizonMinutes')} onChange={(e) => setScenario({ ...scenario, horizonMinutes: parseNumber(e.target.value) })} /><FieldError id="horizon-error" errors={visibleErrors.filter((e) => e.field === 'horizonMinutes')} /></label>
      </div><label><span>Description <small>Optional</small></span><textarea value={scenario.description} maxLength={1200} rows={3} aria-invalid={visibleErrors.some((e) => e.field === 'description')} aria-describedby={visibleErrors.some((e) => e.field === 'description') ? 'description-error' : undefined} onBlur={() => touch('description')} onChange={(e) => setScenario({ ...scenario, description: e.target.value })} /><div className="field-meta"><FieldError id="description-error" errors={visibleErrors.filter((e) => e.field === 'description')} /><span>{scenario.description.length}/1,000</span></div></label></section>
      <BuilderSection step="2" title="Timed objects" description="Each object has its own stable client ID and state-changing events." action={<button className="button secondary small" type="button" disabled={scenario.objects.length >= 25} onClick={() => setScenario({ ...scenario, objects: [...scenario.objects, createObject()] })}><Plus /> Add object</button>}>
        {scenario.objects.length === 0 ? <div className="empty-state large"><h3>No objects yet</h3><p>Add a subscription, cache, session, token, invitation, or credential.</p></div> : scenario.objects.map((object, index) => <ObjectCard key={object.clientId} object={object} index={index} errors={visibleErrors} onTouch={touch} onChange={(value) => setScenario({ ...scenario, objects: scenario.objects.map((item) => item.clientId === object.clientId ? value : item) })} onRemove={() => removeObject(object.clientId)} />)}
        <FieldError id="objects-error" errors={visibleErrors.filter((e) => e.field === 'objects')} />
      </BuilderSection>
      <BuilderSection step="3" title="Invariants" description="Choose one of the three frozen policy templates." action={<button className="button secondary small" type="button" disabled={scenario.invariants.length >= 10} onClick={() => setScenario({ ...scenario, invariants: [...scenario.invariants, createInvariant()] })}><Plus /> Add invariant</button>}>
        {scenario.invariants.length === 0 ? <div className="empty-state large"><h3>No invariant configured</h3><p>Add a rule before viewing the result preview.</p></div> : scenario.invariants.map((invariant, index) => <InvariantCard key={invariant.formId} invariant={invariant} index={index} objects={scenario.objects} errors={visibleErrors} onTouch={touch} onChange={(value) => setScenario({ ...scenario, invariants: scenario.invariants.map((item) => item.formId === invariant.formId ? value : item) })} onRemove={() => setScenario({ ...scenario, invariants: scenario.invariants.filter((item) => item.formId !== invariant.formId) })} />)}
        <FieldError id="invariants-error" errors={visibleErrors.filter((e) => e.field === 'invariants')} />
      </BuilderSection>
    </div><aside className="builder-aside"><TimelinePreview scenario={scenario} compact /></aside></div>
    <div className="sticky-actions"><div><strong>{errors.length === 0 ? 'Ready for the static preview' : `${errors.length} ${errors.length === 1 ? 'issue' : 'issues'} to resolve`}</strong><span>{errors.length ? 'Preview will show every field that needs attention.' : 'No scenario or analysis API will be called.'}</span></div><button className="button primary" type="button" aria-disabled={errors.length > 0} onClick={preview}>View result preview <ArrowRight aria-hidden="true" /></button></div>
  </div>
}
function BuilderSection({ step, title, description, action, children }: { step: string; title: string; description: string; action: React.ReactNode; children: React.ReactNode }) { return <section className="builder-section"><div className="section-heading"><div><p className="eyebrow">Step {step}</p><h2>{title}</h2><p>{description}</p></div>{action}</div><div className="card-stack">{children}</div></section> }
