import { useQuery } from '@tanstack/react-query'
import { api } from '@/lib/api'
import type { ContainersResponse, ControlResponse, LogsResponse } from '@/types/api'
import { useMutationHelper } from '@/lib/mutations'

export function useContainers(filter?: string, search?: string) {
  return useQuery({
    queryKey: ['containers', filter, search],
    queryFn: async () => {
      const params = new URLSearchParams()
      if (filter && filter !== 'all') params.set('filter', filter)
      if (search) params.set('search', search)
      const url = params.toString() ? `/containers?${params}` : '/containers'
      const response = await api.get<ContainersResponse>(url)
      return response.data
    },
    refetchInterval: 30000,
  })
}

export function useContainerLogs(name: string, lines: number = 100) {
  return useQuery({
    queryKey: ['containers', name, 'logs', lines],
    queryFn: async () => {
      const response = await api.get<LogsResponse>(`/containers/${name}/logs?lines=${lines}`)
      return response.data
    },
    enabled: !!name,
    refetchInterval: 30000,
  })
}

export function useContainerControl() {
  return useMutationHelper({
    mutationFn: async ({ container, action }: { container: string; action: 'start' | 'stop' | 'restart' | 'rebuild' }) => {
      const response = await api.post<ControlResponse>(`/containers/${container}/control`, { action })
      return response.data
    },
    successMessage: (vars, data) => {
      if (data.success) {
        return `${vars.action} ${vars.container}`
      } else {
        return data.error || `Failed to ${vars.action} ${vars.container}`
      }
    },
    errorMessage: (_error, vars) => `Failed to ${vars.action} ${vars.container}`,
    invalidate: [['containers']],
  })
}

export function useBulkControl() {
  return useMutationHelper({
    mutationFn: async ({ containers, action }: { containers: string[]; action: 'start' | 'stop' }) => {
      const response = await api.post<ControlResponse>('/bulk', { containers, action })
      return response.data
    },
    successMessage: (vars, data) => {
      if (data.success) {
        return `Bulk ${vars.action} completed`
      } else {
        return data.error || `Bulk ${vars.action} failed`
      }
    },
    errorMessage: (_error, vars) => `Bulk ${vars.action} failed`,
    invalidate: [['containers']],
  })
}
