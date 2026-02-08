import { useMemo, useState } from 'react'
import { createFileRoute } from '@tanstack/react-router'
import { Loader2, Layers, Plus, CheckCircle2, XCircle } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { useStacks, useDeleteStack, useEnableStack, useDisableStack, useCreateNetwork, useRemoveNetwork } from '@/features/stacks/queries'
import { StackTable } from '@/components/stacks/stack-table'
import { StackGrid } from '@/components/stacks/stack-grid'
import { CreateStackDialog } from '@/components/stacks/create-stack-dialog'
import { ListToolbar } from '@/components/ui/list-toolbar'
import { StatCard } from '@/components/ui/stat-card'
import { EmptyState } from '@/components/ui/empty-state'
import { useListControls } from '@/hooks/use-list-controls'
import type { Stack } from '@/types/api'

export const Route = createFileRoute('/stacks/')({
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
  const stacks = useMemo(() => data ?? [], [data])

  const deleteMutation = useDeleteStack()
  const enableMutation = useEnableStack()
  const disableMutation = useDisableStack()
  const createNetworkMutation = useCreateNetwork()
  const removeNetworkMutation = useRemoveNetwork()

  const [createOpen, setCreateOpen] = useState(false)

  const controls = useListControls({
    storageKey: 'stacks',
    items: stacks,
    searchFn,
    filterFns,
    sortFns,
    defaultSort: 'name',
    defaultView: 'grid',
  })

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

  const handleClone = (name: string) => { void name }

  const handleRename = (name: string) => { void name }

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
      <div className="space-y-6">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-bold">Stacks</h1>
            <p className="text-muted-foreground">No stacks created yet</p>
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
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">Stacks</h1>
          <p className="text-muted-foreground">
            Manage all {stats.total} stack{stats.total !== 1 ? 's' : ''} in your environment
          </p>
        </div>
        <Button size="sm" onClick={handleCreateStack}>
          <Plus className="size-4" /> Create Stack
        </Button>
      </div>

      <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
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
    </div>
  )
}
