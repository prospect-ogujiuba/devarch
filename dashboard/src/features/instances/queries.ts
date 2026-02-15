import { useQuery } from '@tanstack/react-query'
import { api, getErrorMessage } from '@/lib/api'
import type {
  Instance,
  InstanceDetail,
  EffectiveConfig,
  InstanceDeletePreview,
  InstancePort,
  InstanceVolume,
  InstanceEnvVar,
  InstanceLabel,
  InstanceDomain,
  InstanceHealthcheck,
  InstanceDependency,
  ResourceLimits,
  ResourceLimitsResponse,
  ServiceConfigMount,
} from '@/types/api'
import { useMutationHelper } from '@/lib/mutations'

export function useInstances(stackName: string) {
  return useQuery({
    queryKey: ['stacks', stackName, 'instances'],
    queryFn: async () => {
      const response = await api.get<Instance[]>(`/stacks/${stackName}/instances`)
      return Array.isArray(response.data) ? response.data : []
    },
    enabled: !!stackName,
    refetchInterval: 30000,
  })
}

export function useInstance(stackName: string, instanceId: string) {
  return useQuery({
    queryKey: ['stacks', stackName, 'instances', instanceId],
    queryFn: async () => {
      const response = await api.get<InstanceDetail>(`/stacks/${stackName}/instances/${instanceId}`)
      const data = response.data
      return {
        ...data,
        ports: data.ports ?? [],
        volumes: data.volumes ?? [],
        env_vars: data.env_vars ?? [],
        env_files: data.env_files ?? [],
        networks: data.networks ?? [],
        config_mounts: data.config_mounts ?? [],
        labels: data.labels ?? [],
        domains: data.domains ?? [],
        healthcheck: data.healthcheck ?? null,
        dependencies: data.dependencies ?? [],
        config_files: data.config_files ?? [],
      }
    },
    enabled: !!stackName && !!instanceId,
    refetchInterval: 30000,
  })
}

export function useEffectiveConfig(stackName: string, instanceId: string) {
  return useQuery({
    queryKey: ['stacks', stackName, 'instances', instanceId, 'effective-config'],
    queryFn: async () => {
      const response = await api.get<EffectiveConfig>(`/stacks/${stackName}/instances/${instanceId}/effective-config`)
      return response.data
    },
    enabled: !!stackName && !!instanceId,
  })
}

export function useInstanceDeletePreview(stackName: string, instanceId: string) {
  return useQuery({
    queryKey: ['stacks', stackName, 'instances', instanceId, 'delete-preview'],
    queryFn: async () => {
      const response = await api.get<InstanceDeletePreview>(`/stacks/${stackName}/instances/${instanceId}/delete-preview`)
      const data = response.data
      return {
        ...data,
        instance_id: data.instance_id || data.instance_name || instanceId,
      }
    },
    enabled: !!stackName && !!instanceId,
  })
}

interface CreateInstanceRequest {
  instance_id: string
  template_service_id: number
  description?: string
}

export function useCreateInstance(stackName: string) {
  return useMutationHelper({
    mutationFn: async (data: CreateInstanceRequest) => {
      const response = await api.post(`/stacks/${stackName}/instances`, data)
      return response.data
    },
    successMessage: 'Instance created',
    errorMessage: (error) => getErrorMessage(error, 'Failed to create instance'),
    invalidate: [
      ['stacks', stackName, 'instances'],
      ['stacks', stackName],
      ['stacks'],
    ],
  })
}

interface UpdateInstanceRequest {
  description?: string
  enabled?: boolean
}

export function useUpdateInstance(stackName: string, instanceId: string) {
  return useMutationHelper({
    mutationFn: async (data: UpdateInstanceRequest) => {
      const response = await api.put(`/stacks/${stackName}/instances/${instanceId}`, data)
      return response.data
    },
    successMessage: 'Instance updated',
    errorMessage: (error) => getErrorMessage(error, 'Failed to update instance'),
    invalidate: [
      ['stacks', stackName, 'instances', instanceId],
      ['stacks', stackName, 'instances'],
      ['stacks', stackName],
    ],
  })
}

export function useDeleteInstance(stackName: string, instanceId: string) {
  return useMutationHelper({
    mutationFn: async () => {
      const response = await api.delete(`/stacks/${stackName}/instances/${instanceId}`)
      return response.data
    },
    successMessage: 'Instance deleted',
    errorMessage: (error) => getErrorMessage(error, 'Failed to delete instance'),
    invalidate: [
      ['stacks', stackName, 'instances'],
      ['stacks', stackName],
      ['stacks'],
    ],
  })
}

