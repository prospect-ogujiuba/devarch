import { Trash2, ExternalLink } from 'lucide-react'
import { CardContent } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { CopyButton } from '@/components/ui/copy-button'
import { FieldRow } from '@/components/ui/field-row'
import { EditableCard } from '@/components/ui/editable-card'
import { useEditableSection } from '@/hooks/use-editable-section'
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
  const section = useEditableSection(() => toPortDrafts(ports))
  const updatePorts = useUpdatePorts()

  const save = () => {
    updatePorts.mutate(
      {
        name,
        data: {
          ports: section.drafts.map((d) => ({
            host_ip: d.host_ip,
            host_port: parseInt(d.host_port, 10) || 0,
            container_port: parseInt(d.container_port, 10) || 0,
            protocol: d.protocol || 'tcp',
          })),
        },
      },
      { onSuccess: () => section.setEditing(false) },
    )
  }

  return (
    <EditableCard
      title="Ports"
      editing={section.editing}
      onEdit={section.startEdit}
      onCancel={section.cancel}
      onSave={save}
      onAdd={() => section.add({ host_ip: '0.0.0.0', host_port: '', container_port: '', protocol: 'tcp' })}
      isPending={updatePorts.isPending}
    >
      <CardContent className={section.editing ? 'space-y-2' : undefined}>
        {section.editing ? (
          <>
            {section.drafts.map((d, i) => (
              <FieldRow key={i}>
                <Input className="w-full sm:w-28" value={d.host_ip} onChange={(e) => section.update(i, { host_ip: e.target.value })} placeholder="Host IP" />
                <Input className="w-full sm:w-24" type="number" value={d.host_port} onChange={(e) => section.update(i, { host_port: e.target.value })} placeholder="Host" />
                <span className="text-muted-foreground hidden sm:inline">:</span>
                <Input className="w-full sm:w-24" type="number" value={d.container_port} onChange={(e) => section.update(i, { container_port: e.target.value })} placeholder="Container" />
                <Select value={d.protocol} onValueChange={(v) => section.update(i, { protocol: v })}>
                  <SelectTrigger className="w-full sm:w-20"><SelectValue /></SelectTrigger>
                  <SelectContent>
                    <SelectItem value="tcp">tcp</SelectItem>
                    <SelectItem value="udp">udp</SelectItem>
                  </SelectContent>
                </Select>
                <Button variant="ghost" size="icon-sm" onClick={() => section.remove(i)}><Trash2 className="size-4 text-destructive" /></Button>
              </FieldRow>
            ))}
            {section.drafts.length === 0 && <p className="text-muted-foreground text-sm">No ports</p>}
          </>
        ) : (
          <>
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
          </>
        )}
      </CardContent>
    </EditableCard>
  )
}
