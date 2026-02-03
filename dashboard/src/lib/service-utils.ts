import type { Service } from '@/types/api'

export function getServiceStatus(s: Service): string {
  const raw = s.status?.status ?? 'stopped'
  if (raw === 'exited' || raw === 'dead' || raw === 'created') return 'stopped'
  return raw
}
