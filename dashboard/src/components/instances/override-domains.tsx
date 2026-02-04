import { useState } from 'react'
import { Plus, Trash2, X, Check, RotateCcw } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle, CardAction } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { useUpdateInstanceDomains } from '@/features/instances/queries'
import type { InstanceDetail, Service, InstanceDomain } from '@/types/api'

interface Props {
  instance: InstanceDetail
  templateData: Service
  stackName: string
  instanceId: string
}

interface DomainDraft {
  domain: string
  proxy_port?: number
}

function toDomainDrafts(domains: InstanceDomain[]): DomainDraft[] {
  return domains.map((d) => ({
    domain: d.domain,
    proxy_port: d.proxy_port,
  }))
}

export function OverrideDomains({ instance, templateData, stackName, instanceId }: Props) {
  const [editing, setEditing] = useState(false)
  const [drafts, setDrafts] = useState<DomainDraft[]>([])
  const updateDomains = useUpdateInstanceDomains(stackName, instanceId)

  const templateDomains = templateData.domains ?? []
  const overrideDomains = instance.domains ?? []

  const startEdit = () => {
    setDrafts(toDomainDrafts(overrideDomains))
    setEditing(true)
  }

  const cancel = () => setEditing(false)

  const save = () => {
    updateDomains.mutate(drafts, { onSuccess: () => setEditing(false) })
  }

  const resetAll = () => {
    updateDomains.mutate([], { onSuccess: () => setEditing(false) })
  }

  const add = () => setDrafts([...drafts, { domain: '', proxy_port: undefined }])
  const remove = (i: number) => setDrafts(drafts.filter((_, idx) => idx !== i))
  const update = (i: number, field: keyof DomainDraft, value: string | number | undefined) => {
    const next = [...drafts]
    next[i] = { ...next[i], [field]: value }
    setDrafts(next)
  }

  const isDirty = JSON.stringify(drafts) !== JSON.stringify(toDomainDrafts(overrideDomains))

  if (editing) {
    return (
      <Card>
        <CardHeader className="flex flex-row items-center justify-between">
          <CardTitle className="text-base">Domain Overrides</CardTitle>
          <div className="flex gap-1">
            <Button variant="outline" size="sm" onClick={add}>
              <Plus className="size-4" /> Add
            </Button>
            {overrideDomains.length > 0 && (
              <Button variant="outline" size="sm" onClick={resetAll}>
                <RotateCcw className="size-4" /> Reset All
              </Button>
            )}
            <Button variant="ghost" size="icon-sm" onClick={cancel}>
              <X className="size-4" />
            </Button>
            <Button variant="default" size="icon-sm" onClick={save} disabled={updateDomains.isPending || !isDirty}>
              <Check className="size-4" />
            </Button>
          </div>
        </CardHeader>
        <CardContent className="space-y-4">
          {templateDomains.length > 0 && (
            <div>
              <div className="text-xs font-medium text-muted-foreground mb-2">Template Domains (read-only)</div>
              <div className="space-y-2">
                {templateDomains.map((domain, i) => (
                  <div key={i} className="text-sm text-muted-foreground flex items-center gap-2">
                    <Badge variant="outline" className="text-muted-foreground">
                      {domain.domain}
                    </Badge>
                    {domain.proxy_port && <span className="text-xs">→ :{domain.proxy_port}</span>}
                  </div>
                ))}
              </div>
            </div>
          )}

          <div>
            <div className="text-xs font-medium mb-2">Override Domains</div>
            <div className="space-y-2">
              {drafts.map((d, i) => (
                <div key={i} className="border-l-2 border-blue-500 pl-3 flex gap-2 items-center">
                  <Input
                    className="flex-1"
                    value={d.domain}
                    onChange={(e) => update(i, 'domain', e.target.value)}
                    placeholder="example.local"
                  />
                  <Input
                    className="w-32"
                    type="number"
                    value={d.proxy_port ?? ''}
                    onChange={(e) => update(i, 'proxy_port', e.target.value ? parseInt(e.target.value, 10) : undefined)}
                    placeholder="Proxy port"
                  />
                  <Button variant="ghost" size="icon-sm" onClick={() => remove(i)}>
                    <Trash2 className="size-4 text-destructive" />
                  </Button>
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
        <CardTitle className="text-base">Domain Overrides</CardTitle>
        <CardAction>
          <Button variant="outline" size="sm" onClick={startEdit}>
            Edit
          </Button>
        </CardAction>
      </CardHeader>
      <CardContent className="space-y-4">
        {templateDomains.length > 0 && (
          <div>
            <div className="text-xs font-medium text-muted-foreground mb-2">Template Domains</div>
            <div className="space-y-2">
              {templateDomains.map((domain, i) => (
                <div key={i} className="text-sm text-muted-foreground flex items-center gap-2">
                  <Badge variant="outline" className="text-muted-foreground">
                    {domain.domain}
                  </Badge>
                  {domain.proxy_port && <span className="text-xs">→ :{domain.proxy_port}</span>}
                </div>
              ))}
            </div>
          </div>
        )}

        {overrideDomains.length > 0 ? (
          <div>
            <div className="text-xs font-medium mb-2">Overrides</div>
            <div className="space-y-2">
              {overrideDomains.map((domain, i) => (
                <div key={i} className="text-sm border-l-2 border-blue-500 pl-2 flex items-center gap-2">
                  <Badge variant="default">
                    {domain.domain}
                  </Badge>
                  {domain.proxy_port && <span className="text-xs">→ :{domain.proxy_port}</span>}
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
