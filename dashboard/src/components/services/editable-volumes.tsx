import { useState } from 'react'
import { Pencil, Plus, Trash2, X, Check } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle, CardAction } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Checkbox } from '@/components/ui/checkbox'
import { useUpdateVolumes } from '@/features/services/queries'
import type { ServiceVolume } from '@/types/api'

interface Props {
  name: string
  volumes: ServiceVolume[]
}

interface VolDraft {
  volume_type: string
  source: string
  target: string
  read_only: boolean
  is_external: boolean
}

export function EditableVolumes({ name, volumes }: Props) {
  const [editing, setEditing] = useState(false)
  const [drafts, setDrafts] = useState<VolDraft[]>([])
  const mutation = useUpdateVolumes()

  const startEdit = () => {
    setDrafts(volumes.map((v) => ({ volume_type: v.volume_type, source: v.source, target: v.target, read_only: v.read_only, is_external: v.is_external })))
    setEditing(true)
  }

  const save = () => {
    mutation.mutate({ name, data: { volumes: drafts } }, { onSuccess: () => setEditing(false) })
  }

  const add = () => setDrafts([...drafts, { volume_type: 'bind', source: '', target: '', read_only: false, is_external: false }])
  const remove = (i: number) => setDrafts(drafts.filter((_, idx) => idx !== i))

  if (editing) {
    return (
      <Card>
        <CardHeader className="flex flex-row items-center justify-between">
          <CardTitle className="text-base">Volumes</CardTitle>
          <div className="flex gap-1">
            <Button variant="outline" size="sm" onClick={add}><Plus className="size-4" /> Add</Button>
            <Button variant="ghost" size="icon-sm" onClick={() => setEditing(false)}><X className="size-4" /></Button>
            <Button variant="default" size="icon-sm" onClick={save} disabled={mutation.isPending}><Check className="size-4" /></Button>
          </div>
        </CardHeader>
        <CardContent className="space-y-2">
          {drafts.map((d, i) => (
            <div key={i} className="flex gap-2 items-center">
              <Select value={d.volume_type} onValueChange={(v) => { const next = [...drafts]; next[i] = { ...d, volume_type: v }; setDrafts(next) }}>
                <SelectTrigger className="w-24"><SelectValue /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="bind">bind</SelectItem>
                  <SelectItem value="volume">volume</SelectItem>
                  <SelectItem value="tmpfs">tmpfs</SelectItem>
                </SelectContent>
              </Select>
              <Input className="flex-1" value={d.source} onChange={(e) => { const next = [...drafts]; next[i] = { ...d, source: e.target.value }; setDrafts(next) }} placeholder="Source" />
              <span className="text-muted-foreground">:</span>
              <Input className="flex-1" value={d.target} onChange={(e) => { const next = [...drafts]; next[i] = { ...d, target: e.target.value }; setDrafts(next) }} placeholder="Target" />
              <div className="flex items-center gap-1">
                <Checkbox checked={d.read_only} onCheckedChange={(v) => { const next = [...drafts]; next[i] = { ...d, read_only: !!v }; setDrafts(next) }} />
                <span className="text-xs">RO</span>
              </div>
              <div className="flex items-center gap-1">
                <Checkbox checked={d.is_external} onCheckedChange={(v) => { const next = [...drafts]; next[i] = { ...d, is_external: !!v }; setDrafts(next) }} />
                <span className="text-xs">Ext</span>
              </div>
              <Button variant="ghost" size="icon-sm" onClick={() => remove(i)}><Trash2 className="size-4 text-destructive" /></Button>
            </div>
          ))}
          {drafts.length === 0 && <p className="text-muted-foreground text-sm">No volumes</p>}
        </CardContent>
      </Card>
    )
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base">Volumes</CardTitle>
        <CardAction>
          <Button variant="ghost" size="icon-sm" onClick={startEdit}><Pencil className="size-4" /></Button>
        </CardAction>
      </CardHeader>
      <CardContent>
        {volumes.length > 0 ? (
          <div className="space-y-1">
            {volumes.map((vol, i) => (
              <div key={i} className="text-sm font-mono text-muted-foreground">
                {vol.source}:{vol.target}{vol.read_only ? ' (ro)' : ''}
              </div>
            ))}
          </div>
        ) : (
          <p className="text-muted-foreground">No volumes mounted</p>
        )}
      </CardContent>
    </Card>
  )
}
