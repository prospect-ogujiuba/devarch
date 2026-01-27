import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '@/lib/api'
import type { Service, ServiceLogsResponse } from '@/types/api'
import { toast } from 'sonner'

interface ServicesResult {
  services: Service[]
  total: number
}

export function useServices() {
  return useQuery({
    queryKey: ['services'],
    queryFn: async (): Promise<ServicesResult> => {
      const response = await api.get<Service[]>('/services?include=status&limit=500')
      return {
        services: response.data,
        total: parseInt(response.headers['x-total-count'] ?? '0', 10) || response.data.length,
      }
    },
  })
}

export function useService(name: string) {
  return useQuery({
    queryKey: ['services', name],
    queryFn: async () => {
      const response = await api.get<Service>(`/services/${name}?include=all`)
      return response.data
    },
    enabled: !!name,
  })
}

export function useServiceLogs(name: string, tail: number = 100) {
  return useQuery({
    queryKey: ['services', name, 'logs', tail],
    queryFn: async () => {
      const response = await api.get(`/services/${name}/logs?tail=${tail}`, {
        responseType: 'text',
      })
      return response.data as string
    },
    enabled: !!name,
    refetchInterval: 5000,
  })
}

export function useServiceCompose(name: string) {
  return useQuery({
    queryKey: ['services', name, 'compose'],
    queryFn: async () => {
      const response = await api.get(`/services/${name}/compose`, {
        headers: { Accept: 'text/yaml' },
        responseType: 'text',
      })
      return response.data as string
    },
    enabled: !!name,
  })
}

export function useStartService() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (name: string) => {
      const response = await api.post(`/services/${name}/start`)
      return response.data
    },
    onSuccess: (_data, name) => {
      toast.success(`Started ${name}`)
      queryClient.invalidateQueries({ queryKey: ['services'] })
      queryClient.invalidateQueries({ queryKey: ['status'] })
    },
    onError: (_error, name) => {
      toast.error(`Failed to start ${name}`)
    },
  })
}

export function useStopService() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (name: string) => {
      const response = await api.post(`/services/${name}/stop`)
      return response.data
    },
    onSuccess: (_data, name) => {
      toast.success(`Stopped ${name}`)
      queryClient.invalidateQueries({ queryKey: ['services'] })
      queryClient.invalidateQueries({ queryKey: ['status'] })
    },
    onError: (_error, name) => {
      toast.error(`Failed to stop ${name}`)
    },
  })
}

export function useRestartService() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (name: string) => {
      const response = await api.post(`/services/${name}/restart`)
      return response.data
    },
    onSuccess: (_data, name) => {
      toast.success(`Restarted ${name}`)
      queryClient.invalidateQueries({ queryKey: ['services'] })
      queryClient.invalidateQueries({ queryKey: ['status'] })
    },
    onError: (_error, name) => {
      toast.error(`Failed to restart ${name}`)
    },
  })
}
