import { useEffect, useMemo, useRef, useState } from 'react'
import { createFileRoute } from '@tanstack/react-router'
import { z } from 'zod'
import { Loader2, Layers, Plus, CheckCircle2, XCircle } from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  useStacks,
  useDeleteStack,
  useEnableStack,
  useDisableStack,
  useCreateNetwork,
  useRemoveNetwork,
  useStartStack,
  useStopStack,
  useRestartStack,
} from '@/features/stacks/queries'
import { StackTable } from '@/components/stacks/stack-table'
import { StackGrid } from '@/components/stacks/stack-grid'
import { CreateStackDialog } from '@/components/stacks/create-stack-dialog'
import { CloneStackDialog } from '@/components/stacks/clone-stack-dialog'
import { RenameStackDialog } from '@/components/stacks/rename-stack-dialog'
import { ListToolbar } from '@/components/ui/list-toolbar'
import { StatCard } from '@/components/ui/stat-card'
import { EmptyState } from '@/components/ui/empty-state'
import { useListControls } from '@/hooks/use-list-controls'
import type { Stack } from '@/types/api'

export const Route = createFileRoute('/stacks/')({
  validateSearch: z.object({
    q: z.string().optional(),
    sort: z.enum(['name', 'status', 'instances', 'created']).optional(),
    dir: z.enum(['asc', 'desc']).optional(),
    view: z.enum(['table', 'grid']).optional(),
  }),
  component: StacksPage,
})

const searchFn = (s: Stack, q: string) => {
  const lower = q.toLowerCase()
  return (
    s.name.toLowerCase().includes(lower) ||
    (s.description?.toLowerCase() ?? '').includes(lower)
  )
}

const filterFns = {
  status: (s: Stack, value: string) => {
    if (value === 'enabled') return s.enabled
    if (value === 'disabled') return !s.enabled
    return true
  },
}

const sortFns: Record<string, (a: Stack, b: Stack) => number> = {
  name: (a, b) => a.name.localeCompare(b.name),
  status: (a, b) => Number(b.enabled) - Number(a.enabled),
  instances: (a, b) => a.instance_count - b.instance_count,
  created: (a, b) => new Date(a.created_at).getTime() - new Date(b.created_at).getTime(),
}

const sortOptions = [
  { value: 'name', label: 'Name' },
  { value: 'status', label: 'Status' },
  { value: 'instances', label: 'Instances' },
  { value: 'created', label: 'Created' },
]

