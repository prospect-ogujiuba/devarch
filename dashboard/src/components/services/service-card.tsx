import { Link } from '@tanstack/react-router'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { StatusBadge } from './status-badge'
import { ActionButton } from './action-button'
import type { Service } from '@/types/api'

interface ServiceCardProps {
  service: Service
}

export function ServiceCard({ service }: ServiceCardProps) {
  return (
    <Card className="py-4 hover:border-primary/50 transition-colors">
      <CardHeader className="pb-2">
        <div className="flex items-start justify-between">
          <Link
            to="/services/$name"
            params={{ name: service.name }}
            className="hover:underline"
          >
            <CardTitle className="text-base">{service.name}</CardTitle>
          </Link>
          <StatusBadge status={service.status} />
        </div>
      </CardHeader>
      <CardContent className="space-y-3">
        <div className="flex items-center gap-2 text-sm text-muted-foreground">
          <Badge variant="outline" className="text-xs">
            {service.category}
          </Badge>
          {service.ports.length > 0 && (
            <span className="text-xs">
              :{service.ports[0].host}
            </span>
          )}
        </div>
        <div className="text-xs text-muted-foreground truncate">
          {service.image}
        </div>
        <div className="pt-2">
          <ActionButton name={service.name} status={service.status} />
        </div>
      </CardContent>
    </Card>
  )
}
