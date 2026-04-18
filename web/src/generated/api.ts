export type JsonValue =
  | string
  | number
  | boolean
  | null
  | JsonValue[]
  | { [key: string]: JsonValue }

export type EnvValue = string | number | boolean | { secretRef: string }

export interface ApiErrorEnvelope {
  error: {
    code: string
    message: string
    details?: unknown
  }
}

export interface AdapterCapabilities {
  inspect?: boolean
  apply?: boolean
  logs?: boolean
  exec?: boolean
  network?: boolean
}

export interface WorkspaceSummary {
  name: string
  displayName?: string
  description?: string
  provider?: string
  capabilities?: AdapterCapabilities
  resourceCount: number
}

export interface WorkspaceDetail extends WorkspaceSummary {
  manifestPath: string
  resourceKeys?: string[]
}

export interface RuntimePreferences {
  provider?: string
  isolatedNetwork?: boolean
  namingStrategy?: string
}

export interface CatalogConfig {
  sources?: string[]
}

export interface Policies {
  autoWire?: boolean
  secretSource?: string
}

export interface PortBinding {
  host?: number
  container: number
  protocol?: string
  hostIP?: string
}

export interface VolumeMount {
  source?: string
  target: string
  readOnly?: boolean
  kind?: string
  type?: string
}

export interface ImportContract {
  contract: string
  from?: string
  alias?: string
}

export interface ExportContract {
  contract: string
  env?: Record<string, string>
}

export interface HealthCheck {
  test?: string[]
  interval?: string
  timeout?: string
  retries?: number
  startPeriod?: string
}

export interface ManifestSourceRef {
  type: string
  path: string
  service?: string
}

export interface ManifestResource {
  template?: string
  source?: ManifestSourceRef
  enabled?: boolean
  env?: Record<string, EnvValue>
  ports?: PortBinding[]
  volumes?: VolumeMount[]
  dependsOn?: string[]
  imports?: ImportContract[]
  exports?: ExportContract[]
  health?: HealthCheck
  domains?: string[]
  develop?: Record<string, JsonValue>
  overrides?: Record<string, JsonValue>
}

export interface WorkspaceManifest {
  apiVersion?: string
  kind?: string
  metadata: {
    name: string
    displayName?: string
    description?: string
  }
  runtime?: RuntimePreferences
  catalog?: CatalogConfig
  policies?: Policies
  resources: Record<string, ManifestResource>
}

export interface ResolveTemplateRef {
  name: string
  path?: string
}

export interface ResolveRuntime {
  image?: string
  build?: {
    context: string
    dockerfile?: string
    target?: string
    args?: Record<string, EnvValue>
  }
  command?: string[]
  entrypoint?: string[]
  workingDir?: string
}

export interface ResolveResource {
  key: string
  enabled: boolean
  host: string
  template?: ResolveTemplateRef
  source?: ManifestSourceRef
  runtime?: ResolveRuntime
  env?: Record<string, EnvValue>
  ports?: PortBinding[]
  volumes?: VolumeMount[]
  dependsOn?: string[]
  imports?: ImportContract[]
  exports?: ExportContract[]
  health?: HealthCheck
  domains?: string[]
  develop?: Record<string, JsonValue>
  overrides?: Record<string, JsonValue>
}

export interface ResolveGraph {
  workspace: {
    name: string
    displayName?: string
    description?: string
    runtime?: RuntimePreferences
    policies?: Policies
    catalogSources?: string[]
  }
  resources: ResolveResource[]
}

export interface ContractLink {
  consumer: string
  contract: string
  alias?: string
  provider: string
  source: string
  injectedEnv?: Record<string, EnvValue>
}

export interface ContractDiagnostic {
  severity: string
  code: string
  consumer?: string
  contract?: string
  provider?: string
  providers?: string[]
  envKey?: string
  message: string
}

export interface ContractsResult {
  links?: ContractLink[]
  diagnostics?: ContractDiagnostic[]
}

