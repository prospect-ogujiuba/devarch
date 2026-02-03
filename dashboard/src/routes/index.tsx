import { useMemo } from 'react'
import { createFileRoute, Link } from '@tanstack/react-router'
import { Server, Play, Square, Activity, Loader2, Cpu, MemoryStick, FolderOpen } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { ResourceBar } from '@/components/ui/resource-bar'
import { StatCard } from '@/components/ui/stat-card'
import { FilterBar, type FilterOption } from '@/components/ui/filter-bar'
import { ListToolbar } from '@/components/ui/list-toolbar'
import { EmptyState } from '@/components/ui/empty-state'
import { useStatusOverview } from '@/features/status/queries'
import { useServices } from '@/features/services/queries'
import { useStartCategory, useStopCategory } from '@/features/categories/queries'
import { useListControls } from '@/hooks/use-list-controls'
import type { CategoryOverview } from '@/types/api'
import { titleCase } from '@/lib/utils'

export const Route = createFileRoute('/')({
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

function OverviewPage() {
  const { data: status, isLoading } = useStatusOverview()
  const { data: servicesData } = useServices()
  const services = useMemo(() => servicesData?.services ?? [], [servicesData])
  const categories = useMemo(() => status?.categories ?? [], [status])

  const controls = useListControls({
    storageKey: 'overview',
    items: categories,
    searchFn,
    filterFns,
    sortFns,
    defaultSort: 'name',
    defaultView: 'grid',
  })

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
    <div className="space-y-8">
      <div>
        <h1 className="text-2xl font-bold">Overview</h1>
        <p className="text-muted-foreground">Monitor and manage your development services</p>
      </div>

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-6">
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
          <EmptyState icon={FolderOpen} message="No categories match your filters" />
        ) : controls.viewMode === 'table' ? (
          <div className="rounded-lg border">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b bg-muted/50">
                  <th className="text-left p-3 font-medium">Category</th>
                  <th className="text-left p-3 font-medium">Running</th>
                  <th className="text-left p-3 font-medium">Progress</th>
                  <th className="text-right p-3 font-medium">Actions</th>
                </tr>
              </thead>
              <tbody>
                {controls.filtered.map((category) => (
                  <OverviewCategoryRow key={category.name} category={category} />
                ))}
              </tbody>
            </table>
          </div>
        ) : (
          <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
            {controls.filtered.map((category) => (
              <CategoryQuickCard key={category.name} category={category} />
            ))}
          </div>
        )}

        <div className="text-sm text-muted-foreground">
          Showing {controls.filtered.length} of {categories.length} categories
        </div>
      </div>
    </div>
  )
}

function OverviewCategoryRow({ category }: { category: CategoryOverview }) {
  const startMutation = useStartCategory()
  const stopMutation = useStopCategory()
  const isLoading = startMutation.isPending || stopMutation.isPending
  const running = category.running_services ?? 0
  const total = category.total_services ?? 0
  const pct = total > 0 ? (running / total) * 100 : 0

  return (
    <tr className="border-b last:border-0">
      <td className="p-3">
        <Link to="/services" search={{ category: category.name }} className="font-medium hover:underline">
          {titleCase(category.name)}
        </Link>
      </td>
      <td className="p-3">{running}/{total}</td>
      <td className="p-3">
        <ResourceBar value={pct} className="w-28" />
      </td>
      <td className="p-3 text-right">
        <div className="flex items-center justify-end gap-1">
          {running < total && (
            <Button variant="ghost" size="icon-sm" onClick={() => startMutation.mutate(category.name)} disabled={isLoading}>
              {startMutation.isPending ? <Loader2 className="size-4 animate-spin" /> : <Play className="size-4" />}
            </Button>
          )}
          {running > 0 && (
            <Button variant="ghost" size="icon-sm" onClick={() => stopMutation.mutate(category.name)} disabled={isLoading}>
              {stopMutation.isPending ? <Loader2 className="size-4 animate-spin" /> : <Square className="size-4" />}
            </Button>
          )}
        </div>
      </td>
    </tr>
  )
}

function CategoryQuickCard({ category }: { category: CategoryOverview }) {
  const startMutation = useStartCategory()
  const stopMutation = useStopCategory()
  const isLoading = startMutation.isPending || stopMutation.isPending

  const serviceCount = category.total_services ?? 0
  const runningCount = category.running_services ?? 0
  const allRunning = runningCount === serviceCount && serviceCount > 0
  const allStopped = runningCount === 0

  return (
    <Card className="py-4">
      <CardHeader className="pb-2">
        <div className="flex items-center justify-between">
          <Link to="/services" search={{ category: category.name }} className="hover:underline">
            <CardTitle className="text-base capitalize">{category.name}</CardTitle>
          </Link>
          <span className="text-sm text-muted-foreground">
            {runningCount}/{serviceCount}
          </span>
        </div>
      </CardHeader>
      <CardContent>
        <div className="flex items-center gap-2">
          <ResourceBar
            value={serviceCount > 0 ? (runningCount / serviceCount) * 100 : 0}
            className="flex-1"
          />
          <div className="flex items-center gap-1">
            {!allRunning && (
              <Button variant="ghost" size="icon-sm" onClick={() => startMutation.mutate(category.name)} disabled={isLoading}>
                {startMutation.isPending ? <Loader2 className="size-4 animate-spin" /> : <Play className="size-4" />}
              </Button>
            )}
            {!allStopped && (
              <Button variant="ghost" size="icon-sm" onClick={() => stopMutation.mutate(category.name)} disabled={isLoading}>
                {stopMutation.isPending ? <Loader2 className="size-4 animate-spin" /> : <Square className="size-4" />}
              </Button>
            )}
          </div>
        </div>
      </CardContent>
    </Card>
  )
}
