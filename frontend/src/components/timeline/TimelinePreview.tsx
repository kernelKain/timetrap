import type { ScenarioDraft } from '../../types/forms'
import { positionForMinute, sortedEvents, validitySegments } from '../../lib/timeline'

export function TimelinePreview({ scenario, compact = false }: { scenario: ScenarioDraft; compact?: boolean }) {
  const horizon = scenario.horizonMinutes
  return <section className={`panel timeline-panel ${compact ? 'compact' : ''}`} aria-labelledby="timeline-heading">
    <div className="section-heading"><div><p className="eyebrow">Visual model</p><h2 id="timeline-heading">Scenario preview</h2></div><span className="status-label">Static preview · no verdict</span></div>
    {!scenario.objects.length ? <div className="empty-state">Add an object to begin the timeline.</div> : <div className="timeline-scroll"><div className="timeline" style={{ minWidth: compact ? 620 : 760 }}>
      <div className="timeline-axis"><span>0m</span><span>{horizon === '' ? '—' : `${Math.round(horizon / 2)}m`}</span><span>{horizon === '' ? 'Horizon' : `${horizon}m`}</span></div>
      {scenario.objects.map((object) => <div className="timeline-lane" key={object.clientId}>
        <div className="lane-label"><strong>{object.name || 'Untitled object'}</strong><span>{object.kind.replaceAll('_', ' ')}</span></div>
        <div className="lane-track">
          {validitySegments(object.events, horizon).map((segment, index) => <div className="validity-bar" key={`${segment.start}-${segment.end}-${index}`} style={{ left: `${positionForMinute(segment.start, horizon)}%`, width: `${Math.max(0, positionForMinute(segment.end, horizon) - positionForMinute(segment.start, horizon))}%` }}><span className="sr-only">Valid from minute {segment.start} to {segment.end}</span></div>)}
          {sortedEvents(object.events).map((event) => event.atMinute === '' ? null : <div className={`event-marker ${event.type}`} key={event.formId} style={{ left: `${positionForMinute(event.atMinute, horizon)}%` }}><span className="marker-dot" /><span className="marker-label">{event.type} {event.atMinute}</span></div>)}
        </div>
      </div>)}
    </div></div>}
    <p className="timeline-note">Bars show derived validity windows for orientation only. The Go analyzer remains authoritative.</p>
  </section>
}
