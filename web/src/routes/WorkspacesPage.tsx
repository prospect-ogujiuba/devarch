import { useEffect, useMemo, useState } from 'react'
import { Link, useNavigate, useParams, useSearchParams } from 'react-router-dom'
import { parse, stringify } from 'yaml'
import { api } from '../lib/api'
import {
  formatCapabilityList,
  formatDateTime,
  formatEnvValue,
  formatLogChunk,
  formatPort,
  formatVolume,
  summarizeOperation,
} from '../lib/format'
import { getDefaultLogTail, setSelectedWorkspace } from '../lib/settings'
import type {
  AdapterCapabilities,
  ApplyResult,
  DesiredResource,
  Diagnostic,
  EnvValue,
  EventEnvelope,
  PortBinding,
  ResolveResource,
  SnapshotResource,
} from '../generated/api'
import { useWorkspaceEvents } from '../features/activity/useWorkspaceEvents'
import { useWorkspaceBundle, useWorkspaceLogs, useWorkspaceResourceOptions, useWorkspaces } from '../features/workspaces/hooks'
import { Card } from '../components/Card'
import { CodeEditor } from '../components/CodeEditor'
import { EmptyState } from '../components/EmptyState'
import { ErrorPanel } from '../components/ErrorPanel'
import { EventFeed } from '../components/EventFeed'
import { LoadingBlock } from '../components/LoadingBlock'
import { StatusBadge, toneFromStatus } from '../components/StatusBadge'
import { classNames } from '../lib/utils'

type WorkspaceTab = 'overview' | 'resources' | 'graph' | 'plan' | 'logs' | 'raw'

const workspaceTabs: Array<{ id: WorkspaceTab; label: string }> = [
  { id: 'overview', label: 'Overview' },
  { id: 'resources', label: 'Resources' },
  { id: 'graph', label: 'Graph' },
  { id: 'plan', label: 'Plan' },
  { id: 'logs', label: 'Logs' },
  { id: 'raw', label: 'Raw Config' },
]

function normalizeTab(value: string | null): WorkspaceTab {
  if (workspaceTabs.some((tab) => tab.id === value)) {
    return value as WorkspaceTab
  }
  return 'overview'
}

function listOrDash(values?: string[]) {
  return values && values.length > 0 ? values.join(', ') : '—'
}

function envEntries(values?: Record<string, EnvValue>) {
  return values ? Object.entries(values) : []
}

function diagnosticsForResource(resourceKey: string, diagnostics?: Diagnostic[]) {
  return (diagnostics ?? []).filter((item) => item.resource === resourceKey)
}

