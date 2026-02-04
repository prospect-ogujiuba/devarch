import { useState } from 'react'
import { Plus, Trash2, X, Check, RotateCcw, Eye, EyeOff } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle, CardAction } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Checkbox } from '@/components/ui/checkbox'
import { useUpdateInstanceEnvVars } from '@/features/instances/queries'
import type { InstanceDetail, Service, InstanceEnvVar } from '@/types/api'

interface Props {
  instance: InstanceDetail
  templateData: Service
  stackName: string
  instanceId: string
}

interface EnvVarDraft {
  key: string
  value: string
  is_secret: boolean
}

function toEnvVarDrafts(envVars: InstanceEnvVar[]): EnvVarDraft[] {
  return envVars.map((e) => ({
    key: e.key,
    value: e.value ?? '',
    is_secret: e.is_secret,
  }))
}

export function OverrideEnvVars({ instance, templateData, stackName, instanceId }: Props) {
  const [editing, setEditing] = useState(false)
  const [drafts, setDrafts] = useState<EnvVarDraft[]>([])
  const [showSecrets, setShowSecrets] = useState<Record<number, boolean>>({})
  const updateEnvVars = useUpdateInstanceEnvVars(stackName, instanceId)

  const templateEnvVars = templateData.env_vars ?? []
  const overrideEnvVars = instance.env_vars ?? []

  const startEdit = () => {
    setDrafts(toEnvVarDrafts(overrideEnvVars))
    setEditing(true)
  }

  const cancel = () => setEditing(false)

  const save = () => {
    updateEnvVars.mutate(drafts, { onSuccess: () => setEditing(false) })
  }

  const resetAll = () => {
    updateEnvVars.mutate([], { onSuccess: () => setEditing(false) })
  }

  const add = () => setDrafts([...drafts, { key: '', value: '', is_secret: false }])
  const remove = (i: number) => setDrafts(drafts.filter((_, idx) => idx !== i))
  const update = (i: number, field: keyof EnvVarDraft, value: string | boolean) => {
    const next = [...drafts]
    next[i] = { ...next[i], [field]: value }
    setDrafts(next)
  }

  const toggleShowSecret = (i: number) => {
    setShowSecrets((prev) => ({ ...prev, [i]: !prev[i] }))
  }

  const isDirty = JSON.stringify(drafts) !== JSON.stringify(toEnvVarDrafts(overrideEnvVars))

  const getTemplatePlaceholder = (key: string): string | undefined => {
    const templateVar = templateEnvVars.find((e) => e.key === key)
    return templateVar?.value
  }

  if (editing) {
    return (
      <Card>
        <CardHeader className="flex flex-row items-center justify-between">
          <CardTitle className="text-base">Environment Variable Overrides</CardTitle>
          <div className="flex gap-1">
            <Button variant="outline" size="sm" onClick={add}>
              <Plus className="size-4" /> Add
            </Button>
            {overrideEnvVars.length > 0 && (
              <Button variant="outline" size="sm" onClick={resetAll}>
                <RotateCcw className="size-4" /> Reset All
              </Button>
            )}
            <Button variant="ghost" size="icon-sm" onClick={cancel}>
              <X className="size-4" />
            </Button>
            <Button variant="default" size="icon-sm" onClick={save} disabled={updateEnvVars.isPending || !isDirty}>
              <Check className="size-4" />
            </Button>
          </div>
        </CardHeader>
        <CardContent className="space-y-4">
          {templateEnvVars.length > 0 && (
            <div>
              <div className="text-xs font-medium text-muted-foreground mb-2">Template Variables (read-only)</div>
              <div className="space-y-2">
                {templateEnvVars.map((env, i) => (
                  <div key={i} className="text-sm text-muted-foreground flex items-center gap-2">
                    <code className="bg-muted px-2 py-1 rounded">{env.key}</code>
                    <span>=</span>
                    <span>{env.is_secret ? '••••••••' : env.value ?? ''}</span>
                    {env.is_secret && <Badge variant="secondary" className="text-xs">secret</Badge>}
                  </div>
                ))}
              </div>
            </div>
          )}

          <div>
            <div className="text-xs font-medium mb-2">Override Variables</div>
            <div className="space-y-3">
              {drafts.map((d, i) => {
                const templatePlaceholder = getTemplatePlaceholder(d.key)
                return (
                  <div key={i} className="border-l-2 border-blue-500 pl-3 space-y-2">
                    <div className="flex gap-2 items-center">
                      <Input
                        className="w-48"
                        value={d.key}
                        onChange={(e) => update(i, 'key', e.target.value)}
                        placeholder="KEY"
                      />
                      <span className="text-muted-foreground">=</span>
                      <div className="flex-1 relative">
                        <Input
                          type={d.is_secret && !showSecrets[i] ? 'password' : 'text'}
                          value={d.value}
                          onChange={(e) => update(i, 'value', e.target.value)}
                          placeholder={templatePlaceholder ? `Template: ${templatePlaceholder}` : 'Value'}
                          className={templatePlaceholder ? 'italic' : ''}
                        />
                      </div>
                      {d.is_secret && (
                        <Button variant="ghost" size="icon-sm" onClick={() => toggleShowSecret(i)}>
                          {showSecrets[i] ? <EyeOff className="size-4" /> : <Eye className="size-4" />}
                        </Button>
                      )}
                      <Button variant="ghost" size="icon-sm" onClick={() => remove(i)}>
                        <Trash2 className="size-4 text-destructive" />
                      </Button>
                    </div>
                    <label className="flex items-center gap-2 text-sm">
                      <Checkbox checked={d.is_secret} onCheckedChange={(checked) => update(i, 'is_secret', !!checked)} />
                      Secret
                    </label>
                  </div>
                )
              })}
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
        <CardTitle className="text-base">Environment Variable Overrides</CardTitle>
        <CardAction>
          <Button variant="outline" size="sm" onClick={startEdit}>
            Edit
          </Button>
        </CardAction>
      </CardHeader>
      <CardContent className="space-y-4">
        {templateEnvVars.length > 0 && (
          <div>
            <div className="text-xs font-medium text-muted-foreground mb-2">Template Variables</div>
            <div className="space-y-2">
              {templateEnvVars.map((env, i) => (
                <div key={i} className="text-sm text-muted-foreground flex items-center gap-2">
                  <code className="bg-muted px-2 py-1 rounded">{env.key}</code>
                  <span>=</span>
                  <span>{env.is_secret ? '••••••••' : env.value ?? ''}</span>
                  {env.is_secret && <Badge variant="secondary" className="text-xs">secret</Badge>}
                </div>
              ))}
            </div>
          </div>
        )}

        {overrideEnvVars.length > 0 ? (
          <div>
            <div className="text-xs font-medium mb-2">Overrides</div>
            <div className="space-y-2">
              {overrideEnvVars.map((env, i) => (
                <div key={i} className="text-sm border-l-2 border-blue-500 pl-2 flex items-center gap-2">
                  <code className="bg-muted px-2 py-1 rounded">{env.key}</code>
                  <span>=</span>
                  <span>{env.is_secret ? '••••••••' : env.value ?? ''}</span>
                  {env.is_secret && <Badge variant="secondary" className="text-xs">secret</Badge>}
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
