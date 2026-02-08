import { X, CheckSquare } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { LifecycleButtons } from '@/components/ui/entity-actions'
import { useBulkServiceControl } from '@/features/services/queries'

interface BulkActionsToolbarProps {
  selected: Set<string>
  totalCount: number
  onSelectAll: () => void
  onClear: () => void
}

export function BulkActionsToolbar({ selected, totalCount, onSelectAll, onClear }: BulkActionsToolbarProps) {
  const bulkControl = useBulkServiceControl()
  const names = Array.from(selected)

  if (selected.size === 0) return null

  return (
    <div className="sticky top-0 z-10 flex flex-wrap items-center gap-3 rounded-lg border bg-background/95 backdrop-blur px-4 py-2 shadow-sm">
      <span className="shrink-0 min-w-0 text-sm font-medium">
        {selected.size} selected
      </span>
      <LifecycleButtons
        isRunning={false}
        onStart={() => bulkControl.mutate({ names, action: 'start' })}
        onStop={() => bulkControl.mutate({ names, action: 'stop' })}
        onRestart={() => bulkControl.mutate({ names, action: 'restart' })}
        isPending={bulkControl.isPending}
        showAll
      />
      <div className="ml-auto flex items-center gap-1">
        <Button variant="ghost" size="sm" onClick={onSelectAll}>
          <CheckSquare className="size-4" />
          All ({totalCount})
        </Button>
        <Button variant="ghost" size="sm" onClick={onClear}>
          <X className="size-4" />
          Clear
        </Button>
      </div>
    </div>
  )
}
