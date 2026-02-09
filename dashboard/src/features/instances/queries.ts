import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
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
import { toast } from 'sonner'

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
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (data: CreateInstanceRequest) => {
      const response = await api.post(`/stacks/${stackName}/instances`, data)
      return response.data
    },
    onSuccess: () => {
      toast.success('Instance created')
      queryClient.invalidateQueries({ queryKey: ['stacks', stackName, 'instances'] })
      queryClient.invalidateQueries({ queryKey: ['stacks', stackName] })
      queryClient.invalidateQueries({ queryKey: ['stacks'] })
    },
    onError: (error) => {
      toast.error(getErrorMessage(error, 'Failed to create instance'))
    },
  })
}

interface UpdateInstanceRequest {
  description?: string
  enabled?: boolean
}

export function useUpdateInstance(stackName: string, instanceId: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (data: UpdateInstanceRequest) => {
      const response = await api.put(`/stacks/${stackName}/instances/${instanceId}`, data)
      return response.data
    },
    onSuccess: () => {
      toast.success('Instance updated')
      queryClient.invalidateQueries({ queryKey: ['stacks', stackName, 'instances', instanceId] })
      queryClient.invalidateQueries({ queryKey: ['stacks', stackName, 'instances'] })
      queryClient.invalidateQueries({ queryKey: ['stacks', stackName] })
    },
    onError: (error) => {
      toast.error(getErrorMessage(error, 'Failed to update instance'))
    },
  })
}

export function useDeleteInstance(stackName: string, instanceId: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async () => {
      const response = await api.delete(`/stacks/${stackName}/instances/${instanceId}`)
      return response.data
    },
    onSuccess: () => {
      toast.success('Instance deleted')
      queryClient.invalidateQueries({ queryKey: ['stacks', stackName, 'instances'] })
      queryClient.invalidateQueries({ queryKey: ['stacks', stackName] })
      queryClient.invalidateQueries({ queryKey: ['stacks'] })
    },
    onError: (error) => {
      toast.error(getErrorMessage(error, 'Failed to delete instance'))
    },
  })
}

export function useDuplicateInstance(stackName: string, instanceId: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (newInstanceId: string) => {
      const response = await api.post(`/stacks/${stackName}/instances/${instanceId}/duplicate`, {
        instance_id: newInstanceId,
      })
      return response.data
    },
    onSuccess: () => {
      toast.success('Instance duplicated')
      queryClient.invalidateQueries({ queryKey: ['stacks', stackName, 'instances'] })
      queryClient.invalidateQueries({ queryKey: ['stacks', stackName] })
      queryClient.invalidateQueries({ queryKey: ['stacks'] })
    },
    onError: (error) => {
      toast.error(getErrorMessage(error, 'Failed to duplicate instance'))
    },
  })
}

interface RenameInstanceRequest {
  instance_id: string
}

export function useRenameInstance(stackName: string, instanceId: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (data: RenameInstanceRequest) => {
      const response = await api.put(`/stacks/${stackName}/instances/${instanceId}/rename`, data)
      return response.data
    },
    onSuccess: () => {
      toast.success('Instance renamed')
      queryClient.invalidateQueries({ queryKey: ['stacks', stackName, 'instances'] })
      queryClient.invalidateQueries({ queryKey: ['stacks', stackName] })
      queryClient.invalidateQueries({ queryKey: ['stacks'] })
    },
    onError: (error) => {
      toast.error(getErrorMessage(error, 'Failed to rename instance'))
    },
  })
}

export function useStopInstance(stackName: string, instanceId: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async () => {
      const response = await api.post(`/stacks/${stackName}/instances/${instanceId}/stop`)
      return response.data
    },
    onSuccess: () => {
      toast.success(`Stopped ${instanceId}`)
      queryClient.invalidateQueries({ queryKey: ['stacks', stackName] })
      queryClient.invalidateQueries({ queryKey: ['stacks'] })
    },
    onError: (error) => {
      toast.error(getErrorMessage(error, `Failed to stop ${instanceId}`))
    },
  })
}

export function useStartInstance(stackName: string, instanceId: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async () => {
      const response = await api.post(`/stacks/${stackName}/instances/${instanceId}/start`)
      return response.data
    },
    onSuccess: () => {
      toast.success(`Started ${instanceId}`)
      queryClient.invalidateQueries({ queryKey: ['stacks', stackName] })
      queryClient.invalidateQueries({ queryKey: ['stacks'] })
    },
    onError: (error) => {
      toast.error(getErrorMessage(error, `Failed to start ${instanceId}`))
    },
  })
}

export function useRestartInstance(stackName: string, instanceId: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async () => {
      const response = await api.post(`/stacks/${stackName}/instances/${instanceId}/restart`)
      return response.data
    },
    onSuccess: () => {
      toast.success(`Restarted ${instanceId}`)
      queryClient.invalidateQueries({ queryKey: ['stacks', stackName] })
      queryClient.invalidateQueries({ queryKey: ['stacks'] })
    },
    onError: (error) => {
      toast.error(getErrorMessage(error, `Failed to restart ${instanceId}`))
    },
  })
}

