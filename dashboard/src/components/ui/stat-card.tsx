import type { LucideIcon } from 'lucide-react'
import { cn } from '@/lib/utils'

interface StatCardProps {
  icon: LucideIcon
  label: string
  value: string | number
  color?: string
  className?: string
}

export function StatCard({ icon: Icon, label, value, color, className }: StatCardProps) {
  return (
    <div className={cn('flex min-w-0 items-center gap-2 rounded-md border px-3 py-2 sm:gap-3 sm:rounded-lg sm:px-4 sm:py-3', className)}>
      <Icon className={cn('size-4 shrink-0 text-muted-foreground sm:size-5', color)} />
      <div className="min-w-0">
        <div className="truncate text-[11px] text-muted-foreground sm:text-xs">{label}</div>
        <div className="text-base font-semibold leading-tight sm:text-lg">{value}</div>
      </div>
    </div>
  )
}
