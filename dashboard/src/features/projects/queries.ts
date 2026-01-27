import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '@/lib/api'
import type { Project, ProjectScanResponse } from '@/types/api'

export function useProjects() {
  return useQuery({
    queryKey: ['projects'],
    queryFn: async () => {
      const response = await api.get<Project[]>('/projects')
      return response.data
    },
  })
}

export function useProject(name: string) {
  return useQuery({
    queryKey: ['projects', name],
    queryFn: async () => {
      const response = await api.get<Project>(`/projects/${name}`)
      return response.data
    },
    enabled: !!name,
  })
}

export function useScanProjects() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async () => {
      const response = await api.post<ProjectScanResponse>('/projects/scan')
      return response.data
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['projects'] })
    },
  })
}
