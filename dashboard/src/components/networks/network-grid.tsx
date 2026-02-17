import { Link } from '@tanstack/react-router'
import { Network, Trash2, AlertTriangle, Shield } from 'lucide-react'
import { Card, CardContent, CardFooter, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Checkbox } from '@/components/ui/checkbox'
import { EmptyState } from '@/components/ui/empty-state'
import type { NetworkInfo } from '@/types/api'

interface NetworkGridProps {
  networks: NetworkInfo[]
  selected: Set<string>
  onToggleSelect: (name: string) => void
  onRemove: (name: string) => void
}

export function NetworkGrid({ networks, selected, onToggleSelect, onRemove }: NetworkGridProps) {
  if (networks.length === 0) {
    return <EmptyState icon={Network} message="No networks match your filters" />
  }

  return (
    <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
      {networks.map((net) => (
        <Card key={net.name} className="py-4 hover:border-primary/50 transition-colors">
          <CardHeader className="pb-2">
            <div className="flex items-start justify-between gap-2">
              <Checkbox
                checked={selected.has(net.name)}
                onCheckedChange={() => onToggleSelect(net.name)}
                onClick={(e) => e.stopPropagation()}
              />
              <CardTitle className="text-base font-mono truncate flex-1">{net.name}</CardTitle>
              <div className="flex gap-1 shrink-0">
                {!net.managed && (
                  <Badge variant="outline" className="text-muted-foreground text-xs">
                    <Shield className="size-3 mr-1" />
                    External
                  </Badge>
                )}
                {net.orphaned && (
                  <Badge variant="outline" className="text-yellow-600 border-yellow-600 text-xs">
                    <AlertTriangle className="size-3 mr-1" />
                    Orphaned
                  </Badge>
                )}
              </div>
            </div>
          </CardHeader>
          <CardContent className="space-y-3">
            <div className="flex items-center gap-2 text-sm">
              {net.stack_name ? (
                net.orphaned ? (
                  <span className="text-muted-foreground">{net.stack_name}</span>
                ) : (
                  <Link
                    to="/stacks/$name"
                    params={{ name: net.stack_name }}
                    className="text-blue-500 hover:underline"
                  >
                    {net.stack_name}
                  </Link>
                )
              ) : (
                <span className="text-muted-foreground">No stack</span>
              )}
            </div>
            <div className="flex items-center gap-3 text-sm text-muted-foreground">
              <span>{net.driver}</span>
              <span>{net.container_count} container{net.container_count !== 1 ? 's' : ''}</span>
            </div>
            {net.created && net.created !== '0001-01-01T00:00:00Z' && (
              <div className="text-xs text-muted-foreground">
                Created {new Date(net.created).toLocaleDateString()}
              </div>
            )}
          </CardContent>
          <CardFooter className="pt-2">
            <Button
              variant="outline"
              size="sm"
              disabled={net.container_count > 0 || net.name === 'podman'}
              onClick={(e) => {
                e.preventDefault()
                onRemove(net.name)
              }}
              title={net.name === 'podman' ? 'Default network' : net.container_count > 0 ? 'Has connected containers' : 'Remove network'}
              className="w-full"
            >
              <Trash2 className="size-3" />
              Remove
            </Button>
          </CardFooter>
        </Card>
      ))}
    </div>
  )
}
