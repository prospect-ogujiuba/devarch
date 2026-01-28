import { LayoutGrid, List, LayoutList } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'

export type ViewMode = 'table' | 'grid' | 'card'

interface ViewSwitcherProps {
  value: ViewMode
  onChange: (mode: ViewMode) => void
}

const views: { mode: ViewMode; icon: typeof List; label: string }[] = [
  { mode: 'table', icon: List, label: 'Table' },
  { mode: 'grid', icon: LayoutGrid, label: 'Grid' },
  { mode: 'card', icon: LayoutList, label: 'Card' },
]

export function ViewSwitcher({ value, onChange }: ViewSwitcherProps) {
  return (
    <div className="flex items-center rounded-md border">
      {views.map(({ mode, icon: Icon, label }) => (
        <Button
          key={mode}
          variant="ghost"
          size="sm"
          className={cn(
            'rounded-none first:rounded-l-md last:rounded-r-md px-2.5',
            value === mode && 'bg-muted',
          )}
          onClick={() => onChange(mode)}
          title={label}
        >
          <Icon className="size-4" />
        </Button>
      ))}
    </div>
  )
}
