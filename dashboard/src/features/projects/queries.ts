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
    refetchInterval: 30000,
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

export function useLinkStack(name: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (stackId: number | null) => {
      const response = await api.put<Project>(`/projects/${name}/stack`, { stack_id: stackId })
      return response.data
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['projects', name] })
      queryClient.invalidateQueries({ queryKey: ['projects', name, 'services'] })
      queryClient.invalidateQueries({ queryKey: ['projects', name, 'status'] })
    },
  })
}

export function useProjectServiceControl(projectName: string) {
  const queryClient = useQueryClient()

  const invalidate = () => {
    queryClient.invalidateQueries({ queryKey: ['projects', projectName] })
    queryClient.invalidateQueries({ queryKey: ['projects', projectName, 'services'] })
    queryClient.invalidateQueries({ queryKey: ['projects', projectName, 'status'] })
  }

  const startService = useMutation({
    mutationFn: async (service: string) => {
      const response = await api.post<ControlResponse>(`/projects/${projectName}/services/${service}/start`)
      return response.data
    },
    onSettled: invalidate,
  })

  const stopService = useMutation({
    mutationFn: async (service: string) => {
      const response = await api.post<ControlResponse>(`/projects/${projectName}/services/${service}/stop`)
      return response.data
    },
    onSettled: invalidate,
  })

  const restartService = useMutation({
    mutationFn: async (service: string) => {
      const response = await api.post<ControlResponse>(`/projects/${projectName}/services/${service}/restart`)
      return response.data
    },
    onSettled: invalidate,
  })

  return { startService, stopService, restartService }
}

export function useProjectControl(name: string) {
  const queryClient = useQueryClient()

  const invalidateAll = () => {
    queryClient.invalidateQueries({ queryKey: ['projects', name] })
    queryClient.invalidateQueries({ queryKey: ['projects', name, 'services'] })
    queryClient.invalidateQueries({ queryKey: ['projects', name, 'status'] })
  }

  const start = useMutation({
    mutationFn: async () => {
      const response = await api.post<ControlResponse>(`/projects/${name}/start`)
      return response.data
    },
    onSettled: invalidateAll,
  })

  const stop = useMutation({
    mutationFn: async () => {
      const response = await api.post<ControlResponse>(`/projects/${name}/stop`)
      return response.data
    },
    onSettled: invalidateAll,
  })

  const restart = useMutation({
    mutationFn: async () => {
      const response = await api.post<ControlResponse>(`/projects/${name}/restart`)
      return response.data
    },
    onSettled: invalidateAll,
  })

  return { start, stop, restart }
}
