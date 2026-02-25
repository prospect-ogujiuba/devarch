import { useRef } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { isAxiosError } from 'axios'
import { api, getErrorMessage } from '@/lib/api'
import type { Stack, DeletePreview, NetworkStatus, StackCompose, StackPlan, ApplyResult, ImportResult, Wire } from '@/types/api'
import { toast } from 'sonner'
import { useMutationHelper } from '@/lib/mutations'

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
  return useMutationHelper({
    mutationFn: async (data: CreateStackRequest) => {
      const response = await api.post('/stacks', data)
      return response.data
    },
    successMessage: 'Stack created',
    errorMessage: (error) => getErrorMessage(error, 'Failed to create stack'),
    invalidate: [['stacks']],
  })
}

interface UpdateStackRequest {
  description: string
}

export function useUpdateStack() {
  const queryClient = useQueryClient()
  return useMutationHelper({
    mutationFn: async ({ name, data }: { name: string; data: UpdateStackRequest }) => {
      const response = await api.put(`/stacks/${name}`, data)
      return response.data
    },
    successMessage: 'Stack updated',
    errorMessage: (error) => getErrorMessage(error, 'Failed to update stack'),
    invalidate: [['stacks']],
    onSuccess: (_data, { name }) => {
      queryClient.invalidateQueries({ queryKey: ['stacks', name] })
    },
  })
}

export function useDeleteStack() {
  return useMutationHelper({
    mutationFn: async (name: string) => {
      const response = await api.delete(`/stacks/${name}`)
      return response.data
    },
    successMessage: 'Stack deleted',
    errorMessage: (error) => getErrorMessage(error, 'Failed to delete stack'),
    invalidate: [['stacks']],
  })
}

export function useEnableStack() {
  return useMutationHelper({
    mutationFn: async (name: string) => {
      const response = await api.post(`/stacks/${name}/enable`)
      return response.data
    },
    loadingMessage: (name) => `Enabling ${name}...`,
    successMessage: (name) => `Enabled ${name}`,
    errorMessage: (error, name) => getErrorMessage(error, `Failed to enable ${name}`),
    invalidate: [['stacks']],
  })
}

export function useDisableStack() {
  return useMutationHelper({
    mutationFn: async (name: string) => {
      const response = await api.post(`/stacks/${name}/disable`)
      return response.data
    },
    loadingMessage: (name) => `Disabling ${name}...`,
    successMessage: (name) => `Disabled ${name}`,
    errorMessage: (error, name) => getErrorMessage(error, `Failed to disable ${name}`),
    invalidate: [['stacks']],
  })
}

export function useCloneStack() {
  const queryClient = useQueryClient()
  return useMutationHelper({
    mutationFn: async ({ name, newName }: { name: string; newName: string }) => {
      const response = await api.post(`/stacks/${name}/clone`, { name: newName })
      return response.data
    },
    successMessage: 'Stack cloned',
    errorMessage: (error) => getErrorMessage(error, 'Failed to clone stack'),
    invalidate: [['stacks']],
    onSuccess: (_data, { name, newName }) => {
      queryClient.invalidateQueries({ queryKey: ['stacks', name] })
      queryClient.invalidateQueries({ queryKey: ['stacks', newName] })
      queryClient.invalidateQueries({ queryKey: ['stacks', newName, 'instances'] })
    },
  })
}

export function useRenameStack() {
  return useMutationHelper({
    mutationFn: async ({ name, newName }: { name: string; newName: string }) => {
      const response = await api.post(`/stacks/${name}/rename`, { name: newName })
      return response.data
    },
    successMessage: 'Stack renamed',
    errorMessage: (error) => getErrorMessage(error, 'Failed to rename stack'),
    invalidate: [['stacks']],
  })
}

export function useRestoreStack() {
  return useMutationHelper({
    mutationFn: async (name: string) => {
      const response = await api.post(`/stacks/trash/${name}/restore`)
      return response.data
    },
    successMessage: (name) => `Restored ${name}`,
    errorMessage: (error) => getErrorMessage(error, 'Failed to restore stack'),
    invalidate: [['stacks'], ['stacks', 'trash']],
  })
}

export function usePermanentDeleteStack() {
  return useMutationHelper({
    mutationFn: async (name: string) => {
      const response = await api.delete(`/stacks/trash/${name}`)
      return response.data
    },
    successMessage: 'Stack permanently deleted',
    errorMessage: (error) => getErrorMessage(error, 'Failed to permanently delete stack'),
    invalidate: [['stacks'], ['stacks', 'trash']],
  })
}

