import { Trash2 } from 'lucide-react'
import { CardContent } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { Checkbox } from '@/components/ui/checkbox'
import { FieldRow, FieldRowActions } from '@/components/ui/field-row'
import { EditableCard } from '@/components/ui/editable-card'
import { useEditableSection } from '@/hooks/use-editable-section'
import { useUpdateConfigMounts } from '@/features/services/queries'
import type { ServiceConfigMount } from '@/types/api'

interface Props {
  name: string
  configMounts: ServiceConfigMount[]
}

interface MountDraft {
  source_path: string
  target_path: string
  readonly: boolean
  config_file_id?: number | null
}

export function EditableConfigMounts({ name, configMounts }: Props) {
  const section = useEditableSection<MountDraft>(() =>
    configMounts.map((m) => ({
      source_path: m.source_path,
      target_path: m.target_path,
      readonly: m.readonly,
      config_file_id: m.config_file_id,
    })),
  )
  const mutation = useUpdateConfigMounts()

  const save = () => {
    mutation.mutate({ name, data: { config_mounts: section.drafts } }, { onSuccess: () => section.setEditing(false) })
  }

  return (
    <EditableCard
      title="Config Mounts"
      editing={section.editing}
      onEdit={section.startEdit}
      onCancel={section.cancel}
      onSave={save}
      onAdd={() => section.add({ source_path: '', target_path: '', readonly: false, config_file_id: null })}
      isPending={mutation.isPending}
    >
      <CardContent className={section.editing ? 'space-y-2' : undefined}>
        {section.editing ? (
          <>
            {section.drafts.map((d, i) => (
              <FieldRow key={i}>
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
                <FieldRowActions>
                  <div className="flex items-center gap-1">
                    <Checkbox checked={d.readonly} onCheckedChange={(v) => section.update(i, { readonly: !!v })} />
                    <span className="text-xs">RO</span>
                  </div>
                  <Button variant="ghost" size="icon-sm" onClick={() => section.remove(i)}>
                    <Trash2 className="size-4 text-destructive" />
                  </Button>
                </FieldRowActions>
              </FieldRow>
            ))}
            {section.drafts.length === 0 && <p className="text-muted-foreground text-sm">No config mounts</p>}
          </>
        ) : configMounts.length === 0 ? (
          <p className="text-sm text-muted-foreground">No config mounts configured</p>
        ) : (
          <div className="space-y-2">
            {configMounts.map((mount, i) => (
              <div key={i} className="text-sm flex items-center gap-2">
                <code className="font-mono text-muted-foreground flex-1">
                  {mount.source_path} → {mount.target_path}
                </code>
                <div className="flex items-center gap-1">
                  {mount.config_file_id ? (
                    <Badge variant="default" className="bg-green-600 hover:bg-green-700">
                      resolved
                    </Badge>
                  ) : (
                    <Badge variant="secondary" className="bg-amber-600 hover:bg-amber-700">
                      unresolved
                    </Badge>
                  )}
                  {mount.readonly && (
                    <Badge variant="outline" className="text-xs">
                      ro
                    </Badge>
                  )}
                </div>
              </div>
            ))}
          </div>
        )}
      </CardContent>
    </EditableCard>
  )
}