export function WorkspacesPage() {
  const { workspaceName } = useParams<{ workspaceName: string }>()
  const navigate = useNavigate()
  const [searchParams, setSearchParams] = useSearchParams()
  const activeTab = normalizeTab(searchParams.get('tab'))

  const workspaceList = useWorkspaces()
  const workspaceBundle = useWorkspaceBundle(workspaceName)
  const resourceOptions = useWorkspaceResourceOptions(workspaceBundle.data)

  const [selectedResource, setSelectedResource] = useState('')
  const [tail, setTail] = useState(getDefaultLogTail())
  const workspaceLogs = useWorkspaceLogs(workspaceName, selectedResource, tail, activeTab === 'logs' && !!selectedResource)
  const { events, error: eventsError } = useWorkspaceEvents(workspaceName)

  const [applyResult, setApplyResult] = useState<ApplyResult | null>(null)
  const [applyError, setApplyError] = useState<string | null>(null)
  const [applyLoading, setApplyLoading] = useState(false)
  const [copyLabel, setCopyLabel] = useState('Copy YAML')
  const [rawDraft, setRawDraft] = useState('')

  useEffect(() => {
    if (!workspaceName && workspaceList.data.length > 0) {
      navigate(`/workspaces/${workspaceList.data[0].name}`, { replace: true })
    }
  }, [navigate, workspaceList.data, workspaceName])

  useEffect(() => {
    if (workspaceName) {
      setSelectedWorkspace(workspaceName)
      setApplyResult(null)
      setApplyError(null)
      setCopyLabel('Copy YAML')
      setTail(getDefaultLogTail())
    }
  }, [workspaceName])

  useEffect(() => {
    if (resourceOptions.length === 0) {
      setSelectedResource('')
      return
    }
    if (!resourceOptions.includes(selectedResource)) {
      setSelectedResource(resourceOptions[0])
    }
  }, [resourceOptions, selectedResource])

  const manifestYaml = useMemo(() => {
    return workspaceBundle.data?.manifest ? stringify(workspaceBundle.data.manifest) : ''
  }, [workspaceBundle.data?.manifest])

  useEffect(() => {
    setRawDraft(manifestYaml)
  }, [manifestYaml])

  const draftError = useMemo(() => {
    if (!rawDraft.trim()) {
      return 'Manifest draft cannot be empty.'
    }
    try {
      parse(rawDraft)
      return null
    } catch (error) {
      return error instanceof Error ? error.message : 'Draft YAML is invalid.'
    }
  }, [rawDraft])

  const desiredResources = workspaceBundle.data?.status?.desired?.resources ?? []
  const graphResources = workspaceBundle.data?.graph?.graph.resources ?? []
  const snapshotResources = workspaceBundle.data?.status?.snapshot?.resources ?? []
  const resources = useMemo(() => mergeResourceViews(desiredResources, graphResources), [desiredResources, graphResources])

  const setTab = (nextTab: WorkspaceTab) => {
    const nextSearch = new URLSearchParams(searchParams)
    nextSearch.set('tab', nextTab)
    setSearchParams(nextSearch, { replace: true })
  }

  const selectWorkspace = (name: string) => {
    navigate({
      pathname: `/workspaces/${name}`,
      search: activeTab === 'overview' ? '' : `?tab=${activeTab}`,
    })
  }

  const runApply = async () => {
    if (!workspaceName) return
    setApplyLoading(true)
    setApplyError(null)
    try {
      const result = await api.applyWorkspace(workspaceName)
      setApplyResult(result)
      await workspaceBundle.reload()
    } catch (error) {
      setApplyError(error instanceof Error ? error.message : 'Apply failed')
      setApplyResult(null)
    } finally {
      setApplyLoading(false)
    }
  }

  const copyYaml = async () => {
    if (!rawDraft) return
    if (typeof navigator === 'undefined' || !navigator.clipboard) {
      setCopyLabel('Clipboard unavailable')
      return
    }
    await navigator.clipboard.writeText(rawDraft)
    setCopyLabel('Copied')
    window.setTimeout(() => setCopyLabel('Copy YAML'), 1000)
  }

  const selectedSummary = workspaceList.data.find((workspace) => workspace.name === workspaceName)

  return (
    <div className="page stack-lg">
      <header className="page__header">
        <div>
          <div className="page__eyebrow">Phase 5 · Workspace flow</div>
          <h1 className="page__title">Workspaces</h1>
          <p className="page__subtitle">
            Open a workspace, inspect the effective graph, preview the plan, trigger apply, and observe logs without drifting back into V1 admin sprawl.
          </p>
        </div>
      </header>

      <div className="workspace-layout">
        <aside className="workspace-layout__sidebar">
          <Card
            title="Workspace list"
            description="Workspaces are the primary entry point in V2."
            actions={
              <button className="button button--ghost" onClick={() => workspaceList.reload()} type="button">
                Refresh
              </button>
            }
          >
            {workspaceList.loading ? <LoadingBlock label="Loading workspaces…" /> : null}
            {workspaceList.error ? <ErrorPanel title="Workspace list unavailable" message={workspaceList.error} /> : null}
            {!workspaceList.loading && workspaceList.data.length === 0 ? (
              <EmptyState title="No workspaces found" description="Point devarchd at a workspace root to populate this list." />
            ) : null}
            <div className="workspace-list">
              {workspaceList.data.map((workspace) => {
                const active = workspace.name === workspaceName
                return (
                  <button
                    key={workspace.name}
                    className={classNames('workspace-list__item', active && 'workspace-list__item--active')}
                    type="button"
                    onClick={() => selectWorkspace(workspace.name)}
                  >
                    <div className="workspace-list__title-row">
                      <strong>{workspace.displayName || workspace.name}</strong>
                      <StatusBadge tone={toneFromStatus(workspace.provider)}>{workspace.provider || 'auto'}</StatusBadge>
                    </div>
                    <div className="workspace-list__meta">{workspace.resourceCount} resources</div>
                    <p className="workspace-list__description">{workspace.description || 'No description provided.'}</p>
                  </button>
                )
              })}
            </div>
          </Card>
        </aside>

        <section className="workspace-layout__detail">
          {!workspaceName ? (
            <EmptyState title="Select a workspace" description="Choose a workspace from the list to open its detail flow." />
          ) : workspaceBundle.loading && !workspaceBundle.data ? (
            <LoadingBlock label="Loading workspace detail…" />
          ) : workspaceBundle.error ? (
            <ErrorPanel title="Workspace detail unavailable" message={workspaceBundle.error} />
          ) : !workspaceBundle.data ? (
            <EmptyState title="Workspace unavailable" description="The selected workspace could not be loaded." />
          ) : (
            <div className="stack-lg">
              <Card
                title={workspaceBundle.data.detail?.displayName || workspaceBundle.data.detail?.name || selectedSummary?.displayName || workspaceName}
                description={workspaceBundle.data.detail?.description || selectedSummary?.description || 'Developer-oriented manifest and runtime view.'}
                actions={
                  <div className="button-group">
                    <button className="button button--ghost" onClick={() => workspaceBundle.reload()} type="button">
                      Refresh tabs
                    </button>
                    <Link className="button button--ghost" to="/activity">
                      Open activity
                    </Link>
                  </div>
                }
              >
                <div className="stat-grid">
                  <div className="stat-card">
                    <span className="stat-card__label">Provider</span>
                    <strong>{workspaceBundle.data.detail?.provider || 'auto'}</strong>
                  </div>
                  <div className="stat-card">
                    <span className="stat-card__label">Resources</span>
                    <strong>{workspaceBundle.data.detail?.resourceCount ?? resourceOptions.length}</strong>
                  </div>
                  <div className="stat-card">
                    <span className="stat-card__label">Capabilities</span>
                    <strong>{listOrDash(formatCapabilityList(workspaceBundle.data.detail?.capabilities))}</strong>
                  </div>
                </div>
              </Card>

              <nav className="tab-nav" aria-label="Workspace detail tabs">
                {workspaceTabs.map((tab) => (
                  <button
                    key={tab.id}
                    type="button"
                    className={classNames('tab-nav__button', activeTab === tab.id && 'tab-nav__button--active')}
                    onClick={() => setTab(tab.id)}
                  >
                    {tab.label}
                  </button>
                ))}
              </nav>

              {activeTab === 'overview' ? (
                <OverviewTab
                  manifestPath={workspaceBundle.data.detail?.manifestPath}
                  statusError={workspaceBundle.data.errors.status}
                  capabilities={workspaceBundle.data.detail?.capabilities}
                  desiredResources={desiredResources}
                  snapshotResources={snapshotResources}
                  diagnostics={workspaceBundle.data.status?.desired?.diagnostics}
                />
              ) : null}

              {activeTab === 'resources' ? (
                <ResourcesTab resources={resources} snapshotResources={snapshotResources} diagnostics={workspaceBundle.data.status?.desired?.diagnostics} />
              ) : null}

              {activeTab === 'graph' ? (
                <GraphTab
                  resources={graphResources}
                  contractLinks={workspaceBundle.data.graph?.contracts?.links}
                  contractDiagnostics={workspaceBundle.data.graph?.contracts?.diagnostics}
                  graphError={workspaceBundle.data.errors.graph}
                />
              ) : null}

              {activeTab === 'plan' ? (
                <PlanTab
                  plan={workspaceBundle.data.plan}
                  planError={workspaceBundle.data.errors.plan}
                  applyLoading={applyLoading}
                  applyError={applyError}
                  applyResult={applyResult}
                  onApply={runApply}
                />
              ) : null}

              {activeTab === 'logs' ? (
                <LogsTab
                  selectedResource={selectedResource}
                  setSelectedResource={setSelectedResource}
                  resourceOptions={resourceOptions}
                  tail={tail}
                  setTail={setTail}
                  logs={workspaceLogs.data}
                  logsLoading={workspaceLogs.loading}
                  logsError={workspaceLogs.error}
                  reloadLogs={workspaceLogs.reload}
                  events={events}
                  eventsError={eventsError}
                />
              ) : null}

              {activeTab === 'raw' ? (
                <RawTab
                  rawDraft={rawDraft}
                  setRawDraft={setRawDraft}
                  draftError={draftError}
                  copyLabel={copyLabel}
                  onCopy={copyYaml}
                  onReset={() => setRawDraft(manifestYaml)}
                  manifestError={workspaceBundle.data.errors.manifest}
                />
              ) : null}
            </div>
          )}
        </section>
      </div>
    </div>
  )
}

