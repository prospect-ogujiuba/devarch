import { useQuery } from '@tanstack/react-query'
import { api, getErrorMessage } from '@/lib/api'
import type { ProxyConfigResult, ProxyTypeInfo } from '@/types/api'
import { useMutationHelper } from '@/lib/mutations'

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
  return useMutationHelper({
    mutationFn: async (proxyType: string) => {
      const response = await api.post<ProxyConfigResult>(
        `/services/${serviceName}/proxy-config`,
        { proxy_type: proxyType },
      )
      return response.data
    },
    errorMessage: (error) => getErrorMessage(error, 'Failed to generate proxy config'),
  })
}

export function useGenerateStackProxyConfig(stackName: string) {
  return useMutationHelper({
    mutationFn: async (proxyType: string) => {
      const response = await api.post<ProxyConfigResult>(
        `/stacks/${stackName}/proxy-config`,
        { proxy_type: proxyType },
      )
      return response.data
    },
    errorMessage: (error) => getErrorMessage(error, 'Failed to generate proxy config'),
  })
}

export function useGenerateProjectProxyConfig(projectName: string) {
  return useMutationHelper({
    mutationFn: async (proxyType: string) => {
      const response = await api.post<ProxyConfigResult>(
        `/projects/${projectName}/proxy-config`,
        { proxy_type: proxyType },
      )
      return response.data
    },
    errorMessage: (error) => getErrorMessage(error, 'Failed to generate proxy config'),
  })
}