export function useStopStack() {
  return useMutationHelper({
    mutationFn: async (name: string) => {
      const response = await api.post(`/stacks/${name}/stop`)
      return response.data
    },
    loadingMessage: (name) => `Stopping ${name}...`,
    successMessage: (name) => `Stopped ${name}`,
    errorMessage: (error, name) => getErrorMessage(error, `Failed to stop ${name}`),
    invalidate: [['stacks']],
  })
}

export function useStartStack() {
  return useMutationHelper({
    mutationFn: async (name: string) => {
      const response = await api.post(`/stacks/${name}/start`)
      return response.data
    },
    loadingMessage: (name) => `Starting ${name}...`,
    successMessage: (name) => `Started ${name}`,
    errorMessage: (error, name) => getErrorMessage(error, `Failed to start ${name}`),
    invalidate: [['stacks']],
  })
}

export function useRestartStack() {
  return useMutationHelper({
    mutationFn: async (name: string) => {
      const response = await api.post(`/stacks/${name}/restart`)
      return response.data
    },
    loadingMessage: (name) => `Restarting ${name}...`,
    successMessage: (name) => `Restarted ${name}`,
    errorMessage: (error, name) => getErrorMessage(error, `Failed to restart ${name}`),
    invalidate: [['stacks']],
  })
}

export function useCreateNetwork() {
  const queryClient = useQueryClient()
  return useMutationHelper({
    mutationFn: async (name: string) => {
      const response = await api.post(`/stacks/${name}/network`)
      return response.data
    },
    loadingMessage: (name) => `Creating network for ${name}...`,
    successMessage: (name) => `Network created for ${name}`,
    errorMessage: (error, name) => getErrorMessage(error, `Failed to create network for ${name}`),
    invalidate: [['stacks']],
    onSuccess: (_data, name) => {
      queryClient.invalidateQueries({ queryKey: ['stacks', name, 'network'] })
      queryClient.invalidateQueries({ queryKey: ['stacks', name] })
    },
  })
}

export function useRemoveNetwork() {
  const queryClient = useQueryClient()
  return useMutationHelper({
    mutationFn: async (name: string) => {
      const response = await api.delete(`/stacks/${name}/network`)
      return response.data
    },
    loadingMessage: (name) => `Removing network for ${name}...`,
    successMessage: (name) => `Network removed for ${name}`,
    errorMessage: (error, name) => getErrorMessage(error, `Failed to remove network for ${name}`),
    invalidate: [['stacks']],
    onSuccess: (_data, name) => {
      queryClient.invalidateQueries({ queryKey: ['stacks', name, 'network'] })
      queryClient.invalidateQueries({ queryKey: ['stacks', name] })
    },
  })
}

export function useGeneratePlan(name: string) {
  const toastIdRef = useRef<string | number | undefined>(undefined)
  return useMutation({
    mutationFn: async () => {
      const response = await api.get<StackPlan>(`/stacks/${name}/plan`)
      return response.data
    },
    onMutate: () => {
      toastIdRef.current = toast.loading('Generating deploy plan...')
    },
    onSuccess: () => {
      toast.success('Deploy plan ready', { id: toastIdRef.current })
      toastIdRef.current = undefined
    },
    onError: (error) => {
      toast.error(getErrorMessage(error, 'Failed to generate plan'), { id: toastIdRef.current })
      toastIdRef.current = undefined
    },
  })
}

export function useApplyPlan(name: string) {
  const queryClient = useQueryClient()
  const toastIdRef = useRef<string | number | undefined>(undefined)
  return useMutation({
    mutationFn: async ({ token }: { token: string }) => {
      const response = await api.post<ApplyResult>(`/stacks/${name}/apply`, { token })
      return response.data
    },
    onMutate: () => {
      toastIdRef.current = toast.loading('Deploying stack...')
    },
    onSuccess: () => {
      toast.success('Stack deployed', { id: toastIdRef.current })
      toastIdRef.current = undefined
      queryClient.invalidateQueries({ queryKey: ['stacks', name] })
      queryClient.invalidateQueries({ queryKey: ['stacks'] })
      queryClient.invalidateQueries({ queryKey: ['stacks', name, 'compose'] })
      queryClient.invalidateQueries({ queryKey: ['stacks', name, 'network'] })
    },
    onError: (error) => {
      if (error && typeof error === 'object' && 'response' in error) {
        const resp = (error as Record<string, unknown>).response as Record<string, unknown> | undefined
        if (!resp) {
          toast.error(getErrorMessage(error, 'Apply failed'), { id: toastIdRef.current })
          toastIdRef.current = undefined
          return
        }
        if (resp.status === 409) {
          toast.error('Plan is stale or another operation in progress. Regenerate plan.', { id: toastIdRef.current })
          toastIdRef.current = undefined
          return
        }
      }
      toast.error(getErrorMessage(error, 'Apply failed'), { id: toastIdRef.current })
      toastIdRef.current = undefined
    },
  })
}

