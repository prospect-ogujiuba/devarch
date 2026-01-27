import { Badge } from '@/components/ui/badge'
import { cn } from '@/lib/utils'

interface StatusBadgeProps {
  status: string
  className?: string
}

const variants: Record<string, 'success' | 'destructive' | 'warning' | 'secondary'> = {
  running: 'success',
  stopped: 'secondary',
  starting: 'warning',
  error: 'destructive',
}

const labels: Record<string, string> = {
  running: 'Running',
  stopped: 'Stopped',
  starting: 'Starting',
  error: 'Error',
}

export function StatusBadge({ status, className }: StatusBadgeProps) {

  return (
    <Badge variant={variants[status] ?? 'secondary'} className={cn('gap-1.5', className)}>
      <span className={cn(
        'size-1.5 rounded-full',
        status === 'running' && 'bg-success animate-pulse',
        status === 'stopped' && 'bg-muted-foreground',
        status === 'starting' && 'bg-warning animate-pulse',
        status === 'error' && 'bg-destructive',
      )} />
      {labels[status] ?? status}
    </Badge>
  )
}
