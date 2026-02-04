import { useState } from 'react'
import { Plus, Trash2, X, Check, RotateCcw, Lock } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle, CardAction } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
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
  const [editing, setEditing] = useState(false)
  const [drafts, setDrafts] = useState<LabelDraft[]>([])
  const [error, setError] = useState('')
  const updateLabels = useUpdateInstanceLabels(stackName, instanceId)

  const templateLabels = templateData.labels ?? []
  const overrideLabels = instance.labels ?? []

  const startEdit = () => {
    setDrafts(toLabelDrafts(overrideLabels))
    setError('')
    setEditing(true)
  }

  const cancel = () => setEditing(false)

  const save = () => {
    const hasSystemLabel = drafts.some((d) => d.key.startsWith('devarch.'))
    if (hasSystemLabel) {
      setError('Cannot create labels with "devarch." prefix (system-managed)')
      return
    }
    updateLabels.mutate(drafts, { onSuccess: () => setEditing(false) })
  }

  const resetAll = () => {
    updateLabels.mutate([], { onSuccess: () => setEditing(false) })
  }

  const add = () => {
    setDrafts([...drafts, { key: '', value: '' }])
    setError('')
  }

  const remove = (i: number) => setDrafts(drafts.filter((_, idx) => idx !== i))
  const update = (i: number, field: keyof LabelDraft, value: string) => {
    const next = [...drafts]
    next[i] = { ...next[i], [field]: value }
    setDrafts(next)
    setError('')
  }

  const isDirty = JSON.stringify(drafts) !== JSON.stringify(toLabelDrafts(overrideLabels))

  const systemLabels = [...templateLabels, ...overrideLabels].filter((l) => l.key.startsWith('devarch.'))
  const customTemplateLabels = templateLabels.filter((l) => !l.key.startsWith('devarch.'))
  const customOverrideLabels = overrideLabels.filter((l) => !l.key.startsWith('devarch.'))

  if (editing) {
    return (
      <Card>
        <CardHeader className="flex flex-row items-center justify-between">
          <CardTitle className="text-base">Label Overrides</CardTitle>
          <div className="flex gap-1">
            <Button variant="outline" size="sm" onClick={add}>
              <Plus className="size-4" /> Add
            </Button>
            {customOverrideLabels.length > 0 && (
              <Button variant="outline" size="sm" onClick={resetAll}>
                <RotateCcw className="size-4" /> Reset All
              </Button>
            )}
            <Button variant="ghost" size="icon-sm" onClick={cancel}>
              <X className="size-4" />
            </Button>
            <Button variant="default" size="icon-sm" onClick={save} disabled={updateLabels.isPending || !isDirty}>
              <Check className="size-4" />
            </Button>
          </div>
        </CardHeader>
        <CardContent className="space-y-4">
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
              {drafts.map((d, i) => (
                <div key={i} className="border-l-2 border-blue-500 pl-3">
                  <div className="flex gap-2 items-center">
                    <Input
                      className="flex-1"
                      value={d.key}
                      onChange={(e) => update(i, 'key', e.target.value)}
                      placeholder="key"
                    />
                    <span className="text-muted-foreground">=</span>
                    <Input
                      className="flex-1"
                      value={d.value}
                      onChange={(e) => update(i, 'value', e.target.value)}
                      placeholder="value"
                    />
                    <Button variant="ghost" size="icon-sm" onClick={() => remove(i)}>
                      <Trash2 className="size-4 text-destructive" />
                    </Button>
                  </div>
                </div>
              ))}
              {drafts.length === 0 && <p className="text-muted-foreground text-sm">No overrides</p>}
            </div>
          </div>
        </CardContent>
      </Card>
    )
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base">Label Overrides</CardTitle>
        <CardAction>
          <Button variant="outline" size="sm" onClick={startEdit}>
            Edit
          </Button>
        </CardAction>
      </CardHeader>
      <CardContent className="space-y-4">
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
      </CardContent>
    </Card>
  )
}