export function useUpdateInstancePorts(stackName: string, instanceId: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (ports: Omit<InstancePort, 'id' | 'instance_id'>[]) => {
      const response = await api.put(`/stacks/${stackName}/instances/${instanceId}/ports`, { ports })
      return response.data
    },
    onSuccess: () => {
      toast.success('Ports updated')
      queryClient.invalidateQueries({ queryKey: ['stacks', stackName, 'instances', instanceId] })
      queryClient.invalidateQueries({ queryKey: ['stacks', stackName, 'instances', instanceId, 'effective-config'] })
      queryClient.invalidateQueries({ queryKey: ['stacks', stackName, 'instances'] })
    },
    onError: (error) => {
      toast.error(getErrorMessage(error, 'Failed to update ports'))
    },
  })
}

export function useUpdateInstanceVolumes(stackName: string, instanceId: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (volumes: Omit<InstanceVolume, 'id' | 'instance_id'>[]) => {
      const response = await api.put(`/stacks/${stackName}/instances/${instanceId}/volumes`, { volumes })
      return response.data
    },
    onSuccess: () => {
      toast.success('Volumes updated')
      queryClient.invalidateQueries({ queryKey: ['stacks', stackName, 'instances', instanceId] })
      queryClient.invalidateQueries({ queryKey: ['stacks', stackName, 'instances', instanceId, 'effective-config'] })
      queryClient.invalidateQueries({ queryKey: ['stacks', stackName, 'instances'] })
    },
    onError: (error) => {
      toast.error(getErrorMessage(error, 'Failed to update volumes'))
    },
  })
}

export function useUpdateInstanceEnvVars(stackName: string, instanceId: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (env_vars: Omit<InstanceEnvVar, 'id' | 'instance_id'>[]) => {
      const response = await api.put(`/stacks/${stackName}/instances/${instanceId}/env-vars`, { env_vars })
      return response.data
    },
    onSuccess: () => {
      toast.success('Environment variables updated')
      queryClient.invalidateQueries({ queryKey: ['stacks', stackName, 'instances', instanceId] })
      queryClient.invalidateQueries({ queryKey: ['stacks', stackName, 'instances', instanceId, 'effective-config'] })
      queryClient.invalidateQueries({ queryKey: ['stacks', stackName, 'instances'] })
    },
    onError: (error) => {
      toast.error(getErrorMessage(error, 'Failed to update environment variables'))
    },
  })
}

export function useUpdateInstanceLabels(stackName: string, instanceId: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (labels: Omit<InstanceLabel, 'id' | 'instance_id'>[]) => {
      const response = await api.put(`/stacks/${stackName}/instances/${instanceId}/labels`, { labels })
      return response.data
    },
    onSuccess: () => {
      toast.success('Labels updated')
      queryClient.invalidateQueries({ queryKey: ['stacks', stackName, 'instances', instanceId] })
      queryClient.invalidateQueries({ queryKey: ['stacks', stackName, 'instances', instanceId, 'effective-config'] })
      queryClient.invalidateQueries({ queryKey: ['stacks', stackName, 'instances'] })
    },
    onError: (error) => {
      toast.error(getErrorMessage(error, 'Failed to update labels'))
    },
  })
}

export function useUpdateInstanceDomains(stackName: string, instanceId: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (domains: Omit<InstanceDomain, 'id' | 'instance_id'>[]) => {
      const response = await api.put(`/stacks/${stackName}/instances/${instanceId}/domains`, { domains })
      return response.data
    },
    onSuccess: () => {
      toast.success('Domains updated')
      queryClient.invalidateQueries({ queryKey: ['stacks', stackName, 'instances', instanceId] })
      queryClient.invalidateQueries({ queryKey: ['stacks', stackName, 'instances', instanceId, 'effective-config'] })
      queryClient.invalidateQueries({ queryKey: ['stacks', stackName, 'instances'] })
    },
    onError: (error) => {
      toast.error(getErrorMessage(error, 'Failed to update domains'))
    },
  })
}

export function useUpdateInstanceHealthcheck(stackName: string, instanceId: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (healthcheck: Omit<InstanceHealthcheck, 'id' | 'instance_id'> | null) => {
      const response = await api.put(`/stacks/${stackName}/instances/${instanceId}/healthcheck`, healthcheck)
      return response.data
    },
    onSuccess: () => {
      toast.success('Healthcheck updated')
      queryClient.invalidateQueries({ queryKey: ['stacks', stackName, 'instances', instanceId] })
      queryClient.invalidateQueries({ queryKey: ['stacks', stackName, 'instances', instanceId, 'effective-config'] })
      queryClient.invalidateQueries({ queryKey: ['stacks', stackName, 'instances'] })
    },
    onError: (error) => {
      toast.error(getErrorMessage(error, 'Failed to update healthcheck'))
    },
  })
}

