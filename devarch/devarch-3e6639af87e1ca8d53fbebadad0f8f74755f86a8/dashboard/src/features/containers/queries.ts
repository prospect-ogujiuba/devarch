import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '@/lib/api'
import type { ContainersResponse, ControlResponse, LogsResponse } from '@/types/api'
import { toast } from 'sonner'

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
    refetchInterval: 5000,
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
    refetchInterval: 3000,
  })
}

export function useContainerControl() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async ({ container, action }: { container: string; action: 'start' | 'stop' | 'restart' | 'rebuild' }) => {
      const response = await api.post<ControlResponse>(`/containers/${container}/control`, { action })
      return response.data
    },
    onSuccess: (data, { container, action }) => {
      if (data.success) {
        toast.success(`${action} ${container}`)
      } else {
        toast.error(data.error || `Failed to ${action} ${container}`)
      }
      queryClient.invalidateQueries({ queryKey: ['containers'] })
    },
    onError: (_error, { container, action }) => {
      toast.error(`Failed to ${action} ${container}`)
    },
  })
}

export function useBulkControl() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async ({ containers, action }: { containers: string[]; action: 'start' | 'stop' }) => {
      const response = await api.post<ControlResponse>('/bulk', { containers, action })
      return response.data
    },
    onSuccess: (data, { action }) => {
      if (data.success) {
        toast.success(`Bulk ${action} completed`)
      } else {
        toast.error(data.error || `Bulk ${action} failed`)
      }
      queryClient.invalidateQueries({ queryKey: ['containers'] })
    },
    onError: (_error, { action }) => {
      toast.error(`Bulk ${action} failed`)
    },
  })
}
