import { useMemo, useCallback } from 'react'
import { createFileRoute } from '@tanstack/react-router'
import { z } from 'zod'
import { Server, Play, Square, Activity, Loader2, Cpu, MemoryStick, FolderOpen } from 'lucide-react'
import { StatCard } from '@/components/ui/stat-card'
import { FilterBar, type FilterOption } from '@/components/ui/filter-bar'
import { ListToolbar } from '@/components/ui/list-toolbar'
import { EmptyState } from '@/components/ui/empty-state'
import { PaginationControls } from '@/components/ui/pagination-controls'
import { CategoryCard } from '@/components/categories/category-card'
import { CategoryTable } from '@/components/categories/category-table'
import { useStatusOverview } from '@/features/status/queries'
import { useServices } from '@/features/services/queries'
import { useUrlSyncedListControls } from '@/hooks/use-url-synced-list-controls'
import { useUrlPagination } from '@/hooks/use-url-pagination'
import { DEFAULT_PAGE_SIZE, PAGE_SIZE_OPTIONS } from '@/lib/pagination'
import type { CategoryOverview, Category } from '@/types/api'

export const Route = createFileRoute('/')({
  validateSearch: z.object({
    q: z.string().optional(),
    status: z.string().optional(),
    sort: z.enum(['name', 'services', 'running']).optional(),
    dir: z.enum(['asc', 'desc']).optional(),
    view: z.enum(['table', 'grid']).optional(),
    page: z.string().optional(),
    size: z.string().optional(),
  }),
  component: OverviewPage,
})

const searchFn = (c: CategoryOverview, q: string) =>
  c.name.toLowerCase().includes(q.toLowerCase())

const filterFns = {
  status: (c: CategoryOverview, value: string) => {
    const running = c.running_services ?? 0
    const total = c.total_services ?? 0
    if (value === 'running') return running === total && total > 0
    if (value === 'partial') return running > 0 && running < total
    if (value === 'stopped') return running === 0
    return true
  },
}

const sortFns: Record<string, (a: CategoryOverview, b: CategoryOverview) => number> = {
  name: (a, b) => a.name.localeCompare(b.name),
  services: (a, b) => (a.total_services ?? 0) - (b.total_services ?? 0),
  running: (a, b) => (a.running_services ?? 0) - (b.running_services ?? 0),
}

const sortOptions = [
  { value: 'name', label: 'Name' },
  { value: 'services', label: 'Services' },
  { value: 'running', label: 'Running' },
]

function toCategory(c: CategoryOverview): Category {
  return {
    id: 0,
    name: c.name,
    startup_order: 0,
    service_count: c.total_services ?? 0,
    runningCount: c.running_services ?? 0,
  }
}

