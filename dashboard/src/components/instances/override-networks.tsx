import { Trash2 } from 'lucide-react'
import { CardContent } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { EditableCard } from '@/components/ui/editable-card'
import { useOverrideSection } from '@/hooks/use-override-section'
import { useUpdateInstanceNetworks } from '@/features/instances/queries'
import type { InstanceDetail, Service } from '@/types/api'

interface Props {
  instance: InstanceDetail
  templateData: Service
  stackName: string
  instanceId: string
}

export function OverrideNetworks({ instance, templateData, stackName, instanceId }: Props) {
  const templateNetworks = templateData.networks ?? []
  const overrideNetworks = instance.networks ?? []

  const section = useOverrideSection(() => [...overrideNetworks])
  const updateNetworks = useUpdateInstanceNetworks(stackName, instanceId)

  const save = () => {
    const draftSet = new Set(section.drafts)
    const preserved = templateNetworks.filter((n) => !draftSet.has(n))
    updateNetworks.mutate([...preserved, ...section.drafts], {
      onSuccess: () => section.setEditing(false),
    })
  }

  const resetAll = () => {
    updateNetworks.mutate([], { onSuccess: () => section.setEditing(false) })
  }

  return (
    <EditableCard
      title="Network Overrides"
      editing={section.editing}
      onEdit={section.startEdit}
      onCancel={section.cancel}
      onSave={save}
      onAdd={() => section.add('')}
      isPending={updateNetworks.isPending}
      isDirty={section.isDirty}
      onResetAll={resetAll}
      showResetAll={overrideNetworks.length > 0}
      editVariant="text"
    >
      <CardContent className="space-y-4">
        {section.editing ? (
          <>
            {templateNetworks.length > 0 && (
              <div>
                <div className="text-xs font-medium text-muted-foreground mb-2">Template Networks (read-only)</div>
                <div className="space-y-2">
                  {templateNetworks.map((network, i) => (
                    <div key={i} className="text-sm text-muted-foreground">
                      <Badge variant="outline" className="text-muted-foreground">
                        {network}
                      </Badge>
                    </div>
                  ))}
                </div>
              </div>
            )}

            <div>
              <div className="text-xs font-medium mb-2">Override Networks</div>
              <div className="space-y-2">
                {section.drafts.map((network, i) => (
                  <div key={i} className="border-l-2 border-blue-500 pl-3 flex gap-2 items-center">
                    <Input
                      className="flex-1"
                      value={network}
                      onChange={(e) => section.setDrafts(section.drafts.map((n, idx) => idx === i ? e.target.value : n))}
                      placeholder="network-name"
                    />
                    <Button variant="ghost" size="icon-sm" onClick={() => section.remove(i)}>
                      <Trash2 className="size-4 text-destructive" />
                    </Button>
                  </div>
                ))}
                {section.drafts.length === 0 && <p className="text-muted-foreground text-sm">No overrides</p>}
              </div>
            </div>
          </>
        ) : (
          <>
            {templateNetworks.length > 0 && (
              <div>
                <div className="text-xs font-medium text-muted-foreground mb-2">Template Networks</div>
                <div className="space-y-2">
                  {templateNetworks.map((network, i) => (
                    <div key={i} className="text-sm text-muted-foreground">
                      <Badge variant="outline" className="text-muted-foreground">
                        {network}
                      </Badge>
                    </div>
                  ))}
                </div>
              </div>
            )}

            {overrideNetworks.length > 0 ? (
              <div>
                <div className="text-xs font-medium mb-2">Overrides</div>
                <div className="space-y-2">
                  {overrideNetworks.map((network, i) => (
                    <div key={i} className="text-sm border-l-2 border-blue-500 pl-2">
                      <Badge variant="default">
                        {network}
                      </Badge>
                    </div>
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
