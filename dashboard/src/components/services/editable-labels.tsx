import { Trash2 } from 'lucide-react'
import { CardContent } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { FieldRow } from '@/components/ui/field-row'
import { EditableCard } from '@/components/ui/editable-card'
import { useEditableSection } from '@/hooks/use-editable-section'
import { useUpdateLabels } from '@/features/services/queries'
import type { ServiceLabel } from '@/types/api'

interface Props {
  name: string
  labels: ServiceLabel[]
}

interface LabelDraft {
  key: string
  value: string
}

export function EditableLabels({ name, labels }: Props) {
  const section = useEditableSection<LabelDraft>(() => labels.map((l) => ({ key: l.key, value: l.value })))
  const mutation = useUpdateLabels()

  const save = () => {
    mutation.mutate({ name, data: { labels: section.drafts } }, { onSuccess: () => section.setEditing(false) })
  }

  if (!section.editing && labels.length === 0) return null

  return (
    <EditableCard
      title="Labels"
      editing={section.editing}
      onEdit={section.startEdit}
      onCancel={section.cancel}
      onSave={save}
      onAdd={() => section.add({ key: '', value: '' })}
      isPending={mutation.isPending}
    >
      <CardContent className={section.editing ? 'space-y-2' : undefined}>
        {section.editing ? (
          <>
            {section.drafts.map((d, i) => (
              <FieldRow key={i}>
                <Input className="flex-1" value={d.key} onChange={(e) => section.update(i, { key: e.target.value })} placeholder="key" />
                <span className="text-muted-foreground hidden sm:inline">=</span>
                <Input className="flex-1" value={d.value} onChange={(e) => section.update(i, { value: e.target.value })} placeholder="value" />
                <Button variant="ghost" size="icon-sm" onClick={() => section.remove(i)}><Trash2 className="size-4 text-destructive" /></Button>
              </FieldRow>
            ))}
            {section.drafts.length === 0 && <p className="text-muted-foreground text-sm">No labels</p>}
          </>
        ) : (
          <div className="space-y-1">
            {labels.map((l, i) => (
              <div key={i} className="text-sm font-mono text-muted-foreground">
                {l.key}={l.value}
              </div>
            ))}
          </div>
        )}
      </CardContent>
    </EditableCard>
  )
}
