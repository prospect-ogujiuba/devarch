import { useEffect, useMemo, useRef } from 'react'
import { createFileRoute } from '@tanstack/react-router'
import { z } from 'zod'
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
  validateSearch: z.object({
    q: z.string().optional(),
    status: z.string().optional(),
    sort: z.enum(['name', 'services', 'running', 'order']).optional(),
    dir: z.enum(['asc', 'desc']).optional(),
    view: z.enum(['table', 'grid']).optional(),
  }),
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
  const routeSearch = Route.useSearch()
  const navigate = Route.useNavigate()
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

  const {
    search,
    setSearch,
    filters,
    setFilter,
    sortBy,
    setSortBy,
    sortDir,
    setSortDir,
    viewMode,
    setViewMode,
  } = controls
  const syncingFromUrlRef = useRef(false)

  useEffect(() => {
    syncingFromUrlRef.current = true
    setSearch(routeSearch.q ?? '')
    setFilter('status', routeSearch.status ?? 'all')
    setSortBy(routeSearch.sort ?? 'order')
    setSortDir(routeSearch.dir ?? 'asc')
    setViewMode(routeSearch.view ?? 'grid')
  }, [
    routeSearch.q,
    routeSearch.status,
    routeSearch.sort,
    routeSearch.dir,
    routeSearch.view,
    setSearch,
    setFilter,
    setSortBy,
    setSortDir,
    setViewMode,
  ])

  useEffect(() => {
    if (syncingFromUrlRef.current) {
      syncingFromUrlRef.current = false
      return
    }

    const nextQ = search || undefined
    const nextStatus = filters.status && filters.status !== 'all' ? filters.status : undefined
    const nextSort =
      sortBy !== 'order' && sortOptions.some((option) => option.value === sortBy)
        ? (sortBy as typeof routeSearch.sort)
        : undefined
    const nextDir = sortDir === 'asc' ? undefined : sortDir
    const nextView = viewMode === 'grid' ? undefined : viewMode

    if (
      routeSearch.q === nextQ
      && routeSearch.status === nextStatus
      && routeSearch.sort === nextSort
      && routeSearch.dir === nextDir
      && routeSearch.view === nextView
    ) {
      return
    }

    navigate({
      search: (prev) => ({
        ...prev,
        q: nextQ,
        status: nextStatus,
        sort: nextSort,
        dir: nextDir,
        view: nextView,
      }),
      replace: true,
    })
  }, [
    search,
    filters.status,
    sortBy,
    sortDir,
    viewMode,
    routeSearch.q,
    routeSearch.status,
    routeSearch.sort,
    routeSearch.dir,
    routeSearch.view,
    navigate,
  ])

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
    <div className="space-y-5 sm:space-y-6">
      <div>
        <h1 className="text-xl font-bold sm:text-2xl">Categories</h1>
        <p className="text-sm text-muted-foreground sm:text-base">
          {totalRunning} of {totalServices} services running across {items.length} categories
        </p>
      </div>

      <div className="grid grid-cols-2 gap-3 sm:grid-cols-3">
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
