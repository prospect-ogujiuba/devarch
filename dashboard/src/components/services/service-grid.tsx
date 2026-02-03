import { Link } from '@tanstack/react-router'
import { Server } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { ResourceBar } from '@/components/ui/resource-bar'
import { StatusBadge } from './status-badge'
import { ActionButton } from './action-button'
import { CopyButton } from '@/components/ui/copy-button'
import { EmptyState } from '@/components/ui/empty-state'
import { getServiceStatus } from '@/lib/service-utils'
import type { Service } from '@/types/api'
import { formatUptime, computeUptime } from '@/lib/format'
import { titleCase, cn } from '@/lib/utils'

interface ServiceGridProps {
  services: Service[]
  selected: Set<string>
  onToggleSelect: (name: string) => void
}

export function ServiceGrid({ services, selected, onToggleSelect }: ServiceGridProps) {
  if (services.length === 0) {
    return <EmptyState icon={Server} message="No services match your filters" />
  }

  return (
    <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
      {services.map((service) => {
        const status = getServiceStatus(service)
        const isSelected = selected.has(service.name)
        const uptime = computeUptime(service.status?.started_at)
        const cpuPct = service.metrics?.cpu_percentage ?? 0
        const memPct = service.metrics?.memory_percentage ?? 0
        const domain = service.domains?.[0]?.domain

        return (
          <Card
            key={service.name}
            className={cn(
              'py-4 hover:border-primary/50 transition-colors relative',
              isSelected && 'border-primary ring-1 ring-primary/30',
            )}
          >
            <div className="absolute top-3 right-3">
              <input
                type="checkbox"
                checked={isSelected}
                onChange={() => onToggleSelect(service.name)}
                className="size-4 rounded border-muted-foreground"
              />
            </div>
            <CardHeader className="pb-2">
              <div className="flex items-start justify-between pr-6">
                <Link to="/services/$name" params={{ name: service.name }} className="hover:underline">
                  <CardTitle className="text-base">{titleCase(service.name)}</CardTitle>
                </Link>
                <StatusBadge status={status} />
              </div>
            </CardHeader>
            <CardContent className="space-y-3">
              <div className="flex items-center gap-2 text-sm text-muted-foreground">
                <Badge variant="outline" className="text-xs">
                  {titleCase(service.category?.name ?? '')}
                </Badge>
                {service.ports && service.ports.length > 0 && (
                  <span className="text-xs">:{service.ports[0].host_port}</span>
                )}
                {status === 'running' && uptime > 0 && (
                  <span className="text-xs">{formatUptime(uptime)}</span>
                )}
              </div>

              {domain && (
                <div className="flex items-center gap-1 text-xs">
                  <a
                    href={`http://${domain}`}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-primary hover:underline truncate"
                  >
                    {domain}
                  </a>
                  <CopyButton value={domain} />
                </div>
              )}

              {status === 'running' && (cpuPct > 0 || memPct > 0) && (
                <div className="space-y-1.5">
                  <ResourceBar value={cpuPct} label="CPU" />
                  <ResourceBar value={memPct} label="Memory" />
                </div>
              )}

              <div className="pt-1">
                <ActionButton name={service.name} status={status} size="icon-sm" />
              </div>
            </CardContent>
          </Card>
        )
      })}
    </div>
  )
}