function OverviewTab({
  manifestPath,
  statusError,
  capabilities,
  desiredResources,
  snapshotResources,
  diagnostics,
}: {
  manifestPath?: string
  statusError?: string
  capabilities?: AdapterCapabilities
  desiredResources: DesiredResource[]
  snapshotResources: SnapshotResource[]
  diagnostics?: Diagnostic[]
}) {
  return (
    <div className="detail-grid">
      <Card title="Workspace metadata" description="Core manifest identity and runtime boundary.">
        <dl className="key-value-grid">
          <div>
            <dt>Manifest path</dt>
            <dd>{manifestPath || '—'}</dd>
          </div>
          <div>
            <dt>Capabilities</dt>
            <dd>{listOrDash(formatCapabilityList(capabilities))}</dd>
          </div>
          <div>
            <dt>Desired resources</dt>
            <dd>{desiredResources.length}</dd>
          </div>
          <div>
            <dt>Observed resources</dt>
            <dd>{snapshotResources.length}</dd>
          </div>
        </dl>
      </Card>

      <Card title="Runtime snapshot" description="Observed state remains derived, not canonical.">
        {statusError ? <ErrorPanel title="Runtime status unavailable" message={statusError} /> : null}
        {!statusError && snapshotResources.length === 0 ? (
          <EmptyState title="No snapshot resources" description="Status has not observed runtime resources yet." />
        ) : null}
        <div className="stack">
          {snapshotResources.map((resource) => (
            <div key={resource.key} className="row-card">
              <div className="row-card__title-row">
                <strong>{resource.key}</strong>
                <StatusBadge tone={toneFromStatus(resource.state?.status)}>{resource.state?.status || 'unknown'}</StatusBadge>
              </div>
              <div className="row-card__meta">{resource.runtimeName}</div>
            </div>
          ))}
        </div>
      </Card>

      <Card title="Diagnostics" description="Blocking and warning diagnostics carried into the UI.">
        {!(diagnostics && diagnostics.length > 0) ? (
          <EmptyState title="No workspace diagnostics" description="The workspace currently has no top-level diagnostics." />
        ) : (
          <ul className="stack list-reset">
            {diagnostics.map((diagnostic) => (
              <li key={`${diagnostic.code}-${diagnostic.resource ?? 'workspace'}`} className="row-card">
                <div className="row-card__title-row">
                  <strong>{diagnostic.code}</strong>
                  <StatusBadge tone={toneFromStatus(diagnostic.severity)}>{diagnostic.severity}</StatusBadge>
                </div>
                <p>{diagnostic.message}</p>
              </li>
            ))}
          </ul>
        )}
      </Card>
    </div>
  )
}

