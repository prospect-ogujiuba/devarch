import { useEffect, useState } from 'react'
import type { EventEnvelope } from '../../generated/api'
import { parseEventEnvelope, workspaceEventKinds, workspaceEventsUrl } from '../../lib/api'

export function useWorkspaceEvents(workspaceName?: string, maxEvents = 40) {
  const [events, setEvents] = useState<EventEnvelope[]>([])
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    setEvents([])
    setError(null)

    if (!workspaceName || typeof EventSource === 'undefined') {
      return
    }

    const source = new EventSource(workspaceEventsUrl(workspaceName))
    const handler = (event: MessageEvent) => {
      try {
        const envelope = parseEventEnvelope(event.data)
        setEvents((current) => [envelope, ...current].slice(0, maxEvents))
      } catch (nextError) {
        setError(nextError instanceof Error ? nextError.message : 'Failed to decode workspace event')
      }
    }

    workspaceEventKinds.forEach((kind) => source.addEventListener(kind, handler as EventListener))
    source.onerror = () => {
      setError('Workspace activity stream disconnected. Refresh to reconnect.')
    }

    return () => {
      workspaceEventKinds.forEach((kind) => source.removeEventListener(kind, handler as EventListener))
      source.close()
    }
  }, [maxEvents, workspaceName])

  return { events, error }
}