export function useDuplicateInstance(stackName: string, instanceId: string) {
  return useMutationHelper({
    mutationFn: async (newInstanceId: string) => {
      const response = await api.post(`/stacks/${stackName}/instances/${instanceId}/duplicate`, {
        instance_id: newInstanceId,
      })
      return response.data
    },
    successMessage: 'Instance duplicated',
    errorMessage: (error) => getErrorMessage(error, 'Failed to duplicate instance'),
    invalidate: [
      ['stacks', stackName, 'instances'],
      ['stacks', stackName],
      ['stacks'],
    ],
  })
}

interface RenameInstanceRequest {
  instance_id: string
}

export function useRenameInstance(stackName: string, instanceId: string) {
  return useMutationHelper({
    mutationFn: async (data: RenameInstanceRequest) => {
      const response = await api.put(`/stacks/${stackName}/instances/${instanceId}/rename`, data)
      return response.data
    },
    successMessage: 'Instance renamed',
    errorMessage: (error) => getErrorMessage(error, 'Failed to rename instance'),
    invalidate: [
      ['stacks', stackName, 'instances'],
      ['stacks', stackName],
      ['stacks'],
    ],
  })
}

export function useStopInstance(stackName: string, instanceId: string) {
  return useMutationHelper({
    mutationFn: async () => {
      const response = await api.post(`/stacks/${stackName}/instances/${instanceId}/stop`)
      return response.data
    },
    successMessage: `Stopped ${instanceId}`,
    errorMessage: (error) => getErrorMessage(error, `Failed to stop ${instanceId}`),
    invalidate: [
      ['stacks', stackName],
      ['stacks'],
    ],
  })
}

export function useStartInstance(stackName: string, instanceId: string) {
  return useMutationHelper({
    mutationFn: async () => {
      const response = await api.post(`/stacks/${stackName}/instances/${instanceId}/start`)
      return response.data
    },
    successMessage: `Started ${instanceId}`,
    errorMessage: (error) => getErrorMessage(error, `Failed to start ${instanceId}`),
    invalidate: [
      ['stacks', stackName],
      ['stacks'],
    ],
  })
}

export function useRestartInstance(stackName: string, instanceId: string) {
  return useMutationHelper({
    mutationFn: async () => {
      const response = await api.post(`/stacks/${stackName}/instances/${instanceId}/restart`)
      return response.data
    },
    successMessage: `Restarted ${instanceId}`,
    errorMessage: (error) => getErrorMessage(error, `Failed to restart ${instanceId}`),
    invalidate: [
      ['stacks', stackName],
      ['stacks'],
    ],
  })
}

export function useUpdateInstancePorts(stackName: string, instanceId: string) {
  return useMutationHelper({
    mutationFn: async (ports: Omit<InstancePort, 'id' | 'instance_id'>[]) => {
      const response = await api.put(`/stacks/${stackName}/instances/${instanceId}/ports`, { ports })
      return response.data
    },
    successMessage: 'Ports updated',
    errorMessage: (error) => getErrorMessage(error, 'Failed to update ports'),
    invalidate: [
      ['stacks', stackName, 'instances', instanceId],
      ['stacks', stackName, 'instances', instanceId, 'effective-config'],
      ['stacks', stackName, 'instances'],
    ],
  })
}

export function useUpdateInstanceVolumes(stackName: string, instanceId: string) {
  return useMutationHelper({
    mutationFn: async (volumes: Omit<InstanceVolume, 'id' | 'instance_id'>[]) => {
      const response = await api.put(`/stacks/${stackName}/instances/${instanceId}/volumes`, { volumes })
      return response.data
    },
    successMessage: 'Volumes updated',
    errorMessage: (error) => getErrorMessage(error, 'Failed to update volumes'),
    invalidate: [
      ['stacks', stackName, 'instances', instanceId],
      ['stacks', stackName, 'instances', instanceId, 'effective-config'],
      ['stacks', stackName, 'instances'],
    ],
  })
}

