import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api, getErrorMessage } from '@/lib/api'
import type { Stack, DeletePreview, NetworkStatus, StackCompose, StackPlan, ApplyResult, ImportResult, Wire } from '@/types/api'
import { toast } from 'sonner'

interface UnresolvedContract {
  instance: string
  contract_name: string
  contract_type: string
  required: boolean
  reason: string
  available_providers?: { instance_id: string, instance_name: string }[]
}

interface WireListResponse {
  wires: Wire[]
  unresolved: UnresolvedContract[]
}

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
    onError: (error) => {
      toast.error(getErrorMessage(error, 'Failed to create stack'))
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
    onError: (error) => {
      toast.error(getErrorMessage(error, 'Failed to update stack'))
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
    onError: (error) => {
      toast.error(getErrorMessage(error, 'Failed to delete stack'))
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
    onError: (error, name) => {
      toast.error(getErrorMessage(error, `Failed to enable ${name}`))
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
    onError: (error, name) => {
      toast.error(getErrorMessage(error, `Failed to disable ${name}`))
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
    onError: (error) => {
      toast.error(getErrorMessage(error, 'Failed to clone stack'))
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
    onError: (error) => {
      toast.error(getErrorMessage(error, 'Failed to rename stack'))
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
    onError: (error) => {
      toast.error(getErrorMessage(error, 'Failed to restore stack'))
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
    onError: (error) => {
      toast.error(getErrorMessage(error, 'Failed to permanently delete stack'))
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
    onError: (error, name) => {
      toast.error(getErrorMessage(error, `Failed to stop ${name}`))
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
    onError: (error, name) => {
      toast.error(getErrorMessage(error, `Failed to start ${name}`))
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
    onError: (error, name) => {
      toast.error(getErrorMessage(error, `Failed to restart ${name}`))
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
    onError: (error, name) => {
      toast.error(getErrorMessage(error, `Failed to create network for ${name}`))
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
    onError: (error, name) => {
      toast.error(getErrorMessage(error, `Failed to remove network for ${name}`))
    },
  })
}

export function useGeneratePlan(name: string) {
  return useMutation({
    mutationFn: async () => {
      const response = await api.get<StackPlan>(`/stacks/${name}/plan`)
      return response
    },
    onError: (error) => {
      toast.error(getErrorMessage(error, 'Failed to generate plan'))
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
    onError: (error) => {
      if (error && typeof error === 'object' && 'response' in error) {
        const resp = (error as Record<string, unknown>).response as Record<string, unknown> | undefined
        if (!resp) {
          toast.error(getErrorMessage(error, 'Apply failed'))
          return
        }
        if (resp.status === 409) {
          toast.error('Plan is stale or another operation in progress. Regenerate plan.')
          return
        }
      }
      toast.error(getErrorMessage(error, 'Apply failed'))
    },
  })
}

export function useExportStack(name: string) {
  return useMutation({
    mutationFn: async () => {
      const response = await api.get(`/stacks/${name}/export`, {
        responseType: 'blob',
      })
      return response.data
    },
    onSuccess: (blob) => {
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = `${name}-devarch.yml`
      a.click()
      URL.revokeObjectURL(url)
    },
    onError: () => {
      toast.error('Export failed')
    },
  })
}

export function useImportStack() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (file: File) => {
      const formData = new FormData()
      formData.append('file', file)
      const response = await api.post<ImportResult>('/stacks/import', formData, {
        headers: {
          'Content-Type': 'multipart/form-data',
        },
      })
      return response.data
    },
    onSuccess: (result) => {
      const createdCount = result.created.length
      const updatedCount = result.updated.length
      toast.success(`Imported stack. Created: ${createdCount}, Updated: ${updatedCount}`)
      queryClient.invalidateQueries({ queryKey: ['stacks'] })
      queryClient.invalidateQueries({ queryKey: ['stacks', result.stack_name] })
    },
    onError: (error) => {
      toast.error(getErrorMessage(error, 'Import failed'))
    },
  })
}

export function useStackWires(name: string) {
  return useQuery({
    queryKey: ['stacks', name, 'wires'],
    queryFn: async () => {
      const response = await api.get<WireListResponse>(`/stacks/${name}/wires`)
      const data = response.data

      const wires = (data.wires ?? []).map((wire) => {
        const consumerName = wire.consumer_instance_name || wire.consumer_instance
        const providerName = wire.provider_instance_name || wire.provider_instance
        return {
          ...wire,
          consumer_instance_name: consumerName,
          provider_instance_name: providerName,
          contract_name: wire.contract_name,
          consumer_contract_type: wire.consumer_contract_type || wire.contract_type,
        }
      })

      const unresolved = (data.unresolved ?? []).map((item) => {
        const reason = item.reason === 'missing' || item.reason === 'ambiguous'
          ? item.reason
          : ((item.available_providers?.length ?? 0) > 0 ? 'ambiguous' : 'missing')

        return {
          ...item,
          reason,
        }
      })

      return {
        ...data,
        wires,
        unresolved,
      }
    },
    enabled: !!name,
    refetchInterval: 30000,
  })
}

export function useResolveWires(name: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async () => {
      const response = await api.post<{ resolved: number; warnings: string[] }>(`/stacks/${name}/wires/resolve`)
      return response.data
    },
    onSuccess: (data) => {
      if (data.resolved > 0) {
        toast.success(`Resolved ${data.resolved} wire${data.resolved === 1 ? '' : 's'}`)
      } else {
        toast.success('No changes')
      }
      queryClient.invalidateQueries({ queryKey: ['stacks', name, 'wires'] })
    },
    onError: (error) => {
      toast.error(getErrorMessage(error, 'Failed to resolve wires'))
    },
  })
}

export function useCreateWire(name: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (data: { consumer_instance_id: string; provider_instance_id: string; import_contract_name: string }) => {
      const response = await api.post(`/stacks/${name}/wires`, data)
      return response.data
    },
    onSuccess: () => {
      toast.success('Wire created')
      queryClient.invalidateQueries({ queryKey: ['stacks', name, 'wires'] })
    },
    onError: (error) => {
      toast.error(getErrorMessage(error, 'Failed to create wire'))
    },
  })
}

export function useDeleteWire(name: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (wireId: number) => {
      const response = await api.delete(`/stacks/${name}/wires/${wireId}`)
      return response.data
    },
    onSuccess: () => {
      toast.success('Wire disconnected')
      queryClient.invalidateQueries({ queryKey: ['stacks', name, 'wires'] })
    },
    onError: (error) => {
      toast.error(getErrorMessage(error, 'Failed to disconnect wire'))
    },
  })
}

export function useCleanupOrphanedWires(name: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async () => {
      const response = await api.post<{ deleted: number }>(`/stacks/${name}/wires/cleanup`)
      return response.data
    },
    onSuccess: (data) => {
      toast.success(`Cleaned up ${data.deleted} orphaned wire${data.deleted === 1 ? '' : 's'}`)
      queryClient.invalidateQueries({ queryKey: ['stacks', name, 'wires'] })
    },
    onError: (error) => {
      toast.error(getErrorMessage(error, 'Failed to cleanup wires'))
    },
  })
}
