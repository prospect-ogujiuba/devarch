import { useQuery } from '@tanstack/react-query'
import { api, getErrorMessage } from '@/lib/api'
import { useMutationHelper } from '@/lib/mutations'
import type {
  ContainerImage,
  ContainerImageInspect,
  ImageHistoryEntry,
  ImagePullProgress,
  ImagePruneResult,
} from '@/types/api'

export function useImages(all = false) {
  return useQuery({
    queryKey: ['images', { all }],
    queryFn: async () => {
      const response = await api.get<ContainerImage[]>(`/images?all=${all}`)
      return Array.isArray(response.data) ? response.data : []
    },
  })
}

export function useImageInspect(name: string) {
  return useQuery({
    queryKey: ['images', 'inspect', name],
    queryFn: async () => {
      const response = await api.get<ContainerImageInspect>(`/images/inspect?name=${encodeURIComponent(name)}`)
      return response.data
    },
    enabled: !!name,
  })
}

export function useImageHistory(name: string) {
  return useQuery({
    queryKey: ['images', 'history', name],
    queryFn: async () => {
      const response = await api.get<ImageHistoryEntry[]>(`/images/history?name=${encodeURIComponent(name)}`)
      return Array.isArray(response.data) ? response.data : []
    },
    enabled: !!name,
  })
}

export function useRemoveImage() {
  return useMutationHelper({
    mutationFn: async ({ name, force = false }: { name: string; force?: boolean }) => {
      const response = await api.delete(`/images/remove?name=${encodeURIComponent(name)}&force=${force}`)
      return response.data
    },
    successMessage: 'Image removed',
    errorMessage: (error) => getErrorMessage(error, 'Failed to remove image'),
    invalidate: [['images']],
  })
}

export function usePruneImages() {
  return useMutationHelper<ImagePruneResult[], boolean>({
    mutationFn: async (dangling) => {
      const response = await api.post<ImagePruneResult[]>(`/images/prune?dangling=${dangling}`)
      return Array.isArray(response.data) ? response.data : []
    },
    successMessage: (_, data) => {
      const totalSize = data.reduce((acc, r) => acc + r.Size, 0)
      const sizeMB = (totalSize / 1024 / 1024).toFixed(2)
      return `Pruned ${data.length} image(s), freed ${sizeMB} MB`
    },
    errorMessage: (error) => getErrorMessage(error, 'Failed to prune images'),
    invalidate: [['images']],
  })
}

export async function pullImageWithProgress(
  reference: string,
  onProgress: (report: ImagePullProgress) => void,
): Promise<void> {
  const apiKey = localStorage.getItem('devarch-api-key') ?? ''
  const res = await fetch('/api/v1/images/pull', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      ...(apiKey ? { 'X-API-Key': apiKey } : {}),
    },
    body: JSON.stringify({ reference }),
  })
  if (!res.ok) throw new Error(await res.text())
  if (!res.body) return

  const reader = res.body.getReader()
  const decoder = new TextDecoder()
  let buffer = ''

  while (true) {
    const { done, value } = await reader.read()
    if (done) break
    buffer += decoder.decode(value, { stream: true })
    const lines = buffer.split('\n')
    buffer = lines.pop() || ''
    for (const line of lines) {
      if (line.trim()) {
        onProgress(JSON.parse(line))
      }
    }
  }
}
