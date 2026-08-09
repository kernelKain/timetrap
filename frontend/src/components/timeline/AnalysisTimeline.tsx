import { CirclePlus, Diamond, OctagonX, RefreshCw, TriangleAlert } from 'lucide-react'
import type { EventRequest, EventType, ScenarioRequest, TimelineLane, Violation } from '../../types/api'
import { formatDuration, formatInterval, intervalToPosition, minuteToPercent } from '../../lib/timeline'

const eventIcon: Record<EventType, React.ReactNode> = { issue: <CirclePlus />, refresh: <RefreshCw />, revoke: <OctagonX />, expire: <Diamond /> }
const eventVerb: Record<EventType, string> = { issue: 'issued', refresh: 'refreshed', revoke: 'revoked', expire: 'expired' }

export function AnalysisTimeline({ scenario, lanes, violation }: { scenario: ScenarioRequest; lanes?: TimelineLane[]; violation?: Violation }) {
  const laneMap = new Map((lanes ?? []).map((lane) => [lane.objectId, lane]))
  const summary = violation ? `Authorization timeline showing a policy violation from minute ${violation.startMinute} to minute ${violation.endMinute}.` : `Authorization timeline with no server-reported violation inside the ${scenario.horizonMinutes}-minute horizon.`
  return <section className="panel analysis-timeline" aria-labelledby="analysis-timeline-heading"><header><div><p className="eyebrow">Visual timeline</p><h2 id="analysis-timeline-heading">Shared-scale authorization timeline</h2><p>{summary}</p></div><div className="event-legend" aria-label="Event legend"><span className="issue">● Issue</span><span className="refresh">↻ Refresh</span><span className="revoke">■ Revoke</span><span className="expire">◆ Expire</span></div></header>{!lanes?.length ? <div className="timeline-fallback">This older stored analysis does not include normalized validity intervals. Its verdict and evidence remain available above.</div> : <div className="analysis-timeline-scroll"><div className="analysis-timeline-grid"><div className="analysis-axis-label" aria-hidden="true" /><div className="analysis-axis"><span>0m</span><span>{Math.round(scenario.horizonMinutes / 2)}m</span><span>{scenario.horizonMinutes}m</span></div>{scenario.objects.map((object) => <TimelineLaneRow key={object.clientId} object={object} lane={laneMap.get(object.clientId)} horizon={scenario.horizonMinutes} violation={violation} />)}</div></div>}</section>
}

function TimelineLaneRow({ object, lane, horizon, violation }: { object: ScenarioRequest['objects'][number]; lane?: TimelineLane; horizon: number; violation?: Violation }) {
  const findingObjectId = violation?.dependentObjectId ?? violation?.targetObjectId ?? violation?.sourceObjectId
  const involved = Boolean(violation && findingObjectId === object.clientId)
  return <><div className="analysis-lane-label" title={object.name}><strong>{object.name}</strong><span>{object.kind.replaceAll('_', ' ')}</span></div><div className="analysis-lane-track" aria-label={`${object.name} timeline`}>
    {(lane?.intervals ?? []).map((interval, index) => <div className="analysis-validity" key={`${interval.startMinute}-${interval.endMinute}-${index}`} style={intervalToPosition(interval.startMinute, interval.endMinute, horizon)}><span className="sr-only">{object.name} valid from minute {interval.startMinute} to minute {interval.endMinute}</span></div>)}
    {involved && violation && <div className="violation-overlay" style={intervalToPosition(violation.startMinute, violation.endMinute, horizon)}><TriangleAlert aria-hidden="true" /><span>Policy violation</span><small>{formatInterval(violation.startMinute, violation.endMinute, violation.ongoing)} · {formatDuration(violation.durationMinutes)}</small></div>}
    {[...object.events].sort((a, b) => a.atMinute - b.atMinute).map((event, index) => <EventMarker key={`${event.type}-${event.atMinute}-${index}`} event={event} objectName={object.name} horizon={horizon} />)}
  </div></>
}

function EventMarker({ event, objectName, horizon }: { event: EventRequest; objectName: string; horizon: number }) { return <div className={`analysis-event ${event.type}`} style={{ left: `${minuteToPercent(event.atMinute, horizon)}%` }} aria-label={`${objectName} ${eventVerb[event.type]} at minute ${event.atMinute}`} title={`${event.type} · minute ${event.atMinute}`}><span className="event-shape">{eventIcon[event.type]}</span><span>{event.type}<small>{event.atMinute}m</small></span></div> }
