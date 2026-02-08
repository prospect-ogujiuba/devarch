import { cn } from '@/lib/utils'

interface FieldRowProps {
  className?: string
  children: React.ReactNode
}

export function FieldRow({ className, children }: FieldRowProps) {
  return (
    <div className={cn('flex flex-col items-stretch gap-2 sm:flex-row sm:items-center', className)}>
      {children}
    </div>
  )
}

interface FieldRowActionsProps {
  className?: string
  children: React.ReactNode
}

export function FieldRowActions({ className, children }: FieldRowActionsProps) {
  return (
    <div className={cn('flex items-center gap-1 sm:ml-auto', className)}>
      {children}
    </div>
  )
}
