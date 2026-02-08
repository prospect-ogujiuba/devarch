import { Trash2 } from 'lucide-react'
import { CardContent } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { FieldRow } from '@/components/ui/field-row'
import { EditableCard } from '@/components/ui/editable-card'
import { useOverrideSection } from '@/hooks/use-override-section'
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
  const templatePorts = templateData.ports ?? []
  const overridePorts = instance.ports ?? []

  const section = useOverrideSection(() => toPortDrafts(overridePorts))
  const updatePorts = useUpdateInstancePorts(stackName, instanceId)

  const save = () => {
    updatePorts.mutate(
      section.drafts.map((d) => ({
        host_ip: d.host_ip,
        host_port: parseInt(d.host_port, 10) || 0,
        container_port: parseInt(d.container_port, 10) || 0,
        protocol: d.protocol || 'tcp',
      })),
      { onSuccess: () => section.setEditing(false) },
    )
  }

  const resetAll = () => {
    updatePorts.mutate([], { onSuccess: () => section.setEditing(false) })
  }

  return (
    <EditableCard
      title="Port Overrides"
      editing={section.editing}
      onEdit={section.startEdit}
      onCancel={section.cancel}
      onSave={save}
      onAdd={() => section.add({ host_ip: '0.0.0.0', host_port: '', container_port: '', protocol: 'tcp' })}
      isPending={updatePorts.isPending}
      isDirty={section.isDirty}
      onResetAll={resetAll}
      showResetAll={overridePorts.length > 0}
      editVariant="text"
    >
      <CardContent className="space-y-4">
        {section.editing ? (
          <>
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
                {section.drafts.map((d, i) => (
                  <FieldRow key={i} className="border-l-2 border-blue-500 pl-2">
                    <Input className="w-full sm:w-28" value={d.host_ip} onChange={(e) => section.update(i, { host_ip: e.target.value })} placeholder="0.0.0.0" />
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
                    <Button variant="ghost" size="icon-sm" onClick={() => section.remove(i)}>
                      <Trash2 className="size-4 text-destructive" />
                    </Button>
                  </FieldRow>
                ))}
                {section.drafts.length === 0 && <p className="text-muted-foreground text-sm">No overrides</p>}
              </div>
            </div>
          </>
        ) : (
          <>
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
          </>
        )}
      </CardContent>
    </EditableCard>
  )
}
