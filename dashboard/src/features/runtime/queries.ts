import { useQuery } from '@tanstack/react-query'
import { api } from '@/lib/api'
import type { RuntimeStatus, SocketStatus } from '@/types/api'
import { useMutationHelper } from '@/lib/mutations'

export function useRuntimeStatus() {
  return useQuery({
    queryKey: ['runtime', 'status'],
    queryFn: async () => {
      const response = await api.get<RuntimeStatus>('/runtime/status')
      return response.data
    },
    refetchInterval: 30000,
  })
}

export function useSwitchRuntime() {
  return useMutationHelper({
    mutationFn: async (params: { runtime: string; options?: { stop_services?: boolean; preserve_data?: boolean; update_config?: boolean } }) => {
      const response = await api.post('/runtime/switch', params)
      return response.data
    },
    successMessage: (_vars, data) => data.message ?? 'Runtime switched',
    errorMessage: 'Failed to switch runtime',
    invalidate: [['runtime'], ['status']],
  })
}

export function useSocketStatus() {
  return useQuery({
    queryKey: ['socket', 'status'],
    queryFn: async () => {
      const response = await api.get<SocketStatus>('/socket/status')
      return response.data
    },
    refetchInterval: 30000,
  })
}

export function useStartSocket() {
  return useMutationHelper({
    mutationFn: async (params: { type: string; options?: { enable_lingering?: boolean; stop_conflicting?: boolean } }) => {
      const response = await api.post('/socket/start', params)
      return response.data
    },
    successMessage: (_vars, data) => data.message ?? 'Socket started',
    errorMessage: 'Failed to start socket',
    invalidate: [['socket'], ['runtime']],
  })
}
