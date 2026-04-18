import type { ReactNode } from 'react'
import { classNames } from '../lib/utils'

export type BadgeTone = 'neutral' | 'success' | 'warning' | 'danger' | 'info'

interface StatusBadgeProps {
  children: ReactNode
  tone?: BadgeTone
}

export function toneFromStatus(value?: string): BadgeTone {
  const normalized = value?.toLowerCase() ?? ''
  if (normalized.includes('run') || normalized.includes('success') || normalized.includes('healthy')) {
    return 'success'
  }
  if (normalized.includes('warn') || normalized.includes('block') || normalized.includes('pending')) {
    return 'warning'
  }
  if (normalized.includes('error') || normalized.includes('fail') || normalized.includes('unhealthy') || normalized.includes('stopped')) {
    return 'danger'
  }
  if (normalized.includes('info') || normalized.includes('add') || normalized.includes('modify') || normalized.includes('restart')) {
    return 'info'
  }
  return 'neutral'
}

export function StatusBadge({ children, tone = 'neutral' }: StatusBadgeProps) {
  return <span className={classNames('badge', `badge--${tone}`)}>{children}</span>
}
