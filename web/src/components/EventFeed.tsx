import type { EventEnvelope } from '../generated/api'
import { compactJson, eventSummary, formatDateTime } from '../lib/format'
import { EmptyState } from './EmptyState'

interface EventFeedProps {
  events: EventEnvelope[]
  emptyTitle?: string
  emptyDescription?: string
}

export function EventFeed({
  events,
  emptyTitle = 'No activity yet',
  emptyDescription = 'Apply, logs, and exec events will appear here for the selected workspace.',
}: EventFeedProps) {
  if (events.length === 0) {
    return <EmptyState title={emptyTitle} description={emptyDescription} />
  }

  return (
    <ol className="event-feed">
      {events.map((event) => (
        <li key={`${event.sequence}-${event.kind}`} className="event-feed__item">
          <div className="event-feed__meta">
            <strong>{event.kind}</strong>
            <span>{formatDateTime(event.timestamp)}</span>
          </div>
          <p className="event-feed__summary">{eventSummary(event)}</p>
          {event.payload ? <pre className="event-feed__payload">{compactJson(event.payload)}</pre> : null}
        </li>
      ))}
    </ol>
  )
}
