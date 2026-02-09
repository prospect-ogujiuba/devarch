import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api, getErrorMessage } from '@/lib/api'
import type { NetworkInfo } from '@/types/api'
import { toast } from 'sonner'

export function useNetworks() {
  return useQuery({
    queryKey: ['networks'],
    queryFn: async () => {
      const response = await api.get<NetworkInfo[]>('/networks')
      return Array.isArray(response.data) ? response.data : []
    },
    refetchInterval: 30000,
  })
}

export function useCreateNetwork() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (name: string) => {
      const response = await api.post<NetworkInfo>('/networks', { name })
      return response.data
    },
    onSuccess: (_data, name) => {
      toast.success(`Created network ${name}`)
      queryClient.invalidateQueries({ queryKey: ['networks'] })
    },
    onError: (error) => {
      toast.error(getErrorMessage(error, 'Failed to create network'))
    },
  })
}

export function useRemoveNetwork() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (name: string) => {
      const response = await api.delete(`/networks/${name}`)
      return response.data
    },
    onSuccess: (_data, name) => {
      toast.success(`Removed network ${name}`)
      queryClient.invalidateQueries({ queryKey: ['networks'] })
    },
    onError: (error, name) => {
      toast.error(getErrorMessage(error, `Failed to remove ${name}`))
    },
  })
}

export function useBulkRemoveNetworks() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (names: string[]) => {
      const response = await api.post<{ removed: string[]; errors: { name: string; error: string }[] }>('/networks/bulk-remove', { names })
      return response.data
    },
    onSuccess: (data) => {
      if (data.removed.length > 0) {
        toast.success(`Removed ${data.removed.length} network${data.removed.length === 1 ? '' : 's'}`)
      }
      if (data.errors.length > 0) {
        toast.error(`${data.errors.length} network${data.errors.length === 1 ? '' : 's'} failed to remove`)
      }
      queryClient.invalidateQueries({ queryKey: ['networks'] })
    },
    onError: (error) => {
      toast.error(getErrorMessage(error, 'Bulk remove failed'))
    },
  })
}
