import type { EventDraft, ScenarioDraft } from '../types/forms'
export const positionForMinute = (minute: number, horizon: number | '') => Math.min(100, Math.max(0, horizon === '' || horizon <= 0 ? 0 : (minute / horizon) * 100))
export const sortedEvents = (events: EventDraft[]) => [...events].sort((a, b) => (a.atMinute === '' ? Infinity : a.atMinute) - (b.atMinute === '' ? Infinity : b.atMinute))
export function validitySegments(events: EventDraft[], horizon: ScenarioDraft['horizonMinutes']) {
  if (horizon === '') return []
  const segments: Array<{ start: number; end: number }> = []; let start: number | null = null
  sortedEvents(events).forEach((event) => {
    if (event.atMinute === '') return
    if ((event.type === 'issue' || event.type === 'refresh') && start === null) start = event.atMinute
    if ((event.type === 'revoke' || event.type === 'expire') && start !== null) { segments.push({ start, end: event.atMinute }); start = null }
  })
  if (start !== null) segments.push({ start, end: horizon })
  return segments
}
