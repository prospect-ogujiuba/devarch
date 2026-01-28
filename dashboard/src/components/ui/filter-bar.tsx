import { Button } from '@/components/ui/button'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { cn } from '@/lib/utils'

export interface FilterOption {
  value: string
  label: string
  count?: number
}

interface FilterBarProps {
  options: FilterOption[]
  value: string
  onChange: (value: string) => void
  className?: string
  variant?: 'chips' | 'dropdown'
}

export function FilterBar({ options, value, onChange, className, variant = 'chips' }: FilterBarProps) {
  if (variant === 'dropdown') {
    return (
      <div className={className}>
        <Select value={value} onValueChange={onChange}>
          <SelectTrigger className="w-[180px] h-8">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            {options.map((opt) => (
              <SelectItem key={opt.value} value={opt.value}>
                {opt.label}{opt.count !== undefined ? ` (${opt.count})` : ''}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>
    )
  }

  return (
    <>
      <div className={cn('hidden md:flex items-center gap-1', className)}>
        {options.map((opt) => (
          <Button
            key={opt.value}
            variant={value === opt.value ? 'secondary' : 'ghost'}
            size="sm"
            onClick={() => onChange(opt.value)}
            className="h-8"
          >
            {opt.label}
            {opt.count !== undefined && (
              <span className="ml-1.5 text-xs text-muted-foreground">{opt.count}</span>
            )}
          </Button>
        ))}
      </div>
      <div className={cn('md:hidden', className)}>
        <Select value={value} onValueChange={onChange}>
          <SelectTrigger className="w-[160px]">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            {options.map((opt) => (
              <SelectItem key={opt.value} value={opt.value}>
                {opt.label}{opt.count !== undefined ? ` (${opt.count})` : ''}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>
    </>
  )
}
