import { useCallback, useEffect, useMemo, useState } from 'react'
import type {
  LogChunk,
  PlanResult,
  WorkspaceDetail,
  WorkspaceGraphView,
  WorkspaceManifest,
  WorkspaceStatusView,
  WorkspaceSummary,
} from '../../generated/api'
import { api } from '../../lib/api'

function toErrorMessage(error: unknown) {
  return error instanceof Error ? error.message : 'Request failed'
}

export function useWorkspaces() {
  const [data, setData] = useState<WorkspaceSummary[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const load = useCallback(async () => {
    setLoading(true)
    setError(null)
    try {
      setData(await api.listWorkspaces())
    } catch (nextError) {
      setError(toErrorMessage(nextError))
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    void load()
  }, [load])

  return { data, loading, error, reload: load }
}

export interface WorkspaceBundle {
  detail?: WorkspaceDetail
  manifest?: WorkspaceManifest
  graph?: WorkspaceGraphView
  status?: WorkspaceStatusView
  plan?: PlanResult
  errors: Partial<Record<'detail' | 'manifest' | 'graph' | 'status' | 'plan', string>>
}

export function useWorkspaceBundle(name?: string) {
  const [data, setData] = useState<WorkspaceBundle | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const load = useCallback(async () => {
    if (!name) {
      setData(null)
      setError(null)
      setLoading(false)
      return
    }

    setLoading(true)
    setError(null)
    const requests = await Promise.allSettled([
      api.getWorkspace(name),
      api.getWorkspaceManifest(name),
      api.getWorkspaceGraph(name),
      api.getWorkspaceStatus(name),
      api.getWorkspacePlan(name),
    ])

    const nextData: WorkspaceBundle = { errors: {} }

    const [detail, manifest, graph, status, plan] = requests
    if (detail.status === 'fulfilled') nextData.detail = detail.value
    else nextData.errors.detail = toErrorMessage(detail.reason)

    if (manifest.status === 'fulfilled') nextData.manifest = manifest.value
    else nextData.errors.manifest = toErrorMessage(manifest.reason)

    if (graph.status === 'fulfilled') nextData.graph = graph.value
    else nextData.errors.graph = toErrorMessage(graph.reason)

    if (status.status === 'fulfilled') nextData.status = status.value
    else nextData.errors.status = toErrorMessage(status.reason)

    if (plan.status === 'fulfilled') nextData.plan = plan.value
    else nextData.errors.plan = toErrorMessage(plan.reason)

    if (!nextData.detail && !nextData.manifest && !nextData.graph) {
      setError(nextData.errors.detail ?? nextData.errors.manifest ?? nextData.errors.graph ?? 'Workspace failed to load')
      setData(null)
    } else {
      setData(nextData)
      setError(null)
    }
    setLoading(false)
  }, [name])

  useEffect(() => {
    void load()
  }, [load])

  return { data, loading, error, reload: load }
}

export function useWorkspaceLogs(name?: string, resource?: string, tail = 200, enabled = true) {
  const [data, setData] = useState<LogChunk[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const load = useCallback(async () => {
    if (!name || !resource || !enabled) {
      setData([])
      setError(null)
      setLoading(false)
      return
    }

    setLoading(true)
    setError(null)
    try {
      setData(await api.getWorkspaceLogs(name, resource, tail))
    } catch (nextError) {
      setError(toErrorMessage(nextError))
      setData([])
    } finally {
      setLoading(false)
    }
  }, [enabled, name, resource, tail])

  useEffect(() => {
    void load()
  }, [load])

  return { data, loading, error, reload: load }
}

export function useWorkspaceResourceOptions(bundle: WorkspaceBundle | null) {
  return useMemo(() => {
    if (bundle?.detail?.resourceKeys?.length) {
      return bundle.detail.resourceKeys
    }
    if (bundle?.status?.desired?.resources?.length) {
      return bundle.status.desired.resources.map((resource) => resource.key)
    }
    if (bundle?.manifest?.resources) {
      return Object.keys(bundle.manifest.resources).sort()
    }
    return []
  }, [bundle])
}
