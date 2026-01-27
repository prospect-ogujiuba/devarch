import { useState, useMemo } from 'react'
import { Link } from '@tanstack/react-router'
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
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { StatusBadge } from './status-badge'
import { ActionButton } from './action-button'
import type { Service } from '@/types/api'
import { Search } from 'lucide-react'

function getServiceCategory(s: Service): string {
  return s.category?.name ?? ''
}

function getServiceImage(s: Service): string {
  return `${s.image_name}:${s.image_tag}`
}

function getServiceStatus(s: Service): string {
  return s.status?.status ?? 'stopped'
}

interface ServiceTableProps {
  services: Service[]
  categories: string[]
}

export function ServiceTable({ services, categories }: ServiceTableProps) {
  const [search, setSearch] = useState('')
  const [categoryFilter, setCategoryFilter] = useState<string>('all')
  const [statusFilter, setStatusFilter] = useState<string>('all')

  const filteredServices = useMemo(() => {
    return services.filter((service) => {
      const image = getServiceImage(service)
      const category = getServiceCategory(service)
      const status = getServiceStatus(service)
      const matchesSearch = service.name.toLowerCase().includes(search.toLowerCase()) ||
        image.toLowerCase().includes(search.toLowerCase())
      const matchesCategory = categoryFilter === 'all' || category === categoryFilter
      const matchesStatus = statusFilter === 'all' || status === statusFilter
      return matchesSearch && matchesCategory && matchesStatus
    })
  }, [services, search, categoryFilter, statusFilter])

  return (
    <div className="space-y-4">
      <div className="flex items-center gap-3">
        <div className="relative flex-1 max-w-sm">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 size-4 text-muted-foreground" />
          <Input
            placeholder="Search services..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="pl-9"
          />
        </div>
        <Select value={categoryFilter} onValueChange={setCategoryFilter}>
          <SelectTrigger className="w-[180px]">
            <SelectValue placeholder="Category" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">All Categories</SelectItem>
            {categories.map((cat) => (
              <SelectItem key={cat} value={cat}>
                {cat}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
        <Select value={statusFilter} onValueChange={setStatusFilter}>
          <SelectTrigger className="w-[140px]">
            <SelectValue placeholder="Status" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">All Status</SelectItem>
            <SelectItem value="running">Running</SelectItem>
            <SelectItem value="stopped">Stopped</SelectItem>
            <SelectItem value="starting">Starting</SelectItem>
            <SelectItem value="error">Error</SelectItem>
          </SelectContent>
        </Select>
      </div>

      <div className="rounded-lg border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Name</TableHead>
              <TableHead>Category</TableHead>
              <TableHead>Image</TableHead>
              <TableHead>Ports</TableHead>
              <TableHead>Status</TableHead>
              <TableHead className="text-right">Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {filteredServices.length === 0 ? (
              <TableRow>
                <TableCell colSpan={6} className="text-center text-muted-foreground py-8">
                  No services found
                </TableCell>
              </TableRow>
            ) : (
              filteredServices.map((service) => (
                <TableRow key={service.name} className="cursor-pointer">
                  <TableCell>
                    <Link
                      to="/services/$name"
                      params={{ name: service.name }}
                      className="font-medium hover:underline"
                    >
                      {service.name}
                    </Link>
                  </TableCell>
                  <TableCell>
                    <Badge variant="outline">{getServiceCategory(service)}</Badge>
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
                    <StatusBadge status={getServiceStatus(service)} />
                  </TableCell>
                  <TableCell className="text-right">
                    <ActionButton name={service.name} status={getServiceStatus(service)} size="icon-sm" />
                  </TableCell>
                </TableRow>
              ))
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
