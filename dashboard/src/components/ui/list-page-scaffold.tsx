import type { ReactNode } from 'react'
import type { LucideIcon } from 'lucide-react'
import { Loader2 } from 'lucide-react'
import { StatCard } from '@/components/ui/stat-card'
import { ListToolbar } from '@/components/ui/list-toolbar'
import { EmptyState } from '@/components/ui/empty-state'
import type { SortOption } from '@/components/ui/sort-controls'
import type { ViewMode } from '@/components/ui/view-switcher'

interface StatCardConfig {
  icon: LucideIcon
  label: string
  value: string | number
  color?: string
}

interface ListControls {
  search: string
  setSearch: (value: string) => void
  sortBy: string
  setSortBy: (value: string) => void
  sortDir: 'asc' | 'desc'
  setSortDir: (dir: 'asc' | 'desc') => void
  viewMode: ViewMode
  setViewMode: (mode: ViewMode) => void
  filters: Record<string, string>
  setFilter: (key: string, value: string) => void
  filtered: unknown[]
  total: number
}

interface ListPageScaffoldProps<T> {
  title: string
  subtitle?: string
  isLoading?: boolean
  statCards?: StatCardConfig[]
  statGridClassName?: string
  controls: ListControls
  sortOptions: SortOption[]
  searchPlaceholder?: string
  actionButton?: ReactNode
  selectionSlot?: ReactNode
  filterChildren?: ReactNode
  emptyIcon: LucideIcon
  emptyMessage: string
  emptyAction?: { label: string; onClick: () => void }
  items: T[]
  tableView: (items: T[]) => ReactNode
  gridView: (items: T[]) => ReactNode
  gridClassName?: string
  children?: ReactNode
}

export function ListPageScaffold<T>({
  title,
  subtitle,
  isLoading,
  statCards,
  statGridClassName,
  controls,
  sortOptions,
  searchPlaceholder = 'Search...',
  actionButton,
  selectionSlot,
  filterChildren,
  emptyIcon,
  emptyMessage,
  emptyAction,
  items,
  tableView,
  gridView,
  children,
}: ListPageScaffoldProps<T>) {
  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <Loader2 className="size-8 animate-spin text-muted-foreground" />
      </div>
    )
  }

  return (
    <div className="space-y-5 sm:space-y-6">
      <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 className="text-xl font-bold sm:text-2xl">{title}</h1>
          {subtitle && (
            <p className="text-sm text-muted-foreground sm:text-base">{subtitle}</p>
          )}
        </div>
        {actionButton}
      </div>

      {statCards && statCards.length > 0 && (
        <div className={statGridClassName ?? 'grid grid-cols-2 gap-3 sm:grid-cols-3'}>
          {statCards.map((card) => (
            <StatCard
              key={card.label}
              icon={card.icon}
              label={card.label}
              value={card.value}
              color={card.color}
            />
          ))}
        </div>
      )}

      {selectionSlot}

      <ListToolbar
        search={controls.search}
        onSearchChange={controls.setSearch}
        searchPlaceholder={searchPlaceholder}
        sortOptions={sortOptions}
        sortBy={controls.sortBy}
        sortDir={controls.sortDir}
        onSortByChange={controls.setSortBy}
        onSortDirChange={controls.setSortDir}
        viewMode={controls.viewMode}
        onViewModeChange={controls.setViewMode}
      >
        {filterChildren}
      </ListToolbar>

      {items.length === 0 ? (
        <EmptyState icon={emptyIcon} message={emptyMessage} action={emptyAction} />
      ) : controls.viewMode === 'table' ? (
        tableView(items)
      ) : (
        gridView(items)
      )}

      {children}
    </div>
  )
}
