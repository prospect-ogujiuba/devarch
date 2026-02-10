import { Trash2 } from 'lucide-react'
import { CardContent } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { FieldRow, FieldRowActions } from '@/components/ui/field-row'
import { EditableCard } from '@/components/ui/editable-card'
import { useEditableSection } from '@/hooks/use-editable-section'
import { useUpdateEnvFiles } from '@/features/services/queries'

interface Props {
  name: string
  envFiles: string[]
}

export function EditableEnvFiles({ name, envFiles }: Props) {
  const section = useEditableSection<string>(() => [...envFiles])
  const mutation = useUpdateEnvFiles()

  const save = () => {
    mutation.mutate({ name, data: { env_files: section.drafts } }, { onSuccess: () => section.setEditing(false) })
  }

  return (
    <EditableCard
      title="Env Files"
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
            {section.drafts.map((path, i) => (
              <FieldRow key={i}>
                <Input
                  className="flex-1"
                  value={path}
                  onChange={(e) => section.setDrafts((prev) => prev.map((p, idx) => (idx === i ? e.target.value : p)))}
                  placeholder="Path to env file"
                />
                <FieldRowActions>
                  <Button variant="ghost" size="icon-sm" onClick={() => section.remove(i)}>
                    <Trash2 className="size-4 text-destructive" />
                  </Button>
                </FieldRowActions>
              </FieldRow>
            ))}
            {section.drafts.length === 0 && <p className="text-muted-foreground text-sm">No env files</p>}
          </>
        ) : envFiles.length === 0 ? (
          <p className="text-sm text-muted-foreground">No env files configured</p>
        ) : (
          <div className="flex flex-wrap gap-2">
            {envFiles.map((path, i) => (
              <Badge key={i} variant="outline">
                {path}
              </Badge>
            ))}
          </div>
        )}
      </CardContent>
    </EditableCard>
  )
}
