import { useState } from 'react'
import { CardContent } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { EditableCard } from '@/components/ui/editable-card'
import { useEditableSection } from '@/hooks/use-editable-section'
import { useResourceLimits, useUpdateResourceLimits } from '@/features/instances/queries'
import type { ResourceLimits as ResourceLimitsType } from '@/types/api'

interface Props {
  stackName: string
  instanceId: string
}

interface ResourceLimitsDraft {
  cpu_limit: string
  cpu_reservation: string
  memory_limit: string
  memory_reservation: string
}

function toLimitsDraft(limits: ResourceLimitsType | null): ResourceLimitsDraft {
  return {
    cpu_limit: limits?.cpu_limit ?? '',
    cpu_reservation: limits?.cpu_reservation ?? '',
    memory_limit: limits?.memory_limit ?? '',
    memory_reservation: limits?.memory_reservation ?? '',
  }
}

function toResourceLimits(draft: ResourceLimitsDraft): ResourceLimitsType {
  const result: ResourceLimitsType = {}
  if (draft.cpu_limit.trim()) result.cpu_limit = draft.cpu_limit.trim()
  if (draft.cpu_reservation.trim()) result.cpu_reservation = draft.cpu_reservation.trim()
  if (draft.memory_limit.trim()) result.memory_limit = draft.memory_limit.trim()
  if (draft.memory_reservation.trim()) result.memory_reservation = draft.memory_reservation.trim()
  return result
}

export function ResourceLimits({ stackName, instanceId }: Props) {
  const { data: limits, isLoading } = useResourceLimits(stackName, instanceId)
  const updateLimits = useUpdateResourceLimits(stackName, instanceId)
  const [warnings, setWarnings] = useState<string[]>([])

  const section = useEditableSection<ResourceLimitsDraft>(
    () => [toLimitsDraft(limits ?? null)]
  )

  const draft = section.drafts[0] ?? toLimitsDraft(null)

  const save = () => {
    const payload = toResourceLimits(draft)
    updateLimits.mutate(payload, {
      onSuccess: (response) => {
        setWarnings(response.warnings ?? [])
        section.setEditing(false)
      },
    })
  }

  const clearAll = () => {
    updateLimits.mutate({}, {
      onSuccess: () => {
        setWarnings([])
        section.setEditing(false)
      },
    })
  }

  const currentLimits = limits ?? null
  const isDirty = JSON.stringify(draft) !== JSON.stringify(toLimitsDraft(currentLimits))

  const hasLimits = limits && (
    limits.cpu_limit || limits.cpu_reservation || limits.memory_limit || limits.memory_reservation
  )

  if (isLoading) {
    return <div className="text-muted-foreground">Loading...</div>
  }

  return (
    <EditableCard
      title="Resource Limits"
      editing={section.editing}
      onEdit={section.startEdit}
      onCancel={() => {
        section.cancel()
        setWarnings([])
      }}
      onSave={save}
      isPending={updateLimits.isPending}
      isDirty={isDirty}
      onResetAll={clearAll}
      showResetAll={!!hasLimits}
      editVariant="text"
    >
      <CardContent className="space-y-4">
        {section.editing ? (
          <>
            <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
              <div className="space-y-2">
                <label className="text-xs font-medium">CPU Limit</label>
                <Input
                  value={draft.cpu_limit}
                  onChange={(e) => section.update(0, { cpu_limit: e.target.value })}
                  placeholder="e.g., 2.0"
                />
                <p className="text-xs text-muted-foreground">Maximum CPU cores</p>
              </div>
              <div className="space-y-2">
                <label className="text-xs font-medium">CPU Reservation</label>
                <Input
                  value={draft.cpu_reservation}
                  onChange={(e) => section.update(0, { cpu_reservation: e.target.value })}
                  placeholder="e.g., 0.5"
                />
                <p className="text-xs text-muted-foreground">Guaranteed CPU cores</p>
              </div>
              <div className="space-y-2">
                <label className="text-xs font-medium">Memory Limit</label>
                <Input
                  value={draft.memory_limit}
                  onChange={(e) => section.update(0, { memory_limit: e.target.value })}
                  placeholder="e.g., 1g"
                />
                <p className="text-xs text-muted-foreground">Maximum memory (e.g., 512m, 1g)</p>
              </div>
              <div className="space-y-2">
                <label className="text-xs font-medium">Memory Reservation</label>
                <Input
                  value={draft.memory_reservation}
                  onChange={(e) => section.update(0, { memory_reservation: e.target.value })}
                  placeholder="e.g., 512m"
                />
                <p className="text-xs text-muted-foreground">Guaranteed memory (e.g., 256m, 512m)</p>
              </div>
            </div>

            {warnings.length > 0 && (
              <div className="space-y-1">
                {warnings.map((warning, i) => (
                  <p key={i} className="text-sm text-amber-600">{warning}</p>
                ))}
              </div>
            )}

            <div className="flex gap-2">
              <Button variant="outline" size="sm" onClick={clearAll} disabled={updateLimits.isPending}>
                Clear All
              </Button>
            </div>
          </>
        ) : (
          <>
            {hasLimits ? (
              <div className="grid grid-cols-1 gap-3 text-sm sm:grid-cols-2">
                {limits.cpu_limit && (
                  <div className="flex flex-col">
                    <span className="text-muted-foreground text-xs">CPU Limit</span>
                    <span className="font-mono">{limits.cpu_limit} cores</span>
                  </div>
                )}
                {limits.cpu_reservation && (
                  <div className="flex flex-col">
                    <span className="text-muted-foreground text-xs">CPU Reservation</span>
                    <span className="font-mono">{limits.cpu_reservation} cores</span>
                  </div>
                )}
                {limits.memory_limit && (
                  <div className="flex flex-col">
                    <span className="text-muted-foreground text-xs">Memory Limit</span>
                    <span className="font-mono">{limits.memory_limit}</span>
                  </div>
                )}
                {limits.memory_reservation && (
                  <div className="flex flex-col">
                    <span className="text-muted-foreground text-xs">Memory Reservation</span>
                    <span className="font-mono">{limits.memory_reservation}</span>
                  </div>
                )}
              </div>
            ) : (
              <p className="text-muted-foreground text-sm">No resource limits configured</p>
            )}
          </>
        )}
      </CardContent>
    </EditableCard>
  )
}
