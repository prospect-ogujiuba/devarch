import { useState } from 'react'
import { Trash2, Lock } from 'lucide-react'
import { CardContent } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { FieldRow } from '@/components/ui/field-row'
import { EditableCard } from '@/components/ui/editable-card'
import { useOverrideSection } from '@/hooks/use-override-section'
import { useUpdateInstanceLabels } from '@/features/instances/queries'
import type { InstanceDetail, Service, InstanceLabel } from '@/types/api'

interface Props {
  instance: InstanceDetail
  templateData: Service
  stackName: string
  instanceId: string
}

interface LabelDraft {
  key: string
  value: string
}

function toLabelDrafts(labels: InstanceLabel[]): LabelDraft[] {
  return labels.map((l) => ({
    key: l.key,
    value: l.value,
  }))
}

export function OverrideLabels({ instance, templateData, stackName, instanceId }: Props) {
  const templateLabels = templateData.labels ?? []
  const overrideLabels = instance.labels ?? []

  const section = useOverrideSection(() => toLabelDrafts(overrideLabels))
  const [error, setError] = useState('')
  const updateLabels = useUpdateInstanceLabels(stackName, instanceId)

  const handleStartEdit = () => {
    section.startEdit()
    setError('')
  }

  const save = () => {
    const hasSystemLabel = section.drafts.some((d) => d.key.startsWith('devarch.'))
    if (hasSystemLabel) {
      setError('Cannot create labels with "devarch." prefix (system-managed)')
      return
    }
    updateLabels.mutate(section.drafts, { onSuccess: () => section.setEditing(false) })
  }

  const resetAll = () => {
    updateLabels.mutate([], { onSuccess: () => section.setEditing(false) })
  }

  const handleAdd = () => {
    section.add({ key: '', value: '' })
    setError('')
  }

  const handleUpdate = (i: number, patch: Partial<LabelDraft>) => {
    section.update(i, patch)
    setError('')
  }

  const systemLabels = [...templateLabels, ...overrideLabels].filter((l) => l.key.startsWith('devarch.'))
  const customTemplateLabels = templateLabels.filter((l) => !l.key.startsWith('devarch.'))
  const customOverrideLabels = overrideLabels.filter((l) => !l.key.startsWith('devarch.'))

  return (
    <EditableCard
      title="Label Overrides"
      editing={section.editing}
      onEdit={handleStartEdit}
      onCancel={section.cancel}
      onSave={save}
      onAdd={handleAdd}
      isPending={updateLabels.isPending}
      isDirty={section.isDirty}
      onResetAll={resetAll}
      showResetAll={customOverrideLabels.length > 0}
      editVariant="text"
    >
      <CardContent className="space-y-4">
        {section.editing ? (
          <>
            {error && (
              <div className="text-sm text-destructive bg-destructive/10 p-2 rounded">
                {error}
              </div>
            )}

            {systemLabels.length > 0 && (
              <div>
                <div className="text-xs font-medium text-muted-foreground mb-2 flex items-center gap-2">
                  <Lock className="size-3" />
                  System Labels (read-only)
                </div>
                <div className="space-y-2">
                  {systemLabels.map((label, i) => (
                    <div key={i} className="text-sm text-muted-foreground flex items-center gap-2">
                      <code className="bg-muted px-2 py-1 rounded">{label.key}</code>
                      <span>=</span>
                      <span>{label.value}</span>
                    </div>
                  ))}
                </div>
              </div>
            )}

            {customTemplateLabels.length > 0 && (
              <div>
                <div className="text-xs font-medium text-muted-foreground mb-2">Template Labels (read-only)</div>
                <div className="space-y-2">
                  {customTemplateLabels.map((label, i) => (
                    <div key={i} className="text-sm text-muted-foreground flex items-center gap-2">
                      <code className="bg-muted px-2 py-1 rounded">{label.key}</code>
                      <span>=</span>
                      <span>{label.value}</span>
                    </div>
                  ))}
                </div>
              </div>
            )}

            <div>
              <div className="text-xs font-medium mb-2">Override Labels</div>
              <div className="space-y-2">
                {section.drafts.map((d, i) => (
                  <div key={i} className="border-l-2 border-blue-500 pl-3">
                    <FieldRow>
                      <Input
                        className="flex-1"
                        value={d.key}
                        onChange={(e) => handleUpdate(i, { key: e.target.value })}
                        placeholder="key"
                      />
                      <span className="text-muted-foreground hidden sm:inline">=</span>
                      <Input
                        className="flex-1"
                        value={d.value}
                        onChange={(e) => handleUpdate(i, { value: e.target.value })}
                        placeholder="value"
                      />
                      <Button variant="ghost" size="icon-sm" onClick={() => section.remove(i)}>
                        <Trash2 className="size-4 text-destructive" />
                      </Button>
                    </FieldRow>
                  </div>
                ))}
                {section.drafts.length === 0 && <p className="text-muted-foreground text-sm">No overrides</p>}
              </div>
            </div>
          </>
        ) : (
          <>
            {systemLabels.length > 0 && (
              <div>
                <div className="text-xs font-medium text-muted-foreground mb-2 flex items-center gap-2">
                  <Lock className="size-3" />
                  System Labels
                </div>
                <div className="space-y-2">
                  {systemLabels.map((label, i) => (
                    <div key={i} className="text-sm text-muted-foreground flex items-center gap-2">
                      <code className="bg-muted px-2 py-1 rounded">{label.key}</code>
                      <span>=</span>
                      <span>{label.value}</span>
                    </div>
                  ))}
                </div>
              </div>
            )}

            {customTemplateLabels.length > 0 && (
              <div>
                <div className="text-xs font-medium text-muted-foreground mb-2">Template Labels</div>
                <div className="space-y-2">
                  {customTemplateLabels.map((label, i) => (
                    <div key={i} className="text-sm text-muted-foreground flex items-center gap-2">
                      <code className="bg-muted px-2 py-1 rounded">{label.key}</code>
                      <span>=</span>
                      <span>{label.value}</span>
                    </div>
                  ))}
                </div>
              </div>
            )}

            {customOverrideLabels.length > 0 ? (
              <div>
                <div className="text-xs font-medium mb-2">Overrides</div>
                <div className="space-y-2">
                  {customOverrideLabels.map((label, i) => (
                    <div key={i} className="text-sm border-l-2 border-blue-500 pl-2 flex items-center gap-2">
                      <code className="bg-muted px-2 py-1 rounded">{label.key}</code>
                      <span>=</span>
                      <span>{label.value}</span>
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
