import { useMemo, useCallback, useState } from 'react'
import { createFileRoute } from '@tanstack/react-router'
import { z } from 'zod'
import { Server, Play, Square, Plus } from 'lucide-react'
import { useCategories } from '@/features/categories/queries'
import { CategoryCard } from '@/components/categories/category-card'
import { CategoryTable } from '@/components/categories/category-table'
import { CreateCategoryDialog } from '@/components/categories/create-category-dialog'
import { FilterBar, type FilterOption } from '@/components/ui/filter-bar'
import { ListPageScaffold } from '@/components/ui/list-page-scaffold'
import { PaginationControls } from '@/components/ui/pagination-controls'
import { Button } from '@/components/ui/button'
import { useUrlSyncedListControls } from '@/hooks/use-url-synced-list-controls'
import { useUrlPagination } from '@/hooks/use-url-pagination'
import { DEFAULT_PAGE_SIZE, PAGE_SIZE_OPTIONS } from '@/lib/pagination'
import type { Category } from '@/types/api'

export const Route = createFileRoute('/categories/')({
  validateSearch: z.object({
    q: z.string().optional(),
    status: z.string().optional(),
    sort: z.enum(['name', 'services', 'running', 'order']).optional(),
    dir: z.enum(['asc', 'desc']).optional(),
    view: z.enum(['table', 'grid']).optional(),
    page: z.string().optional(),
    size: z.string().optional(),
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
  const [createOpen, setCreateOpen] = useState(false)

  const controls = useUrlSyncedListControls(
    { storageKey: 'categories', items, searchFn, filterFns, sortFns, defaultSort: 'name', defaultView: 'grid' },
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

  const totalServices = items.reduce((acc, cat) => acc + (cat.service_count ?? 0), 0)
  const totalRunning = items.reduce((acc, cat) => acc + (cat.runningCount ?? 0), 0)

  const statusOptions: FilterOption[] = [
    { value: 'all', label: 'All', count: statusCounts.all },
    { value: 'running', label: 'All Running', count: statusCounts.running },
    { value: 'partial', label: 'Partial', count: statusCounts.partial },
    { value: 'stopped', label: 'Stopped', count: statusCounts.stopped },
  ]

  return (
    <>
    <ListPageScaffold
      title="Categories"
      subtitle={`${totalRunning} of ${totalServices} services running across ${items.length} categories`}
      isLoading={isLoading}
      actionButton={
        <Button onClick={() => setCreateOpen(true)}>
          <Plus className="size-4" />
          Create
        </Button>
      }
      statCards={[
        { icon: Server, label: 'Categories', value: items.length },
        { icon: Play, label: 'Services Running', value: totalRunning, color: 'text-green-500' },
        { icon: Square, label: 'Services Stopped', value: totalServices - totalRunning, color: 'text-muted-foreground' },
      ]}
      controls={{
        ...controls,
        setSearch: handleSearchChange,
        setSortBy: handleSortByChange,
        setSortDir: handleSortDirChange,
        setViewMode: handleViewModeChange,
      }}
      sortOptions={sortOptions}
      searchPlaceholder="Search categories..."
      emptyIcon={Server}
      emptyMessage="No categories match your filters"
      items={controls.filtered}
      filterChildren={
        <FilterBar
          options={statusOptions}
          value={controls.filters.status ?? 'all'}
          onChange={handleStatusFilterChange}
        />
      }
      tableView={() => <CategoryTable categories={pagination.pagedItems} />}
      gridView={() => (
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          {pagination.pagedItems.map((category) => (
            <CategoryCard key={category.name} category={category} />
          ))}
        </div>
      )}
    >
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
    </ListPageScaffold>
    <CreateCategoryDialog open={createOpen} onOpenChange={setCreateOpen} />
    </>
  )
}