function ResourcesTab({
  resources,
  snapshotResources,
  diagnostics,
}: {
  resources: Array<{ desired?: DesiredResource; graph?: ResolveResource }>
  snapshotResources: SnapshotResource[]
  diagnostics?: Diagnostic[]
}) {
  if (resources.length === 0) {
    return <EmptyState title="No resources" description="The selected workspace does not expose any resources yet." />
  }

  return (
    <div className="stack">
      {resources.map(({ desired, graph }) => {
        const key = desired?.key ?? graph?.key ?? 'resource'
        const snapshot = snapshotResources.find((resource) => resource.key === key)
        const resourceDiagnostics = diagnosticsForResource(key, diagnostics).concat(desired?.diagnostics ?? [])

        return (
          <Card key={key} title={key} description={graph?.template?.name || desired?.templateName || graph?.source?.type || 'Workspace resource'}>
            <div className="resource-grid">
              <div>
                <div className="field-label">Runtime name</div>
                <div>{desired?.runtimeName || snapshot?.runtimeName || '—'}</div>
              </div>
              <div>
                <div className="field-label">Host</div>
                <div>{graph?.host || desired?.logicalHost || '—'}</div>
              </div>
              <div>
                <div className="field-label">State</div>
                <div>
                  <StatusBadge tone={toneFromStatus(snapshot?.state?.status)}>{snapshot?.state?.status || 'unobserved'}</StatusBadge>
                </div>
              </div>
              <div>
                <div className="field-label">Depends on</div>
                <div>{listOrDash(desired?.dependsOn || graph?.dependsOn)}</div>
              </div>
              <div>
                <div className="field-label">Ports</div>
                <div>{listOrDash((graph?.ports || desired?.spec.ports)?.map(formatMixedPort))}</div>
              </div>
              <div>
                <div className="field-label">Volumes</div>
                <div>{listOrDash((graph?.volumes || desired?.spec.volumes)?.map(formatVolume))}</div>
              </div>
              <div>
                <div className="field-label">Domains</div>
                <div>{listOrDash(desired?.domains || graph?.domains)}</div>
              </div>
              <div>
                <div className="field-label">Image</div>
                <div>{desired?.spec.image || graph?.runtime?.image || '—'}</div>
              </div>
            </div>

            <div className="detail-grid detail-grid--tight">
              <Card title="Contracts" description="Imports and exports stay compact and explicit.">
                <div className="stack-sm">
                  <div>
                    <div className="field-label">Imports</div>
                    <div>{listOrDash((graph?.imports || []).map((item) => `${item.contract}${item.from ? ` ← ${item.from}` : ''}`))}</div>
                  </div>
                  <div>
                    <div className="field-label">Exports</div>
                    <div>{listOrDash((graph?.exports || []).map((item) => item.contract))}</div>
                  </div>
                </div>
              </Card>

              <Card title="Environment" description="Declared and injected env values in one place.">
                {envEntries(desired?.spec.env || graph?.env).length === 0 ? (
                  <EmptyState title="No env values" description="This resource does not expose env values in the current view." />
                ) : (
                  <dl className="key-value-grid key-value-grid--single">
                    {envEntries(desired?.spec.env || graph?.env).map(([envKey, envValue]) => (
                      <div key={envKey}>
                        <dt>{envKey}</dt>
                        <dd>{formatEnvValue(envValue)}</dd>
                      </div>
                    ))}
                  </dl>
                )}
              </Card>
            </div>

            {resourceDiagnostics.length > 0 ? (
              <Card title="Resource diagnostics" description="Warnings and errors tied to this resource.">
                <ul className="stack list-reset">
                  {resourceDiagnostics.map((diagnostic, index) => (
                    <li key={`${diagnostic.code}-${index}`} className="row-card">
                      <div className="row-card__title-row">
                        <strong>{diagnostic.code}</strong>
                        <StatusBadge tone={toneFromStatus(diagnostic.severity)}>{diagnostic.severity}</StatusBadge>
                      </div>
                      <p>{diagnostic.message}</p>
                    </li>
                  ))}
                </ul>
              </Card>
            ) : null}
          </Card>
        )
      })}
    </div>
  )
}

