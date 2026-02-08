import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '@/lib/api'
import type { Stack, DeletePreview, NetworkStatus, StackCompose, StackPlan, ApplyResult } from '@/types/api'
import { toast } from 'sonner'

// Queries
export function useStacks() {
  return useQuery({
    queryKey: ['stacks'],
    queryFn: async () => {
      const response = await api.get<Stack[]>('/stacks')
      return Array.isArray(response.data) ? response.data : []
    },
    refetchInterval: 30000,
  })
}

export function useStack(name: string) {
  return useQuery({
    queryKey: ['stacks', name],
    queryFn: async () => {
      const response = await api.get<Stack>(`/stacks/${name}`)
      return response.data
    },
    enabled: !!name,
    refetchInterval: 30000,
  })
}

export function useTrashStacks() {
  return useQuery({
    queryKey: ['stacks', 'trash'],
    queryFn: async () => {
      const response = await api.get<Stack[]>('/stacks/trash')
      return Array.isArray(response.data) ? response.data : []
    },
  })
}

export function useDeletePreview(name: string) {
  return useQuery({
    queryKey: ['stacks', name, 'delete-preview'],
    queryFn: async () => {
      const response = await api.get<DeletePreview>(`/stacks/${name}/delete-preview`)
      return response.data
    },
    enabled: !!name,
  })
}

export function useStackNetwork(name: string) {
  return useQuery({
    queryKey: ['stacks', name, 'network'],
    queryFn: async () => {
      const response = await api.get<NetworkStatus>(`/stacks/${name}/network`)
      const data = response.data
      return {
        ...data,
        containers: Array.isArray(data.containers) ? data.containers : [],
        labels: data.labels ?? {},
      }
    },
    enabled: !!name,
    refetchInterval: 30000,
  })
}

export function useStackCompose(name: string) {
  return useQuery({
    queryKey: ['stacks', name, 'compose'],
    queryFn: async () => {
      const response = await api.get<StackCompose>(`/stacks/${name}/compose`)
      return response.data
    },
    enabled: !!name,
  })
}

// Mutations
interface CreateStackRequest {
  name: string
  description: string
}

export function useCreateStack() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (data: CreateStackRequest) => {
      const response = await api.post('/stacks', data)
      return response.data
    },
    onSuccess: () => {
      toast.success('Stack created')
      queryClient.invalidateQueries({ queryKey: ['stacks'] })
    },
    onError: (error: any) => {
      toast.error(error.response?.data || 'Failed to create stack')
    },
  })
}

interface UpdateStackRequest {
  description: string
}

export function useUpdateStack() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async ({ name, data }: { name: string; data: UpdateStackRequest }) => {
      const response = await api.put(`/stacks/${name}`, data)
      return response.data
    },
    onSuccess: (_data, { name }) => {
      toast.success('Stack updated')
      queryClient.invalidateQueries({ queryKey: ['stacks', name] })
      queryClient.invalidateQueries({ queryKey: ['stacks'] })
    },
    onError: (error: any) => {
      toast.error(error.response?.data || 'Failed to update stack')
    },
  })
}

export function useDeleteStack() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (name: string) => {
      const response = await api.delete(`/stacks/${name}`)
      return response.data
    },
    onSuccess: () => {
      toast.success('Stack deleted')
      queryClient.invalidateQueries({ queryKey: ['stacks'] })
    },
    onError: (error: any) => {
      toast.error(error.response?.data || 'Failed to delete stack')
    },
  })
}

export function useEnableStack() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (name: string) => {
      const response = await api.post(`/stacks/${name}/enable`)
      return response.data
    },
    onSuccess: (_data, name) => {
      toast.success(`Enabled ${name}`)
      queryClient.invalidateQueries({ queryKey: ['stacks'] })
    },
    onError: (error: any, name) => {
      toast.error(error.response?.data || `Failed to enable ${name}`)
    },
  })
}

export function useDisableStack() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (name: string) => {
      const response = await api.post(`/stacks/${name}/disable`)
      return response.data
    },
    onSuccess: (_data, name) => {
      toast.success(`Disabled ${name}`)
      queryClient.invalidateQueries({ queryKey: ['stacks'] })
    },
    onError: (error: any, name) => {
      toast.error(error.response?.data || `Failed to disable ${name}`)
    },
  })
}

export function useCloneStack() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async ({ name, newName }: { name: string; newName: string }) => {
      const response = await api.post(`/stacks/${name}/clone`, { name: newName })
      return response.data
    },
    onSuccess: (_data, { name, newName }) => {
      toast.success('Stack cloned')
      queryClient.invalidateQueries({ queryKey: ['stacks'] })
      queryClient.invalidateQueries({ queryKey: ['stacks', name] })
      queryClient.invalidateQueries({ queryKey: ['stacks', newName] })
      queryClient.invalidateQueries({ queryKey: ['stacks', newName, 'instances'] })
    },
    onError: (error: any) => {
      toast.error(error.response?.data || 'Failed to clone stack')
    },
  })
}