export interface WorkspaceGraphView {
  graph: ResolveGraph
  contracts?: ContractsResult
}

export interface Diagnostic {
  severity: string
  code: string
  workspace?: string
  resource?: string
  contract?: string
  provider?: string
  providers?: string[]
  envKey?: string
  message: string
}

export interface DesiredNetwork {
  name: string
  labels?: Record<string, string>
}

export interface DesiredResource {
  key: string
  enabled: boolean
  logicalHost: string
  runtimeName: string
  templateName?: string
  source?: ManifestSourceRef
  declaredEnv?: Record<string, EnvValue>
  injectedEnv?: Record<string, EnvValue>
  dependsOn?: string[]
  domains?: string[]
  overrideLabels?: Record<string, string>
  diagnostics?: Diagnostic[]
  spec: {
    image?: string
    build?: {
      context: string
      dockerfile?: string
      target?: string
      args?: Record<string, EnvValue>
    }
    command?: string[]
    entrypoint?: string[]
    workingDir?: string
    env?: Record<string, EnvValue>
    ports?: Array<{
      container: number
      published?: number
      protocol?: string
      hostIP?: string
    }>
    volumes?: VolumeMount[]
    health?: HealthCheck
    projectSource?: {
      hostPath: string
      containerPath: string
    }
    developWatch?: Array<{
      path: string
      resolvedPath?: string
      target: string
      action?: string
    }>
    labels?: Record<string, string>
  }
}

export interface DesiredWorkspace {
  name: string
  displayName?: string
  description?: string
  provider?: string
  namingStrategy?: string
  network?: DesiredNetwork
  resources?: DesiredResource[]
  diagnostics?: Diagnostic[]
  capabilities?: AdapterCapabilities
}

export interface SnapshotNetwork {
  name: string
  id?: string
  driver?: string
  labels?: Record<string, string>
}

export interface ResourceState {
  status?: string
  running?: boolean
  health?: string
  exitCode?: number
  restartCount?: number
  startedAt?: string
  finishedAt?: string
  error?: string
}

export interface SnapshotResource {
  key: string
  runtimeName: string
  logicalHost?: string
  id?: string
  state?: ResourceState
  spec: Record<string, JsonValue>
}

export interface Snapshot {
  workspace: {
    name: string
    provider?: string
    network?: SnapshotNetwork
  }
  resources?: SnapshotResource[]
}

export interface WorkspaceStatusView {
  desired: DesiredWorkspace
  snapshot?: Snapshot
}

export type ActionScope = 'workspace' | 'resource'
export type ActionKind = 'add' | 'modify' | 'remove' | 'restart' | 'noop'

export interface PlanAction {
  scope: ActionScope
  target: string
  runtimeName?: string
  kind: ActionKind
  reasons?: string[]
}

export interface PlanResult {
  workspace: string
  provider?: string
  blocked?: boolean
  diagnostics?: Diagnostic[]
  actions?: PlanAction[]
}

export interface ApplyOperation {
  scope: ActionScope
  target: string
  runtimeName?: string
  kind: ActionKind
  status: string
  message?: string
}

export interface ApplyResult {
  workspace: string
  provider?: string
  startedAt: string
  finishedAt: string
  operations?: ApplyOperation[]
  snapshot?: Snapshot
}

export interface LogChunk {
  timestamp?: string
  stream?: string
  line: string
}

export interface TemplateSummary {
  name: string
  description?: string
  tags?: string[]
}

export interface TemplateDetail {
  apiVersion?: string
  kind?: string
  name: string
  description?: string
  tags?: string[]
  runtime?: Record<string, JsonValue>
  env?: Record<string, EnvValue>
  ports?: PortBinding[]
  volumes?: VolumeMount[]
  imports?: ImportContract[]
  exports?: ExportContract[]
  health?: HealthCheck
  develop?: Record<string, JsonValue>
}

export interface EventEnvelope {
  sequence: number
  workspace: string
  resource?: string
  kind: string
  timestamp: string
  payload?: Record<string, JsonValue>
}