function GraphTab({
  resources,
  contractLinks,
  contractDiagnostics,
  graphError,
}: {
  resources: ResolveResource[]
  contractLinks?: Array<{ consumer: string; contract: string; provider: string; source: string; alias?: string }>
  contractDiagnostics?: Array<{ code: string; severity: string; message: string; consumer?: string }>
  graphError?: string
}) {
  return (
    <div className="detail-grid">
      <Card title="Dependency graph" description="Resolved dependency edges and runtime hosts.">
        {graphError ? <ErrorPanel title="Graph unavailable" message={graphError} /> : null}
        {resources.length === 0 ? <EmptyState title="No resolved graph" description="Workspace graph data is not available yet." /> : null}
        <div className="stack">
          {resources.map((resource) => (
            <div key={resource.key} className="row-card">
              <div className="row-card__title-row">
                <strong>{resource.key}</strong>
                <StatusBadge tone={resource.enabled ? 'success' : 'warning'}>{resource.enabled ? 'enabled' : 'disabled'}</StatusBadge>
              </div>
              <div className="row-card__meta">host: {resource.host}</div>
              <p>depends on: {listOrDash(resource.dependsOn)}</p>
              <p>imports: {listOrDash((resource.imports || []).map((item) => item.contract))}</p>
            </div>
          ))}
        </div>
      </Card>

      <Card title="Contract links" description="Resolved provider/consumer links from the contract solver.">
        {!(contractLinks && contractLinks.length > 0) ? (
          <EmptyState title="No contract links" description="This workspace has no resolved contract links in the current view." />
        ) : (
          <div className="stack list-reset">
            {contractLinks.map((link) => (
              <div key={`${link.consumer}-${link.contract}-${link.provider}`} className="row-card">
                <div className="row-card__title-row">
                  <strong>{link.contract}</strong>
                  <StatusBadge tone="info">{link.consumer} → {link.provider}</StatusBadge>
                </div>
                <p>source: {link.source}</p>
                {link.alias ? <p>alias: {link.alias}</p> : null}
              </div>
            ))}
          </div>
        )}
      </Card>

      <Card title="Contract diagnostics" description="Ambiguity and unresolved states stay visible before apply.">
        {!(contractDiagnostics && contractDiagnostics.length > 0) ? (
          <EmptyState title="No contract diagnostics" description="Contract resolution completed without surfaced diagnostics." />
        ) : (
          <div className="stack">
            {contractDiagnostics.map((diagnostic, index) => (
              <div key={`${diagnostic.code}-${index}`} className="row-card">
                <div className="row-card__title-row">
                  <strong>{diagnostic.code}</strong>
                  <StatusBadge tone={toneFromStatus(diagnostic.severity)}>{diagnostic.severity}</StatusBadge>
                </div>
                <p>{diagnostic.message}</p>
              </div>
            ))}
          </div>
        )}
      </Card>
    </div>
  )
}

