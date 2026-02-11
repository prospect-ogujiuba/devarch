import { useQuery } from '@tanstack/react-query'
import { api } from '@/lib/api'
import type { Category, Service } from '@/types/api'
import { useMutationHelper } from '@/lib/mutations'

export function useCategories() {
  return useQuery({
    queryKey: ['categories'],
    queryFn: async () => {
      const response = await api.get<Category[]>('/categories')
      return Array.isArray(response.data) ? response.data : []
    },
    refetchInterval: 30000,
  })
}

export function useCategoryServices(category: string) {
  return useQuery({
    queryKey: ['categories', category, 'services'],
    queryFn: async () => {
      const response = await api.get<Service[]>(`/categories/${category}/services`)
      return response.data
    },
    enabled: !!category,
  })
}

export function useStartCategory() {
  return useMutationHelper({
    mutationFn: async (name: string) => {
      const response = await api.post(`/categories/${name}/start`)
      return response.data
    },
    successMessage: (_vars) => `Started all services in ${_vars}`,
    errorMessage: (_error, vars) => `Failed to start ${vars} category`,
    invalidate: [['services'], ['categories'], ['status']],
  })
}

export function useStopCategory() {
  return useMutationHelper({
    mutationFn: async (name: string) => {
      const response = await api.post(`/categories/${name}/stop`)
      return response.data
    },
    successMessage: (_vars) => `Stopped all services in ${_vars}`,
    errorMessage: (_error, vars) => `Failed to stop ${vars} category`,
    invalidate: [['services'], ['categories'], ['status']],
  })
}
