import { Link } from '@tanstack/react-router'
import { Trash2, AlertTriangle, Shield } from 'lucide-react'
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
import { Checkbox } from '@/components/ui/checkbox'
import type { NetworkInfo } from '@/types/api'

interface NetworkTableProps {
  networks: NetworkInfo[]
  selected: Set<string>
  onToggleSelect: (name: string) => void
  onToggleAll: () => void
  onRemove: (name: string) => void
}

export function NetworkTable({ networks, selected, onToggleSelect, onToggleAll, onRemove }: NetworkTableProps) {
  const allSelected = networks.length > 0 && networks.every((n) => selected.has(n.name))

  return (
    <div className="space-y-4">
      <div className="rounded-lg border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead className="w-10">
                <Checkbox
                  checked={allSelected}
                  onCheckedChange={onToggleAll}
                />
              </TableHead>
              <TableHead>Name</TableHead>
              <TableHead>Stack</TableHead>
              <TableHead>Driver</TableHead>
              <TableHead>Containers</TableHead>
              <TableHead>Created</TableHead>
              <TableHead className="text-right">Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {networks.length === 0 ? (
              <TableRow>
                <TableCell colSpan={7} className="text-center text-muted-foreground py-8">
                  No networks found
                </TableCell>
              </TableRow>
            ) : (
              networks.map((net) => (
                <TableRow key={net.name}>
                  <TableCell>
                    <Checkbox
                      checked={selected.has(net.name)}
                      onCheckedChange={() => onToggleSelect(net.name)}
                      disabled={net.name === 'podman'}
                    />
                  </TableCell>
                  <TableCell>
                    <div className="flex items-center gap-2">
                      <span className="font-mono text-sm">{net.name}</span>
                      {!net.managed && (
                        <Badge variant="outline" className="text-muted-foreground text-xs">
                          <Shield className="size-3 mr-1" />
                          External
                        </Badge>
                      )}
                    </div>
                  </TableCell>
                  <TableCell>
                    {net.stack_name ? (
                      net.orphaned ? (
                        <div className="flex items-center gap-1.5">
                          <span className="text-muted-foreground">{net.stack_name}</span>
                          <Badge variant="outline" className="text-yellow-600 border-yellow-600">
                            <AlertTriangle className="size-3 mr-1" />
                            Orphaned
                          </Badge>
                        </div>
                      ) : (
                        <Link
                          to="/stacks/$name"
                          params={{ name: net.stack_name }}
                          className="font-medium hover:underline"
                        >
                          {net.stack_name}
                        </Link>
                      )
                    ) : (
                      <span className="text-muted-foreground">-</span>
                    )}
                  </TableCell>
                  <TableCell className="text-muted-foreground">{net.driver}</TableCell>
                  <TableCell>{net.container_count}</TableCell>
                  <TableCell className="text-muted-foreground">
                    {net.created && net.created !== '0001-01-01T00:00:00Z'
                      ? new Date(net.created).toLocaleDateString()
                      : '-'}
                  </TableCell>
                  <TableCell className="text-right">
                    <Button
                      variant="ghost"
                      size="icon-sm"
                      disabled={net.container_count > 0 || net.name === 'podman'}
                      onClick={() => onRemove(net.name)}
                      title={net.name === 'podman' ? 'Default network' : net.container_count > 0 ? 'Has connected containers' : 'Remove network'}
                    >
                      <Trash2 className="size-4" />
                    </Button>
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </div>

      <div className="text-sm text-muted-foreground">
        Showing {networks.length} network{networks.length !== 1 ? 's' : ''}
      </div>
    </div>
  )
}
