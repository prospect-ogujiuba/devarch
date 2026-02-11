import { useQuery } from '@tanstack/react-query'
import { api, getErrorMessage } from '@/lib/api'
import type { NetworkInfo } from '@/types/api'
import { useMutationHelper } from '@/lib/mutations'

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
  return useMutationHelper({
    mutationFn: async (name: string) => {
      const response = await api.post<NetworkInfo>('/networks', { name })
      return response.data
    },
    successMessage: (_vars) => `Created network ${_vars}`,
    errorMessage: (error) => getErrorMessage(error, 'Failed to create network'),
    invalidate: [['networks']],
  })
}

export function useRemoveNetwork() {
  return useMutationHelper({
    mutationFn: async (name: string) => {
      const response = await api.delete(`/networks/${name}`)
      return response.data
    },
    successMessage: (_vars) => `Removed network ${_vars}`,
    errorMessage: (error, vars) => getErrorMessage(error, `Failed to remove ${vars}`),
    invalidate: [['networks']],
  })
}

export function useBulkRemoveNetworks() {
  return useMutationHelper({
    mutationFn: async (names: string[]) => {
      const response = await api.post<{ removed: string[]; errors: { name: string; error: string }[] }>('/networks/bulk-remove', { names })
      return response.data
    },
    successMessage: (_vars, data) => {
      if (data.removed.length > 0 && data.errors.length === 0) {
        return `Removed ${data.removed.length} network${data.removed.length === 1 ? '' : 's'}`
      } else if (data.removed.length > 0 && data.errors.length > 0) {
        return `Removed ${data.removed.length}, ${data.errors.length} failed`
      } else {
        return `${data.errors.length} network${data.errors.length === 1 ? '' : 's'} failed to remove`
      }
    },
    errorMessage: (error) => getErrorMessage(error, 'Bulk remove failed'),
    invalidate: [['networks']],
  })
}
