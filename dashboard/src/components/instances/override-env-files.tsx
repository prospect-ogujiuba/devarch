import { Trash2 } from 'lucide-react'
import { CardContent } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { EditableCard } from '@/components/ui/editable-card'
import { useOverrideSection } from '@/hooks/use-override-section'
import { useUpdateInstanceEnvFiles } from '@/features/instances/queries'
import type { InstanceDetail, Service } from '@/types/api'

interface Props {
  instance: InstanceDetail
  templateData: Service
  stackName: string
  instanceId: string
}

export function OverrideEnvFiles({ instance, templateData, stackName, instanceId }: Props) {
  const templateEnvFiles = templateData.env_files ?? []
  const overrideEnvFiles = instance.env_files ?? []

  const section = useOverrideSection(() => [...overrideEnvFiles])
  const updateEnvFiles = useUpdateInstanceEnvFiles(stackName, instanceId)

  const save = () => {
    const draftSet = new Set(section.drafts)
    const preserved = templateEnvFiles.filter((f) => !draftSet.has(f))
    updateEnvFiles.mutate([...preserved, ...section.drafts], {
      onSuccess: () => section.setEditing(false),
    })
  }

  const resetAll = () => {
    updateEnvFiles.mutate([], { onSuccess: () => section.setEditing(false) })
  }

  return (
    <EditableCard
      title="Env Files Overrides"
      editing={section.editing}
      onEdit={section.startEdit}
      onCancel={section.cancel}
      onSave={save}
      onAdd={() => section.add('')}
      isPending={updateEnvFiles.isPending}
      isDirty={section.isDirty}
      onResetAll={resetAll}
      showResetAll={overrideEnvFiles.length > 0}
      editVariant="text"
    >
      <CardContent className="space-y-4">
        {section.editing ? (
          <>
            {templateEnvFiles.length > 0 && (
              <div>
                <div className="text-xs font-medium text-muted-foreground mb-2">Template Env Files (read-only)</div>
                <div className="space-y-2">
                  {templateEnvFiles.map((file, i) => (
                    <div key={i} className="text-sm text-muted-foreground">
                      <Badge variant="outline" className="text-muted-foreground">
                        {file}
                      </Badge>
                    </div>
                  ))}
                </div>
              </div>
            )}

            <div>
              <div className="text-xs font-medium mb-2">Override Env Files</div>
              <div className="space-y-2">
                {section.drafts.map((file, i) => (
                  <div key={i} className="border-l-2 border-blue-500 pl-3 flex gap-2 items-center">
                    <Input
                      className="flex-1"
                      value={file}
                      onChange={(e) => section.setDrafts(section.drafts.map((f, idx) => idx === i ? e.target.value : f))}
                      placeholder=".env"
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
            {templateEnvFiles.length > 0 && (
              <div>
                <div className="text-xs font-medium text-muted-foreground mb-2">Template Env Files</div>
                <div className="space-y-2">
                  {templateEnvFiles.map((file, i) => (
                    <div key={i} className="text-sm text-muted-foreground">
                      <Badge variant="outline" className="text-muted-foreground">
                        {file}
                      </Badge>
                    </div>
                  ))}
                </div>
              </div>
            )}

            {overrideEnvFiles.length > 0 ? (
              <div>
                <div className="text-xs font-medium mb-2">Overrides</div>
                <div className="space-y-2">
                  {overrideEnvFiles.map((file, i) => (
                    <div key={i} className="text-sm border-l-2 border-blue-500 pl-2">
                      <Badge variant="default">
                        {file}
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