function PlanTab({
  plan,
  planError,
  applyLoading,
  applyError,
  applyResult,
  onApply,
}: {
  plan?: { blocked?: boolean; diagnostics?: Diagnostic[]; actions?: Array<{ scope: string; target: string; runtimeName?: string; kind: string; reasons?: string[] }> }
  planError?: string
  applyLoading: boolean
  applyError: string | null
  applyResult: ApplyResult | null
  onApply: () => void
}) {
  return (
    <div className="detail-grid">
      <Card
        title="Plan actions"
        description="Review planned actions and reasons before apply."
        actions={
          <button className="button" type="button" onClick={onApply} disabled={applyLoading}>
            {applyLoading ? 'Applying…' : 'Apply workspace'}
          </button>
        }
      >
        {planError ? <ErrorPanel title="Plan unavailable" message={planError} /> : null}
        {applyError ? <ErrorPanel title="Apply failed" message={applyError} /> : null}
        {plan?.blocked ? <div className="panel panel--warning">Apply is currently blocked by diagnostics.</div> : null}
        {!(plan?.actions && plan.actions.length > 0) ? (
          <EmptyState title="No plan actions" description="No planned actions are available for this workspace yet." />
        ) : (
          <ul className="stack list-reset">
            {plan.actions.map((action) => (
              <li key={`${action.scope}-${action.target}-${action.kind}`} className="row-card">
                <div className="row-card__title-row">
                  <strong>{action.target}</strong>
                  <StatusBadge tone={toneFromStatus(action.kind)}>{action.kind}</StatusBadge>
                </div>
                <div className="row-card__meta">{action.scope}{action.runtimeName ? ` · ${action.runtimeName}` : ''}</div>
                <ul className="list-reset stack-sm">
                  {(action.reasons || []).map((reason) => (
                    <li key={reason}>• {reason}</li>
                  ))}
                </ul>
              </li>
            ))}
          </ul>
        )}
      </Card>

      <Card title="Plan diagnostics" description="Planner diagnostics stay attached to the plan output.">
        {!(plan?.diagnostics && plan.diagnostics.length > 0) ? (
          <EmptyState title="No plan diagnostics" description="The planner did not return additional diagnostics." />
        ) : (
          <div className="stack">
            {plan.diagnostics.map((diagnostic, index) => (
              <div key={`${diagnostic.code}-${index}`} className="row-card">
                <div className="row-card__title-row">
                  <strong>{diagnostic.code}</strong>
                  <StatusBadge tone={toneFromStatus(diagnostic.severity)}>{diagnostic.severity}</StatusBadge>
                </div>
                <p>{diagnostic.message}</p>
              </div>
            ))}
          </div>
        )}
      </Card>

      <Card title="Latest apply result" description="The UI shows the latest apply execution result returned by the thin API.">
        {!applyResult ? (
          <EmptyState title="No apply result yet" description="Run apply to populate the latest operation summary." />
        ) : (
          <div className="stack">
            <div className="panel">
              Started {formatDateTime(applyResult.startedAt)} · Finished {formatDateTime(applyResult.finishedAt)}
            </div>
            <ul className="stack list-reset">
              {(applyResult.operations || []).map((operation) => (
                <li key={`${operation.scope}-${operation.target}-${operation.kind}`} className="row-card">
                  <div className="row-card__title-row">
                    <strong>{summarizeOperation(operation)}</strong>
                    <StatusBadge tone={toneFromStatus(operation.status)}>{operation.status}</StatusBadge>
                  </div>
                  {operation.message ? <p>{operation.message}</p> : null}
                </li>
              ))}
            </ul>
          </div>
        )}
      </Card>
    </div>
  )
}

