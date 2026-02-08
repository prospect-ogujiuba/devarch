import { Trash2 } from 'lucide-react'
import { CardContent } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Checkbox } from '@/components/ui/checkbox'
import { FieldRow, FieldRowActions } from '@/components/ui/field-row'
import { EditableCard } from '@/components/ui/editable-card'
import { useEditableSection } from '@/hooks/use-editable-section'
import { useUpdateVolumes } from '@/features/services/queries'
import type { ServiceVolume } from '@/types/api'

interface Props {
  name: string
  volumes: ServiceVolume[]
}

interface VolDraft {
  volume_type: string
  source: string
  target: string
  read_only: boolean
  is_external: boolean
}

export function EditableVolumes({ name, volumes }: Props) {
  const section = useEditableSection<VolDraft>(() =>
    volumes.map((v) => ({ volume_type: v.volume_type, source: v.source, target: v.target, read_only: v.read_only, is_external: v.is_external })),
  )
  const mutation = useUpdateVolumes()

  const save = () => {
    mutation.mutate({ name, data: { volumes: section.drafts } }, { onSuccess: () => section.setEditing(false) })
  }

  return (
    <EditableCard
      title="Volumes"
      editing={section.editing}
      onEdit={section.startEdit}
      onCancel={section.cancel}
      onSave={save}
      onAdd={() => section.add({ volume_type: 'bind', source: '', target: '', read_only: false, is_external: false })}
      isPending={mutation.isPending}
    >
      <CardContent className={section.editing ? 'space-y-2' : undefined}>
        {section.editing ? (
          <>
            {section.drafts.map((d, i) => (
              <FieldRow key={i}>
                <Select value={d.volume_type} onValueChange={(v) => section.update(i, { volume_type: v })}>
                  <SelectTrigger className="w-full sm:w-24"><SelectValue /></SelectTrigger>
                  <SelectContent>
                    <SelectItem value="bind">bind</SelectItem>
                    <SelectItem value="volume">volume</SelectItem>
                    <SelectItem value="tmpfs">tmpfs</SelectItem>
                  </SelectContent>
                </Select>
                <Input className="flex-1" value={d.source} onChange={(e) => section.update(i, { source: e.target.value })} placeholder="Source" />
                <span className="text-muted-foreground hidden sm:inline">:</span>
                <Input className="flex-1" value={d.target} onChange={(e) => section.update(i, { target: e.target.value })} placeholder="Target" />
                <FieldRowActions>
                  <div className="flex items-center gap-1">
                    <Checkbox checked={d.read_only} onCheckedChange={(v) => section.update(i, { read_only: !!v })} />
                    <span className="text-xs">RO</span>
                  </div>
                  <div className="flex items-center gap-1">
                    <Checkbox checked={d.is_external} onCheckedChange={(v) => section.update(i, { is_external: !!v })} />
                    <span className="text-xs">Ext</span>
                  </div>
                  <Button variant="ghost" size="icon-sm" onClick={() => section.remove(i)}><Trash2 className="size-4 text-destructive" /></Button>
                </FieldRowActions>
              </FieldRow>
            ))}
            {section.drafts.length === 0 && <p className="text-muted-foreground text-sm">No volumes</p>}
          </>
        ) : (
          <>
            {volumes.length > 0 ? (
              <div className="space-y-1">
                {volumes.map((vol, i) => (
                  <div key={i} className="text-sm font-mono text-muted-foreground">
                    {vol.source}:{vol.target}{vol.read_only ? ' (ro)' : ''}
                  </div>
                ))}
              </div>
            ) : (
              <p className="text-muted-foreground">No volumes mounted</p>
            )}
          </>
        )}
      </CardContent>
    </EditableCard>
  )
}
