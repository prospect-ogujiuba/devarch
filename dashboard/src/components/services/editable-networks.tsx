import { Trash2 } from 'lucide-react'
import { CardContent } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { FieldRow, FieldRowActions } from '@/components/ui/field-row'
import { EditableCard } from '@/components/ui/editable-card'
import { useEditableSection } from '@/hooks/use-editable-section'
import { useUpdateNetworks } from '@/features/services/queries'

interface Props {
  name: string
  networks: string[]
}

export function EditableNetworks({ name, networks }: Props) {
  const section = useEditableSection<string>(() => [...networks])
  const mutation = useUpdateNetworks()

  const save = () => {
    mutation.mutate({ name, data: { networks: section.drafts } }, { onSuccess: () => section.setEditing(false) })
  }

  return (
    <EditableCard
      title="Networks"
      editing={section.editing}
      onEdit={section.startEdit}
      onCancel={section.cancel}
      onSave={save}
      onAdd={() => section.add('')}
      isPending={mutation.isPending}
    >
      <CardContent className={section.editing ? 'space-y-2' : undefined}>
        {section.editing ? (
          <>
            {section.drafts.map((network, i) => (
              <FieldRow key={i}>
                <Input
                  className="flex-1"
                  value={network}
                  onChange={(e) => section.setDrafts((prev) => prev.map((n, idx) => (idx === i ? e.target.value : n)))}
                  placeholder="Network name"
                />
                <FieldRowActions>
                  <Button variant="ghost" size="icon-sm" onClick={() => section.remove(i)}>
                    <Trash2 className="size-4 text-destructive" />
                  </Button>
                </FieldRowActions>
              </FieldRow>
            ))}
            {section.drafts.length === 0 && <p className="text-muted-foreground text-sm">No networks</p>}
          </>
        ) : networks.length === 0 ? (
          <p className="text-sm text-muted-foreground">No networks configured</p>
        ) : (
          <div className="flex flex-wrap gap-2">
            {networks.map((network, i) => (
              <Badge key={i} variant="outline">
                {network}
              </Badge>
            ))}
          </div>
        )}
      </CardContent>
    </EditableCard>
  )
}