function LogsTab({
  selectedResource,
  setSelectedResource,
  resourceOptions,
  tail,
  setTail,
  logs,
  logsLoading,
  logsError,
  reloadLogs,
  events,
  eventsError,
}: {
  selectedResource: string
  setSelectedResource: (value: string) => void
  resourceOptions: string[]
  tail: number
  setTail: (value: number) => void
  logs: Array<{ timestamp?: string; stream?: string; line: string }>
  logsLoading: boolean
  logsError: string | null
  reloadLogs: () => void
  events: EventEnvelope[]
  eventsError: string | null
}) {
  return (
    <div className="detail-grid">
      <Card
        title="Resource logs"
        description="Fetch log chunks from the JSON logs endpoint and pair them with workspace activity."
        actions={
          <div className="toolbar toolbar--compact">
            <select value={selectedResource} onChange={(event) => setSelectedResource(event.target.value)}>
              {resourceOptions.map((resource) => (
                <option key={resource} value={resource}>
                  {resource}
                </option>
              ))}
            </select>
            <label className="input-with-label">
              <span>Tail</span>
              <input
                type="number"
                min={1}
                value={tail}
                onChange={(event) => setTail(Math.max(1, Number.parseInt(event.target.value || '1', 10)))}
              />
            </label>
            <button className="button button--ghost" type="button" onClick={reloadLogs}>
              Reload
            </button>
          </div>
        }
      >
        {logsError ? <ErrorPanel title="Logs unavailable" message={logsError} /> : null}
        {logsLoading ? <LoadingBlock label="Loading logs…" /> : null}
        {!logsLoading && logs.length === 0 ? (
          <EmptyState title="No log lines" description="Select a resource and fetch logs to populate this view." />
        ) : null}
        {logs.length > 0 ? <pre className="log-view">{logs.map(formatLogChunk).join('\n')}</pre> : null}
      </Card>

      <Card title="Workspace activity" description="Workspace-scoped SSE feed from apply, logs, and exec events.">
        {eventsError ? <ErrorPanel title="Activity stream notice" message={eventsError} /> : null}
        <EventFeed
          events={events}
          emptyTitle="No workspace events"
          emptyDescription="Events appear here when the API publishes apply, logs, or exec activity for the selected workspace."
        />
      </Card>
    </div>
  )
}

