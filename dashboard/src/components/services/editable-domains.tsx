import { useState } from 'react'
import { Pencil, Plus, Trash2, X, Check, ExternalLink } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle, CardAction } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { CopyButton } from '@/components/ui/copy-button'
import { useUpdateDomains } from '@/features/services/queries'
import type { ServiceDomain } from '@/types/api'

interface Props {
  name: string
  domains: ServiceDomain[]
}

interface DomainDraft {
  domain: string
  proxy_port: string
}

export function EditableDomains({ name, domains }: Props) {
  const [editing, setEditing] = useState(false)
  const [drafts, setDrafts] = useState<DomainDraft[]>([])
  const mutation = useUpdateDomains()

  const startEdit = () => {
    setDrafts(domains.map((d) => ({ domain: d.domain, proxy_port: d.proxy_port ? String(d.proxy_port) : '' })))
    setEditing(true)
  }

  const save = () => {
    mutation.mutate(
      {
        name,
        data: {
          domains: drafts.map((d) => ({
            domain: d.domain,
            proxy_port: parseInt(d.proxy_port, 10) || 0,
          })),
        },
      },
      { onSuccess: () => setEditing(false) },
    )
  }

  const add = () => setDrafts([...drafts, { domain: '', proxy_port: '' }])
  const remove = (i: number) => setDrafts(drafts.filter((_, idx) => idx !== i))

  if (editing) {
    return (
      <Card>
        <CardHeader className="flex flex-row items-center justify-between">
          <CardTitle className="text-base">Domains</CardTitle>
          <div className="flex gap-1">
            <Button variant="outline" size="sm" onClick={add}><Plus className="size-4" /> Add</Button>
            <Button variant="ghost" size="icon-sm" onClick={() => setEditing(false)}><X className="size-4" /></Button>
            <Button variant="default" size="icon-sm" onClick={save} disabled={mutation.isPending}><Check className="size-4" /></Button>
          </div>
        </CardHeader>
        <CardContent className="space-y-2">
          {drafts.map((d, i) => (
            <div key={i} className="flex gap-2 items-center">
              <Input className="flex-1" value={d.domain} onChange={(e) => { const next = [...drafts]; next[i] = { ...d, domain: e.target.value }; setDrafts(next) }} placeholder="example.test" />
              <Input className="w-24" type="number" value={d.proxy_port} onChange={(e) => { const next = [...drafts]; next[i] = { ...d, proxy_port: e.target.value }; setDrafts(next) }} placeholder="Port" />
              <Button variant="ghost" size="icon-sm" onClick={() => remove(i)}><Trash2 className="size-4 text-destructive" /></Button>
            </div>
          ))}
          {drafts.length === 0 && <p className="text-muted-foreground text-sm">No domains</p>}
        </CardContent>
      </Card>
    )
  }

  if (domains.length === 0) return null

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base">Domains</CardTitle>
        <CardAction>
          <Button variant="ghost" size="icon-sm" onClick={startEdit}><Pencil className="size-4" /></Button>
        </CardAction>
      </CardHeader>
      <CardContent>
        <div className="space-y-2">
          {domains.map((d) => (
            <div key={d.domain} className="flex items-center gap-2">
              <a
                href={`http://${d.domain}`}
                target="_blank"
                rel="noopener noreferrer"
                className="text-primary hover:underline flex items-center gap-1"
              >
                {d.domain}
                <ExternalLink className="size-3" />
              </a>
              <CopyButton value={d.domain} />
              {d.proxy_port ? <span className="text-xs text-muted-foreground">:{d.proxy_port}</span> : null}
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  )
}
