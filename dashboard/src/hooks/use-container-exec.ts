import { fetchWSToken } from '@/lib/api'

export async function buildExecWsUrl(containerName: string, cols: number, rows: number): Promise<string> {
  const token = await fetchWSToken()
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  const host = window.location.host
  const params = new URLSearchParams({
    cmd: '/bin/sh',
    cols: String(cols),
    rows: String(rows),
  })
  if (token) {
    params.set('token', token)
  }
  return `${protocol}//${host}/api/v1/containers/${containerName}/exec?${params}`
}
