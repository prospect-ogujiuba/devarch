import { useState } from 'react'
import { Loader2, Cable, Unplug, AlertTriangle, Plus } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { useStackWires, useResolveWires, useDeleteWire } from '@/features/stacks/queries'
import { CreateWireDialog } from './create-wire-dialog'

interface WiringTabProps {
  stackName: string
}

export function WiringTab({ stackName }: WiringTabProps) {
  const { data, isLoading } = useStackWires(stackName)
  const resolveWires = useResolveWires(stackName)
  const deleteWire = useDeleteWire(stackName)
  const [createWireOpen, setCreateWireOpen] = useState(false)

  const wires = data?.wires ?? []
  const unresolved = data?.unresolved ?? []

  const handleDisconnect = (wireId: number) => {
    deleteWire.mutate(wireId)
  }

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-12">
        <Loader2 className="size-8 animate-spin text-muted-foreground" />
      </div>
    )
  }

  if (wires.length === 0 && unresolved.length === 0) {
    return (
      <Card>
        <CardContent className="py-12">
          <div className="flex flex-col items-center justify-center text-center">
            <div className="rounded-full bg-muted p-3 mb-4">
              <Cable className="size-8 text-muted-foreground" />
            </div>
            <h3 className="font-medium mb-1">No wiring needed</h3>
            <p className="text-sm text-muted-foreground">
              This stack has no import contracts
            </p>
          </div>
        </CardContent>
      </Card>
    )
  }

  return (
    <div className="space-y-4">
      <Card>
        <CardHeader>
          <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
            <CardTitle className="text-base">Wiring</CardTitle>
            <div className="flex w-full items-center gap-2 sm:w-auto">
              <Button
                size="sm"
                variant="outline"
                className="flex-1 sm:flex-none"
                onClick={() => resolveWires.mutate()}
                disabled={resolveWires.isPending}
              >
                {resolveWires.isPending ? (
                  <Loader2 className="size-4 animate-spin" />
                ) : (
                  <Cable className="size-4" />
                )}
                Resolve
              </Button>
              <Button
                size="sm"
                className="flex-1 sm:flex-none"
                onClick={() => setCreateWireOpen(true)}
              >
                <Plus className="size-4" />
                Add Wire
              </Button>
            </div>
          </div>
        </CardHeader>
        <CardContent>
          {wires.length > 0 ? (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Consumer</TableHead>
                  <TableHead className="w-12"></TableHead>
                  <TableHead>Provider</TableHead>
                  <TableHead>Contract</TableHead>
                  <TableHead>Source</TableHead>
                  <TableHead className="w-20">Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {wires.map((wire) => (
                  <TableRow key={wire.id}>
                    <TableCell className="font-mono text-sm">
                      {wire.consumer_instance_name}
                    </TableCell>
                    <TableCell className="text-center text-muted-foreground">
                      →
                    </TableCell>
                    <TableCell className="font-mono text-sm">
                      {wire.provider_instance_name}
                    </TableCell>
                    <TableCell>
                      <div className="flex items-center gap-2">
                        <span className="font-medium text-sm">{wire.contract_name}</span>
                        <Badge variant="outline" className="text-xs">
                          {wire.consumer_contract_type}
                        </Badge>
                      </div>
                    </TableCell>
                    <TableCell>
                      <Badge
                        variant={wire.source === 'explicit' ? 'default' : 'secondary'}
                        className="text-xs"
                      >
                        {wire.source}
                      </Badge>
                    </TableCell>
                    <TableCell>
                      <Button
                        size="sm"
                        variant="ghost"
                        onClick={() => handleDisconnect(wire.id)}
                        disabled={deleteWire.isPending}
                      >
                        <Unplug className="size-4" />
                      </Button>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          ) : (
            <div className="text-center py-8 text-muted-foreground">
              <p>No active wires</p>
            </div>
          )}
        </CardContent>
      </Card>

      {unresolved.length > 0 && (
        <Card className="border-l-4 border-l-amber-500">
          <CardHeader>
            <CardTitle className="text-base flex items-center gap-2">
              <AlertTriangle className="size-4 text-amber-500" />
              Unresolved Contracts ({unresolved.length})
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            {unresolved.map((contract, idx) => (
              <div
                key={idx}
                className="bg-amber-50 dark:bg-amber-950/20 px-4 py-3 rounded-md space-y-2"
              >
                <div className="flex items-start justify-between gap-2">
                  <div className="flex-1 min-w-0">
                    <div className="font-mono text-sm font-medium">
                      {contract.instance}
                    </div>
                    <div className="flex items-center gap-2 mt-1">
                      <span className="text-sm">{contract.contract_name}</span>
                      <Badge variant="outline" className="text-xs">
                        {contract.contract_type}
                      </Badge>
                      {contract.required && (
                        <Badge variant="destructive" className="text-xs">
                          required
                        </Badge>
                      )}
                    </div>
                    <div className="text-sm text-muted-foreground mt-1">
                      {contract.reason === 'missing' && 'No provider available'}
                      {contract.reason === 'ambiguous' &&
                        `${contract.available_providers?.length} providers available`}
                    </div>
                  </div>
                </div>
              </div>
            ))}
          </CardContent>
        </Card>
      )}

      <CreateWireDialog
        stackName={stackName}
        open={createWireOpen}
        onOpenChange={setCreateWireOpen}
        unresolved={unresolved}
      />
    </div>
  )
}
