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
    <div className={cn('flex items-center gap-3 rounded-lg border px-4 py-3', className)}>
      <Icon className={cn('size-5 text-muted-foreground', color)} />
      <div>
        <div className="text-xs text-muted-foreground">{label}</div>
        <div className="text-lg font-semibold leading-tight">{value}</div>
      </div>
    </div>
  )
}
