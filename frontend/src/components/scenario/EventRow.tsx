import { Trash2 } from 'lucide-react'
import type { EventDraft, ValidationIssue } from '../../types/forms'
import { parseNumber } from '../../lib/scenario'
import { FieldError } from '../ui/FieldError'

export function EventRow({ event, path, errors, onChange, onRemove, onTouch }: {
  event: EventDraft; path: string; errors: ValidationIssue[]; onChange: (value: EventDraft) => void; onRemove: () => void; onTouch: (field: string) => void
}) {
  const minuteErrors = errors.filter((error) => error.field === `${path}.atMinute`); const errorId = `${event.formId}-minute-error`
  return <div className="event-row">
    <label><span>Event type</span><select value={event.type} onChange={(e) => onChange({ ...event, type: e.target.value as EventDraft['type'] })}>
      <option value="issue">Issue</option><option value="refresh">Refresh</option><option value="revoke">Revoke</option><option value="expire">Expire</option>
    </select></label>
    <label><span>At minute</span><input type="number" min="0" step="1" value={event.atMinute} aria-invalid={minuteErrors.length > 0} aria-describedby={minuteErrors.length ? errorId : undefined}
      onBlur={() => onTouch(`${path}.atMinute`)} onChange={(e) => onChange({ ...event, atMinute: parseNumber(e.target.value) })} /><FieldError id={errorId} errors={minuteErrors} /></label>
    <button className="icon-button" type="button" onClick={onRemove} aria-label={`Remove ${event.type} event`}><Trash2 aria-hidden="true" /></button>
  </div>
}
