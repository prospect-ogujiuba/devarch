import { Badge } from '@/components/ui/badge'
import { cn } from '@/lib/utils'

interface StatusBadgeProps {
  status: 'running' | 'stopped' | 'starting' | 'error'
  className?: string
}

export function StatusBadge({ status, className }: StatusBadgeProps) {
  const variants: Record<typeof status, 'success' | 'destructive' | 'warning' | 'secondary'> = {
    running: 'success',
    stopped: 'secondary',
    starting: 'warning',
    error: 'destructive',
  }

  const labels: Record<typeof status, string> = {
    running: 'Running',
    stopped: 'Stopped',
    starting: 'Starting',
    error: 'Error',
  }

  return (
    <Badge variant={variants[status]} className={cn('gap-1.5', className)}>
      <span className={cn(
        'size-1.5 rounded-full',
        status === 'running' && 'bg-success animate-pulse',
        status === 'stopped' && 'bg-muted-foreground',
        status === 'starting' && 'bg-warning animate-pulse',
        status === 'error' && 'bg-destructive',
      )} />
      {labels[status]}
    </Badge>
  )
}
