import { Play, Square, RotateCw, X, Loader2, CheckSquare } from 'lucide-react'
import { Button } from '@/components/ui/button'
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
  const isLoading = bulkControl.isPending

  if (selected.size === 0) return null

  return (
    <div className="sticky top-0 z-10 flex items-center gap-3 rounded-lg border bg-background/95 backdrop-blur px-4 py-2 shadow-sm">
      <span className="text-sm font-medium">
        {selected.size} selected
      </span>
      <div className="flex items-center gap-1">
        <Button
          variant="outline"
          size="sm"
          onClick={() => bulkControl.mutate({ names, action: 'start' })}
          disabled={isLoading}
        >
          {isLoading ? <Loader2 className="size-4 animate-spin" /> : <Play className="size-4" />}
          Start
        </Button>
        <Button
          variant="outline"
          size="sm"
          onClick={() => bulkControl.mutate({ names, action: 'stop' })}
          disabled={isLoading}
        >
          {isLoading ? <Loader2 className="size-4 animate-spin" /> : <Square className="size-4" />}
          Stop
        </Button>
        <Button
          variant="outline"
          size="sm"
          onClick={() => bulkControl.mutate({ names, action: 'restart' })}
          disabled={isLoading}
        >
          {isLoading ? <Loader2 className="size-4 animate-spin" /> : <RotateCw className="size-4" />}
          Restart
        </Button>
      </div>
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