export function useUpdateInstanceEnvVars(stackName: string, instanceId: string) {
  return useMutationHelper({
    mutationFn: async (env_vars: Omit<InstanceEnvVar, 'id' | 'instance_id'>[]) => {
      const response = await api.put(`/stacks/${stackName}/instances/${instanceId}/env-vars`, { env_vars })
      return response.data
    },
    successMessage: 'Environment variables updated',
    errorMessage: (error) => getErrorMessage(error, 'Failed to update environment variables'),
    invalidate: [
      ['stacks', stackName, 'instances', instanceId],
      ['stacks', stackName, 'instances', instanceId, 'effective-config'],
      ['stacks', stackName, 'instances'],
    ],
  })
}

export function useUpdateInstanceLabels(stackName: string, instanceId: string) {
  return useMutationHelper({
    mutationFn: async (labels: Omit<InstanceLabel, 'id' | 'instance_id'>[]) => {
      const response = await api.put(`/stacks/${stackName}/instances/${instanceId}/labels`, { labels })
      return response.data
    },
    successMessage: 'Labels updated',
    errorMessage: (error) => getErrorMessage(error, 'Failed to update labels'),
    invalidate: [
      ['stacks', stackName, 'instances', instanceId],
      ['stacks', stackName, 'instances', instanceId, 'effective-config'],
      ['stacks', stackName, 'instances'],
    ],
  })
}

export function useUpdateInstanceDomains(stackName: string, instanceId: string) {
  return useMutationHelper({
    mutationFn: async (domains: Omit<InstanceDomain, 'id' | 'instance_id'>[]) => {
      const response = await api.put(`/stacks/${stackName}/instances/${instanceId}/domains`, { domains })
      return response.data
    },
    successMessage: 'Domains updated',
    errorMessage: (error) => getErrorMessage(error, 'Failed to update domains'),
    invalidate: [
      ['stacks', stackName, 'instances', instanceId],
      ['stacks', stackName, 'instances', instanceId, 'effective-config'],
      ['stacks', stackName, 'instances'],
    ],
  })
}

export function useUpdateInstanceHealthcheck(stackName: string, instanceId: string) {
  return useMutationHelper({
    mutationFn: async (healthcheck: Omit<InstanceHealthcheck, 'id' | 'instance_id'> | null) => {
      const response = await api.put(`/stacks/${stackName}/instances/${instanceId}/healthcheck`, healthcheck)
      return response.data
    },
    successMessage: 'Healthcheck updated',
    errorMessage: (error) => getErrorMessage(error, 'Failed to update healthcheck'),
    invalidate: [
      ['stacks', stackName, 'instances', instanceId],
      ['stacks', stackName, 'instances', instanceId, 'effective-config'],
      ['stacks', stackName, 'instances'],
    ],
  })
}

export function useUpdateInstanceDependencies(stackName: string, instanceId: string) {
  return useMutationHelper({
    mutationFn: async (dependencies: Omit<InstanceDependency, 'id' | 'instance_id'>[]) => {
      const response = await api.put(`/stacks/${stackName}/instances/${instanceId}/dependencies`, { dependencies })
      return response.data
    },
    successMessage: 'Dependencies updated',
    errorMessage: (error) => getErrorMessage(error, 'Failed to update dependencies'),
    invalidate: [
      ['stacks', stackName, 'instances', instanceId],
      ['stacks', stackName, 'instances', instanceId, 'effective-config'],
      ['stacks', stackName, 'instances'],
    ],
  })
}

interface SaveConfigFileRequest {
  content: string
  file_mode: string
}

export function useSaveConfigFile(stackName: string, instanceId: string) {
  return useMutationHelper({
    mutationFn: async ({ filePath, data }: { filePath: string; data: SaveConfigFileRequest }) => {
      const response = await api.put(`/stacks/${stackName}/instances/${instanceId}/files/${filePath}`, data)
      return response.data
    },
    errorMessage: (error) => getErrorMessage(error, 'Failed to save file'),
    invalidate: [
      ['stacks', stackName, 'instances', instanceId],
      ['stacks', stackName, 'instances', instanceId, 'effective-config'],
      ['stacks', stackName, 'instances'],
    ],
  })
}

export function useDeleteConfigFile(stackName: string, instanceId: string) {
  return useMutationHelper({
    mutationFn: async (filePath: string) => {
      const response = await api.delete(`/stacks/${stackName}/instances/${instanceId}/files/${filePath}`)
      return response.data
    },
    errorMessage: (error) => getErrorMessage(error, 'Failed to delete file'),
    invalidate: [
      ['stacks', stackName, 'instances', instanceId],
      ['stacks', stackName, 'instances', instanceId, 'effective-config'],
      ['stacks', stackName, 'instances'],
    ],
  })
}

