import type {
  AdapterCapabilities,
  ApplyOperation,
  EnvValue,
  EventEnvelope,
  LogChunk,
  PortBinding,
  VolumeMount,
} from '../generated/api'

export function formatDateTime(value?: string) {
  if (!value) return '—'
  const parsed = new Date(value)
  return Number.isNaN(parsed.getTime()) ? value : parsed.toLocaleString()
}

export function formatEnvValue(value: EnvValue | undefined) {
  if (value === undefined) return '—'
  if (typeof value === 'string' || typeof value === 'number' || typeof value === 'boolean') {
    return String(value)
  }
  if ('secretRef' in value) {
    return `secretRef:${value.secretRef}`
  }
  return JSON.stringify(value)
}

export function formatPort(binding: PortBinding) {
  const host = binding.host ? `${binding.host}:` : ''
  const hostIP = binding.hostIP ? `${binding.hostIP}:` : ''
  const protocol = binding.protocol ? `/${binding.protocol}` : ''
  return `${hostIP}${host}${binding.container}${protocol}`
}

export function formatVolume(volume: VolumeMount) {
  const source = volume.source ? `${volume.source}:` : ''
  const mode = volume.readOnly ? ':ro' : ''
  return `${source}${volume.target}${mode}`
}

export function formatCapabilityList(capabilities?: AdapterCapabilities) {
  if (!capabilities) return []
  const available: string[] = []
  if (capabilities.inspect) available.push('inspect')
  if (capabilities.apply) available.push('apply')
  if (capabilities.logs) available.push('logs')
  if (capabilities.exec) available.push('exec')
  if (capabilities.network) available.push('network')
  return available
}

export function compactJson(value: unknown) {
  return JSON.stringify(value, null, 2)
}

export function eventSummary(event: EventEnvelope) {
  const payload = event.payload ?? {}
  switch (event.kind) {
    case 'apply.started':
      return `Apply started (${String(payload.totalActions ?? '0')} actions)`
    case 'apply.progress':
      return `${String(payload.action ?? 'action')} ${String(payload.target ?? event.resource ?? 'workspace')} → ${String(payload.status ?? 'unknown')}`
    case 'apply.completed':
      return `Apply completed (${payload.succeeded ? 'success' : 'failure'})`
    case 'logs.started':
      return `Logs requested for ${event.resource ?? 'resource'} (tail ${String(payload.tail ?? '0')})`
    case 'logs.chunk':
      return String(payload.line ?? 'log chunk received')
    case 'logs.completed':
      return `Logs completed for ${event.resource ?? 'resource'}`
    case 'exec.started':
      return `Exec started: ${Array.isArray(payload.command) ? payload.command.join(' ') : 'command'}`
    case 'exec.completed':
      return `Exec completed (exit ${String(payload.exitCode ?? '0')})`
    default:
      return event.kind
  }
}

export function summarizeOperation(operation: ApplyOperation) {
  return `${operation.kind} ${operation.target} → ${operation.status}`
}

export function formatLogChunk(chunk: LogChunk) {
  const stamp = chunk.timestamp ? `[${formatDateTime(chunk.timestamp)}] ` : ''
  const stream = chunk.stream ? `${chunk.stream}: ` : ''
  return `${stamp}${stream}${chunk.line}`
}
