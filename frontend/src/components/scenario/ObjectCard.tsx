import { Plus, Trash2 } from 'lucide-react'
import type { TimedObjectDraft, ValidationIssue } from '../../types/forms'
import { createEvent } from '../../lib/scenario'
import { FieldError } from '../ui/FieldError'
import { EventRow } from './EventRow'

export function ObjectCard({ object, index, errors, onChange, onRemove, onTouch }: {
  object: TimedObjectDraft; index: number; errors: ValidationIssue[]; onChange: (value: TimedObjectDraft) => void; onRemove: () => void; onTouch: (field: string) => void
}) {
  const prefix = `objects[${index}]`; const nameErrors = errors.filter((e) => e.field === `${prefix}.name`); const eventErrors = errors.filter((e) => e.field === `${prefix}.events`)
  return <article className="panel object-card">
    <div className="card-heading"><div><p className="eyebrow">Object {index + 1}</p><h3>{object.name || 'Untitled object'}</h3></div><button className="icon-button danger" type="button" onClick={onRemove} aria-label={`Remove ${object.name || `object ${index + 1}`}`}><Trash2 aria-hidden="true" /></button></div>
    <div className="form-grid two">
      <label><span>Object name</span><input value={object.name} maxLength={80} aria-invalid={nameErrors.length > 0} aria-describedby={nameErrors.length ? `${object.clientId}-name-error` : undefined} onBlur={() => onTouch(`${prefix}.name`)} onChange={(e) => { onTouch(`${prefix}.name`); onChange({ ...object, name: e.target.value }) }} /><FieldError id={`${object.clientId}-name-error`} errors={nameErrors} /></label>
      <label><span>Object kind</span><select value={object.kind} onChange={(e) => onChange({ ...object, kind: e.target.value as TimedObjectDraft['kind'] })}>
        <option value="subscription">Subscription</option><option value="cached_entitlement">Cached entitlement</option><option value="session">Session</option><option value="token">Token</option><option value="invitation">Invitation</option><option value="credential">Credential</option>
      </select></label>
    </div>
    <div className="subsection-heading"><div><h4>Events</h4><p>One state change per minute.</p></div><button className="button secondary small" type="button" disabled={object.events.length >= 50} onClick={() => onChange({ ...object, events: [...object.events, createEvent()] })}><Plus aria-hidden="true" /> Add event</button></div>
    {object.events.length === 0 ? <div className="empty-state">No events yet. Add an issue, refresh, revoke, or expiry.</div> : <div className="event-list">{object.events.map((event, eventIndex) => <EventRow key={event.formId} event={event} path={`${prefix}.events[${eventIndex}]`} errors={errors} onTouch={onTouch} onChange={(value) => onChange({ ...object, events: object.events.map((item) => item.formId === event.formId ? value : item) })} onRemove={() => onChange({ ...object, events: object.events.filter((item) => item.formId !== event.formId) })} />)}</div>}
    <FieldError id={`${object.clientId}-events-error`} errors={eventErrors} />
  </article>
}
