import { useState } from 'react'
import { createFileRoute, Link, useNavigate } from '@tanstack/react-router'
import { ArrowLeft, Plus, Trash2, Loader2 } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Checkbox } from '@/components/ui/checkbox'
import { useCreateService } from '@/features/services/queries'
import { useCategories } from '@/features/categories/queries'
import { useServices } from '@/features/services/queries'

export const Route = createFileRoute('/services/new')({
  component: NewServicePage,
})

interface PortDraft {
  host_ip: string
  host_port: string
  container_port: string
  protocol: string
}

interface VolumeDraft {
  volume_type: string
  source: string
  target: string
  read_only: boolean
  is_external: boolean
}

interface EnvDraft {
  key: string
  value: string
  is_secret: boolean
}

interface LabelDraft {
  key: string
  value: string
}

interface DomainDraft {
  domain: string
  proxy_port: string
}

function NewServicePage() {
  const navigate = useNavigate()
  const createService = useCreateService()
  const { data: categories } = useCategories()
  const { data: servicesData } = useServices()
  const allServices = servicesData?.services ?? []

  const [form, setForm] = useState({
    name: '',
    category_id: '',
    image_name: '',
    image_tag: 'latest',
    restart_policy: 'unless-stopped',
    command: '',
    user_spec: '',
  })

  const [ports, setPorts] = useState<PortDraft[]>([])
  const [volumes, setVolumes] = useState<VolumeDraft[]>([])
  const [envVars, setEnvVars] = useState<EnvDraft[]>([])
  const [dependencies, setDependencies] = useState<string[]>([])
  const [labels, setLabels] = useState<LabelDraft[]>([])
  const [domains, setDomains] = useState<DomainDraft[]>([])
  const [healthcheck, setHealthcheck] = useState({
    enabled: false,
    test: '',
    interval_seconds: 30,
    timeout_seconds: 10,
    retries: 3,
    start_period_seconds: 0,
  })

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    createService.mutate(
      {
        name: form.name,
        category_id: parseInt(form.category_id, 10),
        image_name: form.image_name,
        image_tag: form.image_tag,
        restart_policy: form.restart_policy,
        command: form.command || undefined,
        user_spec: form.user_spec || undefined,
        ports: ports.map((p) => ({
          host_ip: p.host_ip,
          host_port: parseInt(p.host_port, 10) || 0,
          container_port: parseInt(p.container_port, 10) || 0,
          protocol: p.protocol || 'tcp',
        })),
        volumes: volumes.map((v) => ({
          volume_type: v.volume_type,
          source: v.source,
          target: v.target,
          read_only: v.read_only,
          is_external: v.is_external,
        })),
        env_vars: envVars.map((e) => ({
          key: e.key,
          value: e.value,
          is_secret: e.is_secret,
        })),
        dependencies,
      },
      {
        onSuccess: () => navigate({ to: '/services/$name', params: { name: form.name } }),
      },
    )
  }

  const addPort = () => setPorts([...ports, { host_ip: '0.0.0.0', host_port: '', container_port: '', protocol: 'tcp' }])
  const removePort = (i: number) => setPorts(ports.filter((_, idx) => idx !== i))

  const addVolume = () => setVolumes([...volumes, { volume_type: 'bind', source: '', target: '', read_only: false, is_external: false }])
  const removeVolume = (i: number) => setVolumes(volumes.filter((_, idx) => idx !== i))

  const addEnv = () => setEnvVars([...envVars, { key: '', value: '', is_secret: false }])
  const removeEnv = (i: number) => setEnvVars(envVars.filter((_, idx) => idx !== i))

  const addLabel = () => setLabels([...labels, { key: '', value: '' }])
  const removeLabel = (i: number) => setLabels(labels.filter((_, idx) => idx !== i))

  const addDomain = () => setDomains([...domains, { domain: '', proxy_port: '' }])
  const removeDomain = (i: number) => setDomains(domains.filter((_, idx) => idx !== i))

  const toggleDep = (name: string) => {
    setDependencies((prev) =>
      prev.includes(name) ? prev.filter((d) => d !== name) : [...prev, name],
    )
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-6">
      <div className="flex items-center gap-4">
        <Link to="/services" className="text-muted-foreground hover:text-foreground">
          <ArrowLeft className="size-5" />
        </Link>
        <h1 className="text-2xl font-bold flex-1">New Service</h1>
        <Button type="submit" disabled={createService.isPending}>
          {createService.isPending ? <Loader2 className="size-4 animate-spin" /> : 'Create Service'}
        </Button>
      </div>

      <Card>
        <CardHeader>
          <CardTitle className="text-base">Metadata</CardTitle>
        </CardHeader>
        <CardContent className="grid gap-4 sm:grid-cols-2">
          <div className="grid gap-2">
            <label className="text-sm font-medium">Name</label>
            <Input required value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} placeholder="my-service" />
          </div>
          <div className="grid gap-2">
            <label className="text-sm font-medium">Category</label>
            <Select required value={form.category_id} onValueChange={(v) => setForm({ ...form, category_id: v })}>
              <SelectTrigger className="w-full">
                <SelectValue placeholder="Select category" />
              </SelectTrigger>
              <SelectContent>
                {(categories ?? []).map((c) => (
                  <SelectItem key={c.id} value={String(c.id)}>{c.display_name || c.name}</SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
          <div className="grid gap-2">
            <label className="text-sm font-medium">Image Name</label>
            <Input required value={form.image_name} onChange={(e) => setForm({ ...form, image_name: e.target.value })} placeholder="postgres" />
          </div>
          <div className="grid gap-2">
            <label className="text-sm font-medium">Image Tag</label>
            <Input value={form.image_tag} onChange={(e) => setForm({ ...form, image_tag: e.target.value })} placeholder="latest" />
          </div>
          <div className="grid gap-2">
            <label className="text-sm font-medium">Restart Policy</label>
            <Select value={form.restart_policy} onValueChange={(v) => setForm({ ...form, restart_policy: v })}>
              <SelectTrigger className="w-full">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="no">no</SelectItem>
                <SelectItem value="always">always</SelectItem>
                <SelectItem value="unless-stopped">unless-stopped</SelectItem>
                <SelectItem value="on-failure">on-failure</SelectItem>
              </SelectContent>
            </Select>
          </div>
          <div className="grid gap-2">
            <label className="text-sm font-medium">Command</label>
            <Input value={form.command} onChange={(e) => setForm({ ...form, command: e.target.value })} placeholder="Optional" />
          </div>
          <div className="grid gap-2">
            <label className="text-sm font-medium">User</label>
            <Input value={form.user_spec} onChange={(e) => setForm({ ...form, user_spec: e.target.value })} placeholder="Optional" />
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader className="flex flex-row items-center justify-between">
          <CardTitle className="text-base">Ports</CardTitle>
          <Button type="button" variant="outline" size="sm" onClick={addPort}><Plus className="size-4" /> Add</Button>
        </CardHeader>
        <CardContent>
          {ports.length === 0 ? (
            <p className="text-muted-foreground text-sm">No ports configured</p>
          ) : (
            <div className="space-y-2">
              {ports.map((port, i) => (
                <div key={i} className="flex gap-2 items-center">
                  <Input className="w-28" value={port.host_ip} onChange={(e) => { const next = [...ports]; next[i] = { ...port, host_ip: e.target.value }; setPorts(next) }} placeholder="Host IP" />
                  <Input className="w-24" type="number" value={port.host_port} onChange={(e) => { const next = [...ports]; next[i] = { ...port, host_port: e.target.value }; setPorts(next) }} placeholder="Host" />
                  <span className="text-muted-foreground">:</span>
                  <Input className="w-24" type="number" value={port.container_port} onChange={(e) => { const next = [...ports]; next[i] = { ...port, container_port: e.target.value }; setPorts(next) }} placeholder="Container" />
                  <Select value={port.protocol} onValueChange={(v) => { const next = [...ports]; next[i] = { ...port, protocol: v }; setPorts(next) }}>
                    <SelectTrigger className="w-20"><SelectValue /></SelectTrigger>
                    <SelectContent>
                      <SelectItem value="tcp">tcp</SelectItem>
                      <SelectItem value="udp">udp</SelectItem>
                    </SelectContent>
                  </Select>
                  <Button type="button" variant="ghost" size="icon-sm" onClick={() => removePort(i)}><Trash2 className="size-4 text-destructive" /></Button>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      <Card>
        <CardHeader className="flex flex-row items-center justify-between">
          <CardTitle className="text-base">Volumes</CardTitle>
          <Button type="button" variant="outline" size="sm" onClick={addVolume}><Plus className="size-4" /> Add</Button>
        </CardHeader>
        <CardContent>
          {volumes.length === 0 ? (
            <p className="text-muted-foreground text-sm">No volumes configured</p>
          ) : (
            <div className="space-y-2">
              {volumes.map((vol, i) => (
                <div key={i} className="flex gap-2 items-center">
                  <Select value={vol.volume_type} onValueChange={(v) => { const next = [...volumes]; next[i] = { ...vol, volume_type: v }; setVolumes(next) }}>
                    <SelectTrigger className="w-24"><SelectValue /></SelectTrigger>
                    <SelectContent>
                      <SelectItem value="bind">bind</SelectItem>
                      <SelectItem value="volume">volume</SelectItem>
                      <SelectItem value="tmpfs">tmpfs</SelectItem>
                    </SelectContent>
                  </Select>
                  <Input className="flex-1" value={vol.source} onChange={(e) => { const next = [...volumes]; next[i] = { ...vol, source: e.target.value }; setVolumes(next) }} placeholder="Source" />
                  <span className="text-muted-foreground">:</span>
                  <Input className="flex-1" value={vol.target} onChange={(e) => { const next = [...volumes]; next[i] = { ...vol, target: e.target.value }; setVolumes(next) }} placeholder="Target" />
                  <div className="flex items-center gap-1">
                    <Checkbox checked={vol.read_only} onCheckedChange={(v) => { const next = [...volumes]; next[i] = { ...vol, read_only: !!v }; setVolumes(next) }} />
                    <span className="text-xs">RO</span>
                  </div>
                  <Button type="button" variant="ghost" size="icon-sm" onClick={() => removeVolume(i)}><Trash2 className="size-4 text-destructive" /></Button>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      <Card>
        <CardHeader className="flex flex-row items-center justify-between">
          <CardTitle className="text-base">Environment Variables</CardTitle>
          <Button type="button" variant="outline" size="sm" onClick={addEnv}><Plus className="size-4" /> Add</Button>
        </CardHeader>
        <CardContent>
          {envVars.length === 0 ? (
            <p className="text-muted-foreground text-sm">No environment variables</p>
          ) : (
            <div className="space-y-2">
              {envVars.map((env, i) => (
                <div key={i} className="flex gap-2 items-center">
                  <Input className="w-48" value={env.key} onChange={(e) => { const next = [...envVars]; next[i] = { ...env, key: e.target.value }; setEnvVars(next) }} placeholder="KEY" />
                  <span className="text-muted-foreground">=</span>
                  <Input className="flex-1" value={env.value} type={env.is_secret ? 'password' : 'text'} onChange={(e) => { const next = [...envVars]; next[i] = { ...env, value: e.target.value }; setEnvVars(next) }} placeholder="value" />
                  <div className="flex items-center gap-1">
                    <Checkbox checked={env.is_secret} onCheckedChange={(v) => { const next = [...envVars]; next[i] = { ...env, is_secret: !!v }; setEnvVars(next) }} />
                    <span className="text-xs">Secret</span>
                  </div>
                  <Button type="button" variant="ghost" size="icon-sm" onClick={() => removeEnv(i)}><Trash2 className="size-4 text-destructive" /></Button>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle className="text-base">Dependencies</CardTitle>
        </CardHeader>
        <CardContent>
          {allServices.length === 0 ? (
            <p className="text-muted-foreground text-sm">No services available</p>
          ) : (
            <div className="flex flex-wrap gap-2">
              {allServices.map((s) => (
                <Button
                  key={s.name}
                  type="button"
                  variant={dependencies.includes(s.name) ? 'default' : 'outline'}
                  size="sm"
                  onClick={() => toggleDep(s.name)}
                >
                  {s.name}
                </Button>
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      <Card>
        <CardHeader className="flex flex-row items-center justify-between">
          <CardTitle className="text-base">Healthcheck</CardTitle>
          <div className="flex items-center gap-2">
            <Checkbox checked={healthcheck.enabled} onCheckedChange={(v) => setHealthcheck({ ...healthcheck, enabled: !!v })} />
            <span className="text-sm">Enable</span>
          </div>
        </CardHeader>
        {healthcheck.enabled && (
          <CardContent className="grid gap-4 sm:grid-cols-2">
            <div className="grid gap-2 sm:col-span-2">
              <label className="text-sm font-medium">Test Command</label>
              <Input value={healthcheck.test} onChange={(e) => setHealthcheck({ ...healthcheck, test: e.target.value })} placeholder="CMD-SHELL curl -f http://localhost/ || exit 1" />
            </div>
            <div className="grid gap-2">
              <label className="text-sm font-medium">Interval (s)</label>
              <Input type="number" value={healthcheck.interval_seconds} onChange={(e) => setHealthcheck({ ...healthcheck, interval_seconds: parseInt(e.target.value, 10) || 0 })} />
            </div>
            <div className="grid gap-2">
              <label className="text-sm font-medium">Timeout (s)</label>
              <Input type="number" value={healthcheck.timeout_seconds} onChange={(e) => setHealthcheck({ ...healthcheck, timeout_seconds: parseInt(e.target.value, 10) || 0 })} />
            </div>
            <div className="grid gap-2">
              <label className="text-sm font-medium">Retries</label>
              <Input type="number" value={healthcheck.retries} onChange={(e) => setHealthcheck({ ...healthcheck, retries: parseInt(e.target.value, 10) || 0 })} />
            </div>
            <div className="grid gap-2">
              <label className="text-sm font-medium">Start Period (s)</label>
              <Input type="number" value={healthcheck.start_period_seconds} onChange={(e) => setHealthcheck({ ...healthcheck, start_period_seconds: parseInt(e.target.value, 10) || 0 })} />
            </div>
          </CardContent>
        )}
      </Card>

      <Card>
        <CardHeader className="flex flex-row items-center justify-between">
          <CardTitle className="text-base">Labels</CardTitle>
          <Button type="button" variant="outline" size="sm" onClick={addLabel}><Plus className="size-4" /> Add</Button>
        </CardHeader>
        <CardContent>
          {labels.length === 0 ? (
            <p className="text-muted-foreground text-sm">No labels</p>
          ) : (
            <div className="space-y-2">
              {labels.map((label, i) => (
                <div key={i} className="flex gap-2 items-center">
                  <Input className="flex-1" value={label.key} onChange={(e) => { const next = [...labels]; next[i] = { ...label, key: e.target.value }; setLabels(next) }} placeholder="key" />
                  <span className="text-muted-foreground">=</span>
                  <Input className="flex-1" value={label.value} onChange={(e) => { const next = [...labels]; next[i] = { ...label, value: e.target.value }; setLabels(next) }} placeholder="value" />
                  <Button type="button" variant="ghost" size="icon-sm" onClick={() => removeLabel(i)}><Trash2 className="size-4 text-destructive" /></Button>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      <Card>
        <CardHeader className="flex flex-row items-center justify-between">
          <CardTitle className="text-base">Domains</CardTitle>
          <Button type="button" variant="outline" size="sm" onClick={addDomain}><Plus className="size-4" /> Add</Button>
        </CardHeader>
        <CardContent>
          {domains.length === 0 ? (
            <p className="text-muted-foreground text-sm">No domains</p>
          ) : (
            <div className="space-y-2">
              {domains.map((d, i) => (
                <div key={i} className="flex gap-2 items-center">
                  <Input className="flex-1" value={d.domain} onChange={(e) => { const next = [...domains]; next[i] = { ...d, domain: e.target.value }; setDomains(next) }} placeholder="example.test" />
                  <Input className="w-24" type="number" value={d.proxy_port} onChange={(e) => { const next = [...domains]; next[i] = { ...d, proxy_port: e.target.value }; setDomains(next) }} placeholder="Port" />
                  <Button type="button" variant="ghost" size="icon-sm" onClick={() => removeDomain(i)}><Trash2 className="size-4 text-destructive" /></Button>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>
    </form>
  )
}
