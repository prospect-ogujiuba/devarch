import { useState, useMemo, useCallback } from 'react'
import { createFileRoute } from '@tanstack/react-router'
import { Loader2, Server, Play, Square, Cpu, MemoryStick } from 'lucide-react'
import { useServices } from '@/features/services/queries'
import { ServiceTable } from '@/components/services/service-table'
import { ServiceGrid } from '@/components/services/service-grid'
import { BulkActionsToolbar } from '@/components/services/bulk-actions-toolbar'
import { ViewSwitcher, type ViewMode } from '@/components/ui/view-switcher'
import { SortControls, type SortOption } from '@/components/ui/sort-controls'
import { StatCard } from '@/components/ui/stat-card'
import { useLocalStorage } from '@/hooks/use-local-storage'
import { z } from 'zod'

const servicesSearchSchema = z.object({
  search: z.string().optional(),
  category: z.string().optional(),
  status: z.string().optional(),
})

export const Route = createFileRoute('/services/')({
  component: ServicesPage,
  validateSearch: servicesSearchSchema,
})

const sortOptions: SortOption[] = [
  { value: 'name', label: 'Name' },
  { value: 'status', label: 'Status' },
  { value: 'category', label: 'Category' },
  { value: 'cpu', label: 'CPU' },
  { value: 'memory', label: 'Memory' },
]

function ServicesPage() {
  const { data, isLoading } = useServices()
  const search = Route.useSearch()

  const services = useMemo(() => data?.services ?? [], [data])
  const total = data?.total ?? 0
  const categories = useMemo(() => [...new Set(services.map((s) => s.category?.name).filter(Boolean))] as string[], [services])

  const [viewMode, setViewMode] = useLocalStorage<ViewMode>('services-view', 'table')
  const [sortBy, setSortBy] = useLocalStorage('services-sort', 'name')
  const [sortDir, setSortDir] = useLocalStorage<'asc' | 'desc'>('services-sort-dir', 'asc')
  const [selected, setSelected] = useState<Set<string>>(new Set())

  const toggleSelect = useCallback((name: string) => {
    setSelected((prev) => {
      const next = new Set(prev)
      if (next.has(name)) next.delete(name)
      else next.add(name)
      return next
    })
  }, [])

  const selectAll = useCallback(() => {
    setSelected(new Set(services.map((s) => s.name)))
  }, [services])

  const clearSelection = useCallback(() => {
    setSelected(new Set())
  }, [])

  const stats = useMemo(() => {
    let running = 0
    let stopped = 0
    let totalCpu = 0
    let totalMem = 0
    for (const s of services) {
      const st = s.status?.status ?? 'stopped'
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
        <div className="flex items-center gap-2">
          <SortControls
            options={sortOptions}
            sortBy={sortBy}
            sortDir={sortDir}
            onSortByChange={setSortBy}
            onSortDirChange={setSortDir}
          />
          <ViewSwitcher value={viewMode} onChange={setViewMode} />
        </div>
      </div>

      <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-5">
        <StatCard icon={Server} label="Total" value={total} />
        <StatCard icon={Play} label="Running" value={stats.running} color="text-green-500" />
        <StatCard icon={Square} label="Stopped" value={stats.stopped} color="text-muted-foreground" />
        <StatCard icon={Cpu} label="Avg CPU" value={`${stats.avgCpu.toFixed(1)}%`} />
        <StatCard icon={MemoryStick} label="Total Memory" value={`${stats.totalMem.toFixed(0)} MB`} />
      </div>

      <BulkActionsToolbar
        selected={selected}
        totalCount={services.length}
        onSelectAll={selectAll}
        onClear={clearSelection}
      />

      {viewMode === 'table' ? (
        <ServiceTable
          services={services}
          categories={categories}
          searchQuery={search.search}
          categoryFilter={search.category}
          statusFilter={search.status}
          sortBy={sortBy}
          sortDir={sortDir}
          selected={selected}
          onToggleSelect={toggleSelect}
        />
      ) : (
        <ServiceGrid
          services={services}
          selected={selected}
          onToggleSelect={toggleSelect}
        />
      )}
    </div>
  )
}
