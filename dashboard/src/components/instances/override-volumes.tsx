import { useState } from 'react'
import { Plus, Trash2, X, Check, RotateCcw } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle, CardAction } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Checkbox } from '@/components/ui/checkbox'
import { useUpdateInstanceVolumes } from '@/features/instances/queries'
import type { InstanceDetail, Service, InstanceVolume } from '@/types/api'

interface Props {
  instance: InstanceDetail
  templateData: Service
  stackName: string
  instanceId: string
}

interface VolumeDraft {
  volume_type: string
  source: string
  target: string
  read_only: boolean
  is_external: boolean
}

function toVolumeDrafts(volumes: InstanceVolume[]): VolumeDraft[] {
  return volumes.map((v) => ({
    volume_type: v.volume_type,
    source: v.source,
    target: v.target,
    read_only: v.read_only,
    is_external: v.is_external,
  }))
}

export function OverrideVolumes({ instance, templateData, stackName, instanceId }: Props) {
  const [editing, setEditing] = useState(false)
  const [drafts, setDrafts] = useState<VolumeDraft[]>([])
  const updateVolumes = useUpdateInstanceVolumes(stackName, instanceId)

  const templateVolumes = templateData.volumes ?? []
  const overrideVolumes = instance.volumes ?? []

  const startEdit = () => {
    setDrafts(toVolumeDrafts(overrideVolumes))
    setEditing(true)
  }

  const cancel = () => setEditing(false)

  const save = () => {
    updateVolumes.mutate(drafts, { onSuccess: () => setEditing(false) })
  }

  const resetAll = () => {
    updateVolumes.mutate([], { onSuccess: () => setEditing(false) })
  }

  const add = () => setDrafts([...drafts, { volume_type: 'bind', source: '', target: '', read_only: false, is_external: false }])
  const remove = (i: number) => setDrafts(drafts.filter((_, idx) => idx !== i))
  const update = (i: number, field: keyof VolumeDraft, value: string | boolean) => {
    const next = [...drafts]
    next[i] = { ...next[i], [field]: value }
    setDrafts(next)
  }

  const isDirty = JSON.stringify(drafts) !== JSON.stringify(toVolumeDrafts(overrideVolumes))

  if (editing) {
    return (
      <Card>
        <CardHeader className="flex flex-row items-center justify-between">
          <CardTitle className="text-base">Volume Overrides</CardTitle>
          <div className="flex gap-1">
            <Button variant="outline" size="sm" onClick={add}>
              <Plus className="size-4" /> Add
            </Button>
            {overrideVolumes.length > 0 && (
              <Button variant="outline" size="sm" onClick={resetAll}>
                <RotateCcw className="size-4" /> Reset All
              </Button>
            )}
            <Button variant="ghost" size="icon-sm" onClick={cancel}>
              <X className="size-4" />
            </Button>
            <Button variant="default" size="icon-sm" onClick={save} disabled={updateVolumes.isPending || !isDirty}>
              <Check className="size-4" />
            </Button>
          </div>
        </CardHeader>
        <CardContent className="space-y-4">
          {templateVolumes.length > 0 && (
            <div>
              <div className="text-xs font-medium text-muted-foreground mb-2">Template Volumes (read-only)</div>
              <div className="space-y-2">
                {templateVolumes.map((vol, i) => (
                  <div key={i} className="text-sm text-muted-foreground">
                    <Badge variant="outline" className="text-muted-foreground">
                      {vol.volume_type}
                    </Badge>
                    <span className="ml-2">{vol.source} → {vol.target}</span>
                    {vol.read_only && <span className="ml-2 text-xs">(ro)</span>}
                  </div>
                ))}
              </div>
            </div>
          )}

          <div>
            <div className="text-xs font-medium mb-2">Override Volumes</div>
            <div className="space-y-3">
              {drafts.map((d, i) => (
                <div key={i} className="border-l-2 border-blue-500 pl-3 space-y-2">
                  <div className="flex gap-2 items-center">
                    <Select value={d.volume_type} onValueChange={(v) => update(i, 'volume_type', v)}>
                      <SelectTrigger className="w-32">
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="bind">bind</SelectItem>
                        <SelectItem value="volume">volume</SelectItem>
                        <SelectItem value="tmpfs">tmpfs</SelectItem>
                      </SelectContent>
                    </Select>
                    <Input className="flex-1" value={d.source} onChange={(e) => update(i, 'source', e.target.value)} placeholder="Source" />
                    <span className="text-muted-foreground">→</span>
                    <Input className="flex-1" value={d.target} onChange={(e) => update(i, 'target', e.target.value)} placeholder="Target" />
                    <Button variant="ghost" size="icon-sm" onClick={() => remove(i)}>
                      <Trash2 className="size-4 text-destructive" />
                    </Button>
                  </div>
                  <div className="flex gap-4 text-sm">
                    <label className="flex items-center gap-2">
                      <Checkbox checked={d.read_only} onCheckedChange={(checked) => update(i, 'read_only', !!checked)} />
                      Read-only
                    </label>
                    <label className="flex items-center gap-2">
                      <Checkbox checked={d.is_external} onCheckedChange={(checked) => update(i, 'is_external', !!checked)} />
                      External
                    </label>
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
        <CardTitle className="text-base">Volume Overrides</CardTitle>
        <CardAction>
          <Button variant="outline" size="sm" onClick={startEdit}>
            Edit
          </Button>
        </CardAction>
      </CardHeader>
      <CardContent className="space-y-4">
        {templateVolumes.length > 0 && (
          <div>
            <div className="text-xs font-medium text-muted-foreground mb-2">Template Volumes</div>
            <div className="space-y-2">
              {templateVolumes.map((vol, i) => (
                <div key={i} className="text-sm text-muted-foreground">
                  <Badge variant="outline" className="text-muted-foreground">
                    {vol.volume_type}
                  </Badge>
                  <span className="ml-2">{vol.source} → {vol.target}</span>
                  {vol.read_only && <span className="ml-2 text-xs">(ro)</span>}
                </div>
              ))}
            </div>
          </div>
        )}

        {overrideVolumes.length > 0 ? (
          <div>
            <div className="text-xs font-medium mb-2">Overrides</div>
            <div className="space-y-2">
              {overrideVolumes.map((vol, i) => (
                <div key={i} className="text-sm border-l-2 border-blue-500 pl-2">
                  <Badge variant="default">
                    {vol.volume_type}
                  </Badge>
                  <span className="ml-2">{vol.source} → {vol.target}</span>
                  {vol.read_only && <span className="ml-2 text-xs">(ro)</span>}
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
