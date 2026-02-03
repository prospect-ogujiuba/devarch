import { cn } from '@/lib/utils'
import { getResourceBarColor } from '@/lib/format'

interface ResourceBarProps {
  value: number
  max?: number
  className?: string
  label?: string
}

export function ResourceBar({ value, max = 100, className, label }: ResourceBarProps) {
  const pct = max > 0 ? Math.min((value / max) * 100, 100) : 0

  return (
    <div className={cn('space-y-1', className)}>
      {label && (
        <div className="flex justify-between text-xs text-muted-foreground">
          <span>{label}</span>
          <span>{pct.toFixed(1)}%</span>
        </div>
      )}
      <div className="h-1.5 w-full rounded-full bg-muted">
        <div
          className={cn('h-full rounded-full transition-all', getResourceBarColor(pct))}
          style={{ width: `${pct}%` }}
        />
      </div>
    </div>
  )
}
