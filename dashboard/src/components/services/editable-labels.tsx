import { useState } from 'react'
import { Pencil, Plus, Trash2, X, Check } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle, CardAction } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { useUpdateLabels } from '@/features/services/queries'
import type { ServiceLabel } from '@/types/api'

interface Props {
  name: string
  labels: ServiceLabel[]
}

interface LabelDraft {
  key: string
  value: string
}

export function EditableLabels({ name, labels }: Props) {
  const [editing, setEditing] = useState(false)
  const [drafts, setDrafts] = useState<LabelDraft[]>([])
  const mutation = useUpdateLabels()

  const startEdit = () => {
    setDrafts(labels.map((l) => ({ key: l.key, value: l.value })))
    setEditing(true)
  }

  const save = () => {
    mutation.mutate({ name, data: { labels: drafts } }, { onSuccess: () => setEditing(false) })
  }

  const add = () => setDrafts([...drafts, { key: '', value: '' }])
  const remove = (i: number) => setDrafts(drafts.filter((_, idx) => idx !== i))

  if (editing) {
    return (
      <Card>
        <CardHeader className="flex flex-row items-center justify-between">
          <CardTitle className="text-base">Labels</CardTitle>
          <div className="flex gap-1">
            <Button variant="outline" size="sm" onClick={add}><Plus className="size-4" /> Add</Button>
            <Button variant="ghost" size="icon-sm" onClick={() => setEditing(false)}><X className="size-4" /></Button>
            <Button variant="default" size="icon-sm" onClick={save} disabled={mutation.isPending}><Check className="size-4" /></Button>
          </div>
        </CardHeader>
        <CardContent className="space-y-2">
          {drafts.map((d, i) => (
            <div key={i} className="flex gap-2 items-center">
              <Input className="flex-1" value={d.key} onChange={(e) => { const next = [...drafts]; next[i] = { ...d, key: e.target.value }; setDrafts(next) }} placeholder="key" />
              <span className="text-muted-foreground">=</span>
              <Input className="flex-1" value={d.value} onChange={(e) => { const next = [...drafts]; next[i] = { ...d, value: e.target.value }; setDrafts(next) }} placeholder="value" />
              <Button variant="ghost" size="icon-sm" onClick={() => remove(i)}><Trash2 className="size-4 text-destructive" /></Button>
            </div>
          ))}
          {drafts.length === 0 && <p className="text-muted-foreground text-sm">No labels</p>}
        </CardContent>
      </Card>
    )
  }

  if (labels.length === 0) return null

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base">Labels</CardTitle>
        <CardAction>
          <Button variant="ghost" size="icon-sm" onClick={startEdit}><Pencil className="size-4" /></Button>
        </CardAction>
      </CardHeader>
      <CardContent>
        <div className="space-y-1">
          {labels.map((l, i) => (
            <div key={i} className="text-sm font-mono text-muted-foreground">
              {l.key}={l.value}
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  )
}
