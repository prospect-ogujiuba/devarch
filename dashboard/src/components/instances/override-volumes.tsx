import { Trash2 } from 'lucide-react'
import { CardContent } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Checkbox } from '@/components/ui/checkbox'
import { FieldRow } from '@/components/ui/field-row'
import { EditableCard } from '@/components/ui/editable-card'
import { useOverrideSection } from '@/hooks/use-override-section'
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
  const templateVolumes = templateData.volumes ?? []
  const overrideVolumes = instance.volumes ?? []

  const section = useOverrideSection(() => toVolumeDrafts(overrideVolumes))
  const updateVolumes = useUpdateInstanceVolumes(stackName, instanceId)

  const save = () => {
    const draftTargets = new Set(section.drafts.map((d) => d.target))
    const preserved = templateVolumes
      .filter((v) => !draftTargets.has(v.target))
      .map((v) => ({
        volume_type: v.volume_type,
        source: v.source,
        target: v.target,
        read_only: v.read_only,
        is_external: v.is_external,
      }))
    updateVolumes.mutate([...preserved, ...section.drafts], {
      onSuccess: () => section.setEditing(false),
    })
  }

  const resetAll = () => {
    updateVolumes.mutate([], { onSuccess: () => section.setEditing(false) })
  }

  return (
    <EditableCard
      title="Volume Overrides"
      editing={section.editing}
      onEdit={section.startEdit}
      onCancel={section.cancel}
      onSave={save}
      onAdd={() => section.add({ volume_type: 'bind', source: '', target: '', read_only: false, is_external: false })}
      isPending={updateVolumes.isPending}
      isDirty={section.isDirty}
      onResetAll={resetAll}
      showResetAll={overrideVolumes.length > 0}
      editVariant="text"
    >
      <CardContent className="space-y-4">
        {section.editing ? (
          <>
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
                {section.drafts.map((d, i) => (
                  <div key={i} className="border-l-2 border-blue-500 pl-3 space-y-2">
                    <FieldRow>
                      <Select value={d.volume_type} onValueChange={(v) => section.update(i, { volume_type: v })}>
                        <SelectTrigger className="w-full sm:w-32">
                          <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                          <SelectItem value="bind">bind</SelectItem>
                          <SelectItem value="volume">volume</SelectItem>
                          <SelectItem value="tmpfs">tmpfs</SelectItem>
                        </SelectContent>
                      </Select>
                      <Input className="flex-1" value={d.source} onChange={(e) => section.update(i, { source: e.target.value })} placeholder="Source" />
                      <span className="text-muted-foreground hidden sm:inline">→</span>
                      <Input className="flex-1" value={d.target} onChange={(e) => section.update(i, { target: e.target.value })} placeholder="Target" />
                      <Button variant="ghost" size="icon-sm" onClick={() => section.remove(i)}>
                        <Trash2 className="size-4 text-destructive" />
                      </Button>
                    </FieldRow>
                    <div className="flex gap-4 text-sm">
                      <label className="flex items-center gap-2">
                        <Checkbox checked={d.read_only} onCheckedChange={(checked) => section.update(i, { read_only: !!checked })} />
                        Read-only
                      </label>
                      <label className="flex items-center gap-2">
                        <Checkbox checked={d.is_external} onCheckedChange={(checked) => section.update(i, { is_external: !!checked })} />
                        External
                      </label>
                    </div>
                  </div>
                ))}
                {section.drafts.length === 0 && <p className="text-muted-foreground text-sm">No overrides</p>}
              </div>
            </div>
          </>
        ) : (
          <>
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
          </>
        )}
      </CardContent>
    </EditableCard>
  )
}