function RawTab({
  rawDraft,
  setRawDraft,
  draftError,
  copyLabel,
  onCopy,
  onReset,
  manifestError,
}: {
  rawDraft: string
  setRawDraft: (value: string) => void
  draftError: string | null
  copyLabel: string
  onCopy: () => void
  onReset: () => void
  manifestError?: string
}) {
  return (
    <div className="stack">
      <Card
        title="Raw manifest editor"
        description="Advanced users can inspect and edit a canonical manifest draft directly."
        actions={
          <div className="button-group">
            <button className="button button--ghost" type="button" onClick={onReset}>
              Reset draft
            </button>
            <button className="button" type="button" onClick={onCopy}>
              {copyLabel}
            </button>
          </div>
        }
      >
        {manifestError ? <ErrorPanel title="Manifest unavailable" message={manifestError} /> : null}
        <div className="panel panel--warning">
          The Phase 4 API exposes manifest reads but not manifest writes yet, so this editor validates and preserves a local draft only.
        </div>
        {draftError ? <ErrorPanel title="Draft validation failed" message={draftError} /> : null}
        {!draftError ? <div className="panel">Draft YAML parses successfully against the current editor state.</div> : null}
        <CodeEditor value={rawDraft} onChange={setRawDraft} invalid={!!draftError} minRows={22} />
      </Card>
    </div>
  )
}

function formatMixedPort(port: PortBinding | NonNullable<DesiredResource['spec']['ports']>[number]) {
  const published = (port as { published?: number }).published
  const host = (port as PortBinding).host
  return formatPort({
    host: published ?? host,
    container: port.container,
    protocol: port.protocol,
    hostIP: port.hostIP,
  })
}

function mergeResourceViews(desiredResources: DesiredResource[], graphResources: ResolveResource[]) {
  const keys = Array.from(new Set([...desiredResources.map((resource) => resource.key), ...graphResources.map((resource) => resource.key)])).sort()
  return keys.map((key) => ({
    desired: desiredResources.find((resource) => resource.key === key),
    graph: graphResources.find((resource) => resource.key === key),
  }))
}
