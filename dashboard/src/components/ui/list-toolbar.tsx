import { Search } from 'lucide-react'
import { Input } from '@/components/ui/input'
import { ViewSwitcher, type ViewMode } from '@/components/ui/view-switcher'
import { SortControls, type SortOption } from '@/components/ui/sort-controls'

interface ListToolbarProps {
  search: string
  onSearchChange: (value: string) => void
  searchPlaceholder?: string
  sortOptions: SortOption[]
  sortBy: string
  sortDir: 'asc' | 'desc'
  onSortByChange: (value: string) => void
  onSortDirChange: (dir: 'asc' | 'desc') => void
  viewMode: ViewMode
  onViewModeChange: (mode: ViewMode) => void
  children?: React.ReactNode
}

export function ListToolbar({
  search,
  onSearchChange,
  searchPlaceholder = 'Search...',
  sortOptions,
  sortBy,
  sortDir,
  onSortByChange,
  onSortDirChange,
  viewMode,
  onViewModeChange,
  children,
}: ListToolbarProps) {
  return (
    <div className="flex flex-wrap items-center gap-3">
      <div className="relative flex-1 max-w-sm">
        <Search className="absolute left-3 top-1/2 -translate-y-1/2 size-4 text-muted-foreground" />
        <Input
          placeholder={searchPlaceholder}
          value={search}
          onChange={(e) => onSearchChange(e.target.value)}
          className="pl-9"
        />
      </div>
      {children}
      <div className="ml-auto flex items-center gap-2">
        <SortControls
          options={sortOptions}
          sortBy={sortBy}
          sortDir={sortDir}
          onSortByChange={onSortByChange}
          onSortDirChange={onSortDirChange}
        />
        <ViewSwitcher value={viewMode} onChange={onViewModeChange} />
      </div>
    </div>
  )
}
