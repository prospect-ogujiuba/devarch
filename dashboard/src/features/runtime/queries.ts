import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '@/lib/api'
import type { RuntimeStatus, SocketStatus } from '@/types/api'
import { toast } from 'sonner'

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
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (params: { runtime: string; options?: { stop_services?: boolean; preserve_data?: boolean; update_config?: boolean } }) => {
      const response = await api.post('/runtime/switch', params)
      return response.data
    },
    onSuccess: (data) => {
      toast.success(data.message ?? 'Runtime switched')
      queryClient.invalidateQueries({ queryKey: ['runtime'] })
      queryClient.invalidateQueries({ queryKey: ['status'] })
    },
    onError: () => {
      toast.error('Failed to switch runtime')
    },
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
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (params: { type: string; options?: { enable_lingering?: boolean; stop_conflicting?: boolean } }) => {
      const response = await api.post('/socket/start', params)
      return response.data
    },
    onSuccess: (data) => {
      toast.success(data.message ?? 'Socket started')
      queryClient.invalidateQueries({ queryKey: ['socket'] })
      queryClient.invalidateQueries({ queryKey: ['runtime'] })
    },
    onError: () => {
      toast.error('Failed to start socket')
    },
  })
}
