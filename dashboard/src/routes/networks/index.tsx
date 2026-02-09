import { useMemo, useState, useCallback } from 'react'
import { createFileRoute } from '@tanstack/react-router'
import { z } from 'zod'
import { Network, AlertTriangle, Boxes, Plus, Trash2 } from 'lucide-react'
import { useNetworks, useCreateNetwork, useRemoveNetwork, useBulkRemoveNetworks } from '@/features/networks/queries'
import { NetworkTable } from '@/components/networks/network-table'
import { NetworkGrid } from '@/components/networks/network-grid'
import { EmptyState } from '@/components/ui/empty-state'
import { ListPageScaffold } from '@/components/ui/list-page-scaffold'
import { PaginationControls } from '@/components/ui/pagination-controls'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { useUrlSyncedListControls } from '@/hooks/use-url-synced-list-controls'
import { useUrlPagination } from '@/hooks/use-url-pagination'
import { DEFAULT_PAGE_SIZE, PAGE_SIZE_OPTIONS } from '@/lib/pagination'
import type { NetworkInfo } from '@/types/api'

export const Route = createFileRoute('/networks/')({
  validateSearch: z.object({
    q: z.string().optional(),
    sort: z.enum(['name', 'stack', 'containers', 'created']).optional(),
    dir: z.enum(['asc', 'desc']).optional(),
    view: z.enum(['table', 'grid']).optional(),
    status: z.enum(['orphaned', 'managed', 'external']).optional(),
    page: z.string().optional(),
    size: z.string().optional(),
  }),
  component: NetworksPage,
})

const searchFn = (n: NetworkInfo, q: string) => {
  const lower = q.toLowerCase()
  return (
    n.name.toLowerCase().includes(lower) ||
    (n.stack_name?.toLowerCase() ?? '').includes(lower)
  )
}

const filterFns = {
  status: (n: NetworkInfo, value: string) => {
    if (value === 'orphaned') return n.orphaned
    if (value === 'managed') return n.managed
    if (value === 'external') return !n.managed
    return true
  },
}

const sortFns: Record<string, (a: NetworkInfo, b: NetworkInfo) => number> = {
  name: (a, b) => a.name.localeCompare(b.name),
  stack: (a, b) => (a.stack_name ?? '').localeCompare(b.stack_name ?? ''),
  containers: (a, b) => a.container_count - b.container_count,
  created: (a, b) => new Date(a.created).getTime() - new Date(b.created).getTime(),
}

const sortOptions = [
  { value: 'name', label: 'Name' },
  { value: 'stack', label: 'Stack' },
  { value: 'containers', label: 'Containers' },
  { value: 'created', label: 'Created' },
]

