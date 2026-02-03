export interface Container {
  id: string
  name: string
  status: 'running' | 'exited' | 'stopped' | 'not-created' | 'paused'
  image: string
  version: string
  uptime: string
  ports: string[]
  category: string
  color: string
  statusColor: string
  cpu: string
  memory: string
  network: string
  cpuPercentage: number
  memoryUsedMb: number
  memoryLimitMb: number
  memoryPercentage: number
  restartCount: number
  healthStatus: string | null
  uptimeSeconds: number
  startedAt: string | null
  testDomains: string[]
  localhostUrls: string[]
}

export interface ContainersResponse {
  success: boolean
  data: {
    containers: Container[]
    stats: ContainerStats
    count: number
    total: number
  }
  meta?: {
    timestamp: number
    filter: string | null
    search: string | null
    sort: string
    order: string
  }
}

export interface ContainerStats {
  total: number
  running: number
  stopped: number
  notCreated: number
  healthy: number
  unhealthy: number
  starting: number
  totalCpuPercentage: number
  avgCpuPercentage: number
  totalMemoryMb: number
  avgMemoryMb: number
  totalRestarts: number
  maxRestarts: number
  [category: string]: number
}

export interface Category {
  id: number
  name: string
  display_name?: string
  color?: string
  startup_order: number
  service_count: number
  runningCount?: number
  created_at?: string
  updated_at?: string
}

export interface Service {
  id: number
  name: string
  category_id: number
  image_name: string
  image_tag: string
  restart_policy: string
  command?: string
  user_spec?: string
  enabled: boolean
  created_at: string
  updated_at: string
  category?: Category
  ports?: ServicePort[]
  volumes?: ServiceVolume[]
  env_vars?: ServiceEnvVar[]
  dependencies?: string[]
  healthcheck?: ServiceHealthcheck
  labels?: ServiceLabel[]
  domains?: ServiceDomain[]
  status?: ContainerState
  metrics?: ContainerMetrics
}

export interface ServicePort {
  id: number
  service_id: number
  host_ip: string
  host_port: number
  container_port: number
  protocol: string
}

export interface ServiceVolume {
  id: number
  service_id: number
  volume_type: string
  source: string
  target: string
  read_only: boolean
  is_external: boolean
}

export interface ServiceEnvVar {
  id: number
  service_id: number
  key: string
  value?: string
  is_secret: boolean
}

export interface ServiceHealthcheck {
  id: number
  service_id: number
  test: string
  interval_seconds: number
  timeout_seconds: number
  retries: number
  start_period_seconds: number
}

export interface ServiceLabel {
  id: number
  service_id: number
  key: string
  value: string
}

export interface ServiceDomain {
  id: number
  service_id: number
  domain: string
  proxy_port?: number
}

export interface ServiceConfigFile {
  id: number
  service_id: number
  file_path: string
  content: string
  file_mode: string
  is_template: boolean
  created_at: string
  updated_at: string
}

export interface ContainerState {
  id?: number
  service_id?: number
  container_id?: string
  status: string
  health_status?: string
  restart_count: number
  started_at?: string
  finished_at?: string
  exit_code?: number
  error?: string
  updated_at?: string
}

export interface ContainerMetrics {
  id?: number
  service_id?: number
  cpu_percentage: number
  memory_used_mb: number
  memory_limit_mb: number
  memory_percentage: number
  network_rx_bytes: number
  network_tx_bytes: number
  recorded_at?: string
}

export interface StatusOverview {
  total_services: number
  enabled_services: number
  running_services: number
  stopped_services: number
  categories: CategoryOverview[]
  container_runtime: string
  socket_path?: string
}

export interface CategoryOverview {
  name: string
  total_services: number
  running_services: number
}

export interface ServiceLogsResponse {
  logs?: string
}

export interface WebSocketMessage {
  type: string
  timestamp?: string
  data?: Record<string, unknown>
}

export interface ControlResponse {
  success: boolean
  message?: string
  error?: string
  output?: string
}

export interface LogsResponse {
  success: boolean
  data?: {
    container: string
    logs: string
    lines: number
  }
  error?: string
}

export interface RuntimeInfo {
  installed: boolean
  version: string | null
  running: boolean
  responsive: boolean
}

export interface RuntimeStatus {
  current: string
  available: Record<string, RuntimeInfo>
  containers: Record<string, number>
  microservices: {
    running: number
    network: string
    network_exists: boolean
  }
}

export interface SocketInfo {
  active: boolean
  socket_path: string
  exists: boolean
  connectivity: string
}

export interface SocketStatus {
  active: string
  sockets: Record<string, SocketInfo>
  environment: {
    docker_host: string
    user: string
    uid: number
  }
  integration: {
    project_network: string
    network_exists: boolean
    running_services: number
  }
}

export interface ProjectService {
  id: number
  project_id: number
  service_name: string
  container_name?: string
  image?: string
  service_type?: string
  ports: string[]
  depends_on: string[]
}

export interface ProjectServiceStatus {
  name: string
  status: string
  type?: string
}

export interface Project {
  id: number
  name: string
  path: string
  project_type: string
  framework?: string
  language?: string
  package_manager?: string
  description?: string
  version?: string
  license?: string
  entry_point?: string
  has_frontend: boolean
  frontend_framework?: string
  domain?: string
  proxy_port?: number
  compose_path?: string
  service_count: number
  dependencies: Record<string, unknown>
  scripts: Record<string, string>
  git_remote?: string
  git_branch?: string
  last_scanned_at?: string
  created_at: string
  updated_at: string
}

export interface ProjectScanResponse {
  scanned: number
  projects: Project[]
}

export interface Stack {
  id: number
  name: string
  description: string
  network_name: string | null
  enabled: boolean
  instance_count: number
  running_count: number
  created_at: string
  updated_at: string
  deleted_at?: string | null
  instances?: StackInstance[]
}

export interface StackInstance {
  id: number
  stack_id: number
  instance_id: string
  template_service_id: number | null
  container_name: string | null
  enabled: boolean
  created_at: string
  updated_at: string
}

export interface DeletePreview {
  stack_name: string
  instance_count: number
  container_names: string[]
}
