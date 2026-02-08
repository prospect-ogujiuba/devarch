import { Link } from '@tanstack/react-router'
import { MoreHorizontal, Power, PowerOff, Play, Square, RotateCcw, Copy, Edit, Trash2, Globe, Network, Unplug } from 'lucide-react'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { cn } from '@/lib/utils'
import type { Stack } from '@/types/api'

interface StackTableProps {
  stacks: Stack[]
  onEnable: (name: string) => void
  onDisable: (name: string) => void
  onClone: (name: string) => void
  onRename: (name: string) => void
  onStart: (name: string) => void
  onStop: (name: string) => void
  onRestart: (name: string) => void
  onDelete: (name: string) => void
  onCreateNetwork: (name: string) => void
  onRemoveNetwork: (name: string) => void
}

export function StackTable({
  stacks,
  onEnable,
  onDisable,
  onClone,
  onRename,
  onStart,
  onStop,
  onRestart,
  onDelete,
  onCreateNetwork,
  onRemoveNetwork,
}: StackTableProps) {
  return (
    <div className="space-y-4">
      <div className="rounded-lg border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Name</TableHead>
              <TableHead>Description</TableHead>
              <TableHead>Status</TableHead>
              <TableHead>Instances</TableHead>
              <TableHead>Running</TableHead>
              <TableHead>Network</TableHead>
              <TableHead>Created</TableHead>
              <TableHead className="text-right">Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {stacks.length === 0 ? (
              <TableRow>
                <TableCell colSpan={8} className="text-center text-muted-foreground py-8">
                  No stacks found
                </TableCell>
              </TableRow>
            ) : (
              stacks.map((stack) => {
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
                  <TableRow key={stack.name} className="cursor-pointer">
                    <TableCell>
                      <Link
                        to="/stacks/$name"
                        params={{ name: stack.name }}
                        className="font-medium hover:underline"
                      >
                        {stack.name}
                      </Link>
                    </TableCell>
                    <TableCell className="max-w-[300px] truncate text-muted-foreground">
                      {stack.description || '-'}
                    </TableCell>
                    <TableCell>
                      <Badge variant={stack.enabled ? 'default' : 'outline'}>
                        {stack.enabled ? 'Enabled' : 'Disabled'}
                      </Badge>
                    </TableCell>
                    <TableCell>{stack.instance_count}</TableCell>
                    <TableCell className={cn('font-medium', statusColor)}>
                      {runningRatio}
                    </TableCell>
                    <TableCell>
                      {stack.network_name ? (
                        <div className="flex items-center gap-1.5 text-blue-500">
                          <Globe className="size-3" />
                          <span className="font-mono text-xs">{stack.network_name}</span>
                        </div>
                      ) : (
                        <span className="text-muted-foreground">—</span>
                      )}
                    </TableCell>
                    <TableCell className="text-muted-foreground">
                      {new Date(stack.created_at).toLocaleDateString()}
                    </TableCell>
                    <TableCell className="text-right">
                      <DropdownMenu>
                        <DropdownMenuTrigger asChild>
                          <Button variant="ghost" size="icon-sm">
                            <MoreHorizontal className="size-4" />
                          </Button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent align="end">
                          {stack.enabled ? (
                            <DropdownMenuItem onClick={() => onDisable(stack.name)}>
                              <PowerOff className="size-4" />
                              Disable
                            </DropdownMenuItem>
                          ) : (
                            <DropdownMenuItem onClick={() => onEnable(stack.name)}>
                              <Power className="size-4" />
                              Enable
                            </DropdownMenuItem>
                          )}
                          {stack.enabled && stack.running_count === 0 && stack.instance_count > 0 && (
                            <DropdownMenuItem onClick={() => onStart(stack.name)}>
                              <Play className="size-4" />
                              Start
                            </DropdownMenuItem>
                          )}
                          {stack.enabled && stack.running_count > 0 && (
                            <DropdownMenuItem onClick={() => onStop(stack.name)}>
                              <Square className="size-4" />
                              Stop
                            </DropdownMenuItem>
                          )}
                          {stack.enabled && stack.running_count > 0 && (
                            <DropdownMenuItem onClick={() => onRestart(stack.name)}>
                              <RotateCcw className="size-4" />
                              Restart
                            </DropdownMenuItem>
                          )}
                          <DropdownMenuItem onClick={() => onClone(stack.name)}>
                            <Copy className="size-4" />
                            Clone
                          </DropdownMenuItem>
                          <DropdownMenuItem onClick={() => onRename(stack.name)}>
                            <Edit className="size-4" />
                            Rename
                          </DropdownMenuItem>
                          <DropdownMenuItem
                            disabled={stack.network_active}
                            onClick={() => onCreateNetwork(stack.name)}
                          >
                            <Network className="size-4" />
                            Create Network
                          </DropdownMenuItem>
                          <DropdownMenuItem
                            disabled={!stack.network_active}
                            onClick={() => onRemoveNetwork(stack.name)}
                          >
                            <Unplug className="size-4" />
                            Remove Network
                          </DropdownMenuItem>
                          <DropdownMenuSeparator />
                          <DropdownMenuItem onClick={() => onDelete(stack.name)} className="text-destructive">
                            <Trash2 className="size-4" />
                            Delete
                          </DropdownMenuItem>
                        </DropdownMenuContent>
                      </DropdownMenu>
                    </TableCell>
                  </TableRow>
                )
              })
            )}
          </TableBody>
        </Table>
      </div>

      <div className="text-sm text-muted-foreground">
        Showing {stacks.length} stack{stacks.length !== 1 ? 's' : ''}
      </div>
    </div>
  )
}
