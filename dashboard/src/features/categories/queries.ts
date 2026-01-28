import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '@/lib/api'
import type { Category, Service } from '@/types/api'
import { toast } from 'sonner'

export function useCategories() {
  return useQuery({
    queryKey: ['categories'],
    queryFn: async () => {
      const response = await api.get<Category[]>('/categories')
      return Array.isArray(response.data) ? response.data : []
    },
    refetchInterval: 5000,
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
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (name: string) => {
      const response = await api.post(`/categories/${name}/start`)
      return response.data
    },
    onSuccess: (_data, name) => {
      toast.success(`Started all services in ${name}`)
      queryClient.invalidateQueries({ queryKey: ['services'] })
      queryClient.invalidateQueries({ queryKey: ['categories'] })
      queryClient.invalidateQueries({ queryKey: ['status'] })
    },
    onError: (_error, name) => {
      toast.error(`Failed to start ${name} category`)
    },
  })
}

export function useStopCategory() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (name: string) => {
      const response = await api.post(`/categories/${name}/stop`)
      return response.data
    },
    onSuccess: (_data, name) => {
      toast.success(`Stopped all services in ${name}`)
      queryClient.invalidateQueries({ queryKey: ['services'] })
      queryClient.invalidateQueries({ queryKey: ['categories'] })
      queryClient.invalidateQueries({ queryKey: ['status'] })
    },
    onError: (_error, name) => {
      toast.error(`Failed to stop ${name} category`)
    },
  })
}