export function useExportStack(name: string) {
  const toastIdRef = useRef<string | number | undefined>(undefined)
  return useMutation({
    mutationFn: async () => {
      const response = await api.get(`/stacks/${name}/export`, {
        responseType: 'blob',
      })
      return response.data
    },
    onMutate: () => {
      toastIdRef.current = toast.loading('Exporting stack...')
    },
    onSuccess: (blob) => {
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = `${name}-devarch.yml`
      a.click()
      URL.revokeObjectURL(url)
      toast.success('Stack exported', { id: toastIdRef.current })
      toastIdRef.current = undefined
    },
    onError: () => {
      toast.error('Export failed', { id: toastIdRef.current })
      toastIdRef.current = undefined
    },
  })
}

export function useImportStack() {
  const queryClient = useQueryClient()
  const toastIdRef = useRef<string | number | undefined>(undefined)
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
    onMutate: () => {
      toastIdRef.current = toast.loading('Importing stack...')
    },
    onSuccess: (result) => {
      const createdCount = result.created.length
      const updatedCount = result.updated.length
      toast.success(`Imported stack. Created: ${createdCount}, Updated: ${updatedCount}`, { id: toastIdRef.current })
      toastIdRef.current = undefined
      queryClient.invalidateQueries({ queryKey: ['stacks'] })
      queryClient.invalidateQueries({ queryKey: ['stacks', result.stack_name] })
    },
    onError: (error) => {
      if (isAxiosError(error) && error.response?.status === 413) {
        const maxBytes = typeof error.response.data === 'object' && error.response.data !== null
          ? (error.response.data as Record<string, unknown>).max_bytes
          : undefined
        const maxMB = typeof maxBytes === 'number' ? Math.round(maxBytes / (1024 * 1024)) : 256
        toast.error(`Import file is too large. Max allowed is ${maxMB}MB.`, { id: toastIdRef.current })
        toastIdRef.current = undefined
        return
      }
      toast.error(getErrorMessage(error, 'Import failed'), { id: toastIdRef.current })
      toastIdRef.current = undefined
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
  return useMutationHelper({
    mutationFn: async () => {
      const response = await api.post<{ resolved: number; warnings: string[] }>(`/stacks/${name}/wires/resolve`)
      return response.data
    },
    errorMessage: (error) => getErrorMessage(error, 'Failed to resolve wires'),
    onSuccess: (data) => {
      if (data.resolved > 0) {
        toast.success(`Resolved ${data.resolved} wire${data.resolved === 1 ? '' : 's'}`)
      } else {
        toast.success('No changes')
      }
      queryClient.invalidateQueries({ queryKey: ['stacks', name, 'wires'] })
    },
  })
}

export function useCreateWire(name: string) {
  const queryClient = useQueryClient()
  return useMutationHelper({
    mutationFn: async (data: { consumer_instance_id: string; provider_instance_id: string; import_contract_name: string }) => {
      const response = await api.post(`/stacks/${name}/wires`, data)
      return response.data
    },
    successMessage: 'Wire created',
    errorMessage: (error) => getErrorMessage(error, 'Failed to create wire'),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['stacks', name, 'wires'] })
    },
  })
}

export function useDeleteWire(name: string) {
  const queryClient = useQueryClient()
  return useMutationHelper({
    mutationFn: async (wireId: number) => {
      const response = await api.delete(`/stacks/${name}/wires/${wireId}`)
      return response.data
    },
    successMessage: 'Wire disconnected',
    errorMessage: (error) => getErrorMessage(error, 'Failed to disconnect wire'),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['stacks', name, 'wires'] })
    },
  })
}

export function useCleanupOrphanedWires(name: string) {
  const queryClient = useQueryClient()
  return useMutationHelper({
    mutationFn: async () => {
      const response = await api.post<{ deleted: number }>(`/stacks/${name}/wires/cleanup`)
      return response.data
    },
    errorMessage: (error) => getErrorMessage(error, 'Failed to cleanup wires'),
    onSuccess: (data) => {
      toast.success(`Cleaned up ${data.deleted} orphaned wire${data.deleted === 1 ? '' : 's'}`)
      queryClient.invalidateQueries({ queryKey: ['stacks', name, 'wires'] })
    },
  })
}
