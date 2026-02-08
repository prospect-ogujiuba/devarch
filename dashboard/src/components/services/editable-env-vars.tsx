import { useState } from 'react'
import { Trash2, Eye, EyeOff } from 'lucide-react'
import { CardContent } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Checkbox } from '@/components/ui/checkbox'
import { EditableCard } from '@/components/ui/editable-card'
import { useEditableSection } from '@/hooks/use-editable-section'
import { useUpdateEnvVars } from '@/features/services/queries'
import type { ServiceEnvVar } from '@/types/api'

interface Props {
  name: string
  envVars: ServiceEnvVar[]
}

interface EnvDraft {
  key: string
  value: string
  is_secret: boolean
}

export function EditableEnvVars({ name, envVars }: Props) {
  const section = useEditableSection<EnvDraft>(() =>
    envVars.map((e) => ({ key: e.key, value: e.value ?? '', is_secret: e.is_secret })),
  )
  const [revealedSecrets, setRevealedSecrets] = useState<Set<number>>(new Set())
  const mutation = useUpdateEnvVars()

  const toggleSecret = (index: number) => {
    setRevealedSecrets((prev) => {
      const next = new Set(prev)
      if (next.has(index)) next.delete(index)
      else next.add(index)
      return next
    })
  }

  const save = () => {
    mutation.mutate({ name, data: { env_vars: section.drafts } }, { onSuccess: () => section.setEditing(false) })
  }

  return (
    <EditableCard
      title="Environment Variables"
      editing={section.editing}
      onEdit={section.startEdit}
      onCancel={section.cancel}
      onSave={save}
      onAdd={() => section.add({ key: '', value: '', is_secret: false })}
      isPending={mutation.isPending}
    >
      <CardContent className={section.editing ? 'space-y-2' : undefined}>
        {section.editing ? (
          <>
            {section.drafts.map((d, i) => (
              <div key={i} className="flex flex-col items-start gap-2 sm:flex-row sm:items-center">
                <Input className="w-full sm:w-48" value={d.key} onChange={(e) => section.update(i, { key: e.target.value })} placeholder="KEY" />
                <span className="text-muted-foreground">=</span>
                <Input className="flex-1" value={d.value} type={d.is_secret ? 'password' : 'text'} onChange={(e) => section.update(i, { value: e.target.value })} placeholder="value" />
                <div className="flex items-center gap-1 sm:ml-2">
                  <Checkbox checked={d.is_secret} onCheckedChange={(v) => section.update(i, { is_secret: !!v })} />
                  <span className="text-xs">Secret</span>
                </div>
                <Button variant="ghost" size="icon-sm" onClick={() => section.remove(i)}><Trash2 className="size-4 text-destructive" /></Button>
              </div>
            ))}
            {section.drafts.length === 0 && <p className="text-muted-foreground text-sm">No environment variables</p>}
          </>
        ) : (
          <>
            {envVars.length > 0 ? (
              <div className="space-y-1">
                {envVars.map((env, i) => (
                  <div key={i} className="text-sm font-mono flex min-w-0 flex-col gap-1 sm:flex-row sm:items-center sm:gap-2">
                    <span className="text-muted-foreground sm:min-w-[200px]">{env.key}:</span>
                    <span className="break-all">{env.is_secret && !revealedSecrets.has(i) ? '********' : env.value}</span>
                    {env.is_secret && (
                      <Button variant="ghost" size="icon-sm" className="size-6 sm:ml-1" onClick={() => toggleSecret(i)}>
                        {revealedSecrets.has(i) ? <EyeOff className="size-3" /> : <Eye className="size-3" />}
                      </Button>
                    )}
                  </div>
                ))}
              </div>
            ) : (
              <p className="text-muted-foreground">No environment variables</p>
            )}
          </>
        )}
      </CardContent>
    </EditableCard>
  )
}
