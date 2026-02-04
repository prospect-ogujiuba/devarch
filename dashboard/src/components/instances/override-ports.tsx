import { useState } from 'react'
import { Plus, Trash2, X, Check, RotateCcw } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle, CardAction } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { useUpdateInstancePorts } from '@/features/instances/queries'
import type { InstanceDetail, Service, InstancePort } from '@/types/api'

interface Props {
  instance: InstanceDetail
  templateData: Service
  stackName: string
  instanceId: string
}

interface PortDraft {
  host_ip: string
  host_port: string
  container_port: string
  protocol: string
}

function toPortDrafts(ports: InstancePort[]): PortDraft[] {
  return ports.map((p) => ({
    host_ip: p.host_ip,
    host_port: String(p.host_port),
    container_port: String(p.container_port),
    protocol: p.protocol,
  }))
}

export function OverridePorts({ instance, templateData, stackName, instanceId }: Props) {
  const [editing, setEditing] = useState(false)
  const [drafts, setDrafts] = useState<PortDraft[]>([])
  const updatePorts = useUpdateInstancePorts(stackName, instanceId)

  const templatePorts = templateData.ports ?? []
  const overridePorts = instance.ports ?? []

  const startEdit = () => {
    setDrafts(toPortDrafts(overridePorts))
    setEditing(true)
  }

  const cancel = () => setEditing(false)

  const save = () => {
    updatePorts.mutate(
      drafts.map((d) => ({
        host_ip: d.host_ip,
        host_port: parseInt(d.host_port, 10) || 0,
        container_port: parseInt(d.container_port, 10) || 0,
        protocol: d.protocol || 'tcp',
      })),
      { onSuccess: () => setEditing(false) },
    )
  }

  const resetAll = () => {
    updatePorts.mutate([], { onSuccess: () => setEditing(false) })
  }

  const add = () => setDrafts([...drafts, { host_ip: '0.0.0.0', host_port: '', container_port: '', protocol: 'tcp' }])
  const remove = (i: number) => setDrafts(drafts.filter((_, idx) => idx !== i))
  const update = (i: number, field: keyof PortDraft, value: string) => {
    const next = [...drafts]
    next[i] = { ...next[i], [field]: value }
    setDrafts(next)
  }

  const isDirty = JSON.stringify(drafts) !== JSON.stringify(toPortDrafts(overridePorts))

  if (editing) {
    return (
      <Card>
        <CardHeader className="flex flex-row items-center justify-between">
          <CardTitle className="text-base">Port Overrides</CardTitle>
          <div className="flex gap-1">
            <Button variant="outline" size="sm" onClick={add}>
              <Plus className="size-4" /> Add
            </Button>
            {overridePorts.length > 0 && (
              <Button variant="outline" size="sm" onClick={resetAll}>
                <RotateCcw className="size-4" /> Reset All
              </Button>
            )}
            <Button variant="ghost" size="icon-sm" onClick={cancel}>
              <X className="size-4" />
            </Button>
            <Button variant="default" size="icon-sm" onClick={save} disabled={updatePorts.isPending || !isDirty}>
              <Check className="size-4" />
            </Button>
          </div>
        </CardHeader>
        <CardContent className="space-y-4">
          {templatePorts.length > 0 && (
            <div>
              <div className="text-xs font-medium text-muted-foreground mb-2">Template Ports (read-only)</div>
              <div className="space-y-2">
                {templatePorts.map((port, i) => (
                  <div key={i} className="flex gap-2 items-center text-sm text-muted-foreground">
                    <Badge variant="outline" className="text-muted-foreground">
                      {port.host_ip}:{port.host_port}:{port.container_port}/{port.protocol}
                    </Badge>
                  </div>
                ))}
              </div>
            </div>
          )}

          <div>
            <div className="text-xs font-medium mb-2">Override Ports</div>
            <div className="space-y-2">
              {drafts.map((d, i) => (
                <div key={i} className="flex gap-2 items-center border-l-2 border-blue-500 pl-2">
                  <Input className="w-28" value={d.host_ip} onChange={(e) => update(i, 'host_ip', e.target.value)} placeholder="0.0.0.0" />
                  <Input className="w-24" type="number" value={d.host_port} onChange={(e) => update(i, 'host_port', e.target.value)} placeholder="Host" />
                  <span className="text-muted-foreground">:</span>
                  <Input className="w-24" type="number" value={d.container_port} onChange={(e) => update(i, 'container_port', e.target.value)} placeholder="Container" />
                  <Select value={d.protocol} onValueChange={(v) => update(i, 'protocol', v)}>
                    <SelectTrigger className="w-20"><SelectValue /></SelectTrigger>
                    <SelectContent>
                      <SelectItem value="tcp">tcp</SelectItem>
                      <SelectItem value="udp">udp</SelectItem>
                    </SelectContent>
                  </Select>
                  <Button variant="ghost" size="icon-sm" onClick={() => remove(i)}>
                    <Trash2 className="size-4 text-destructive" />
                  </Button>
                </div>
              ))}
              {drafts.length === 0 && <p className="text-muted-foreground text-sm">No overrides</p>}
            </div>
          </div>
        </CardContent>
      </Card>
    )
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base">Port Overrides</CardTitle>
        <CardAction>
          <Button variant="outline" size="sm" onClick={startEdit}>
            Edit
          </Button>
        </CardAction>
      </CardHeader>
      <CardContent className="space-y-4">
        {templatePorts.length > 0 && (
          <div>
            <div className="text-xs font-medium text-muted-foreground mb-2">Template Ports</div>
            <div className="space-y-2">
              {templatePorts.map((port, i) => (
                <Badge key={i} variant="outline" className="text-muted-foreground">
                  {port.host_ip}:{port.host_port}:{port.container_port}/{port.protocol}
                </Badge>
              ))}
            </div>
          </div>
        )}

        {overridePorts.length > 0 ? (
          <div>
            <div className="text-xs font-medium mb-2">Overrides</div>
            <div className="space-y-2">
              {overridePorts.map((port, i) => (
                <Badge key={i} variant="default" className="border-l-2 border-blue-500">
                  {port.host_ip}:{port.host_port}:{port.container_port}/{port.protocol}
                </Badge>
              ))}
            </div>
          </div>
        ) : (
          <p className="text-muted-foreground text-sm">No overrides configured</p>
        )}
      </CardContent>
    </Card>
  )
}
