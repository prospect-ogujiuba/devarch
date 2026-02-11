import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api, getErrorMessage } from '@/lib/api'
import { toast } from 'sonner'
import type { Project, ProjectScanResponse, ProjectServiceStatus, ControlResponse } from '@/types/api'

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

interface CreateProjectRequest {
  name: string
  path: string
  project_type: string
  framework?: string
  language?: string
  description?: string
  domain?: string
  proxy_port?: number
}

export function useCreateProject() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (data: CreateProjectRequest) => {
      const response = await api.post<Project>('/projects', data)
      return response.data
    },
    onSuccess: () => {
      toast.success('Project created')
      queryClient.invalidateQueries({ queryKey: ['projects'] })
      queryClient.invalidateQueries({ queryKey: ['stacks'] })
    },
    onError: (error) => {
      toast.error(getErrorMessage(error, 'Failed to create project'))
    },
  })
}

interface UpdateProjectRequest {
  path?: string
  project_type?: string
  framework?: string
  language?: string
  description?: string
  domain?: string
  proxy_port?: number
}

export function useUpdateProject(name: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (data: UpdateProjectRequest) => {
      const response = await api.put<Project>(`/projects/${name}`, data)
      return response.data
    },
    onSuccess: () => {
      toast.success('Project updated')
      queryClient.invalidateQueries({ queryKey: ['projects', name] })
      queryClient.invalidateQueries({ queryKey: ['projects'] })
    },
    onError: (error) => {
      toast.error(getErrorMessage(error, 'Failed to update project'))
    },
  })
}

export function useDeleteProject(name: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async () => {
      const response = await api.delete(`/projects/${name}`)
      return response.data
    },
    onSuccess: () => {
      toast.success('Project deleted')
      queryClient.invalidateQueries({ queryKey: ['projects'] })
      queryClient.invalidateQueries({ queryKey: ['stacks'] })
      queryClient.invalidateQueries({ queryKey: ['stacks', 'trash'] })
    },
    onError: (error) => {
      toast.error(getErrorMessage(error, 'Failed to delete project'))
    },
  })
}

export function useProjectServiceControl(projectName: string) {
  const queryClient = useQueryClient()

  const invalidate = () => {
    queryClient.invalidateQueries({ queryKey: ['projects', projectName] })
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
