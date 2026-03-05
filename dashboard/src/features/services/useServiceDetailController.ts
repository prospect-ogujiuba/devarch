import { useService, useServiceCompose, useDeleteService, useUpdateService, useStartService, useStopService, useRestartService } from './queries'
import { useGenerateServiceProxyConfig } from '@/features/proxy/queries'
import { computeUptime } from '@/lib/format'

export function useServiceDetailController(serviceName: string) {
  // Query orchestration
  const { data: service, isLoading } = useService(serviceName)
  const { data: composeYaml, isLoading: composeLoading } = useServiceCompose(serviceName)
  const deleteService = useDeleteService()
  const updateService = useUpdateService()
  const startService = useStartService()
  const stopService = useStopService()
  const restartService = useRestartService()
  const generateProxyConfig = useGenerateServiceProxyConfig(serviceName)

  // Derived state
  const status = service?.status?.status ?? 'stopped'
  const image = service ? `${service.image_name}:${service.image_tag}` : ''
  const healthStatus = service?.status?.health_status ?? (service?.healthcheck ? 'configured' : 'none')
  const uptime = computeUptime(service?.status?.started_at)
  const cpuPct = service?.metrics?.cpu_percentage ?? 0
  const memUsed = service?.metrics?.memory_used_mb ?? 0
  const memLimit = service?.metrics?.memory_limit_mb ?? 0
  const rxBytes = service?.metrics?.network_rx_bytes ?? 0
  const txBytes = service?.metrics?.network_tx_bytes ?? 0

  return {
    // Data
    service,
    composeYaml,
    isLoading,
    composeLoading,

    // Derived state
    status,
    image,
    healthStatus,
    uptime,
    cpuPct,
    memUsed,
    memLimit,
    rxBytes,
    txBytes,

    // Mutations
    deleteService,
    updateService,
    startService,
    stopService,
    restartService,
    generateProxyConfig,
  }
}
