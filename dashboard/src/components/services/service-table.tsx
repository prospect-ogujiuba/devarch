import { Link } from '@tanstack/react-router'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Badge } from '@/components/ui/badge'
import { ResourceBar } from '@/components/ui/resource-bar'
import { StatusBadge } from './status-badge'
import { ActionButton } from './action-button'
import { getServiceStatus } from '@/lib/service-utils'
import type { Service } from '@/types/api'
import { titleCase } from '@/lib/utils'

interface ServiceTableProps {
  services: Service[]
  selected: Set<string>
  onToggleSelect: (name: string) => void
}

export function ServiceTable({ services, selected, onToggleSelect }: ServiceTableProps) {
  return (
    <div className="space-y-4">
      <div className="rounded-lg border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead className="w-10">
                <input
                  type="checkbox"
                  checked={services.length > 0 && services.every((s) => selected.has(s.name))}
                  onChange={() => {
                    const allSelected = services.every((s) => selected.has(s.name))
                    for (const s of services) {
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
            {services.length === 0 ? (
              <TableRow>
                <TableCell colSpan={9} className="text-center text-muted-foreground py-8">
                  No services found
                </TableCell>
              </TableRow>
            ) : (
              services.map((service) => {
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
                      <Badge variant="outline">{titleCase(service.category?.name ?? '')}</Badge>
                    </TableCell>
                    <TableCell className="max-w-[200px] truncate text-muted-foreground">
                      {service.image_name}:{service.image_tag}
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
        Showing {services.length} service{services.length !== 1 ? 's' : ''}
      </div>
    </div>
  )
}
