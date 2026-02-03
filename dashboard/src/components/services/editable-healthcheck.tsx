import { useState } from 'react'
import { Pencil, X, Check, Trash2 } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle, CardAction } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { useUpdateHealthcheck } from '@/features/services/queries'
import type { ServiceHealthcheck } from '@/types/api'

interface Props {
  name: string
  healthcheck: ServiceHealthcheck | null
}

interface HcDraft {
  test: string
  interval_seconds: number
  timeout_seconds: number
  retries: number
  start_period_seconds: number
}

const defaultHc: HcDraft = { test: '', interval_seconds: 30, timeout_seconds: 10, retries: 3, start_period_seconds: 0 }

export function EditableHealthcheck({ name, healthcheck }: Props) {
  const [editing, setEditing] = useState(false)
  const [draft, setDraft] = useState<HcDraft>(defaultHc)
  const mutation = useUpdateHealthcheck()

  const startEdit = () => {
    setDraft(healthcheck ? {
      test: healthcheck.test,
      interval_seconds: healthcheck.interval_seconds,
      timeout_seconds: healthcheck.timeout_seconds,
      retries: healthcheck.retries,
      start_period_seconds: healthcheck.start_period_seconds,
    } : defaultHc)
    setEditing(true)
  }

  const save = () => {
    mutation.mutate(
      { name, data: draft.test ? draft as any : null },
      { onSuccess: () => setEditing(false) },
    )
  }

  const removeHc = () => {
    mutation.mutate(
      { name, data: null },
      { onSuccess: () => setEditing(false) },
    )
  }

  if (editing) {
    return (
      <Card>
        <CardHeader className="flex flex-row items-center justify-between">
          <CardTitle className="text-base">Healthcheck</CardTitle>
          <div className="flex gap-1">
            {healthcheck && <Button variant="destructive" size="sm" onClick={removeHc}><Trash2 className="size-4" /> Remove</Button>}
            <Button variant="ghost" size="icon-sm" onClick={() => setEditing(false)}><X className="size-4" /></Button>
            <Button variant="default" size="icon-sm" onClick={save} disabled={mutation.isPending}><Check className="size-4" /></Button>
          </div>
        </CardHeader>
        <CardContent className="grid gap-4 sm:grid-cols-2">
          <div className="grid gap-2 sm:col-span-2">
            <label className="text-sm font-medium">Test Command</label>
            <Input value={draft.test} onChange={(e) => setDraft({ ...draft, test: e.target.value })} placeholder="CMD-SHELL curl -f http://localhost/ || exit 1" />
          </div>
          <div className="grid gap-2">
            <label className="text-sm font-medium">Interval (s)</label>
            <Input type="number" value={draft.interval_seconds} onChange={(e) => setDraft({ ...draft, interval_seconds: parseInt(e.target.value, 10) || 0 })} />
          </div>
          <div className="grid gap-2">
            <label className="text-sm font-medium">Timeout (s)</label>
            <Input type="number" value={draft.timeout_seconds} onChange={(e) => setDraft({ ...draft, timeout_seconds: parseInt(e.target.value, 10) || 0 })} />
          </div>
          <div className="grid gap-2">
            <label className="text-sm font-medium">Retries</label>
            <Input type="number" value={draft.retries} onChange={(e) => setDraft({ ...draft, retries: parseInt(e.target.value, 10) || 0 })} />
          </div>
          <div className="grid gap-2">
            <label className="text-sm font-medium">Start Period (s)</label>
            <Input type="number" value={draft.start_period_seconds} onChange={(e) => setDraft({ ...draft, start_period_seconds: parseInt(e.target.value, 10) || 0 })} />
          </div>
        </CardContent>
      </Card>
    )
  }

  if (!healthcheck) {
    return (
      <Card>
        <CardHeader>
          <CardTitle className="text-base">Healthcheck</CardTitle>
          <CardAction>
            <Button variant="ghost" size="icon-sm" onClick={startEdit}><Pencil className="size-4" /></Button>
          </CardAction>
        </CardHeader>
        <CardContent>
          <p className="text-muted-foreground">No healthcheck configured</p>
        </CardContent>
      </Card>
    )
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base">Healthcheck</CardTitle>
        <CardAction>
          <Button variant="ghost" size="icon-sm" onClick={startEdit}><Pencil className="size-4" /></Button>
        </CardAction>
      </CardHeader>
      <CardContent>
        <div className="grid gap-2 text-sm">
          <div className="flex"><span className="text-muted-foreground w-40">Test:</span> <code>{healthcheck.test}</code></div>
          <div className="flex"><span className="text-muted-foreground w-40">Interval:</span> {healthcheck.interval_seconds}s</div>
          <div className="flex"><span className="text-muted-foreground w-40">Timeout:</span> {healthcheck.timeout_seconds}s</div>
          <div className="flex"><span className="text-muted-foreground w-40">Retries:</span> {healthcheck.retries}</div>
        </div>
      </CardContent>
    </Card>
  )
}
