import { ShieldAlert, Trash2 } from 'lucide-react'
import { TechnicalBadge } from '../ui/TechnicalBadge'
import type { InvariantDraft, ScenarioDraft, ValidationIssue } from '../../types/forms'
import { parseNumber } from '../../lib/scenario'
import { FieldError } from '../ui/FieldError'

const labels = { revoked_access_grace: 'Revoked access grace', dependent_not_outlive_source: 'Dependent must not outlive source', max_validity: 'Maximum validity' }
export function InvariantCard({ invariant, index, objects, errors, onChange, onRemove, onTouch }: {
  invariant: InvariantDraft; index: number; objects: ScenarioDraft['objects']; errors: ValidationIssue[]; onChange: (value: InvariantDraft) => void; onRemove: () => void; onTouch: (field: string) => void
}) {
  const prefix = `invariants[${index}]`
  const field = (name: string) => errors.filter((error) => error.field === `${prefix}.${name}`)
  const select = (name: 'sourceObjectId' | 'dependentObjectId' | 'targetObjectId', label: string) => {
    const fieldErrors = field(name); const id = `${invariant.formId}-${name}`
    return <label><span>{label}</span><select value={invariant[name]} aria-invalid={fieldErrors.length > 0} aria-describedby={fieldErrors.length ? `${id}-error` : undefined} onBlur={() => onTouch(`${prefix}.${name}`)} onChange={(e) => { onTouch(`${prefix}.${name}`); onChange({ ...invariant, [name]: e.target.value }) }}><option value="">Select an object</option>{objects.map((object) => <option key={object.clientId} value={object.clientId}>{object.name || 'Untitled object'}</option>)}</select><FieldError id={`${id}-error`} errors={fieldErrors} /></label>
  }
  return <article className="panel invariant-card">
    <div className="card-heading invariant-heading"><span className="object-icon rule"><ShieldAlert /></span><div><p className="eyebrow">Policy rule {String(index + 1).padStart(2, '0')}</p><h3>{labels[invariant.type]}</h3><TechnicalBadge>{invariant.type}</TechnicalBadge></div><button className="icon-button danger" type="button" onClick={onRemove} aria-label={`Remove invariant ${index + 1}`}><Trash2 aria-hidden="true" /></button></div>
    <label><span>Invariant template</span><select value={invariant.type} onChange={(e) => onChange({ ...invariant, type: e.target.value as InvariantDraft['type'] })}><option value="revoked_access_grace">Revoked access must end within grace</option><option value="dependent_not_outlive_source">Dependent must not outlive source</option><option value="max_validity">Object maximum validity</option></select></label>
    <p className="rule-sentence">{invariant.type === 'revoked_access_grace' ? <>When <strong>{objects.find((item) => item.clientId === invariant.sourceObjectId)?.name || 'source'}</strong> is revoked, <strong>{objects.find((item) => item.clientId === invariant.dependentObjectId)?.name || 'dependent'}</strong> must end within <code>{invariant.graceMinutes === '' ? '—' : invariant.graceMinutes} min</code>.</> : invariant.type === 'dependent_not_outlive_source' ? <><strong>{objects.find((item) => item.clientId === invariant.dependentObjectId)?.name || 'Dependent'}</strong> must not remain valid after <strong>{objects.find((item) => item.clientId === invariant.sourceObjectId)?.name || 'source'}</strong>.</> : <><strong>{objects.find((item) => item.clientId === invariant.targetObjectId)?.name || 'Target'}</strong> may remain valid for at most <code>{invariant.maximumMinutes === '' ? '—' : invariant.maximumMinutes} min</code>.</>}</p>
    <div className="form-grid three">
      {invariant.type === 'max_validity' ? <>
        {select('targetObjectId', 'Target object')}
        <NumberField label="Maximum minutes" value={invariant.maximumMinutes} errors={field('maximumMinutes')} id={`${invariant.formId}-maximum`} onTouch={() => onTouch(`${prefix}.maximumMinutes`)} onChange={(maximumMinutes) => onChange({ ...invariant, maximumMinutes })} />
      </> : <>
        {select('sourceObjectId', 'Source object')}{select('dependentObjectId', 'Dependent object')}
        {invariant.type === 'revoked_access_grace' && <NumberField label="Grace minutes" value={invariant.graceMinutes} errors={field('graceMinutes')} id={`${invariant.formId}-grace`} min="0" onTouch={() => onTouch(`${prefix}.graceMinutes`)} onChange={(graceMinutes) => onChange({ ...invariant, graceMinutes })} />}
      </>}
    </div>
  </article>
}

function NumberField({ label, value, errors, id, min = '1', onChange, onTouch }: { label: string; value: number | ''; errors: ValidationIssue[]; id: string; min?: string; onChange: (value: number | '') => void; onTouch: () => void }) {
  return <label><span>{label}</span><input type="number" min={min} step="1" value={value} aria-invalid={errors.length > 0} aria-describedby={errors.length ? `${id}-error` : undefined} onBlur={onTouch} onChange={(e) => { onTouch(); onChange(parseNumber(e.target.value)) }} /><FieldError id={`${id}-error`} errors={errors} /></label>
}
