import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '@/lib/api'
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
    refetchInterval: 5000,
  })
}

export function useInstance(stackName: string, instanceId: string) {
  return useQuery({
    queryKey: ['stacks', stackName, 'instances', instanceId],
    queryFn: async () => {
      const response = await api.get<InstanceDetail>(`/stacks/${stackName}/instances/${instanceId}`)
      return response.data
    },
    enabled: !!stackName && !!instanceId,
    refetchInterval: 5000,
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
      return response.data
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
    },
    onError: (error: any) => {
      toast.error(error.response?.data || 'Failed to create instance')
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
    onError: (error: any) => {
      toast.error(error.response?.data || 'Failed to update instance')
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
    },
    onError: (error: any) => {
      toast.error(error.response?.data || 'Failed to delete instance')
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
    },
    onError: (error: any) => {
      toast.error(error.response?.data || 'Failed to duplicate instance')
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
    },
    onError: (error: any) => {
      toast.error(error.response?.data || 'Failed to rename instance')
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
    onError: (error: any) => {
      toast.error(error.response?.data || 'Failed to update ports')
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
    onError: (error: any) => {
      toast.error(error.response?.data || 'Failed to update volumes')
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
    onError: (error: any) => {
      toast.error(error.response?.data || 'Failed to update environment variables')
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
    onError: (error: any) => {
      toast.error(error.response?.data || 'Failed to update labels')
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
    onError: (error: any) => {
      toast.error(error.response?.data || 'Failed to update domains')
    },
  })
}

export function useUpdateInstanceHealthcheck(stackName: string, instanceId: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (healthcheck: Omit<InstanceHealthcheck, 'id' | 'instance_id'> | null) => {
      const response = await api.put(`/stacks/${stackName}/instances/${instanceId}/healthcheck`, { healthcheck })
      return response.data
    },
    onSuccess: () => {
      toast.success('Healthcheck updated')
      queryClient.invalidateQueries({ queryKey: ['stacks', stackName, 'instances', instanceId] })
      queryClient.invalidateQueries({ queryKey: ['stacks', stackName, 'instances', instanceId, 'effective-config'] })
      queryClient.invalidateQueries({ queryKey: ['stacks', stackName, 'instances'] })
    },
    onError: (error: any) => {
      toast.error(error.response?.data || 'Failed to update healthcheck')
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
    onError: (error: any) => {
      toast.error(error.response?.data || 'Failed to update dependencies')
    },
  })
}
