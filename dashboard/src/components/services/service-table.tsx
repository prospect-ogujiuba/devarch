import { useMemo } from 'react'
import { Link, useNavigate } from '@tanstack/react-router'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { ResourceBar } from '@/components/ui/resource-bar'
import { FilterBar, type FilterOption } from '@/components/ui/filter-bar'
import { StatusBadge } from './status-badge'
import { ActionButton } from './action-button'
import type { Service } from '@/types/api'
import { Search } from 'lucide-react'
import { titleCase } from '@/lib/utils'

function getServiceCategory(s: Service): string {
  return s.category?.name ?? ''
}

function getServiceImage(s: Service): string {
  return `${s.image_name}:${s.image_tag}`
}

function getServiceStatus(s: Service): string {
  const raw = s.status?.status ?? 'stopped'
  if (raw === 'exited' || raw === 'dead' || raw === 'created') return 'stopped'
  return raw
}

interface ServiceTableProps {
  services: Service[]
  categories: string[]
  searchQuery?: string
  categoryFilter?: string
  statusFilter?: string
  sortBy?: string
  sortDir?: 'asc' | 'desc'
  selected: Set<string>
  onToggleSelect: (name: string) => void
}

export function ServiceTable({
  services,
  categories,
  searchQuery = '',
  categoryFilter = 'all',
  statusFilter = 'all',
  sortBy = 'name',
  sortDir = 'asc',
  selected,
  onToggleSelect,
}: ServiceTableProps) {
  const navigate = useNavigate({ from: '/services' })
  const search = searchQuery

  const updateFilter = (key: string, value: string) => {
    navigate({
      search: (prev) => ({
        ...prev,
        [key]: value || undefined,
      }),
    })
  }

  const filteredServices = useMemo(() => {
    const filtered = services.filter((service) => {
      const image = getServiceImage(service)
      const category = getServiceCategory(service)
      const status = getServiceStatus(service)
      const matchesSearch = !search ||
        service.name.toLowerCase().includes(search.toLowerCase()) ||
        image.toLowerCase().includes(search.toLowerCase())
      const matchesCategory = !categoryFilter || categoryFilter === 'all' || category === categoryFilter
      const matchesStatus = !statusFilter || statusFilter === 'all' || status === statusFilter
      return matchesSearch && matchesCategory && matchesStatus
    })

    return [...filtered].sort((a, b) => {
      let cmp = 0
      switch (sortBy) {
        case 'name':
          cmp = a.name.localeCompare(b.name)
          break
        case 'category':
          cmp = getServiceCategory(a).localeCompare(getServiceCategory(b))
          break
        case 'status':
          cmp = getServiceStatus(a).localeCompare(getServiceStatus(b))
          break
        case 'cpu':
          cmp = (a.metrics?.cpu_percentage ?? 0) - (b.metrics?.cpu_percentage ?? 0)
          break
        case 'memory':
          cmp = (a.metrics?.memory_used_mb ?? 0) - (b.metrics?.memory_used_mb ?? 0)
          break
        default:
          cmp = a.name.localeCompare(b.name)
      }
      return sortDir === 'desc' ? -cmp : cmp
    })
  }, [services, search, categoryFilter, statusFilter, sortBy, sortDir])

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

  return (
    <div className="space-y-4">
      <div className="flex flex-wrap items-center gap-3">
        <div className="relative flex-1 max-w-sm">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 size-4 text-muted-foreground" />
          <Input
            placeholder="Search services..."
            value={search}
            onChange={(e) => updateFilter('search', e.target.value)}
            className="pl-9"
          />
        </div>
        <FilterBar
          options={statusOptions}
          value={statusFilter}
          onChange={(v) => updateFilter('status', v === 'all' ? '' : v)}
        />
        <FilterBar
          options={categoryOptions}
          value={categoryFilter}
          onChange={(v) => updateFilter('category', v === 'all' ? '' : v)}
          variant="dropdown"
        />
      </div>

      <div className="rounded-lg border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead className="w-10">
                <input
                  type="checkbox"
                  checked={filteredServices.length > 0 && filteredServices.every((s) => selected.has(s.name))}
                  onChange={() => {
                    const allSelected = filteredServices.every((s) => selected.has(s.name))
                    for (const s of filteredServices) {
                      if (allSelected || !selected.has(s.name)) {
                        onToggleSelect(s.name)
                      }
                    }
                  }}
                  className="size-4 rounded border-muted-foreground"
                />
              </TableHead>
              <TableHead>Name</TableHead>
              <TableHead>Category</TableHead>
              <TableHead>Image</TableHead>
              <TableHead>Ports</TableHead>
              <TableHead>CPU</TableHead>
              <TableHead>Memory</TableHead>
              <TableHead>Status</TableHead>
              <TableHead className="text-right">Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {filteredServices.length === 0 ? (
              <TableRow>
                <TableCell colSpan={9} className="text-center text-muted-foreground py-8">
                  No services found
                </TableCell>
              </TableRow>
            ) : (
              filteredServices.map((service) => {
                const status = getServiceStatus(service)
                const cpuPct = service.metrics?.cpu_percentage ?? 0
                const memPct = service.metrics?.memory_percentage ?? 0
                return (
                  <TableRow key={service.name} className="cursor-pointer">
                    <TableCell>
                      <input
                        type="checkbox"
                        checked={selected.has(service.name)}
                        onChange={() => onToggleSelect(service.name)}
                        className="size-4 rounded border-muted-foreground"
                      />
                    </TableCell>
                    <TableCell>
                      <Link
                        to="/services/$name"
                        params={{ name: service.name }}
                        className="font-medium hover:underline"
                      >
                        {titleCase(service.name)}
                      </Link>
                    </TableCell>
                    <TableCell>
                      <Badge variant="outline">{titleCase(getServiceCategory(service))}</Badge>
                    </TableCell>
                    <TableCell className="max-w-[200px] truncate text-muted-foreground">
                      {getServiceImage(service)}
                    </TableCell>
                    <TableCell>
                      {service.ports && service.ports.length > 0 ? (
                        <span className="text-sm">
                          {service.ports.map((p) => `${p.host_port}:${p.container_port}`).join(', ')}
                        </span>
                      ) : (
                        <span className="text-muted-foreground">-</span>
                      )}
                    </TableCell>
                    <TableCell>
                      {status === 'running' && cpuPct > 0 ? (
                        <ResourceBar value={cpuPct} className="w-20" />
                      ) : (
                        <span className="text-muted-foreground">-</span>
                      )}
                    </TableCell>
                    <TableCell>
                      {status === 'running' && memPct > 0 ? (
                        <ResourceBar value={memPct} className="w-20" />
                      ) : (
                        <span className="text-muted-foreground">-</span>
                      )}
                    </TableCell>
                    <TableCell>
                      <StatusBadge status={status} />
                    </TableCell>
                    <TableCell className="text-right">
                      <ActionButton name={service.name} status={status} size="icon-sm" />
                    </TableCell>
                  </TableRow>
                )
              })
            )}
          </TableBody>
        </Table>
      </div>

      <div className="text-sm text-muted-foreground">
        Showing {filteredServices.length} of {services.length} services
      </div>
    </div>
  )
}
