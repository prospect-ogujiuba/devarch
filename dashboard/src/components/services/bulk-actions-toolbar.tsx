import { X, CheckSquare, Loader2, Play, Square, RotateCcw } from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
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
    <div className="sticky top-0 z-10 flex items-center gap-2 rounded-lg border bg-background/95 px-3 py-2 shadow-sm backdrop-blur">
      <span className="min-w-0 shrink-0 text-sm font-medium">{selected.size} selected</span>
      <div className="ml-auto flex items-center gap-1">
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="outline" size="sm" disabled={bulkControl.isPending}>
              {bulkControl.isPending ? <Loader2 className="size-4 animate-spin" /> : null}
              Actions
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            <DropdownMenuItem onClick={() => bulkControl.mutate({ names, action: 'start' })}>
              <Play className="size-4" />
              Start
            </DropdownMenuItem>
            <DropdownMenuItem onClick={() => bulkControl.mutate({ names, action: 'stop' })}>
              <Square className="size-4" />
              Stop
            </DropdownMenuItem>
            <DropdownMenuItem onClick={() => bulkControl.mutate({ names, action: 'restart' })}>
              <RotateCcw className="size-4" />
              Restart
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
        <Button variant="ghost" size="sm" onClick={onSelectAll}>
          <CheckSquare className="size-4" />
          All
          <span className="hidden sm:inline"> ({totalCount})</span>
        </Button>
        <Button variant="ghost" size="sm" onClick={onClear}>
          <X className="size-4" />
          Clear
        </Button>
      </div>
    </div>
  )
}
