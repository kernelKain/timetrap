import type { EventDraft, ScenarioDraft } from '../types/forms'
export const minuteToPercent = (minute: number, horizon: number) => horizon <= 0 ? 0 : Math.min(100, Math.max(0, (minute / horizon) * 100))
export const intervalToPosition = (start: number, end: number, horizon: number) => {
  const left = minuteToPercent(start, horizon); const right = minuteToPercent(end, horizon)
  const round = (value: number) => Math.round(value * 1_000_000) / 1_000_000
  return { left: `${round(left)}%`, width: `${round(Math.max(0, right - left))}%` }
}
export const formatDuration = (minutes: number) => `${minutes} ${minutes === 1 ? 'minute' : 'minutes'}`
export const formatInterval = (start: number, end: number, ongoing = false) => ongoing ? `[${start},${end}) — continues to simulation horizon` : `[${start},${end})`
export const positionForMinute = (minute: number, horizon: number | '') => minuteToPercent(minute, horizon === '' ? 0 : horizon)
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
