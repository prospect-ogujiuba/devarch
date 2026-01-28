import { Link } from '@tanstack/react-router'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { ResourceBar } from '@/components/ui/resource-bar'
import { StatusBadge } from './status-badge'
import { ActionButton } from './action-button'
import { CopyButton } from '@/components/ui/copy-button'
import type { Service } from '@/types/api'
import { formatUptime, computeUptime } from '@/lib/format'
import { titleCase } from '@/lib/utils'

interface ServiceCardProps {
  service: Service
}

function getStatus(s: Service): string {
  const raw = s.status?.status ?? 'stopped'
  if (raw === 'exited' || raw === 'dead' || raw === 'created') return 'stopped'
  return raw
}

export function ServiceCard({ service }: ServiceCardProps) {
  const status = getStatus(service)
  const uptime = computeUptime(service.status?.started_at)
  const cpuPct = service.metrics?.cpu_percentage ?? 0
  const memPct = service.metrics?.memory_percentage ?? 0
  const domain = service.domains?.[0]?.domain

  return (
    <Card className="py-4 hover:border-primary/50 transition-colors">
      <CardHeader className="pb-2">
        <div className="flex items-start justify-between">
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
            <a href={`http://${domain}`} target="_blank" rel="noopener noreferrer" className="text-primary hover:underline truncate">
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

        <div className="text-xs text-muted-foreground truncate">
          {service.image_name}:{service.image_tag}
        </div>
        <div className="pt-2">
          <ActionButton name={service.name} status={status} size="icon-sm" />
        </div>
      </CardContent>
    </Card>
  )
}
