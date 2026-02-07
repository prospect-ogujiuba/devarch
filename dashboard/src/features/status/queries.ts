import { useQuery } from '@tanstack/react-query'
import { api } from '@/lib/api'
import type { StatusOverview } from '@/types/api'

export function useStatusOverview() {
  return useQuery({
    queryKey: ['status'],
    queryFn: async () => {
      const response = await api.get<StatusOverview>('/status')
      return response.data
    },
    refetchInterval: 30000,
  })
}