function NetworksPage() {
  const { data, isLoading } = useNetworks()
  const routeSearch = Route.useSearch()
  const navigate = Route.useNavigate()
  const networks = useMemo(() => data ?? [], [data])

  const createMutation = useCreateNetwork()
  const removeMutation = useRemoveNetwork()
  const bulkRemoveMutation = useBulkRemoveNetworks()

  const [selected, setSelected] = useState<Set<string>>(new Set())
  const [createOpen, setCreateOpen] = useState(false)
  const [newName, setNewName] = useState('')

  const controls = useUrlSyncedListControls(
    { storageKey: 'networks', items: networks, searchFn, filterFns, sortFns, defaultSort: 'name', defaultView: 'table' },
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

  const handleRemove = (name: string) => removeMutation.mutate(name)

  const handleCreate = () => {
    if (!newName.trim()) return
    createMutation.mutate(newName.trim(), {
      onSuccess: () => {
        setCreateOpen(false)
        setNewName('')
      },
    })
  }

  const handleToggleSelect = (name: string) => {
    setSelected((prev) => {
      const next = new Set(prev)
      if (next.has(name)) next.delete(name)
      else next.add(name)
      return next
    })
  }

  const handleToggleAll = () => {
    const pageNames = pagination.pagedItems.map((n) => n.name)
    const allSelected = pageNames.every((name) => selected.has(name))
    if (allSelected) {
      setSelected((prev) => {
        const next = new Set(prev)
        for (const name of pageNames) next.delete(name)
        return next
      })
    } else {
      setSelected((prev) => {
        const next = new Set(prev)
        for (const name of pageNames) next.add(name)
        return next
      })
    }
  }

  const handleBulkRemove = () => {
    const names = Array.from(selected)
    bulkRemoveMutation.mutate(names, {
      onSuccess: () => setSelected(new Set()),
    })
  }

  const stats = useMemo(() => {
    let managed = 0
    let orphaned = 0
    let totalContainers = 0
    for (const n of networks) {
      if (n.managed) managed++
      if (n.orphaned) orphaned++
      totalContainers += n.container_count
    }
    return { total: networks.length, managed, orphaned, totalContainers }
  }, [networks])

  if (networks.length === 0 && !isLoading && !controls.search && !controls.filters.status) {
    return (
      <div className="space-y-5 sm:space-y-6">
        <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
          <div>
            <h1 className="text-xl font-bold sm:text-2xl">Networks</h1>
            <p className="text-sm text-muted-foreground sm:text-base">No networks found</p>
          </div>
          <Button size="sm" onClick={() => setCreateOpen(true)}>
            <Plus className="size-4" /> Create Network
          </Button>
        </div>
        <EmptyState
          icon={Network}
          message="Networks are created automatically when you deploy stacks, or you can create one manually"
          action={{ label: 'Create a network', onClick: () => setCreateOpen(true) }}
        />
        <CreateNetworkDialog
          open={createOpen}
          onOpenChange={setCreateOpen}
          name={newName}
          onNameChange={setNewName}
          onCreate={handleCreate}
          isPending={createMutation.isPending}
        />
      </div>
    )
  }

  return (
    <>
      <ListPageScaffold
        title="Networks"
        subtitle={`Manage all ${stats.total} network${stats.total !== 1 ? 's' : ''} in your environment`}
        isLoading={isLoading}
        statCards={[
          { icon: Network, label: 'Total', value: stats.total },
          { icon: Network, label: 'Managed', value: stats.managed },
          { icon: AlertTriangle, label: 'Orphaned', value: stats.orphaned, color: 'text-yellow-500' },
          { icon: Boxes, label: 'Containers', value: stats.totalContainers },
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
        searchPlaceholder="Search networks..."
        actionButton={
          selected.size > 0 ? (
            <div className="flex gap-2">
              <Button
                size="sm"
                variant="destructive"
                onClick={handleBulkRemove}
                disabled={bulkRemoveMutation.isPending}
              >
                <Trash2 className="size-4" />
                Remove {selected.size}
              </Button>
              <Button size="sm" onClick={() => setCreateOpen(true)}>
                <Plus className="size-4" /> Create
              </Button>
            </div>
          ) : (
            <Button size="sm" className="w-full sm:w-auto" onClick={() => setCreateOpen(true)}>
              <Plus className="size-4" /> Create Network
            </Button>
          )
        }
        emptyIcon={Network}
        emptyMessage="No networks match your filters"
        items={controls.filtered}
        tableView={() => (
          <NetworkTable
            networks={pagination.pagedItems}
            selected={selected}
            onToggleSelect={handleToggleSelect}
            onToggleAll={handleToggleAll}
            onRemove={handleRemove}
          />
        )}
        gridView={() => (
          <NetworkGrid
            networks={pagination.pagedItems}
            onRemove={handleRemove}
          />
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
            itemLabel="networks"
          />
        )}
      </ListPageScaffold>

      <CreateNetworkDialog
        open={createOpen}
        onOpenChange={setCreateOpen}
        name={newName}
        onNameChange={setNewName}
        onCreate={handleCreate}
        isPending={createMutation.isPending}
      />
    </>
  )
}

function CreateNetworkDialog({
  open,
  onOpenChange,
  name,
  onNameChange,
  onCreate,
  isPending,
}: {
  open: boolean
  onOpenChange: (open: boolean) => void
  name: string
  onNameChange: (name: string) => void
  onCreate: () => void
  isPending: boolean
}) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Create Network</DialogTitle>
        </DialogHeader>
        <div className="space-y-4 py-2">
          <Input
            placeholder="Network name"
            value={name}
            onChange={(e) => onNameChange(e.target.value)}
            onKeyDown={(e) => { if (e.key === 'Enter') onCreate() }}
            autoFocus
          />
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>Cancel</Button>
          <Button onClick={onCreate} disabled={!name.trim() || isPending}>
            {isPending ? 'Creating...' : 'Create'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
