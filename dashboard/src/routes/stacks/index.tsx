import { useMemo, useState, useCallback } from 'react'
import { createFileRoute } from '@tanstack/react-router'
import { z } from 'zod'
import { Layers, Plus, CheckCircle2, XCircle } from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  useStacks,
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
import { DeleteStackDialog } from '@/components/stacks/delete-stack-dialog'
import { EmptyState } from '@/components/ui/empty-state'
import { ListPageScaffold } from '@/components/ui/list-page-scaffold'
import { PaginationControls } from '@/components/ui/pagination-controls'
import { useUrlSyncedListControls } from '@/hooks/use-url-synced-list-controls'
import { useUrlPagination } from '@/hooks/use-url-pagination'
import { DEFAULT_PAGE_SIZE, PAGE_SIZE_OPTIONS } from '@/lib/pagination'
import type { Stack } from '@/types/api'

export const Route = createFileRoute('/stacks/')({
  validateSearch: z.object({
    q: z.string().optional(),
    sort: z.enum(['name', 'status', 'instances', 'created']).optional(),
    dir: z.enum(['asc', 'desc']).optional(),
    view: z.enum(['table', 'grid']).optional(),
    page: z.string().optional(),
    size: z.string().optional(),
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
  const [deleteTarget, setDeleteTarget] = useState<Stack | null>(null)

  const controls = useUrlSyncedListControls(
    { storageKey: 'stacks', items: stacks, searchFn, filterFns, sortFns, defaultSort: 'name', defaultView: 'grid' },
    { routeSearch, navigate, sortOptions },
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

  const handleEnable = (name: string) => enableMutation.mutate(name)
  const handleDisable = (name: string) => disableMutation.mutate(name)
  const handleDelete = (name: string) => {
    const stack = stacks.find((s) => s.name === name)
    if (stack) setDeleteTarget(stack)
  }
  const handleCreateNetwork = (name: string) => createNetworkMutation.mutate(name)
  const handleRemoveNetwork = (name: string) => removeNetworkMutation.mutate(name)
  const handleStart = (name: string) => startMutation.mutate(name)
  const handleStop = (name: string) => stopMutation.mutate(name)
  const handleRestart = (name: string) => restartMutation.mutate(name)
  const handleClone = (name: string) => {
    const stack = stacks.find((s) => s.name === name)
    if (stack) setCloneTarget(stack)
  }
  const handleRename = (name: string) => {
    const stack = stacks.find((s) => s.name === name)
    if (stack) setRenameTarget(stack)
  }
  const handleCreateStack = () => setCreateOpen(true)

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
          action={{ label: 'Create your first stack', onClick: handleCreateStack }}
        />
        <CreateStackDialog open={createOpen} onOpenChange={setCreateOpen} />
      </div>
    )
  }

  return (
    <>
      <ListPageScaffold
        title="Stacks"
        subtitle={`Manage all ${stats.total} stack${stats.total !== 1 ? 's' : ''} in your environment`}
        isLoading={isLoading}
        statCards={[
          { icon: Layers, label: 'Total Stacks', value: stats.total },
          { icon: CheckCircle2, label: 'Enabled', value: stats.enabled, color: 'text-green-500' },
          { icon: XCircle, label: 'Disabled', value: stats.disabled, color: 'text-muted-foreground' },
          { icon: Layers, label: 'Total Instances', value: stats.totalInstances },
        ]}
        statGridClassName="grid grid-cols-2 gap-3 sm:grid-cols-2 lg:grid-cols-4"
        controls={{
          ...controls,
          setSearch: handleSearchChange,
          setSortBy: handleSortByChange,
          setSortDir: handleSortDirChange,
          setViewMode: handleViewModeChange,
        }}
        sortOptions={sortOptions}
        searchPlaceholder="Search stacks..."
        actionButton={
          <Button size="sm" className="w-full sm:w-auto" onClick={handleCreateStack}>
            <Plus className="size-4" /> Create Stack
          </Button>
        }
        emptyIcon={Layers}
        emptyMessage="No stacks match your filters"
        items={controls.filtered}
        tableView={() => (
          <StackTable
            stacks={pagination.pagedItems}
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
        )}
        gridView={() => (
          <StackGrid
            stacks={pagination.pagedItems}
            onEnable={handleEnable}
            onDisable={handleDisable}
            onDelete={handleDelete}
            onCreateNetwork={handleCreateNetwork}
            onRemoveNetwork={handleRemoveNetwork}
          />
        )}
        showCount={false}
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
            itemLabel="stacks"
          />
        )}
      </ListPageScaffold>

      <CreateStackDialog open={createOpen} onOpenChange={setCreateOpen} />
      {cloneTarget && (
        <CloneStackDialog
          stack={cloneTarget}
          open={Boolean(cloneTarget)}
          onOpenChange={(open) => { if (!open) setCloneTarget(null) }}
        />
      )}
      {renameTarget && (
        <RenameStackDialog
          stack={renameTarget}
          open={Boolean(renameTarget)}
          onOpenChange={(open) => { if (!open) setRenameTarget(null) }}
        />
      )}
      {deleteTarget && (
        <DeleteStackDialog
          stack={deleteTarget}
          open={Boolean(deleteTarget)}
          onOpenChange={(open) => { if (!open) setDeleteTarget(null) }}
          onSuccess={() => setDeleteTarget(null)}
        />
      )}
    </>
  )
}
