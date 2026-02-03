import { useMemo } from 'react'
import { createFileRoute } from '@tanstack/react-router'
import { Loader2, Server, Play, Square } from 'lucide-react'
import { useCategories } from '@/features/categories/queries'
import { CategoryCard } from '@/components/categories/category-card'
import { CategoryTable } from '@/components/categories/category-table'
import { ListToolbar } from '@/components/ui/list-toolbar'
import { FilterBar, type FilterOption } from '@/components/ui/filter-bar'
import { StatCard } from '@/components/ui/stat-card'
import { EmptyState } from '@/components/ui/empty-state'
import { useListControls } from '@/hooks/use-list-controls'
import type { Category } from '@/types/api'

export const Route = createFileRoute('/categories/')({
  component: CategoriesPage,
})

const searchFn = (cat: Category, q: string) =>
  cat.name.toLowerCase().includes(q.toLowerCase())

const filterFns = {
  status: (cat: Category, value: string) => {
    const running = cat.runningCount ?? 0
    const total = cat.service_count ?? 0
    if (value === 'running') return running === total && total > 0
    if (value === 'partial') return running > 0 && running < total
    if (value === 'stopped') return running === 0
    return true
  },
}

const sortFns: Record<string, (a: Category, b: Category) => number> = {
  name: (a, b) => a.name.localeCompare(b.name),
  services: (a, b) => (a.service_count ?? 0) - (b.service_count ?? 0),
  running: (a, b) => (a.runningCount ?? 0) - (b.runningCount ?? 0),
  order: (a, b) => a.startup_order - b.startup_order,
}

const sortOptions = [
  { value: 'name', label: 'Name' },
  { value: 'services', label: 'Services' },
  { value: 'running', label: 'Running' },
  { value: 'order', label: 'Startup Order' },
]

function CategoriesPage() {
  const { data: categories, isLoading } = useCategories()
  const items = useMemo(() => categories ?? [], [categories])

  const controls = useListControls({
    storageKey: 'categories',
    items,
    searchFn,
    filterFns,
    sortFns,
    defaultSort: 'order',
    defaultView: 'grid',
  })

  const statusCounts = useMemo(() => {
    let running = 0
    let partial = 0
    let stopped = 0
    for (const cat of items) {
      const r = cat.runningCount ?? 0
      const t = cat.service_count ?? 0
      if (r === t && t > 0) running++
      else if (r > 0) partial++
      else stopped++
    }
    return { all: items.length, running, partial, stopped }
  }, [items])

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <Loader2 className="size-8 animate-spin text-muted-foreground" />
      </div>
    )
  }

  const totalServices = items.reduce((acc, cat) => acc + (cat.service_count ?? 0), 0)
  const totalRunning = items.reduce((acc, cat) => acc + (cat.runningCount ?? 0), 0)

  const statusOptions: FilterOption[] = [
    { value: 'all', label: 'All', count: statusCounts.all },
    { value: 'running', label: 'All Running', count: statusCounts.running },
    { value: 'partial', label: 'Partial', count: statusCounts.partial },
    { value: 'stopped', label: 'Stopped', count: statusCounts.stopped },
  ]

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold">Categories</h1>
        <p className="text-muted-foreground">
          {totalRunning} of {totalServices} services running across {items.length} categories
        </p>
      </div>

      <div className="grid gap-3 sm:grid-cols-3">
        <StatCard icon={Server} label="Categories" value={items.length} />
        <StatCard icon={Play} label="Services Running" value={totalRunning} color="text-green-500" />
        <StatCard icon={Square} label="Services Stopped" value={totalServices - totalRunning} color="text-muted-foreground" />
      </div>

      <ListToolbar
        search={controls.search}
        onSearchChange={controls.setSearch}
        searchPlaceholder="Search categories..."
        sortOptions={sortOptions}
        sortBy={controls.sortBy}
        sortDir={controls.sortDir}
        onSortByChange={controls.setSortBy}
        onSortDirChange={controls.setSortDir}
        viewMode={controls.viewMode}
        onViewModeChange={controls.setViewMode}
      >
        <FilterBar
          options={statusOptions}
          value={controls.filters.status ?? 'all'}
          onChange={(v) => controls.setFilter('status', v)}
        />
      </ListToolbar>

      {controls.filtered.length === 0 ? (
        <EmptyState icon={Server} message="No categories match your filters" />
      ) : controls.viewMode === 'table' ? (
        <CategoryTable categories={controls.filtered} />
      ) : (
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          {controls.filtered.map((category) => (
            <CategoryCard key={category.name} category={category} />
          ))}
        </div>
      )}

      <div className="text-sm text-muted-foreground">
        Showing {controls.filtered.length} of {items.length} categories
      </div>
    </div>
  )
}
