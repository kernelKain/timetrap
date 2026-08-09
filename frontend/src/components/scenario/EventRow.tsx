import { Ban, CirclePlay, RefreshCw, TimerOff, Trash2 } from 'lucide-react'
import type { EventDraft, ValidationIssue } from '../../types/forms'
import { parseNumber } from '../../lib/scenario'
import { FieldError } from '../ui/FieldError'

export function EventRow({ event, path, errors, onChange, onRemove, onTouch }: {
  event: EventDraft; path: string; errors: ValidationIssue[]; onChange: (value: EventDraft) => void; onRemove: () => void; onTouch: (field: string) => void
}) {
  const minuteErrors = errors.filter((error) => error.field === `${path}.atMinute`); const errorId = `${event.formId}-minute-error`
  const icons = { issue: CirclePlay, refresh: RefreshCw, revoke: Ban, expire: TimerOff }; const EventIcon = icons[event.type]
  return <div className={`event-row ${minuteErrors.length ? 'has-error' : ''}`}>
    <span className={`event-type-icon ${event.type}`} aria-hidden="true"><EventIcon /></span>
    <label><span>Event</span><select value={event.type} onChange={(e) => onChange({ ...event, type: e.target.value as EventDraft['type'] })}>
      <option value="issue">Issue</option><option value="refresh">Refresh</option><option value="revoke">Revoke</option><option value="expire">Expire</option>
    </select></label>
    <span className="event-at" aria-hidden="true">at</span>
    <label><span>Minute</span><input type="number" min="0" step="1" value={event.atMinute} aria-invalid={minuteErrors.length > 0} aria-describedby={minuteErrors.length ? errorId : undefined}
      onBlur={() => onTouch(`${path}.atMinute`)} onChange={(e) => { onTouch(`${path}.atMinute`); onChange({ ...event, atMinute: parseNumber(e.target.value) }) }} /><FieldError id={errorId} errors={minuteErrors} /></label>
    <code className="timestamp-token">t={event.atMinute === '' ? '—' : event.atMinute}m</code>
    <button className="icon-button" type="button" onClick={onRemove} aria-label={`Remove ${event.type} event`}><Trash2 aria-hidden="true" /></button>
  </div>
}