export function useRenameStack() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async ({ name, newName }: { name: string; newName: string }) => {
      const response = await api.post(`/stacks/${name}/rename`, { name: newName })
      return response.data
    },
    onSuccess: () => {
      toast.success('Stack renamed')
      queryClient.invalidateQueries({ queryKey: ['stacks'] })
    },
    onError: (error: any) => {
      toast.error(error.response?.data || 'Failed to rename stack')
    },
  })
}

export function useRestoreStack() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (name: string) => {
      const response = await api.post(`/stacks/trash/${name}/restore`)
      return response.data
    },
    onSuccess: (_data, name) => {
      toast.success(`Restored ${name}`)
      queryClient.invalidateQueries({ queryKey: ['stacks'] })
      queryClient.invalidateQueries({ queryKey: ['stacks', 'trash'] })
    },
    onError: (error: any) => {
      toast.error(error.response?.data || 'Failed to restore stack')
    },
  })
}

export function usePermanentDeleteStack() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (name: string) => {
      const response = await api.delete(`/stacks/trash/${name}`)
      return response.data
    },
    onSuccess: () => {
      toast.success('Stack permanently deleted')
      queryClient.invalidateQueries({ queryKey: ['stacks'] })
      queryClient.invalidateQueries({ queryKey: ['stacks', 'trash'] })
    },
    onError: (error: any) => {
      toast.error(error.response?.data || 'Failed to permanently delete stack')
    },
  })
}

export function useStopStack() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (name: string) => {
      const response = await api.post(`/stacks/${name}/stop`)
      return response.data
    },
    onSuccess: (_data, name) => {
      toast.success(`Stopped ${name}`)
      queryClient.invalidateQueries({ queryKey: ['stacks'] })
    },
    onError: (error: any, name) => {
      toast.error(error.response?.data || `Failed to stop ${name}`)
    },
  })
}

export function useStartStack() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (name: string) => {
      const response = await api.post(`/stacks/${name}/start`)
      return response.data
    },
    onSuccess: (_data, name) => {
      toast.success(`Started ${name}`)
      queryClient.invalidateQueries({ queryKey: ['stacks'] })
    },
    onError: (error: any, name) => {
      toast.error(error.response?.data || `Failed to start ${name}`)
    },
  })
}

export function useRestartStack() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (name: string) => {
      const response = await api.post(`/stacks/${name}/restart`)
      return response.data
    },
    onSuccess: (_data, name) => {
      toast.success(`Restarted ${name}`)
      queryClient.invalidateQueries({ queryKey: ['stacks'] })
    },
    onError: (error: any, name) => {
      toast.error(error.response?.data || `Failed to restart ${name}`)
    },
  })
}

export function useCreateNetwork() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (name: string) => {
      const response = await api.post(`/stacks/${name}/network`)
      return response.data
    },
    onSuccess: (_data, name) => {
      toast.success(`Network created for ${name}`)
      queryClient.invalidateQueries({ queryKey: ['stacks', name, 'network'] })
      queryClient.invalidateQueries({ queryKey: ['stacks', name] })
      queryClient.invalidateQueries({ queryKey: ['stacks'] })
    },
    onError: (error: any, name) => {
      toast.error(error.response?.data || `Failed to create network for ${name}`)
    },
  })
}

export function useRemoveNetwork() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (name: string) => {
      const response = await api.delete(`/stacks/${name}/network`)
      return response.data
    },
    onSuccess: (_data, name) => {
      toast.success(`Network removed for ${name}`)
      queryClient.invalidateQueries({ queryKey: ['stacks', name, 'network'] })
      queryClient.invalidateQueries({ queryKey: ['stacks', name] })
      queryClient.invalidateQueries({ queryKey: ['stacks'] })
    },
    onError: (error: any, name) => {
      toast.error(error.response?.data || `Failed to remove network for ${name}`)
    },
  })
}

export function useGeneratePlan(name: string) {
  return useMutation({
    mutationFn: async () => {
      const response = await api.get<StackPlan>(`/stacks/${name}/plan`)
      return response
    },
    onError: (error: any) => {
      toast.error(error.response?.data || 'Failed to generate plan')
    },
  })
}

export function useApplyPlan(name: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async ({ token }: { token: string }) => {
      const response = await api.post<ApplyResult>(`/stacks/${name}/apply`, { token })
      return response.data
    },
    onSuccess: () => {
      toast.success('Stack deployed')
      queryClient.invalidateQueries({ queryKey: ['stacks', name] })
      queryClient.invalidateQueries({ queryKey: ['stacks'] })
      queryClient.invalidateQueries({ queryKey: ['stacks', name, 'compose'] })
      queryClient.invalidateQueries({ queryKey: ['stacks', name, 'network'] })
    },
    onError: (error: any) => {
      if (error.response?.status === 409) {
        toast.error('Plan is stale or another operation in progress. Regenerate plan.')
      } else {
        toast.error(error.response?.data || 'Apply failed')
      }
    },
  })
}
