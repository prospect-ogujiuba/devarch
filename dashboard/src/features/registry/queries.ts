import { useQuery } from '@tanstack/react-query'
import { api } from '@/lib/api'
import type { RegistryInfo, TagResult, SearchResult, ImageInfo } from '@/types/api'

export function useRegistries() {
  return useQuery({
    queryKey: ['registries'],
    queryFn: async () => {
      const res = await api.get<RegistryInfo[]>('/registries')
      return res.data
    },
    staleTime: 10 * 60 * 1000,
  })
}

export function useImageTags(registryName: string, repository: string, opts?: { enabled?: boolean }) {
  return useQuery({
    queryKey: ['registry-tags', registryName, repository],
    queryFn: async () => {
      const res = await api.get<TagResult[]>(`/registries/${registryName}/images/${repository}/tags?page_size=50`)
      return res.data
    },
    enabled: opts?.enabled !== false && !!registryName && !!repository,
    staleTime: 2 * 60 * 1000,
  })
}

export function useSearchImages(registryName: string, query: string) {
  return useQuery({
    queryKey: ['registry-search', registryName, query || ''],
    queryFn: async () => {
      const url = query
        ? `/registries/${registryName}/search?q=${encodeURIComponent(query)}&page_size=25`
        : `/registries/${registryName}/search?page_size=24`
      const res = await api.get<SearchResult[]>(url)
      return res.data
    },
    enabled: !!registryName,
    staleTime: 5 * 60 * 1000,
  })
}

export function useImageInfo(registryName: string, repository: string, opts?: { enabled?: boolean }) {
  return useQuery({
    queryKey: ['registry-image-info', registryName, repository],
    queryFn: async () => {
      const res = await api.get<ImageInfo>(`/registries/${registryName}/images/${repository}`)
      return res.data
    },
    enabled: opts?.enabled !== false && !!registryName && !!repository,
    staleTime: 10 * 60 * 1000,
  })
}
