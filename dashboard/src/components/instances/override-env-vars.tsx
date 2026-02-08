import { useState } from 'react'
import { Trash2, Eye, EyeOff } from 'lucide-react'
import { CardContent } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Checkbox } from '@/components/ui/checkbox'
import { EditableCard } from '@/components/ui/editable-card'
import { useOverrideSection } from '@/hooks/use-override-section'
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
  const templateEnvVars = templateData.env_vars ?? []
  const overrideEnvVars = instance.env_vars ?? []

  const section = useOverrideSection(() => toEnvVarDrafts(overrideEnvVars))
  const [showSecrets, setShowSecrets] = useState<Record<number, boolean>>({})
  const updateEnvVars = useUpdateInstanceEnvVars(stackName, instanceId)

  const save = () => {
    updateEnvVars.mutate(section.drafts, { onSuccess: () => section.setEditing(false) })
  }

  const resetAll = () => {
    updateEnvVars.mutate([], { onSuccess: () => section.setEditing(false) })
  }

  const toggleShowSecret = (i: number) => {
    setShowSecrets((prev) => ({ ...prev, [i]: !prev[i] }))
  }

  const getTemplatePlaceholder = (key: string): string | undefined => {
    const templateVar = templateEnvVars.find((e) => e.key === key)
    return templateVar?.value
  }

  return (
    <EditableCard
      title="Environment Variable Overrides"
      editing={section.editing}
      onEdit={section.startEdit}
      onCancel={section.cancel}
      onSave={save}
      onAdd={() => section.add({ key: '', value: '', is_secret: false })}
      isPending={updateEnvVars.isPending}
      isDirty={section.isDirty}
      onResetAll={resetAll}
      showResetAll={overrideEnvVars.length > 0}
      editVariant="text"
    >
      <CardContent className="space-y-4">
        {section.editing ? (
          <>
            {templateEnvVars.length > 0 && (
              <div>
                <div className="text-xs font-medium text-muted-foreground mb-2">Template Variables (read-only)</div>
                <div className="space-y-2">
                  {templateEnvVars.map((env, i) => (
                    <div key={i} className="text-sm text-muted-foreground flex flex-wrap items-center gap-2">
                      <code className="bg-muted px-2 py-1 rounded break-all">{env.key}</code>
                      <span>=</span>
                      <span className="break-all">{env.is_secret ? '••••••••' : env.value ?? ''}</span>
                      {env.is_secret && <Badge variant="secondary" className="text-xs">secret</Badge>}
                    </div>
                  ))}
                </div>
              </div>
            )}

            <div>
              <div className="text-xs font-medium mb-2">Override Variables</div>
              <div className="space-y-3">
                {section.drafts.map((d, i) => {
                  const templatePlaceholder = getTemplatePlaceholder(d.key)
                  return (
                    <div key={i} className="border-l-2 border-blue-500 pl-3 space-y-2">
                      <div className="flex flex-col items-start gap-2 sm:flex-row sm:items-center">
                        <Input
                          className="w-full sm:w-48"
                          value={d.key}
                          onChange={(e) => section.update(i, { key: e.target.value })}
                          placeholder="KEY"
                        />
                        <span className="text-muted-foreground">=</span>
                        <div className="flex-1 relative">
                          <Input
                            type={d.is_secret && !showSecrets[i] ? 'password' : 'text'}
                            value={d.value}
                            onChange={(e) => section.update(i, { value: e.target.value })}
                            placeholder={templatePlaceholder ? `Template: ${templatePlaceholder}` : 'Value'}
                            className={templatePlaceholder ? 'italic' : ''}
                          />
                        </div>
                        {d.is_secret && (
                          <Button variant="ghost" size="icon-sm" onClick={() => toggleShowSecret(i)}>
                            {showSecrets[i] ? <EyeOff className="size-4" /> : <Eye className="size-4" />}
                          </Button>
                        )}
                        <Button variant="ghost" size="icon-sm" onClick={() => section.remove(i)}>
                          <Trash2 className="size-4 text-destructive" />
                        </Button>
                      </div>
                      <label className="flex items-center gap-2 text-sm">
                        <Checkbox checked={d.is_secret} onCheckedChange={(checked) => section.update(i, { is_secret: !!checked })} />
                        Secret
                      </label>
                    </div>
                  )
                })}
                {section.drafts.length === 0 && <p className="text-muted-foreground text-sm">No overrides</p>}
              </div>
            </div>
          </>
        ) : (
          <>
            {templateEnvVars.length > 0 && (
              <div>
                <div className="text-xs font-medium text-muted-foreground mb-2">Template Variables</div>
                <div className="space-y-2">
                  {templateEnvVars.map((env, i) => (
                    <div key={i} className="text-sm text-muted-foreground flex flex-wrap items-center gap-2">
                      <code className="bg-muted px-2 py-1 rounded break-all">{env.key}</code>
                      <span>=</span>
                      <span className="break-all">{env.is_secret ? '••••••••' : env.value ?? ''}</span>
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
                    <div key={i} className="text-sm border-l-2 border-blue-500 pl-2 flex flex-wrap items-center gap-2">
                      <code className="bg-muted px-2 py-1 rounded break-all">{env.key}</code>
                      <span>=</span>
                      <span className="break-all">{env.is_secret ? '••••••••' : env.value ?? ''}</span>
                      {env.is_secret && <Badge variant="secondary" className="text-xs">secret</Badge>}
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
