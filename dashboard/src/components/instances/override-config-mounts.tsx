import { Trash2 } from 'lucide-react'
import { CardContent } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Checkbox } from '@/components/ui/checkbox'
import { FieldRow } from '@/components/ui/field-row'
import { EditableCard } from '@/components/ui/editable-card'
import { useOverrideSection } from '@/hooks/use-override-section'
import { useUpdateInstanceConfigMounts } from '@/features/instances/queries'
import type { InstanceDetail, Service, ServiceConfigMount } from '@/types/api'

interface Props {
  instance: InstanceDetail
  templateData: Service
  stackName: string
  instanceId: string
}

interface ConfigMountDraft {
  config_file_id?: number | null
  source_path: string
  target_path: string
  readonly: boolean
}

function toConfigMountDrafts(mounts: ServiceConfigMount[]): ConfigMountDraft[] {
  return mounts.map((m) => ({
    config_file_id: m.config_file_id,
    source_path: m.source_path,
    target_path: m.target_path,
    readonly: m.readonly,
  }))
}

export function OverrideConfigMounts({ instance, templateData, stackName, instanceId }: Props) {
  const templateConfigMounts = templateData.config_mounts ?? []
  const overrideConfigMounts = instance.config_mounts ?? []

  const section = useOverrideSection(() => toConfigMountDrafts(overrideConfigMounts))
  const updateConfigMounts = useUpdateInstanceConfigMounts(stackName, instanceId)

  const save = () => {
    updateConfigMounts.mutate(section.drafts, { onSuccess: () => section.setEditing(false) })
  }

  const resetAll = () => {
    updateConfigMounts.mutate([], { onSuccess: () => section.setEditing(false) })
  }

  return (
    <EditableCard
      title="Config Mount Overrides"
      editing={section.editing}
      onEdit={section.startEdit}
      onCancel={section.cancel}
      onSave={save}
      onAdd={() => section.add({ config_file_id: null, source_path: '', target_path: '', readonly: false })}
      isPending={updateConfigMounts.isPending}
      isDirty={section.isDirty}
      onResetAll={resetAll}
      showResetAll={overrideConfigMounts.length > 0}
      editVariant="text"
    >
      <CardContent className="space-y-4">
        {section.editing ? (
          <>
            {templateConfigMounts.length > 0 && (
              <div>
                <div className="text-xs font-medium text-muted-foreground mb-2">Template Config Mounts (read-only)</div>
                <div className="space-y-2">
                  {templateConfigMounts.map((mount, i) => (
                    <div key={i} className="text-sm text-muted-foreground">
                      <span>{mount.source_path} → {mount.target_path}</span>
                      {mount.readonly && <Badge variant="outline" className="ml-2 text-muted-foreground">ro</Badge>}
                      {mount.config_file_id ? (
                        <Badge variant="outline" className="ml-2 text-muted-foreground">resolved</Badge>
                      ) : (
                        <Badge variant="outline" className="ml-2 text-muted-foreground">unresolved</Badge>
                      )}
                    </div>
                  ))}
                </div>
              </div>
            )}

            <div>
              <div className="text-xs font-medium mb-2">Override Config Mounts</div>
              <div className="space-y-3">
                {section.drafts.map((d, i) => (
                  <div key={i} className="border-l-2 border-blue-500 pl-3 space-y-2">
                    <FieldRow>
                      <Input
                        className="flex-1"
                        value={d.source_path}
                        onChange={(e) => section.update(i, { source_path: e.target.value })}
                        placeholder="Source path"
                      />
                      <span className="text-muted-foreground hidden sm:inline">→</span>
                      <Input
                        className="flex-1"
                        value={d.target_path}
                        onChange={(e) => section.update(i, { target_path: e.target.value })}
                        placeholder="Target path"
                      />
                      <Button variant="ghost" size="icon-sm" onClick={() => section.remove(i)}>
                        <Trash2 className="size-4 text-destructive" />
                      </Button>
                    </FieldRow>
                    <div className="flex gap-4 text-sm">
                      <label className="flex items-center gap-2">
                        <Checkbox checked={d.readonly} onCheckedChange={(checked) => section.update(i, { readonly: !!checked })} />
                        Read-only
                      </label>
                    </div>
                  </div>
                ))}
                {section.drafts.length === 0 && <p className="text-muted-foreground text-sm">No overrides</p>}
              </div>
            </div>
          </>
        ) : (
          <>
            {templateConfigMounts.length > 0 && (
              <div>
                <div className="text-xs font-medium text-muted-foreground mb-2">Template Config Mounts</div>
                <div className="space-y-2">
                  {templateConfigMounts.map((mount, i) => (
                    <div key={i} className="text-sm text-muted-foreground">
                      <span>{mount.source_path} → {mount.target_path}</span>
                      {mount.readonly && <Badge variant="outline" className="ml-2 text-muted-foreground">ro</Badge>}
                      {mount.config_file_id ? (
                        <Badge variant="outline" className="ml-2 text-muted-foreground">resolved</Badge>
                      ) : (
                        <Badge variant="outline" className="ml-2 text-muted-foreground">unresolved</Badge>
                      )}
                    </div>
                  ))}
                </div>
              </div>
            )}

            {overrideConfigMounts.length > 0 ? (
              <div>
                <div className="text-xs font-medium mb-2">Overrides</div>
                <div className="space-y-2">
                  {overrideConfigMounts.map((mount, i) => (
                    <div key={i} className="text-sm border-l-2 border-blue-500 pl-2">
                      <span>{mount.source_path} → {mount.target_path}</span>
                      {mount.readonly && <Badge variant="default" className="ml-2">ro</Badge>}
                      {mount.config_file_id ? (
                        <Badge variant="default" className="ml-2">resolved</Badge>
                      ) : (
                        <Badge variant="default" className="ml-2">unresolved</Badge>
                      )}
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
