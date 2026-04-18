import { useEffect } from 'react'
import { Link, useSearchParams } from 'react-router-dom'
import { Card } from '../components/Card'
import { EmptyState } from '../components/EmptyState'
import { ErrorPanel } from '../components/ErrorPanel'
import { EventFeed } from '../components/EventFeed'
import { LoadingBlock } from '../components/LoadingBlock'
import { StatusBadge, toneFromStatus } from '../components/StatusBadge'
import { useWorkspaceEvents } from '../features/activity/useWorkspaceEvents'
import { useWorkspaces } from '../features/workspaces/hooks'
import { getSelectedWorkspace, setSelectedWorkspace } from '../lib/settings'

export function ActivityPage() {
  const [searchParams, setSearchParams] = useSearchParams()
  const workspaceList = useWorkspaces()
  const selected = searchParams.get('workspace') || getSelectedWorkspace()
  const { events, error } = useWorkspaceEvents(selected)

  useEffect(() => {
    if (!selected && workspaceList.data.length > 0) {
      const fallback = workspaceList.data[0].name
      setSelectedWorkspace(fallback)
      setSearchParams({ workspace: fallback }, { replace: true })
    }
  }, [selected, setSearchParams, workspaceList.data])

  const selectedWorkspace = workspaceList.data.find((workspace) => workspace.name === selected)

  return (
    <div className="page stack-lg">
      <header className="page__header">
        <div>
          <div className="page__eyebrow">Phase 5 · Observe</div>
          <h1 className="page__title">Activity</h1>
          <p className="page__subtitle">
            Watch workspace-scoped apply, logs, and exec events through the shared Phase 4 event stream.
          </p>
        </div>
      </header>

      <Card title="Workspace activity stream" description="The current API exposes events per workspace, so Activity stays workspace-first too.">
        <div className="toolbar">
          {workspaceList.loading ? <LoadingBlock label="Loading workspaces…" /> : null}
          {!workspaceList.loading ? (
            <select
              value={selected}
              onChange={(event) => {
                setSelectedWorkspace(event.target.value)
                setSearchParams({ workspace: event.target.value })
              }}
            >
              {workspaceList.data.map((workspace) => (
                <option key={workspace.name} value={workspace.name}>
                  {workspace.displayName || workspace.name}
                </option>
              ))}
            </select>
          ) : null}
          {selected ? (
            <Link className="button button--ghost" to={`/workspaces/${selected}?tab=logs`}>
              Open logs tab
            </Link>
          ) : null}
        </div>

        {workspaceList.error ? <ErrorPanel title="Workspace list unavailable" message={workspaceList.error} /> : null}
        {!selectedWorkspace && !workspaceList.loading ? (
          <EmptyState title="No workspace selected" description="Choose a workspace to start streaming its activity." />
        ) : null}

        {selectedWorkspace ? (
          <div className="stack-lg">
            <div className="stat-grid">
              <div className="stat-card">
                <span className="stat-card__label">Workspace</span>
                <strong>{selectedWorkspace.displayName || selectedWorkspace.name}</strong>
              </div>
              <div className="stat-card">
                <span className="stat-card__label">Provider</span>
                <strong>{selectedWorkspace.provider || 'auto'}</strong>
              </div>
              <div className="stat-card">
                <span className="stat-card__label">Resources</span>
                <strong>{selectedWorkspace.resourceCount}</strong>
              </div>
            </div>

            <div className="panel">
              <StatusBadge tone={toneFromStatus(selectedWorkspace.provider)}>{selectedWorkspace.provider || 'auto'}</StatusBadge>
              <span className="panel__inline-gap">Events are delivered from `/api/workspaces/{selectedWorkspace.name}/events`.</span>
            </div>

            {error ? <ErrorPanel title="Activity stream notice" message={error} /> : null}
            <EventFeed events={events} />
          </div>
        ) : null}
      </Card>
    </div>
  )
}