export function useResourceLimits(stackName: string, instanceId: string) {
  return useQuery({
    queryKey: ['stacks', stackName, 'instances', instanceId, 'resources'],
    queryFn: async () => {
      try {
        const response = await api.get<ResourceLimits>(`/stacks/${stackName}/instances/${instanceId}/resources`)
        return response.data
      } catch (error: unknown) {
        if (error && typeof error === 'object' && 'response' in error) {
          const axiosError = error as { response?: { status?: number } }
          if (axiosError.response?.status === 404) {
            return null
          }
        }
        throw error
      }
    },
    enabled: !!stackName && !!instanceId,
  })
}

export function useUpdateResourceLimits(stackName: string, instanceId: string) {
  return useMutationHelper({
    mutationFn: async (limits: ResourceLimits) => {
      const response = await api.put<ResourceLimitsResponse>(`/stacks/${stackName}/instances/${instanceId}/resources`, limits)
      return response.data
    },
    successMessage: 'Resource limits updated',
    errorMessage: (error) => getErrorMessage(error, 'Failed to update resource limits'),
    invalidate: [
      ['stacks', stackName, 'instances', instanceId, 'resources'],
      ['stacks', stackName, 'instances', instanceId],
      ['stacks', stackName, 'instances', instanceId, 'effective-config'],
    ],
  })
}

export function useUpdateInstanceEnvFiles(stackName: string, instanceId: string) {
  return useMutationHelper({
    mutationFn: async (env_files: string[]) => {
      const response = await api.put(`/stacks/${stackName}/instances/${instanceId}/env-files`, { env_files })
      return response.data
    },
    successMessage: 'Env files updated',
    errorMessage: (error) => getErrorMessage(error, 'Failed to update env files'),
    invalidate: [
      ['stacks', stackName, 'instances', instanceId],
      ['stacks', stackName, 'instances', instanceId, 'effective-config'],
      ['stacks', stackName, 'instances'],
    ],
  })
}

export function useUpdateInstanceNetworks(stackName: string, instanceId: string) {
  return useMutationHelper({
    mutationFn: async (networks: string[]) => {
      const response = await api.put(`/stacks/${stackName}/instances/${instanceId}/networks`, { networks })
      return response.data
    },
    successMessage: 'Networks updated',
    errorMessage: (error) => getErrorMessage(error, 'Failed to update networks'),
    invalidate: [
      ['stacks', stackName, 'instances', instanceId],
      ['stacks', stackName, 'instances', instanceId, 'effective-config'],
      ['stacks', stackName, 'instances'],
    ],
  })
}

export function useUpdateInstanceConfigMounts(stackName: string, instanceId: string) {
  return useMutationHelper({
    mutationFn: async (config_mounts: Omit<ServiceConfigMount, 'id' | 'service_id'>[]) => {
      const response = await api.put(`/stacks/${stackName}/instances/${instanceId}/config-mounts`, { config_mounts })
      return response.data
    },
    successMessage: 'Config mounts updated',
    errorMessage: (error) => getErrorMessage(error, 'Failed to update config mounts'),
    invalidate: [
      ['stacks', stackName, 'instances', instanceId],
      ['stacks', stackName, 'instances', instanceId, 'effective-config'],
      ['stacks', stackName, 'instances'],
    ],
  })
}

export function useInstanceLogs(stackName: string, instanceId: string, tail: number = 100) {
  return useQuery({
    queryKey: ['stacks', stackName, 'instances', instanceId, 'logs', tail],
    queryFn: async () => {
      const response = await api.get(`/stacks/${stackName}/instances/${instanceId}/logs?tail=${tail}`, {
        responseType: 'text',
      })
      return response.data as string
    },
    enabled: !!stackName && !!instanceId,
    refetchInterval: 30000,
  })
}

export function useInstanceCompose(stackName: string, instanceId: string) {
  return useQuery({
    queryKey: ['stacks', stackName, 'instances', instanceId, 'compose'],
    queryFn: async () => {
      const response = await api.get(`/stacks/${stackName}/instances/${instanceId}/compose`, {
        headers: { Accept: 'text/yaml' },
        responseType: 'text',
      })
      return response.data as string
    },
    enabled: !!stackName && !!instanceId,
  })
}
