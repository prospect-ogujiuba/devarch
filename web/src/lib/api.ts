import type {
  ApiErrorEnvelope,
  ApplyResult,
  EventEnvelope,
  LogChunk,
  PlanResult,
  TemplateDetail,
  TemplateSummary,
  WorkspaceDetail,
  WorkspaceGraphView,
  WorkspaceManifest,
  WorkspaceStatusView,
  WorkspaceSummary,
} from '../generated/api'
import { getApiBase } from './settings'

const JSON_HEADERS = {
  Accept: 'application/json',
  'Content-Type': 'application/json',
}

export class ApiRequestError extends Error {
  status: number
  code?: string
  details?: unknown

  constructor(message: string, status: number, code?: string, details?: unknown) {
    super(message)
    this.name = 'ApiRequestError'
    this.status = status
    this.code = code
    this.details = details
  }
}

function normalizeBase(base: string) {
  const trimmed = base.trim() || '/api'
  return trimmed.endsWith('/') ? trimmed.slice(0, -1) : trimmed
}

export function apiPath(path: string) {
  const normalizedPath = path.startsWith('/') ? path : `/${path}`
  return `${normalizeBase(getApiBase())}${normalizedPath}`
}

export function apiUrl(path: string) {
  const joined = apiPath(path)
  if (/^https?:\/\//i.test(joined)) {
    return joined
  }
  if (typeof window === 'undefined') {
    return joined
  }
  return new URL(joined, window.location.origin).toString()
}

async function parseResponseBody(response: Response) {
  const contentType = response.headers.get('content-type') ?? ''
  if (contentType.includes('application/json')) {
    return response.json()
  }
  return response.text()
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(apiPath(path), {
    ...init,
    headers: {
      ...JSON_HEADERS,
      ...(init?.headers ?? {}),
    },
  })

  const body = await parseResponseBody(response)
  if (!response.ok) {
    const errorEnvelope = typeof body === 'object' && body ? (body as ApiErrorEnvelope) : undefined
    const message =
      errorEnvelope?.error?.message ||
      (typeof body === 'string' && body.trim() ? body : `Request failed with status ${response.status}`)
    throw new ApiRequestError(message, response.status, errorEnvelope?.error?.code, errorEnvelope?.error?.details)
  }

  return body as T
}

export const api = {
  listWorkspaces: () => request<WorkspaceSummary[]>('/workspaces'),
  getWorkspace: (name: string) => request<WorkspaceDetail>(`/workspaces/${encodeURIComponent(name)}`),
  getWorkspaceManifest: (name: string) => request<WorkspaceManifest>(`/workspaces/${encodeURIComponent(name)}/manifest`),
  getWorkspaceGraph: (name: string) => request<WorkspaceGraphView>(`/workspaces/${encodeURIComponent(name)}/graph`),
  getWorkspaceStatus: (name: string) => request<WorkspaceStatusView>(`/workspaces/${encodeURIComponent(name)}/status`),
  getWorkspacePlan: (name: string) => request<PlanResult>(`/workspaces/${encodeURIComponent(name)}/plan`),
  applyWorkspace: (name: string) => request<ApplyResult>(`/workspaces/${encodeURIComponent(name)}/apply`, { method: 'POST' }),
  getWorkspaceLogs: (name: string, resource: string, tail: number) =>
    request<LogChunk[]>(`/workspaces/${encodeURIComponent(name)}/logs?${new URLSearchParams({
      resource,
      tail: String(tail),
    }).toString()}`),
  listTemplates: () => request<TemplateSummary[]>('/catalog/templates'),
  getTemplate: (name: string) => request<TemplateDetail>(`/catalog/templates/${encodeURIComponent(name)}`),
}

export const workspaceEventKinds = [
  'apply.started',
  'apply.progress',
  'apply.completed',
  'logs.started',
  'logs.chunk',
  'logs.completed',
  'exec.started',
  'exec.completed',
] as const

export function workspaceEventsUrl(name: string) {
  return apiUrl(`/workspaces/${encodeURIComponent(name)}/events`)
}

export function parseEventEnvelope(raw: string) {
  return JSON.parse(raw) as EventEnvelope
}
