import { useState } from 'react'
import { X, Check, RotateCcw } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle, CardAction } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { useUpdateInstanceHealthcheck } from '@/features/instances/queries'
import type { InstanceDetail, Service, InstanceHealthcheck } from '@/types/api'

interface Props {
  instance: InstanceDetail
  templateData: Service
  stackName: string
  instanceId: string
}

interface HealthcheckDraft {
  test: string
  interval_seconds: string
  timeout_seconds: string
  retries: string
  start_period_seconds: string
}

function toHealthcheckDraft(hc: InstanceHealthcheck | null): HealthcheckDraft {
  if (!hc) {
    return {
      test: '',
      interval_seconds: '',
      timeout_seconds: '',
      retries: '',
      start_period_seconds: '',
    }
  }
  return {
    test: hc.test,
    interval_seconds: String(hc.interval_seconds),
    timeout_seconds: String(hc.timeout_seconds),
    retries: String(hc.retries),
    start_period_seconds: String(hc.start_period_seconds),
  }
}

export function OverrideHealthcheck({ instance, templateData, stackName, instanceId }: Props) {
  const [editing, setEditing] = useState(false)
  const [draft, setDraft] = useState<HealthcheckDraft>({
    test: '',
    interval_seconds: '',
    timeout_seconds: '',
    retries: '',
    start_period_seconds: '',
  })
  const updateHealthcheck = useUpdateInstanceHealthcheck(stackName, instanceId)

  const templateHealthcheck = templateData.healthcheck ?? null
  const overrideHealthcheck = instance.healthcheck ?? null

  const startEdit = () => {
    setDraft(toHealthcheckDraft(overrideHealthcheck))
    setEditing(true)
  }

  const cancel = () => setEditing(false)

  const save = () => {
    if (!draft.test.trim()) {
      updateHealthcheck.mutate(null, { onSuccess: () => setEditing(false) })
      return
    }
    updateHealthcheck.mutate(
      {
        test: draft.test,
        interval_seconds: parseInt(draft.interval_seconds, 10) || 30,
        timeout_seconds: parseInt(draft.timeout_seconds, 10) || 3,
        retries: parseInt(draft.retries, 10) || 3,
        start_period_seconds: parseInt(draft.start_period_seconds, 10) || 0,
      },
      { onSuccess: () => setEditing(false) },
    )
  }

  const resetAll = () => {
    updateHealthcheck.mutate(null, { onSuccess: () => setEditing(false) })
  }

  const update = (field: keyof HealthcheckDraft, value: string) => {
    setDraft((prev) => ({ ...prev, [field]: value }))
  }

  const isDirty = JSON.stringify(draft) !== JSON.stringify(toHealthcheckDraft(overrideHealthcheck))

  if (editing) {
    return (
      <Card>
        <CardHeader className="flex flex-row items-center justify-between">
          <CardTitle className="text-base">Healthcheck Override</CardTitle>
          <div className="flex gap-1">
            {overrideHealthcheck && (
              <Button variant="outline" size="sm" onClick={resetAll}>
                <RotateCcw className="size-4" /> Reset
              </Button>
            )}
            <Button variant="ghost" size="icon-sm" onClick={cancel}>
              <X className="size-4" />
            </Button>
            <Button variant="default" size="icon-sm" onClick={save} disabled={updateHealthcheck.isPending || !isDirty}>
              <Check className="size-4" />
            </Button>
          </div>
        </CardHeader>
        <CardContent className="space-y-4">
          {templateHealthcheck && (
            <div>
              <div className="text-xs font-medium text-muted-foreground mb-2">Template Healthcheck (read-only)</div>
              <div className="text-sm text-muted-foreground space-y-1">
                <div><span className="font-medium">Test:</span> {templateHealthcheck.test}</div>
                <div><span className="font-medium">Interval:</span> {templateHealthcheck.interval_seconds}s</div>
                <div><span className="font-medium">Timeout:</span> {templateHealthcheck.timeout_seconds}s</div>
                <div><span className="font-medium">Retries:</span> {templateHealthcheck.retries}</div>
                <div><span className="font-medium">Start Period:</span> {templateHealthcheck.start_period_seconds}s</div>
              </div>
            </div>
          )}

          <div className="border-l-2 border-blue-500 pl-3 space-y-3">
            <div>
              <label className="text-sm font-medium">Test Command</label>
              <Textarea
                value={draft.test}
                onChange={(e) => update('test', e.target.value)}
                placeholder={templateHealthcheck?.test ?? 'CMD-SHELL curl -f http://localhost/ || exit 1'}
                className="mt-1"
              />
            </div>
            <div className="grid grid-cols-2 gap-3">
              <div>
                <label className="text-sm font-medium">Interval (seconds)</label>
                <Input
                  type="number"
                  value={draft.interval_seconds}
                  onChange={(e) => update('interval_seconds', e.target.value)}
                  placeholder={templateHealthcheck?.interval_seconds?.toString() ?? '30'}
                  className="mt-1"
                />
              </div>
              <div>
                <label className="text-sm font-medium">Timeout (seconds)</label>
                <Input
                  type="number"
                  value={draft.timeout_seconds}
                  onChange={(e) => update('timeout_seconds', e.target.value)}
                  placeholder={templateHealthcheck?.timeout_seconds?.toString() ?? '3'}
                  className="mt-1"
                />
              </div>
            </div>
            <div className="grid grid-cols-2 gap-3">
              <div>
                <label className="text-sm font-medium">Retries</label>
                <Input
                  type="number"
                  value={draft.retries}
                  onChange={(e) => update('retries', e.target.value)}
                  placeholder={templateHealthcheck?.retries?.toString() ?? '3'}
                  className="mt-1"
                />
              </div>
              <div>
                <label className="text-sm font-medium">Start Period (seconds)</label>
                <Input
                  type="number"
                  value={draft.start_period_seconds}
                  onChange={(e) => update('start_period_seconds', e.target.value)}
                  placeholder={templateHealthcheck?.start_period_seconds?.toString() ?? '0'}
                  className="mt-1"
                />
              </div>
            </div>
          </div>
        </CardContent>
      </Card>
    )
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base">Healthcheck Override</CardTitle>
        <CardAction>
          <Button variant="outline" size="sm" onClick={startEdit}>
            Edit
          </Button>
        </CardAction>
      </CardHeader>
      <CardContent className="space-y-4">
        {templateHealthcheck && (
          <div>
            <div className="text-xs font-medium text-muted-foreground mb-2">Template Healthcheck</div>
            <div className="text-sm text-muted-foreground space-y-1">
              <div><span className="font-medium">Test:</span> {templateHealthcheck.test}</div>
              <div><span className="font-medium">Interval:</span> {templateHealthcheck.interval_seconds}s</div>
              <div><span className="font-medium">Timeout:</span> {templateHealthcheck.timeout_seconds}s</div>
              <div><span className="font-medium">Retries:</span> {templateHealthcheck.retries}</div>
              <div><span className="font-medium">Start Period:</span> {templateHealthcheck.start_period_seconds}s</div>
            </div>
          </div>
        )}

        {overrideHealthcheck ? (
          <div>
            <div className="text-xs font-medium mb-2">Override</div>
            <div className="text-sm border-l-2 border-blue-500 pl-2 space-y-1">
              <div><span className="font-medium">Test:</span> {overrideHealthcheck.test}</div>
              <div><span className="font-medium">Interval:</span> {overrideHealthcheck.interval_seconds}s</div>
              <div><span className="font-medium">Timeout:</span> {overrideHealthcheck.timeout_seconds}s</div>
              <div><span className="font-medium">Retries:</span> {overrideHealthcheck.retries}</div>
              <div><span className="font-medium">Start Period:</span> {overrideHealthcheck.start_period_seconds}s</div>
            </div>
          </div>
        ) : (
          <p className="text-muted-foreground text-sm">No override configured</p>
        )}
      </CardContent>
    </Card>
  )
}
