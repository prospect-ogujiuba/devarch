import { useMemo, useCallback, useEffect } from 'react'
import { createFileRoute, Link } from '@tanstack/react-router'
import { z } from 'zod'
import { Loader2, Server, Play, Square, Cpu, MemoryStick, Plus } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { useServices } from '@/features/services/queries'
import { ServiceTable } from '@/components/services/service-table'
import { ServiceGrid } from '@/components/services/service-grid'
import { BulkActionsToolbar } from '@/components/services/bulk-actions-toolbar'
import { ListToolbar } from '@/components/ui/list-toolbar'
import { FilterBar, type FilterOption } from '@/components/ui/filter-bar'
import { StatCard } from '@/components/ui/stat-card'
import { EmptyState } from '@/components/ui/empty-state'
import { useListControls } from '@/hooks/use-list-controls'
import { getServiceStatus } from '@/lib/service-utils'
import { titleCase } from '@/lib/utils'
import type { Service } from '@/types/api'

const servicesSearchSchema = z.object({
  category: z.string().optional(),
  status: z.string().optional(),
})

export const Route = createFileRoute('/services/')({
  component: ServicesPage,
  validateSearch: servicesSearchSchema,
})

const searchFn = (s: Service, q: string) => {
  const lower = q.toLowerCase()
  return (
    s.name.toLowerCase().includes(lower) ||
    `${s.image_name}:${s.image_tag}`.toLowerCase().includes(lower)
  )
}

const filterFns = {
  status: (s: Service, value: string) => getServiceStatus(s) === value,
  category: (s: Service, value: string) => (s.category?.name ?? '') === value,
}

const sortFns: Record<string, (a: Service, b: Service) => number> = {
  name: (a, b) => a.name.localeCompare(b.name),
  status: (a, b) => getServiceStatus(a).localeCompare(getServiceStatus(b)),
  category: (a, b) => (a.category?.name ?? '').localeCompare(b.category?.name ?? ''),
  cpu: (a, b) => (a.metrics?.cpu_percentage ?? 0) - (b.metrics?.cpu_percentage ?? 0),
  memory: (a, b) => (a.metrics?.memory_used_mb ?? 0) - (b.metrics?.memory_used_mb ?? 0),
}

const sortOptions = [
  { value: 'name', label: 'Name' },
  { value: 'status', label: 'Status' },
  { value: 'category', label: 'Category' },
  { value: 'cpu', label: 'CPU' },
  { value: 'memory', label: 'Memory' },
]

function ServicesPage() {
  const { data, isLoading } = useServices()
  const routeSearch = Route.useSearch()
  const services = useMemo(() => data?.services ?? [], [data])
  const total = data?.total ?? 0

  const controls = useListControls({
    storageKey: 'services',
    items: services,
    searchFn,
    filterFns,
    sortFns,
    defaultSort: 'name',
    defaultView: 'table',
  })

  useEffect(() => {
    if (routeSearch.category) controls.setFilter('category', routeSearch.category)
    if (routeSearch.status) controls.setFilter('status', routeSearch.status)
  }, []) // eslint-disable-line react-hooks/exhaustive-deps

  const handleSelectAll = useCallback(() => {
    controls.selectAll(controls.filtered.map((s) => s.name))
  }, [controls])

  const stats = useMemo(() => {
    let running = 0
    let stopped = 0
    let totalCpu = 0
    let totalMem = 0
    for (const s of services) {
      const st = getServiceStatus(s)
      if (st === 'running') running++
      else stopped++
      totalCpu += s.metrics?.cpu_percentage ?? 0
      totalMem += s.metrics?.memory_used_mb ?? 0
    }
    return {
      running,
      stopped,
      avgCpu: services.length > 0 ? totalCpu / services.length : 0,
      totalMem,
    }
  }, [services])

  const categories = useMemo(
    () => [...new Set(services.map((s) => s.category?.name).filter(Boolean))] as string[],
    [services],
  )

  const statusCounts = useMemo(() => {
    const counts = { all: services.length, running: 0, stopped: 0, error: 0 }
    for (const s of services) {
      const st = getServiceStatus(s)
      if (st === 'running') counts.running++
      else if (st === 'error') counts.error++
      else counts.stopped++
    }
    return counts
  }, [services])

  const statusOptions: FilterOption[] = [
    { value: 'all', label: 'All', count: statusCounts.all },
    { value: 'running', label: 'Running', count: statusCounts.running },
    { value: 'stopped', label: 'Stopped', count: statusCounts.stopped },
    { value: 'error', label: 'Error', count: statusCounts.error },
  ]

  const categoryOptions: FilterOption[] = [
    { value: 'all', label: 'All Categories' },
    ...categories.map((cat) => ({ value: cat, label: titleCase(cat) })),
  ]

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <Loader2 className="size-8 animate-spin text-muted-foreground" />
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">Services</h1>
          <p className="text-muted-foreground">
            Manage all {total} services in your environment
          </p>
        </div>
        <Button asChild size="sm">
          <Link to="/services/new"><Plus className="size-4" /> New Service</Link>
        </Button>
      </div>

      <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-5">
        <StatCard icon={Server} label="Total" value={total} />
        <StatCard icon={Play} label="Running" value={stats.running} color="text-green-500" />
        <StatCard icon={Square} label="Stopped" value={stats.stopped} color="text-muted-foreground" />
        <StatCard icon={Cpu} label="Avg CPU" value={`${stats.avgCpu.toFixed(1)}%`} />
        <StatCard icon={MemoryStick} label="Total Memory" value={`${stats.totalMem.toFixed(0)} MB`} />
      </div>

      <BulkActionsToolbar
        selected={controls.selected}
        totalCount={controls.filtered.length}
        onSelectAll={handleSelectAll}
        onClear={controls.clearSelection}
      />

      <ListToolbar
        search={controls.search}
        onSearchChange={controls.setSearch}
        searchPlaceholder="Search services..."
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
        <FilterBar
          options={categoryOptions}
          value={controls.filters.category ?? 'all'}
          onChange={(v) => controls.setFilter('category', v)}
          variant="dropdown"
        />
      </ListToolbar>

      {controls.filtered.length === 0 ? (
        <EmptyState icon={Server} message="No services match your filters" />
      ) : controls.viewMode === 'table' ? (
        <ServiceTable
          services={controls.filtered}
          selected={controls.selected}
          onToggleSelect={controls.toggleSelect}
        />
      ) : (
        <ServiceGrid
          services={controls.filtered}
          selected={controls.selected}
          onToggleSelect={controls.toggleSelect}
        />
      )}
    </div>
  )
}