export function useUpdateInstanceDependencies(stackName: string, instanceId: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (dependencies: Omit<InstanceDependency, 'id' | 'instance_id'>[]) => {
      const response = await api.put(`/stacks/${stackName}/instances/${instanceId}/dependencies`, { dependencies })
      return response.data
    },
    onSuccess: () => {
      toast.success('Dependencies updated')
      queryClient.invalidateQueries({ queryKey: ['stacks', stackName, 'instances', instanceId] })
      queryClient.invalidateQueries({ queryKey: ['stacks', stackName, 'instances', instanceId, 'effective-config'] })
      queryClient.invalidateQueries({ queryKey: ['stacks', stackName, 'instances'] })
    },
    onError: (error) => {
      toast.error(getErrorMessage(error, 'Failed to update dependencies'))
    },
  })
}

interface SaveConfigFileRequest {
  content: string
  file_mode: string
}

export function useSaveConfigFile(stackName: string, instanceId: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async ({ filePath, data }: { filePath: string; data: SaveConfigFileRequest }) => {
      const response = await api.put(`/stacks/${stackName}/instances/${instanceId}/files/${filePath}`, data)
      return response.data
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['stacks', stackName, 'instances', instanceId] })
      queryClient.invalidateQueries({ queryKey: ['stacks', stackName, 'instances', instanceId, 'effective-config'] })
      queryClient.invalidateQueries({ queryKey: ['stacks', stackName, 'instances'] })
    },
    onError: (error) => {
      toast.error(getErrorMessage(error, 'Failed to save file'))
    },
  })
}

export function useDeleteConfigFile(stackName: string, instanceId: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (filePath: string) => {
      const response = await api.delete(`/stacks/${stackName}/instances/${instanceId}/files/${filePath}`)
      return response.data
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['stacks', stackName, 'instances', instanceId] })
      queryClient.invalidateQueries({ queryKey: ['stacks', stackName, 'instances', instanceId, 'effective-config'] })
      queryClient.invalidateQueries({ queryKey: ['stacks', stackName, 'instances'] })
    },
    onError: (error) => {
      toast.error(getErrorMessage(error, 'Failed to delete file'))
    },
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
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (limits: ResourceLimits) => {
      const response = await api.put<ResourceLimitsResponse>(`/stacks/${stackName}/instances/${instanceId}/resources`, limits)
      return response.data
    },
    onSuccess: () => {
      toast.success('Resource limits updated')
      queryClient.invalidateQueries({ queryKey: ['stacks', stackName, 'instances', instanceId, 'resources'] })
      queryClient.invalidateQueries({ queryKey: ['stacks', stackName, 'instances', instanceId] })
      queryClient.invalidateQueries({ queryKey: ['stacks', stackName, 'instances', instanceId, 'effective-config'] })
    },
    onError: (error) => {
      toast.error(getErrorMessage(error, 'Failed to update resource limits'))
    },
  })
}

export function useUpdateInstanceEnvFiles(stackName: string, instanceId: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (env_files: string[]) => {
      const response = await api.put(`/stacks/${stackName}/instances/${instanceId}/env-files`, { env_files })
      return response.data
    },
    onSuccess: () => {
      toast.success('Env files updated')
      queryClient.invalidateQueries({ queryKey: ['stacks', stackName, 'instances', instanceId] })
      queryClient.invalidateQueries({ queryKey: ['stacks', stackName, 'instances', instanceId, 'effective-config'] })
      queryClient.invalidateQueries({ queryKey: ['stacks', stackName, 'instances'] })
    },
    onError: (error) => {
      toast.error(getErrorMessage(error, 'Failed to update env files'))
    },
  })
}

export function useUpdateInstanceNetworks(stackName: string, instanceId: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (networks: string[]) => {
      const response = await api.put(`/stacks/${stackName}/instances/${instanceId}/networks`, { networks })
      return response.data
    },
    onSuccess: () => {
      toast.success('Networks updated')
      queryClient.invalidateQueries({ queryKey: ['stacks', stackName, 'instances', instanceId] })
      queryClient.invalidateQueries({ queryKey: ['stacks', stackName, 'instances', instanceId, 'effective-config'] })
      queryClient.invalidateQueries({ queryKey: ['stacks', stackName, 'instances'] })
    },
    onError: (error) => {
      toast.error(getErrorMessage(error, 'Failed to update networks'))
    },
  })
}

export function useUpdateInstanceConfigMounts(stackName: string, instanceId: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (config_mounts: Omit<ServiceConfigMount, 'id' | 'service_id'>[]) => {
      const response = await api.put(`/stacks/${stackName}/instances/${instanceId}/config-mounts`, { config_mounts })
      return response.data
    },
    onSuccess: () => {
      toast.success('Config mounts updated')
      queryClient.invalidateQueries({ queryKey: ['stacks', stackName, 'instances', instanceId] })
      queryClient.invalidateQueries({ queryKey: ['stacks', stackName, 'instances', instanceId, 'effective-config'] })
      queryClient.invalidateQueries({ queryKey: ['stacks', stackName, 'instances'] })
    },
    onError: (error) => {
      toast.error(getErrorMessage(error, 'Failed to update config mounts'))
    },
  })
}