function OverviewPage() {
  const { data: status, isLoading } = useStatusOverview()
  const routeSearch = Route.useSearch()
  const navigate = Route.useNavigate()
  const { data: servicesData } = useServices()
  const services = useMemo(() => servicesData?.services ?? [], [servicesData])
  const categories = useMemo(() => status?.categories ?? [], [status])

  const controls = useUrlSyncedListControls(
    { storageKey: 'overview', items: categories, searchFn, filterFns, sortFns, defaultSort: 'name', defaultView: 'grid' },
    { routeSearch, navigate, sortOptions, filterKeys: ['status'] },
  )
  const pagination = useUrlPagination({
    items: controls.filtered,
    routeSearch,
    navigate,
    defaultPageSize: DEFAULT_PAGE_SIZE,
    pageSizeOptions: PAGE_SIZE_OPTIONS,
  })

  const handleSearchChange = useCallback((value: string) => {
    pagination.resetPage()
    controls.setSearch(value)
  }, [controls, pagination])

  const handleSortByChange = useCallback((value: string) => {
    pagination.resetPage()
    controls.setSortBy(value)
  }, [controls, pagination])

  const handleSortDirChange = useCallback((dir: 'asc' | 'desc') => {
    pagination.resetPage()
    controls.setSortDir(dir)
  }, [controls, pagination])

  const handleViewModeChange = useCallback((mode: 'table' | 'grid') => {
    pagination.resetPage()
    controls.setViewMode(mode)
  }, [controls, pagination])

  const handleStatusFilterChange = useCallback((value: string) => {
    pagination.resetPage()
    controls.setFilter('status', value)
  }, [controls, pagination])

  const serviceStats = useMemo(() => {
    let totalCpu = 0
    let totalMem = 0
    let runningWithMetrics = 0
    for (const s of services) {
      if (s.metrics) {
        totalCpu += s.metrics.cpu_percentage
        totalMem += s.metrics.memory_used_mb
        runningWithMetrics++
      }
    }
    return {
      avgCpu: runningWithMetrics > 0 ? totalCpu / runningWithMetrics : 0,
      totalMem,
    }
  }, [services])

  const statusCounts = useMemo(() => {
    let running = 0
    let partial = 0
    let stopped = 0
    for (const c of categories) {
      const r = c.running_services ?? 0
      const t = c.total_services ?? 0
      if (r === t && t > 0) running++
      else if (r > 0) partial++
      else stopped++
    }
    return { all: categories.length, running, partial, stopped }
  }, [categories])

  const statusOptions: FilterOption[] = [
    { value: 'all', label: 'All', count: statusCounts.all },
    { value: 'running', label: 'All Running', count: statusCounts.running },
    { value: 'partial', label: 'Partial', count: statusCounts.partial },
    { value: 'stopped', label: 'Stopped', count: statusCounts.stopped },
  ]

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <Loader2 className="size-8 animate-spin text-muted-foreground" />
      </div>
    )
  }

  return (
    <div className="space-y-5 sm:space-y-8">
      <div>
        <h1 className="text-xl font-bold sm:text-2xl">Overview</h1>
        <p className="text-sm text-muted-foreground sm:text-base">Monitor and manage your development services</p>
      </div>

      <div className="grid grid-cols-2 gap-3 sm:grid-cols-3 lg:grid-cols-6">
        <StatCard icon={Server} label="Total Services" value={status?.total_services ?? 0} />
        <StatCard icon={Play} label="Running" value={status?.running_services ?? 0} color="text-green-500" />
        <StatCard icon={Square} label="Stopped" value={status?.stopped_services ?? 0} color="text-muted-foreground" />
        <StatCard icon={Activity} label="Categories" value={categories.length} />
        <StatCard icon={Cpu} label="Avg CPU" value={`${serviceStats.avgCpu.toFixed(1)}%`} />
        <StatCard icon={MemoryStick} label="Total Memory" value={`${serviceStats.totalMem.toFixed(0)} MB`} />
      </div>

      <div className="space-y-4">
        <h2 className="text-lg font-semibold">Categories</h2>

        <ListToolbar
          search={controls.search}
          onSearchChange={handleSearchChange}
          searchPlaceholder="Search categories..."
          sortOptions={sortOptions}
          sortBy={controls.sortBy}
          sortDir={controls.sortDir}
          onSortByChange={handleSortByChange}
          onSortDirChange={handleSortDirChange}
          viewMode={controls.viewMode}
          onViewModeChange={handleViewModeChange}
        >
          <FilterBar
            options={statusOptions}
            value={controls.filters.status ?? 'all'}
            onChange={handleStatusFilterChange}
          />
        </ListToolbar>

        {controls.filtered.length === 0 ? (
          <EmptyState icon={FolderOpen} message="No categories match your filters" />
        ) : controls.viewMode === 'table' ? (
          <CategoryTable categories={pagination.pagedItems.map(toCategory)} compact />
        ) : (
          <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
            {pagination.pagedItems.map((category) => (
              <CategoryCard key={category.name} category={toCategory(category)} compact />
            ))}
          </div>
        )}

        {pagination.totalItems > 0 && (
          <PaginationControls
            page={pagination.page}
            totalPages={pagination.totalPages}
            totalItems={pagination.totalItems}
            pageSize={pagination.pageSize}
            pageSizeOptions={PAGE_SIZE_OPTIONS}
            onPageChange={pagination.setPage}
            onPageSizeChange={pagination.setPageSize}
            itemLabel="categories"
          />
        )}
      </div>
    </div>
  )
}
