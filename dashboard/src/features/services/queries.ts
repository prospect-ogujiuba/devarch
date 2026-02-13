import { useQuery, useQueryClient } from '@tanstack/react-query'
import { api, getErrorMessage } from '@/lib/api'
import type { Service, ServicePort, ServiceVolume, ServiceEnvVar, ServiceHealthcheck, ServiceLabel, ServiceDomain, ServiceConfigFile, ServiceConfigMount } from '@/types/api'
import { useMutationHelper } from '@/lib/mutations'

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
    refetchInterval: 30000,
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
    refetchInterval: 30000,
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
    refetchInterval: 30000,
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
    refetchInterval: 30000,
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
  return useMutationHelper({
    mutationFn: async (name: string) => {
      const response = await api.post(`/services/${name}/start`)
      return response.data
    },
    successMessage: (vars) => `Started ${vars}`,
    errorMessage: (_error, vars) => `Failed to start ${vars}`,
    invalidate: [['services'], ['status']],
  })
}

export function useStopService() {
  return useMutationHelper({
    mutationFn: async (name: string) => {
      const response = await api.post(`/services/${name}/stop`)
      return response.data
    },
    successMessage: (vars) => `Stopped ${vars}`,
    errorMessage: (_error, vars) => `Failed to stop ${vars}`,
    invalidate: [['services'], ['status']],
  })
}

export function useRestartService() {
  return useMutationHelper({
    mutationFn: async (name: string) => {
      const response = await api.post(`/services/${name}/restart`)
      return response.data
    },
    successMessage: (vars) => `Restarted ${vars}`,
    errorMessage: (_error, vars) => `Failed to restart ${vars}`,
    invalidate: [['services'], ['status']],
  })
}

export function useBulkServiceControl() {
  return useMutationHelper({
    mutationFn: async ({ names, action }: { names: string[]; action: 'start' | 'stop' | 'restart' }) => {
      const results = await Promise.allSettled(
        names.map((name) => api.post(`/services/${name}/${action}`)),
      )
      const failed = results.filter((r) => r.status === 'rejected').length
      return { total: names.length, failed }
    },
    successMessage: (vars, data) => {
      if (data.failed === 0) {
        return `Bulk ${vars.action}: ${data.total} services`
      } else {
        return `Bulk ${vars.action}: ${data.total - data.failed}/${data.total} succeeded`
      }
    },
    errorMessage: (_error, vars) => `Bulk ${vars.action} failed`,
    invalidate: [['services'], ['status']],
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
  labels?: Omit<ServiceLabel, 'id' | 'service_id'>[]
  domains?: Omit<ServiceDomain, 'id' | 'service_id'>[]
  healthcheck?: Omit<ServiceHealthcheck, 'id' | 'service_id'> | null
}

export function useCreateService() {
  return useMutationHelper({
    mutationFn: async (data: CreateServicePayload) => {
      const response = await api.post('/services', data)
      return response.data
    },
    successMessage: 'Service created',
    errorMessage: (error) => getErrorMessage(error, 'Failed to create service'),
    invalidate: [['services']],
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
  return useMutationHelper({
    mutationFn: async ({ name, data }: { name: string; data: UpdateServicePayload }) => {
      const response = await api.put(`/services/${name}`, data)
      return response.data
    },
    successMessage: 'Service updated',
    errorMessage: (error) => getErrorMessage(error, 'Failed to update service'),
    invalidate: [['services']],
    onSuccess: (_data, { name }) => {
      queryClient.invalidateQueries({ queryKey: ['services', name] })
    },
  })
}

export function useDeleteService() {
  return useMutationHelper({
    mutationFn: async (name: string) => {
      const response = await api.delete(`/services/${name}`)
      return response.data
    },
    successMessage: 'Service deleted',
    errorMessage: (error) => getErrorMessage(error, 'Failed to delete service'),
    invalidate: [['services']],
  })
}

export function useImportLibrary() {
  return useMutationHelper({
    mutationFn: async () => {
      const response = await api.post('/services/import-library')
      return response.data
    },
    successMessage: 'Service library imported',
    errorMessage: (error) => getErrorMessage(error, 'Failed to import service library'),
    invalidate: [['services']],
  })
}

function makeSubResourceMutation<T>(resource: string, label: string) {
  return function useSubResource() {
    const queryClient = useQueryClient()
    return useMutationHelper({
      mutationFn: async ({ name, data }: { name: string; data: T }) => {
        const response = await api.put(`/services/${name}/${resource}`, data)
        return response.data
      },
      successMessage: `${label} updated`,
      errorMessage: (error) => getErrorMessage(error, `Failed to update ${label.toLowerCase()}`),
      invalidate: [],
      onSuccess: (_data, { name }) => {
        queryClient.invalidateQueries({ queryKey: ['services', name] })
      },
    })
  }
}

export const useUpdatePorts = makeSubResourceMutation<{ ports: Omit<ServicePort, 'id' | 'service_id'>[] }>('ports', 'Ports')
export const useUpdateVolumes = makeSubResourceMutation<{ volumes: Omit<ServiceVolume, 'id' | 'service_id'>[] }>('volumes', 'Volumes')
export const useUpdateEnvVars = makeSubResourceMutation<{ env_vars: Omit<ServiceEnvVar, 'id' | 'service_id'>[] }>('env-vars', 'Environment variables')
export const useUpdateDependencies = makeSubResourceMutation<{ dependencies: string[] }>('dependencies', 'Dependencies')
export const useUpdateEnvFiles = makeSubResourceMutation<{ env_files: string[] }>('env-files', 'Env files')
export const useUpdateNetworks = makeSubResourceMutation<{ networks: string[] }>('networks', 'Networks')
export const useUpdateConfigMounts = makeSubResourceMutation<{ config_mounts: Omit<ServiceConfigMount, 'id' | 'service_id'>[] }>('config-mounts', 'Config mounts')
export const useUpdateHealthcheck = makeSubResourceMutation<Omit<ServiceHealthcheck, 'id' | 'service_id'> | null>('healthcheck', 'Healthcheck')
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
  return useMutationHelper({
    mutationFn: async ({ name, filePath, content, fileMode }: { name: string; filePath: string; content: string; fileMode?: string }) => {
      const response = await api.put(`/services/${name}/files/${filePath}`, { content, file_mode: fileMode || '0644' })
      return response.data
    },
    successMessage: 'File saved',
    errorMessage: (error) => getErrorMessage(error, 'Failed to save file'),
    invalidate: [],
    onSuccess: (_data, { name }) => {
      queryClient.invalidateQueries({ queryKey: ['services', name, 'files'] })
    },
  })
}

export function useDeleteConfigFile() {
  const queryClient = useQueryClient()
  return useMutationHelper({
    mutationFn: async ({ name, filePath }: { name: string; filePath: string }) => {
      const response = await api.delete(`/services/${name}/files/${filePath}`)
      return response.data
    },
    successMessage: 'File deleted',
    errorMessage: (error) => getErrorMessage(error, 'Failed to delete file'),
    invalidate: [],
    onSuccess: (_data, { name }) => {
      queryClient.invalidateQueries({ queryKey: ['services', name, 'files'] })
    },
  })
}
