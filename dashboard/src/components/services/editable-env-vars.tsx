import { useState } from 'react'
import { Pencil, Plus, Trash2, X, Check, Eye, EyeOff } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle, CardAction } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Checkbox } from '@/components/ui/checkbox'
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
  const [editing, setEditing] = useState(false)
  const [drafts, setDrafts] = useState<EnvDraft[]>([])
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

  const startEdit = () => {
    setDrafts(envVars.map((e) => ({ key: e.key, value: e.value ?? '', is_secret: e.is_secret })))
    setEditing(true)
  }

  const save = () => {
    mutation.mutate({ name, data: { env_vars: drafts } }, { onSuccess: () => setEditing(false) })
  }

  const add = () => setDrafts([...drafts, { key: '', value: '', is_secret: false }])
  const remove = (i: number) => setDrafts(drafts.filter((_, idx) => idx !== i))

  if (editing) {
    return (
      <Card>
        <CardHeader className="flex flex-row items-center justify-between">
          <CardTitle className="text-base">Environment Variables</CardTitle>
          <div className="flex gap-1">
            <Button variant="outline" size="sm" onClick={add}><Plus className="size-4" /> Add</Button>
            <Button variant="ghost" size="icon-sm" onClick={() => setEditing(false)}><X className="size-4" /></Button>
            <Button variant="default" size="icon-sm" onClick={save} disabled={mutation.isPending}><Check className="size-4" /></Button>
          </div>
        </CardHeader>
        <CardContent className="space-y-2">
          {drafts.map((d, i) => (
            <div key={i} className="flex gap-2 items-center">
              <Input className="w-48" value={d.key} onChange={(e) => { const next = [...drafts]; next[i] = { ...d, key: e.target.value }; setDrafts(next) }} placeholder="KEY" />
              <span className="text-muted-foreground">=</span>
              <Input className="flex-1" value={d.value} type={d.is_secret ? 'password' : 'text'} onChange={(e) => { const next = [...drafts]; next[i] = { ...d, value: e.target.value }; setDrafts(next) }} placeholder="value" />
              <div className="flex items-center gap-1">
                <Checkbox checked={d.is_secret} onCheckedChange={(v) => { const next = [...drafts]; next[i] = { ...d, is_secret: !!v }; setDrafts(next) }} />
                <span className="text-xs">Secret</span>
              </div>
              <Button variant="ghost" size="icon-sm" onClick={() => remove(i)}><Trash2 className="size-4 text-destructive" /></Button>
            </div>
          ))}
          {drafts.length === 0 && <p className="text-muted-foreground text-sm">No environment variables</p>}
        </CardContent>
      </Card>
    )
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base">Environment Variables</CardTitle>
        <CardAction>
          <Button variant="ghost" size="icon-sm" onClick={startEdit}><Pencil className="size-4" /></Button>
        </CardAction>
      </CardHeader>
      <CardContent>
        {envVars.length > 0 ? (
          <div className="space-y-1">
            {envVars.map((env, i) => (
              <div key={i} className="text-sm font-mono flex items-center">
                <span className="text-muted-foreground min-w-[200px]">{env.key}:</span>
                <span>{env.is_secret && !revealedSecrets.has(i) ? '********' : env.value}</span>
                {env.is_secret && (
                  <Button variant="ghost" size="icon-sm" className="size-6 ml-1" onClick={() => toggleSecret(i)}>
                    {revealedSecrets.has(i) ? <EyeOff className="size-3" /> : <Eye className="size-3" />}
                  </Button>
                )}
              </div>
            ))}
          </div>
        ) : (
          <p className="text-muted-foreground">No environment variables</p>
        )}
      </CardContent>
    </Card>
  )
}
