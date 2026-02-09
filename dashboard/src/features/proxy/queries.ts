import { useQuery, useMutation } from '@tanstack/react-query'
import { api, getErrorMessage } from '@/lib/api'
import type { ProxyConfigResult, ProxyTypeInfo } from '@/types/api'
import { toast } from 'sonner'

export function useProxyTypes() {
  return useQuery({
    queryKey: ['proxy', 'types'],
    queryFn: async () => {
      const response = await api.get<ProxyTypeInfo[]>('/proxy/types')
      return response.data
    },
    staleTime: Infinity,
  })
}

export function useGenerateServiceProxyConfig(serviceName: string) {
  return useMutation({
    mutationFn: async (proxyType: string) => {
      const response = await api.post<ProxyConfigResult>(
        `/services/${serviceName}/proxy-config`,
        { proxy_type: proxyType },
      )
      return response.data
    },
    onError: (error) => {
      toast.error(getErrorMessage(error, 'Failed to generate proxy config'))
    },
  })
}

export function useGenerateStackProxyConfig(stackName: string) {
  return useMutation({
    mutationFn: async (proxyType: string) => {
      const response = await api.post<ProxyConfigResult>(
        `/stacks/${stackName}/proxy-config`,
        { proxy_type: proxyType },
      )
      return response.data
    },
    onError: (error) => {
      toast.error(getErrorMessage(error, 'Failed to generate proxy config'))
    },
  })
}

export function useGenerateProjectProxyConfig(projectName: string) {
  return useMutation({
    mutationFn: async (proxyType: string) => {
      const response = await api.post<ProxyConfigResult>(
        `/projects/${projectName}/proxy-config`,
        { proxy_type: proxyType },
      )
      return response.data
    },
    onError: (error) => {
      toast.error(getErrorMessage(error, 'Failed to generate proxy config'))
    },
  })
}
