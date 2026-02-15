import { useQuery, useQueryClient } from '@tanstack/react-query'
import { api, getErrorMessage } from '@/lib/api'
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

interface CreateCategoryPayload {
  name: string
  display_name?: string
  color?: string
  startup_order?: number
}

export function useCreateCategory() {
  return useMutationHelper({
    mutationFn: async (data: CreateCategoryPayload) => {
      const response = await api.post('/categories', data)
      return response.data
    },
    successMessage: 'Category created',
    errorMessage: (error) => getErrorMessage(error, 'Failed to create category'),
    invalidate: [['categories']],
  })
}

interface UpdateCategoryPayload {
  name?: string
  display_name?: string
  color?: string
  startup_order?: number
}

export function useUpdateCategory() {
  const queryClient = useQueryClient()
  return useMutationHelper({
    mutationFn: async ({ name, data }: { name: string; data: UpdateCategoryPayload }) => {
      const response = await api.put(`/categories/${name}`, data)
      return response.data
    },
    successMessage: 'Category updated',
    errorMessage: (error) => getErrorMessage(error, 'Failed to update category'),
    invalidate: [['categories']],
    onSuccess: (_data, { name }) => {
      queryClient.invalidateQueries({ queryKey: ['categories', name] })
    },
  })
}

export function useDeleteCategory() {
  return useMutationHelper({
    mutationFn: async (name: string) => {
      const response = await api.delete(`/categories/${name}`)
      return response.data
    },
    successMessage: 'Category deleted',
    errorMessage: (error) => getErrorMessage(error, 'Failed to delete category'),
    invalidate: [['categories']],
  })
}
