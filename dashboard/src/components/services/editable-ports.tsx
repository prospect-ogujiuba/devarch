import { useState } from 'react'
import { Pencil, Plus, Trash2, X, Check, ExternalLink } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle, CardAction } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { CopyButton } from '@/components/ui/copy-button'
import { useUpdatePorts } from '@/features/services/queries'
import type { ServicePort } from '@/types/api'

interface Props {
  name: string
  ports: ServicePort[]
}

interface PortDraft {
  host_ip: string
  host_port: string
  container_port: string
  protocol: string
}

function toPortDrafts(ports: ServicePort[]): PortDraft[] {
  return ports.map((p) => ({
    host_ip: p.host_ip,
    host_port: String(p.host_port),
    container_port: String(p.container_port),
    protocol: p.protocol,
  }))
}

export function EditablePorts({ name, ports }: Props) {
  const [editing, setEditing] = useState(false)
  const [drafts, setDrafts] = useState<PortDraft[]>([])
  const updatePorts = useUpdatePorts()

  const startEdit = () => {
    setDrafts(toPortDrafts(ports))
    setEditing(true)
  }

  const cancel = () => setEditing(false)

  const save = () => {
    updatePorts.mutate(
      {
        name,
        data: {
          ports: drafts.map((d) => ({
            host_ip: d.host_ip,
            host_port: parseInt(d.host_port, 10) || 0,
            container_port: parseInt(d.container_port, 10) || 0,
            protocol: d.protocol || 'tcp',
          })),
        },
      },
      { onSuccess: () => setEditing(false) },
    )
  }

  const add = () => setDrafts([...drafts, { host_ip: '0.0.0.0', host_port: '', container_port: '', protocol: 'tcp' }])
  const remove = (i: number) => setDrafts(drafts.filter((_, idx) => idx !== i))
  const update = (i: number, field: keyof PortDraft, value: string) => {
    const next = [...drafts]
    next[i] = { ...next[i], [field]: value }
    setDrafts(next)
  }

  if (editing) {
    return (
      <Card>
        <CardHeader className="flex flex-row items-center justify-between">
          <CardTitle className="text-base">Ports</CardTitle>
          <div className="flex gap-1">
            <Button variant="outline" size="sm" onClick={add}><Plus className="size-4" /> Add</Button>
            <Button variant="ghost" size="icon-sm" onClick={cancel}><X className="size-4" /></Button>
            <Button variant="default" size="icon-sm" onClick={save} disabled={updatePorts.isPending}><Check className="size-4" /></Button>
          </div>
        </CardHeader>
        <CardContent className="space-y-2">
          {drafts.map((d, i) => (
            <div key={i} className="flex gap-2 items-center">
              <Input className="w-28" value={d.host_ip} onChange={(e) => update(i, 'host_ip', e.target.value)} placeholder="Host IP" />
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
              <Button variant="ghost" size="icon-sm" onClick={() => remove(i)}><Trash2 className="size-4 text-destructive" /></Button>
            </div>
          ))}
          {drafts.length === 0 && <p className="text-muted-foreground text-sm">No ports</p>}
        </CardContent>
      </Card>
    )
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base">Ports</CardTitle>
        <CardAction>
          <Button variant="ghost" size="icon-sm" onClick={startEdit}><Pencil className="size-4" /></Button>
        </CardAction>
      </CardHeader>
      <CardContent>
        {ports.length > 0 ? (
          <div className="space-y-2">
            {ports.map((port, i) => {
              const url = `http://localhost:${port.host_port}`
              return (
                <div key={i} className="flex items-center gap-2">
                  <Badge variant="outline">
                    {port.host_ip ? `${port.host_ip}:` : ''}{port.host_port}:{port.container_port}/{port.protocol}
                  </Badge>
                  <a href={url} target="_blank" rel="noopener noreferrer" className="text-xs text-primary hover:underline flex items-center gap-1">
                    {url}
                    <ExternalLink className="size-3" />
                  </a>
                  <CopyButton value={url} />
                </div>
              )
            })}
          </div>
        ) : (
          <p className="text-muted-foreground">No ports exposed</p>
        )}
      </CardContent>
    </Card>
  )
}