function StacksPage() {
  const { data, isLoading } = useStacks()
  const routeSearch = Route.useSearch()
  const navigate = Route.useNavigate()
  const stacks = useMemo(() => data ?? [], [data])

  const deleteMutation = useDeleteStack()
  const enableMutation = useEnableStack()
  const disableMutation = useDisableStack()
  const createNetworkMutation = useCreateNetwork()
  const removeNetworkMutation = useRemoveNetwork()
  const startMutation = useStartStack()
  const stopMutation = useStopStack()
  const restartMutation = useRestartStack()

  const [createOpen, setCreateOpen] = useState(false)
  const [cloneTarget, setCloneTarget] = useState<Stack | null>(null)
  const [renameTarget, setRenameTarget] = useState<Stack | null>(null)

  const controls = useListControls({
    storageKey: 'stacks',
    items: stacks,
    searchFn,
    filterFns,
    sortFns,
    defaultSort: 'name',
    defaultView: 'grid',
  })

  const {
    search,
    setSearch,
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
    setSortBy(routeSearch.sort ?? 'name')
    setSortDir(routeSearch.dir ?? 'asc')
    setViewMode(routeSearch.view ?? 'grid')
  }, [
    routeSearch.q,
    routeSearch.sort,
    routeSearch.dir,
    routeSearch.view,
    setSearch,
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
    const nextSort =
      sortBy !== 'name' && sortOptions.some((option) => option.value === sortBy)
        ? (sortBy as typeof routeSearch.sort)
        : undefined
    const nextDir = sortDir === 'asc' ? undefined : sortDir
    const nextView = viewMode === 'grid' ? undefined : viewMode

    if (
      routeSearch.q === nextQ
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
        sort: nextSort,
        dir: nextDir,
        view: nextView,
      }),
      replace: true,
    })
  }, [
    search,
    sortBy,
    sortDir,
    viewMode,
    routeSearch.q,
    routeSearch.sort,
    routeSearch.dir,
    routeSearch.view,
    navigate,
  ])

  const stats = useMemo(() => {
    let enabled = 0
    let disabled = 0
    let totalInstances = 0
    for (const s of stacks) {
      if (s.enabled) enabled++
      else disabled++
      totalInstances += s.instance_count
    }
    return { total: stacks.length, enabled, disabled, totalInstances }
  }, [stacks])

  const handleEnable = (name: string) => {
    enableMutation.mutate(name)
  }

  const handleDisable = (name: string) => {
    disableMutation.mutate(name)
  }

  const handleDelete = (name: string) => {
    if (confirm(`Delete stack "${name}"? This will soft-delete the stack and stop all containers.`)) {
      deleteMutation.mutate(name)
    }
  }

  const handleCreateNetwork = (name: string) => {
    createNetworkMutation.mutate(name)
  }

  const handleRemoveNetwork = (name: string) => {
    removeNetworkMutation.mutate(name)
  }

  const handleStart = (name: string) => {
    startMutation.mutate(name)
  }

  const handleStop = (name: string) => {
    stopMutation.mutate(name)
  }

  const handleRestart = (name: string) => {
    restartMutation.mutate(name)
  }

  const handleClone = (name: string) => {
    const stack = stacks.find((s) => s.name === name)
    if (!stack) return
    setCloneTarget(stack)
  }

  const handleRename = (name: string) => {
    const stack = stacks.find((s) => s.name === name)
    if (!stack) return
    setRenameTarget(stack)
  }

  const handleCreateStack = () => {
    setCreateOpen(true)
  }

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <Loader2 className="size-8 animate-spin text-muted-foreground" />
      </div>
    )
  }

  if (stacks.length === 0 && !controls.search && !controls.filters.status) {
    return (
      <div className="space-y-5 sm:space-y-6">
        <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
          <div>
            <h1 className="text-xl font-bold sm:text-2xl">Stacks</h1>
            <p className="text-sm text-muted-foreground sm:text-base">No stacks created yet</p>
          </div>
        </div>
        <EmptyState
          icon={Layers}
          message="Stacks let you group related services into isolated environments"
          action={{
            label: 'Create your first stack',
            onClick: handleCreateStack,
          }}
        />
        <CreateStackDialog open={createOpen} onOpenChange={setCreateOpen} />
      </div>
    )
  }

  return (
    <div className="space-y-5 sm:space-y-6">
      <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 className="text-xl font-bold sm:text-2xl">Stacks</h1>
          <p className="text-sm text-muted-foreground sm:text-base">
            Manage all {stats.total} stack{stats.total !== 1 ? 's' : ''} in your environment
          </p>
        </div>
        <Button size="sm" className="w-full sm:w-auto" onClick={handleCreateStack}>
          <Plus className="size-4" /> Create Stack
        </Button>
      </div>

      <div className="grid grid-cols-2 gap-3 sm:grid-cols-2 lg:grid-cols-4">
        <StatCard icon={Layers} label="Total Stacks" value={stats.total} />
        <StatCard icon={CheckCircle2} label="Enabled" value={stats.enabled} color="text-green-500" />
        <StatCard icon={XCircle} label="Disabled" value={stats.disabled} color="text-muted-foreground" />
        <StatCard icon={Layers} label="Total Instances" value={stats.totalInstances} />
      </div>

      <ListToolbar
        search={controls.search}
        onSearchChange={controls.setSearch}
        searchPlaceholder="Search stacks..."
        sortOptions={sortOptions}
        sortBy={controls.sortBy}
        sortDir={controls.sortDir}
        onSortByChange={controls.setSortBy}
        onSortDirChange={controls.setSortDir}
        viewMode={controls.viewMode}
        onViewModeChange={controls.setViewMode}
      />

      {controls.filtered.length === 0 ? (
        <EmptyState icon={Layers} message="No stacks match your filters" />
      ) : controls.viewMode === 'table' ? (
        <StackTable
          stacks={controls.filtered}
          onEnable={handleEnable}
          onDisable={handleDisable}
          onClone={handleClone}
          onRename={handleRename}
          onStart={handleStart}
          onStop={handleStop}
          onRestart={handleRestart}
          onDelete={handleDelete}
          onCreateNetwork={handleCreateNetwork}
          onRemoveNetwork={handleRemoveNetwork}
        />
      ) : (
        <StackGrid
          stacks={controls.filtered}
          onEnable={handleEnable}
          onDisable={handleDisable}
          onDelete={handleDelete}
          onCreateNetwork={handleCreateNetwork}
          onRemoveNetwork={handleRemoveNetwork}
        />
      )}

      <CreateStackDialog open={createOpen} onOpenChange={setCreateOpen} />
      {cloneTarget && (
        <CloneStackDialog
          stack={cloneTarget}
          open={Boolean(cloneTarget)}
          onOpenChange={(open) => {
            if (!open) setCloneTarget(null)
          }}
        />
      )}
      {renameTarget && (
        <RenameStackDialog
          stack={renameTarget}
          open={Boolean(renameTarget)}
          onOpenChange={(open) => {
            if (!open) setRenameTarget(null)
          }}
        />
      )}
    </div>
  )
}
