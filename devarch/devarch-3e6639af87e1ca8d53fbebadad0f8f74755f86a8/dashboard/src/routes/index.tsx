import { useMemo } from 'react'
import { createFileRoute, Link } from '@tanstack/react-router'
import { Server, Play, Square, Activity, Loader2, Cpu, MemoryStick, Search } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { ResourceBar } from '@/components/ui/resource-bar'
import { StatCard } from '@/components/ui/stat-card'
import { FilterBar, type FilterOption } from '@/components/ui/filter-bar'
import { ViewSwitcher, type ViewMode } from '@/components/ui/view-switcher'
import { useStatusOverview } from '@/features/status/queries'
import { useServices } from '@/features/services/queries'
import { useStartCategory, useStopCategory } from '@/features/categories/queries'
import { useLocalStorage } from '@/hooks/use-local-storage'
import { useDebounce } from '@/hooks/use-debounce'
import type { CategoryOverview } from '@/types/api'
import { titleCase } from '@/lib/utils'
import { useState } from 'react'

export const Route = createFileRoute('/')({
  component: OverviewPage,
})

function OverviewPage() {
  const { data: status, isLoading } = useStatusOverview()
  const { data: servicesData } = useServices()
  const services = useMemo(() => servicesData?.services ?? [], [servicesData])

  const [viewMode, setViewMode] = useLocalStorage<ViewMode>('overview-view', 'grid')
  const [searchRaw, setSearch] = useState('')
  const search = useDebounce(searchRaw, 200)
  const [statusFilter, setStatusFilter] = useState('all')

  const categories = useMemo(() => status?.categories ?? [], [status])

  const filteredCategories = useMemo(() => {
    let result = categories
    if (search) {
      result = result.filter((c) => c.name.toLowerCase().includes(search.toLowerCase()))
    }
    if (statusFilter !== 'all') {
      result = result.filter((c) => {
        const running = c.running_services ?? 0
        const total = c.total_services ?? 0
        if (statusFilter === 'running') return running === total && total > 0
        if (statusFilter === 'partial') return running > 0 && running < total
        if (statusFilter === 'stopped') return running === 0
        return true
      })
    }
    return result
  }, [categories, search, statusFilter])

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
        <div className="flex flex-wrap items-center gap-3">
          <h2 className="text-lg font-semibold">Categories</h2>
          <div className="relative flex-1 max-w-sm">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 size-4 text-muted-foreground" />
            <Input
              placeholder="Search categories..."
              value={searchRaw}
              onChange={(e) => setSearch(e.target.value)}
              className="pl-9"
            />
          </div>
          <FilterBar
            options={statusOptions}
            value={statusFilter}
            onChange={setStatusFilter}
          />
          <div className="ml-auto">
            <ViewSwitcher value={viewMode} onChange={setViewMode} />
          </div>
        </div>

        {viewMode === 'table' ? (
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
                {filteredCategories.map((category) => (
                  <OverviewCategoryRow key={category.name} category={category} />
                ))}
              </tbody>
            </table>
          </div>
        ) : (
          <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
            {filteredCategories.map((category) => (
              <CategoryQuickCard key={category.name} category={category} />
            ))}
          </div>
        )}

        {filteredCategories.length === 0 && (
          <p className="text-center text-muted-foreground py-8">No categories match your filters</p>
        )}

        <div className="text-sm text-muted-foreground">
          Showing {filteredCategories.length} of {categories.length} categories
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
