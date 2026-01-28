import { ArrowUpDown, ArrowUp, ArrowDown } from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'

export interface SortOption {
  value: string
  label: string
}

interface SortControlsProps {
  options: SortOption[]
  sortBy: string
  sortDir: 'asc' | 'desc'
  onSortByChange: (value: string) => void
  onSortDirChange: (dir: 'asc' | 'desc') => void
}

export function SortControls({
  options,
  sortBy,
  sortDir,
  onSortByChange,
  onSortDirChange,
}: SortControlsProps) {
  return (
    <div className="flex items-center gap-1">
      <Select value={sortBy} onValueChange={onSortByChange}>
        <SelectTrigger className="w-[140px] h-9">
          <ArrowUpDown className="size-3.5 mr-1" />
          <SelectValue />
        </SelectTrigger>
        <SelectContent>
          {options.map((opt) => (
            <SelectItem key={opt.value} value={opt.value}>
              {opt.label}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
      <Button
        variant="outline"
        size="sm"
        className="px-2"
        onClick={() => onSortDirChange(sortDir === 'asc' ? 'desc' : 'asc')}
      >
        {sortDir === 'asc' ? <ArrowUp className="size-4" /> : <ArrowDown className="size-4" />}
      </Button>
    </div>
  )
}
