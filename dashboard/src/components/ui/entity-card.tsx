import * as React from 'react'
import { cn } from '@/lib/utils'

interface EntityCardProps extends React.ComponentProps<'div'> {
  cursor?: 'default' | 'pointer'
}

export function EntityCard({ className, cursor = 'default', ...props }: EntityCardProps) {
  return (
    <div
      className={cn(
        'rounded-lg border bg-card text-card-foreground shadow-sm',
        'hover:border-primary/50 transition-colors',
        cursor === 'pointer' && 'cursor-pointer',
        className,
      )}
      {...props}
    />
  )
}
