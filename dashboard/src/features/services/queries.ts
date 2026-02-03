import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '@/lib/api'
import type { Service, ServicePort, ServiceVolume, ServiceEnvVar, ServiceHealthcheck, ServiceLabel, ServiceDomain, ServiceConfigFile } from '@/types/api'
import { toast } from 'sonner'

interface ServicesResult {
  services: Service[]
  total: number
}

export function useServices() {
  return useQuery({
    queryKey: ['services'],
    queryFn: async (): Promise<ServicesResult> => {
      const response = await api.get<Service[]>('/services?include=status,metrics&limit=500')
      const services = Array.isArray(response.data) ? response.data : []
      return {
        services,
        total: parseInt(response.headers['x-total-count'] ?? '0', 10) || services.length,
      }
    },
    refetchInterval: 5000,
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
    refetchInterval: 5000,
  })
}

export function useServiceMetrics(name: string) {
  return useQuery({
    queryKey: ['services', name, 'metrics'],
    queryFn: async () => {
      const response = await api.get<Service>(`/services/${name}?include=metrics`)
      return response.data?.metrics ?? null
    },
    enabled: !!name,
    refetchInterval: 3000,
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

export function useBulkServiceControl() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async ({ names, action }: { names: string[]; action: 'start' | 'stop' | 'restart' }) => {
      const results = await Promise.allSettled(
        names.map((name) => api.post(`/services/${name}/${action}`)),
      )
      const failed = results.filter((r) => r.status === 'rejected').length
      return { total: names.length, failed }
    },
    onSuccess: ({ total, failed }, { action }) => {
      if (failed === 0) {
        toast.success(`Bulk ${action}: ${total} services`)
      } else {
        toast.warning(`Bulk ${action}: ${total - failed}/${total} succeeded`)
      }
      queryClient.invalidateQueries({ queryKey: ['services'] })
      queryClient.invalidateQueries({ queryKey: ['status'] })
    },
    onError: (_error, { action }) => {
      toast.error(`Bulk ${action} failed`)
    },
  })
}

interface CreateServicePayload {
  name: string
  category_id: number
  image_name: string
  image_tag: string
  restart_policy: string
  command?: string
  user_spec?: string
  ports?: Omit<ServicePort, 'id' | 'service_id'>[]
  volumes?: Omit<ServiceVolume, 'id' | 'service_id'>[]
  env_vars?: Omit<ServiceEnvVar, 'id' | 'service_id'>[]
  dependencies?: string[]
}

export function useCreateService() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (data: CreateServicePayload) => {
      const response = await api.post('/services', data)
      return response.data
    },
    onSuccess: () => {
      toast.success('Service created')
      queryClient.invalidateQueries({ queryKey: ['services'] })
    },
    onError: (error: any) => {
      toast.error(error.response?.data || 'Failed to create service')
    },
  })
}

interface UpdateServicePayload {
  image_name?: string
  image_tag?: string
  restart_policy?: string
  command?: string
  user_spec?: string
}

export function useUpdateService() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async ({ name, data }: { name: string; data: UpdateServicePayload }) => {
      const response = await api.put(`/services/${name}`, data)
      return response.data
    },
    onSuccess: (_data, { name }) => {
      toast.success('Service updated')
      queryClient.invalidateQueries({ queryKey: ['services', name] })
      queryClient.invalidateQueries({ queryKey: ['services'] })
    },
    onError: (error: any) => {
      toast.error(error.response?.data || 'Failed to update service')
    },
  })
}

export function useDeleteService() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (name: string) => {
      const response = await api.delete(`/services/${name}`)
      return response.data
    },
    onSuccess: () => {
      toast.success('Service deleted')
      queryClient.invalidateQueries({ queryKey: ['services'] })
    },
    onError: (error: any) => {
      toast.error(error.response?.data || 'Failed to delete service')
    },
  })
}

function makeSubResourceMutation<T>(resource: string, label: string) {
  return function useSubResource() {
    const queryClient = useQueryClient()
    return useMutation({
      mutationFn: async ({ name, data }: { name: string; data: T }) => {
        const response = await api.put(`/services/${name}/${resource}`, data)
        return response.data
      },
      onSuccess: (_data, { name }) => {
        toast.success(`${label} updated`)
        queryClient.invalidateQueries({ queryKey: ['services', name] })
      },
      onError: (error: any) => {
        toast.error(error.response?.data || `Failed to update ${label.toLowerCase()}`)
      },
    })
  }
}

export const useUpdatePorts = makeSubResourceMutation<{ ports: Omit<ServicePort, 'id' | 'service_id'>[] }>('ports', 'Ports')
export const useUpdateVolumes = makeSubResourceMutation<{ volumes: Omit<ServiceVolume, 'id' | 'service_id'>[] }>('volumes', 'Volumes')
export const useUpdateEnvVars = makeSubResourceMutation<{ env_vars: Omit<ServiceEnvVar, 'id' | 'service_id'>[] }>('env-vars', 'Environment variables')
export const useUpdateDependencies = makeSubResourceMutation<{ dependencies: string[] }>('dependencies', 'Dependencies')
export const useUpdateHealthcheck = makeSubResourceMutation<ServiceHealthcheck | null>('healthcheck', 'Healthcheck')
export const useUpdateLabels = makeSubResourceMutation<{ labels: Omit<ServiceLabel, 'id' | 'service_id'>[] }>('labels', 'Labels')
export const useUpdateDomains = makeSubResourceMutation<{ domains: Omit<ServiceDomain, 'id' | 'service_id'>[] }>('domains', 'Domains')

export function useServiceConfigFiles(name: string) {
  return useQuery({
    queryKey: ['services', name, 'files'],
    queryFn: async () => {
      const response = await api.get<ServiceConfigFile[]>(`/services/${name}/files`)
      return Array.isArray(response.data) ? response.data : []
    },
    enabled: !!name,
  })
}

export function useServiceConfigFile(name: string, filePath: string) {
  return useQuery({
    queryKey: ['services', name, 'files', filePath],
    queryFn: async () => {
      const response = await api.get<ServiceConfigFile>(`/services/${name}/files/${filePath}`)
      return response.data
    },
    enabled: !!name && !!filePath,
  })
}

export function useSaveConfigFile() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async ({ name, filePath, content, fileMode }: { name: string; filePath: string; content: string; fileMode?: string }) => {
      const response = await api.put(`/services/${name}/files/${filePath}`, { content, file_mode: fileMode || '0644' })
      return response.data
    },
    onSuccess: (_data, { name }) => {
      toast.success('File saved')
      queryClient.invalidateQueries({ queryKey: ['services', name, 'files'] })
    },
    onError: (error: any) => {
      toast.error(error.response?.data || 'Failed to save file')
    },
  })
}

export function useDeleteConfigFile() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async ({ name, filePath }: { name: string; filePath: string }) => {
      const response = await api.delete(`/services/${name}/files/${filePath}`)
      return response.data
    },
    onSuccess: (_data, { name }) => {
      toast.success('File deleted')
      queryClient.invalidateQueries({ queryKey: ['services', name, 'files'] })
    },
    onError: (error: any) => {
      toast.error(error.response?.data || 'Failed to delete file')
    },
  })
}
