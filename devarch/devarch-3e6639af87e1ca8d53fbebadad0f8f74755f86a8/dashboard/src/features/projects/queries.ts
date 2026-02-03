import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '@/lib/api'
import type { Project, ProjectScanResponse, ProjectService, ProjectServiceStatus, ControlResponse } from '@/types/api'

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

export function useProjectServices(name: string) {
  return useQuery({
    queryKey: ['projects', name, 'services'],
    queryFn: async () => {
      const response = await api.get<ProjectService[]>(`/projects/${name}/services`)
      return response.data
    },
    enabled: !!name,
  })
}

export function useProjectStatus(name: string, enabled = true) {
  return useQuery({
    queryKey: ['projects', name, 'status'],
    queryFn: async () => {
      const response = await api.get<ProjectServiceStatus[]>(`/projects/${name}/status`)
      return response.data
    },
    enabled: !!name && enabled,
    refetchInterval: 5000,
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

export function useProjectControl(name: string) {
  const queryClient = useQueryClient()

  const start = useMutation({
    mutationFn: async () => {
      const response = await api.post<ControlResponse>(`/projects/${name}/start`)
      return response.data
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['projects', name, 'status'] })
    },
  })

  const stop = useMutation({
    mutationFn: async () => {
      const response = await api.post<ControlResponse>(`/projects/${name}/stop`)
      return response.data
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['projects', name, 'status'] })
    },
  })

  const restart = useMutation({
    mutationFn: async () => {
      const response = await api.post<ControlResponse>(`/projects/${name}/restart`)
      return response.data
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['projects', name, 'status'] })
    },
  })

  return { start, stop, restart }
}
