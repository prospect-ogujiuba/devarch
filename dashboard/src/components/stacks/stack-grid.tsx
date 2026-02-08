import { Link } from '@tanstack/react-router'
import { Layers, Trash2, Power, PowerOff, Globe, Network, Unplug } from 'lucide-react'
import { Card, CardContent, CardFooter, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { EmptyState } from '@/components/ui/empty-state'
import { cn } from '@/lib/utils'
import type { Stack } from '@/types/api'

interface StackGridProps {
  stacks: Stack[]
  onEnable: (name: string) => void
  onDisable: (name: string) => void
  onDelete: (name: string) => void
  onCreateNetwork: (name: string) => void
  onRemoveNetwork: (name: string) => void
}

export function StackGrid({ stacks, onEnable, onDisable, onDelete, onCreateNetwork, onRemoveNetwork }: StackGridProps) {
  if (stacks.length === 0) {
    return <EmptyState icon={Layers} message="No stacks match your filters" />
  }

  return (
    <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
      {stacks.map((stack) => {
        const runningRatio = stack.instance_count > 0
          ? `${stack.running_count}/${stack.instance_count}`
          : '0/0'

        const statusColor =
          stack.running_count === stack.instance_count && stack.instance_count > 0
            ? 'text-green-500'
            : stack.running_count > 0
            ? 'text-yellow-500'
            : 'text-muted-foreground'

        return (
          <Card
            key={stack.name}
            className="py-4 hover:border-primary/50 transition-colors"
          >
            <CardHeader className="pb-2">
              <div className="flex items-start justify-between">
                <Link
                  to="/stacks/$name"
                  params={{ name: stack.name }}
                  className="hover:underline flex-1"
                >
                  <CardTitle className="text-base">{stack.name}</CardTitle>
                </Link>
                <Badge variant={stack.enabled ? 'default' : 'outline'} className="text-xs">
                  {stack.enabled ? 'Enabled' : 'Disabled'}
                </Badge>
              </div>
            </CardHeader>
            <CardContent className="space-y-3">
              <p className="text-sm text-muted-foreground line-clamp-2">
                {stack.description || 'No description'}
              </p>
              <div className="flex items-center gap-3 text-sm">
                <span className="text-muted-foreground">
                  {stack.instance_count} {stack.instance_count === 1 ? 'instance' : 'instances'}
                </span>
                <span className={cn('font-medium', statusColor)}>
                  {runningRatio} running
                </span>
              </div>
              {stack.network_name && (
                <div className="flex items-center gap-1.5 text-xs text-blue-500">
                  <Globe className="size-3" />
                  <span className="font-mono">{stack.network_name}</span>
                </div>
              )}
            </CardContent>
            <CardFooter className="flex gap-2 pt-2">
              {stack.enabled ? (
                <Button
                  variant="outline"
                  size="sm"
                  onClick={(e) => {
                    e.preventDefault()
                    onDisable(stack.name)
                  }}
                  className="flex-1"
                >
                  <PowerOff className="size-3" />
                  Disable
                </Button>
              ) : (
                <Button
                  variant="outline"
                  size="sm"
                  onClick={(e) => {
                    e.preventDefault()
                    onEnable(stack.name)
                  }}
                  className="flex-1"
                >
                  <Power className="size-3" />
                  Enable
                </Button>
              )}
              <Button
                variant="outline"
                size="icon-sm"
                disabled={stack.network_active}
                onClick={(e) => {
                  e.preventDefault()
                  onCreateNetwork(stack.name)
                }}
                title="Create network"
              >
                <Network className="size-3" />
              </Button>
              <Button
                variant="outline"
                size="icon-sm"
                disabled={!stack.network_active}
                onClick={(e) => {
                  e.preventDefault()
                  onRemoveNetwork(stack.name)
                }}
                title="Remove network"
              >
                <Unplug className="size-3" />
              </Button>
              <Button
                variant="outline"
                size="icon-sm"
                onClick={(e) => {
                  e.preventDefault()
                  onDelete(stack.name)
                }}
              >
                <Trash2 className="size-3" />
              </Button>
            </CardFooter>
          </Card>
        )
      })}
    </div>
  )
}
