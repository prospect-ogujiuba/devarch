import { useState } from 'react'
import { Plus, Trash2, X, Check, RotateCcw } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle, CardAction } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { useUpdateInstanceDependencies } from '@/features/instances/queries'
import type { InstanceDetail, Service, InstanceDependency } from '@/types/api'

interface Props {
  instance: InstanceDetail
  templateData: Service
  stackName: string
  instanceId: string
}

interface DependencyDraft {
  depends_on: string
  condition: string
}

function toDrafts(deps: InstanceDependency[]): DependencyDraft[] {
  return deps.map((d) => ({
    depends_on: d.depends_on,
    condition: d.condition,
  }))
}

const CONDITIONS = ['service_started', 'service_healthy', 'service_completed_successfully'] as const

export function OverrideDependencies({ instance, templateData, stackName, instanceId }: Props) {
  const [editing, setEditing] = useState(false)
  const [drafts, setDrafts] = useState<DependencyDraft[]>([])
  const updateDeps = useUpdateInstanceDependencies(stackName, instanceId)

  const templateDeps = templateData.dependencies ?? []
  const overrideDeps = instance.dependencies ?? []

  const startEdit = () => {
    setDrafts(toDrafts(overrideDeps))
    setEditing(true)
  }

  const cancel = () => setEditing(false)

  const save = () => {
    updateDeps.mutate(drafts, { onSuccess: () => setEditing(false) })
  }

  const resetAll = () => {
    updateDeps.mutate([], { onSuccess: () => setEditing(false) })
  }

  const add = () => setDrafts([...drafts, { depends_on: '', condition: 'service_started' }])
  const remove = (i: number) => setDrafts(drafts.filter((_, idx) => idx !== i))
  const update = (i: number, field: keyof DependencyDraft, value: string) => {
    const next = [...drafts]
    next[i] = { ...next[i], [field]: value }
    setDrafts(next)
  }

  const isDirty = JSON.stringify(drafts) !== JSON.stringify(toDrafts(overrideDeps))

  if (editing) {
    return (
      <Card>
        <CardHeader className="flex flex-row items-center justify-between">
          <CardTitle className="text-base">Dependency Overrides</CardTitle>
          <div className="flex gap-1">
            <Button variant="outline" size="sm" onClick={add}>
              <Plus className="size-4" /> Add
            </Button>
            {overrideDeps.length > 0 && (
              <Button variant="outline" size="sm" onClick={resetAll}>
                <RotateCcw className="size-4" /> Reset All
              </Button>
            )}
            <Button variant="ghost" size="icon-sm" onClick={cancel}>
              <X className="size-4" />
            </Button>
            <Button variant="default" size="icon-sm" onClick={save} disabled={updateDeps.isPending || !isDirty}>
              <Check className="size-4" />
            </Button>
          </div>
        </CardHeader>
        <CardContent className="space-y-4">
          {templateDeps.length > 0 && (
            <div>
              <div className="text-xs font-medium text-muted-foreground mb-2">Template Dependencies (read-only)</div>
              <div className="space-y-2">
                {templateDeps.map((dep, i) => (
                  <div key={i} className="text-sm text-muted-foreground flex items-center gap-2">
                    <Badge variant="outline" className="text-muted-foreground">
                      {dep}
                    </Badge>
                  </div>
                ))}
              </div>
            </div>
          )}

          <div>
            <div className="text-xs font-medium mb-2">Override Dependencies</div>
            <div className="space-y-2">
              {drafts.map((d, i) => (
                <div key={i} className="border-l-2 border-blue-500 pl-3 flex gap-2 items-center">
                  <Input
                    className="flex-1"
                    value={d.depends_on}
                    onChange={(e) => update(i, 'depends_on', e.target.value)}
                    placeholder="service-name"
                  />
                  <select
                    className="h-9 rounded-md border border-input bg-background px-3 text-sm"
                    value={d.condition}
                    onChange={(e) => update(i, 'condition', e.target.value)}
                  >
                    {CONDITIONS.map((c) => (
                      <option key={c} value={c}>{c.replace('service_', '')}</option>
                    ))}
                  </select>
                  <Button variant="ghost" size="icon-sm" onClick={() => remove(i)}>
                    <Trash2 className="size-4 text-destructive" />
                  </Button>
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
        <CardTitle className="text-base">Dependency Overrides</CardTitle>
        <CardAction>
          <Button variant="outline" size="sm" onClick={startEdit}>
            Edit
          </Button>
        </CardAction>
      </CardHeader>
      <CardContent className="space-y-4">
        {templateDeps.length > 0 && (
          <div>
            <div className="text-xs font-medium text-muted-foreground mb-2">Template Dependencies</div>
            <div className="space-y-2">
              {templateDeps.map((dep, i) => (
                <div key={i} className="text-sm text-muted-foreground flex items-center gap-2">
                  <Badge variant="outline" className="text-muted-foreground">
                    {dep}
                  </Badge>
                </div>
              ))}
            </div>
          </div>
        )}

        {overrideDeps.length > 0 ? (
          <div>
            <div className="text-xs font-medium mb-2">Overrides</div>
            <div className="space-y-2">
              {overrideDeps.map((dep, i) => (
                <div key={i} className="text-sm border-l-2 border-blue-500 pl-2 flex items-center gap-2">
                  <Badge variant="default">
                    {dep.depends_on}
                  </Badge>
                  <span className="text-xs text-muted-foreground">{dep.condition}</span>
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
