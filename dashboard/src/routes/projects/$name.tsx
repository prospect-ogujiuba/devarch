import { createFileRoute, Link } from '@tanstack/react-router'
import { z } from 'zod'
import { ArrowLeft, Loader2, GitBranch, Package, Code2, ExternalLink, Puzzle, Palette, Server, Unlink, Layers } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Tabs, TabsContent } from '@/components/ui/tabs'
import { ResponsiveTabsList } from '@/components/ui/responsive-tabs-list'
import { LifecycleButtons } from '@/components/ui/entity-actions'
import { ProjectLogo } from '@/components/projects/project-logo'
import { Button } from '@/components/ui/button'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { useProject, useProjectServices, useProjectStatus, useProjectControl, useProjectServiceControl, useLinkStack } from '@/features/projects/queries'
import { useStacks } from '@/features/stacks/queries'
import { useState } from 'react'

export const Route = createFileRoute('/projects/$name')({
  validateSearch: z.object({
    tab: z.string().optional(),
  }),
  component: ProjectDetailPage,
})

function ProjectDetailPage() {
  const { name } = Route.useParams()
  const routeSearch = Route.useSearch()
  const navigate = Route.useNavigate()
  const { data: project, isLoading } = useProject(name)
  const { data: services } = useProjectServices(name)
  const hasStack = !!project?.stack_id
  const { data: statuses } = useProjectStatus(name, hasStack || !!project?.compose_path)
  const { start, stop, restart } = useProjectControl(name)
  const svcControl = useProjectServiceControl(name)

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <Loader2 className="size-8 animate-spin text-muted-foreground" />
      </div>
    )
  }

  if (!project) {
    return (
      <div className="space-y-4">
        <Link to="/projects" className="inline-flex items-center gap-2 text-muted-foreground hover:text-foreground">
          <ArrowLeft className="size-4" />
          Back to Projects
        </Link>
        <div className="text-center py-12">
          <p className="text-muted-foreground">Project not found</p>
        </div>
      </div>
    )
  }

  const deps = project.dependencies ?? {}
  const scripts = project.scripts ?? {}
  const scriptEntries = Object.entries(scripts)
  const isWordPress = project.project_type === 'wordpress'
  const plugins = Array.isArray(deps.plugins) ? (deps.plugins as string[]) : []
  const themes = Array.isArray(deps.themes) ? (deps.themes as string[]) : []
  const phpDeps = (typeof deps.php === 'object' && deps.php !== null && !Array.isArray(deps.php))
    ? Object.entries(deps.php as Record<string, string>)
    : []

  const flatDeps = isWordPress
    ? phpDeps
    : Object.entries(deps).filter(([, v]) => typeof v === 'string') as [string, string][]

  const statusMap = new Map(statuses?.map(s => [s.name, s]) ?? [])
  const hasCompose = !!project.compose_path
  const hasServices = hasStack || hasCompose
  const tabItems = [
    ...(hasServices ? [{ value: 'services', label: `Services (${services?.length ?? 0})` }] : []),
    { value: 'info', label: 'Info' },
    { value: 'deps', label: `Dependencies (${flatDeps.length})` },
    ...(isWordPress && plugins.length > 0 ? [{ value: 'plugins', label: `Plugins (${plugins.length})` }] : []),
    ...(isWordPress && themes.length > 0 ? [{ value: 'themes', label: `Themes (${themes.length})` }] : []),
    { value: 'scripts', label: `Scripts (${scriptEntries.length})` },
    { value: 'git', label: 'Git' },
  ]
  const tabValues = tabItems.map((t) => t.value)
  const defaultTab = hasServices ? 'services' : 'info'
  const activeTab = routeSearch.tab && tabValues.includes(routeSearch.tab) ? routeSearch.tab : defaultTab

  return (
    <div className="space-y-6">
      <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:gap-4">
        <Link to="/projects" className="text-muted-foreground hover:text-foreground">
          <ArrowLeft className="size-5" />
        </Link>
        <ProjectLogo
          projectType={project.project_type}
          framework={project.framework}
          language={project.language}
          className="size-8"
        />
        <div className="flex-1">
          <div className="flex items-center gap-3">
            <h1 className="text-2xl font-bold">{project.name}</h1>
            <Badge variant="outline" className="capitalize">{project.project_type}</Badge>
            {project.version && (
              <Badge variant="secondary" className="font-mono text-xs">v{project.version}</Badge>
            )}
            {project.service_count > 0 && (
              <Badge variant="secondary">{project.service_count} services</Badge>
            )}
          </div>
          {project.description && (
            <p className="text-muted-foreground">{project.description}</p>
          )}
        </div>
        <div className="grid w-full grid-cols-2 gap-2 sm:flex sm:w-auto sm:items-center">
          {hasServices && (
            <LifecycleButtons
              isRunning={false}
              onStart={() => start.mutate()}
              onStop={() => stop.mutate()}
              onRestart={() => restart.mutate()}
              isStartPending={start.isPending}
              isStopPending={stop.isPending}
              isRestartPending={restart.isPending}
              showAll
              className="col-span-2"
              buttonClassName="w-full sm:w-auto"
            />
          )}
          {project.domain && (
            <a
              href={`https://${project.domain}`}
              target="_blank"
              rel="noopener noreferrer"
              className="col-span-2 inline-flex items-center justify-center gap-1 rounded-md border px-3 py-2 text-sm text-primary hover:underline sm:col-span-1 sm:justify-start sm:border-0 sm:px-0 sm:py-0"
            >
              <ExternalLink className="size-4" />
              {project.domain}
            </a>
          )}
        </div>
      </div>

      <div className="grid gap-4 md:grid-cols-4">
        {project.framework && (
          <Card className="py-4">
            <CardHeader className="pb-2">
              <CardTitle className="text-sm font-medium text-muted-foreground flex items-center gap-2">
                <Code2 className="size-4" />
                Framework
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-lg font-semibold">{project.framework}</div>
            </CardContent>
          </Card>
        )}
        {project.language && (
          <Card className="py-4">
            <CardHeader className="pb-2">
              <CardTitle className="text-sm font-medium text-muted-foreground flex items-center gap-2">
                <Code2 className="size-4" />
                Language
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-lg font-semibold capitalize">{project.language}</div>
            </CardContent>
          </Card>
        )}
        {project.package_manager && (
          <Card className="py-4">
            <CardHeader className="pb-2">
              <CardTitle className="text-sm font-medium text-muted-foreground flex items-center gap-2">
                <Package className="size-4" />
                Package Manager
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-lg font-semibold">{project.package_manager}</div>
            </CardContent>
          </Card>
        )}
        {project.git_branch && (
          <Card className="py-4">
            <CardHeader className="pb-2">
              <CardTitle className="text-sm font-medium text-muted-foreground flex items-center gap-2">
                <GitBranch className="size-4" />
                Branch
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-lg font-semibold">{project.git_branch}</div>
            </CardContent>
          </Card>
        )}
      </div>

      <Tabs
        value={activeTab}
        onValueChange={(tab) => navigate({ search: (prev) => ({ ...prev, tab }) })}
        className="space-y-4"
      >
        <ResponsiveTabsList
          tabs={tabItems}
          value={activeTab}
          onValueChange={(tab) => navigate({ search: (prev) => ({ ...prev, tab }) })}
        />

        {hasServices && (
          <TabsContent value="services">
            {hasStack ? (
              <StackServicesTab services={services} statusMap={statusMap} stackName={project.stack_name} svcControl={svcControl} />
            ) : (
              <LegacyServicesTab services={services} statusMap={statusMap} svcControl={svcControl} />
            )}
          </TabsContent>
        )}

        <TabsContent value="info" className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle className="text-base">Details</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="grid gap-2 text-sm">
                <Row label="Path" value={project.path} mono />
                <Row label="Type" value={project.project_type} />
                {project.framework && <Row label="Framework" value={project.framework} />}
                {project.language && <Row label="Language" value={project.language} />}
                {project.version && <Row label="Version" value={project.version} />}
                {project.license && <Row label="License" value={project.license} />}
                {project.entry_point && <Row label="Entry Point" value={project.entry_point} mono />}
                <Row label="Has Frontend" value={project.has_frontend ? 'Yes' : 'No'} />
                {project.frontend_framework && <Row label="Frontend" value={project.frontend_framework} />}
                {project.domain && <Row label="Domain" value={project.domain} />}
                {project.proxy_port && <Row label="Proxy Port" value={String(project.proxy_port)} />}
                {project.compose_path && <Row label="Compose" value={project.compose_path} mono />}
                {project.service_count > 0 && <Row label="Services" value={String(project.service_count)} />}
                {isWordPress && Boolean(deps.db_name) && <Row label="Database" value={String(deps.db_name)} />}
                {isWordPress && Boolean(deps.table_prefix) && <Row label="Table Prefix" value={String(deps.table_prefix)} mono />}
                {project.last_scanned_at && (
                  <Row label="Last Scanned" value={new Date(project.last_scanned_at).toLocaleString()} />
                )}
              </div>
            </CardContent>
          </Card>

          <StackLinkCard projectName={name} stackId={project.stack_id} stackName={project.stack_name} />
        </TabsContent>

        <TabsContent value="deps">
          <Card>
            <CardHeader>
              <CardTitle className="text-base">Dependencies</CardTitle>
            </CardHeader>
            <CardContent>
              {flatDeps.length > 0 ? (
                <div className="grid gap-1">
                  {flatDeps.map(([depName, version]) => (
                    <div key={depName} className="flex items-center justify-between text-sm py-1 border-b border-border/50 last:border-0">
                      <span className="font-mono">{depName}</span>
                      <Badge variant="secondary" className="font-mono text-xs">{String(version)}</Badge>
                    </div>
                  ))}
                </div>
              ) : (
                <p className="text-muted-foreground">No dependencies found</p>
              )}
            </CardContent>
          </Card>
        </TabsContent>

        {isWordPress && plugins.length > 0 && (
          <TabsContent value="plugins">
            <Card>
              <CardHeader>
                <CardTitle className="text-base flex items-center gap-2">
                  <Puzzle className="size-4" />
                  Plugins ({plugins.length})
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="grid gap-1">
                  {plugins.map((plugin) => (
                    <div key={plugin} className="flex items-center text-sm py-1.5 border-b border-border/50 last:border-0">
                      <span className="font-mono">{plugin}</span>
                    </div>
                  ))}
                </div>
              </CardContent>
            </Card>
          </TabsContent>
        )}

        {isWordPress && themes.length > 0 && (
          <TabsContent value="themes">
            <Card>
              <CardHeader>
                <CardTitle className="text-base flex items-center gap-2">
                  <Palette className="size-4" />
                  Themes ({themes.length})
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="grid gap-1">
                  {themes.map((theme) => (
                    <div key={theme} className="flex items-center text-sm py-1.5 border-b border-border/50 last:border-0">
                      <span className="font-mono">{theme}</span>
                    </div>
                  ))}
                </div>
              </CardContent>
            </Card>
          </TabsContent>
        )}

        <TabsContent value="scripts">
          <Card>
            <CardHeader>
              <CardTitle className="text-base">Scripts</CardTitle>
            </CardHeader>
            <CardContent>
              {scriptEntries.length > 0 ? (
                <div className="grid gap-2">
                  {scriptEntries.map(([scriptName, cmd]) => (
                    <div key={scriptName} className="text-sm py-1 border-b border-border/50 last:border-0">
                      <span className="font-semibold">{scriptName}</span>
                      <pre className="text-xs text-muted-foreground mt-0.5 whitespace-pre-wrap">{cmd}</pre>
                    </div>
                  ))}
                </div>
              ) : (
                <p className="text-muted-foreground">No scripts found</p>
              )}
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="git">
          <Card>
            <CardHeader>
              <CardTitle className="text-base">Git Info</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="grid gap-2 text-sm">
                {project.git_remote ? (
                  <>
                    <Row label="Remote" value={project.git_remote} mono />
                    <Row label="Branch" value={project.git_branch ?? '-'} />
                  </>
                ) : (
                  <p className="text-muted-foreground">No git repository detected</p>
                )}
              </div>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  )
}

type SvcControl = ReturnType<typeof useProjectServiceControl>

function StackServicesTab({ services, statusMap, stackName, svcControl }: {
  services: unknown[] | undefined
  statusMap: Map<string, { status: string }>
  stackName?: string
  svcControl: SvcControl
}) {
  const svcs = services as { id: number; instance_id: string; container_name: string; enabled: boolean; template_name: string; image: string }[] | undefined

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base flex items-center gap-2">
          <Layers className="size-4" />
          Stack Instances
          {stackName && (
            <Link to="/stacks/$name" params={{ name: stackName }} className="text-sm font-normal text-primary hover:underline ml-2">
              {stackName}
            </Link>
          )}
        </CardTitle>
      </CardHeader>
      <CardContent>
        {svcs && svcs.length > 0 ? (
          <div className="grid gap-2">
            {svcs.map((svc) => {
              const st = statusMap.get(svc.container_name)
              const isRunning = st?.status === 'running'
              return (
                <div key={svc.id} className="flex items-center justify-between py-2 border-b border-border/50 last:border-0">
                  <div className="flex items-center gap-3">
                    <span className="font-mono font-medium">{svc.instance_id}</span>
                    <Badge variant="outline" className="text-xs">{svc.template_name}</Badge>
                    <span className="text-xs text-muted-foreground font-mono">{svc.image}</span>
                    {!svc.enabled && <Badge variant="secondary" className="text-xs">disabled</Badge>}
                  </div>
                  <div className="flex items-center gap-2">
                    {st && <ServiceStatusBadge status={st.status} />}
                    {svc.enabled && (
                      <LifecycleButtons
                        isRunning={isRunning}
                        onStart={() => svcControl.startService.mutate(svc.instance_id)}
                        onStop={() => svcControl.stopService.mutate(svc.instance_id)}
                        onRestart={() => svcControl.restartService.mutate(svc.instance_id)}
                        showRestart={isRunning}
                        size="icon-sm"
                      />
                    )}
                  </div>
                </div>
              )
            })}
          </div>
        ) : (
          <p className="text-muted-foreground">No instances in this stack</p>
        )}
      </CardContent>
    </Card>
  )
}

function LegacyServicesTab({ services, statusMap, svcControl }: {
  services: { id: number; service_name: string; service_type?: string; image?: string; ports?: string[] }[] | undefined
  statusMap: Map<string, { status: string }>
  svcControl: SvcControl
}) {
  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base flex items-center gap-2">
          <Server className="size-4" />
          Compose Services
        </CardTitle>
      </CardHeader>
      <CardContent>
        {services && services.length > 0 ? (
          <div className="grid gap-2">
            {services.map((svc) => {
              const st = statusMap.get(svc.service_name)
              const isRunning = st?.status === 'running'
              return (
                <div key={svc.id} className="flex items-center justify-between py-2 border-b border-border/50 last:border-0">
                  <div className="flex items-center gap-3">
                    <span className="font-mono font-medium">{svc.service_name}</span>
                    {svc.service_type && (
                      <Badge variant="outline" className="text-xs">{svc.service_type}</Badge>
                    )}
                    {svc.image && (
                      <span className="text-xs text-muted-foreground font-mono">{svc.image}</span>
                    )}
                  </div>
                  <div className="flex items-center gap-2">
                    {svc.ports && svc.ports.length > 0 && (
                      <span className="text-xs text-muted-foreground">{svc.ports.join(', ')}</span>
                    )}
                    {st && <ServiceStatusBadge status={st.status} />}
                    <LifecycleButtons
                      isRunning={isRunning}
                      onStart={() => svcControl.startService.mutate(svc.service_name)}
                      onStop={() => svcControl.stopService.mutate(svc.service_name)}
                      onRestart={() => svcControl.restartService.mutate(svc.service_name)}
                      showRestart={isRunning}
                      size="icon-sm"
                    />
                  </div>
                </div>
              )
            })}
          </div>
        ) : (
          <p className="text-muted-foreground">No services found</p>
        )}
      </CardContent>
    </Card>
  )
}

function StackLinkCard({ projectName, stackId, stackName }: {
  projectName: string
  stackId?: number
  stackName?: string
}) {
  const { data: stacks } = useStacks()
  const linkStack = useLinkStack(projectName)
  const [selectedStackId, setSelectedStackId] = useState<string>('')

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base flex items-center gap-2">
          <Layers className="size-4" />
          Linked Stack
        </CardTitle>
      </CardHeader>
      <CardContent>
        {stackId && stackName ? (
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <Link to="/stacks/$name" params={{ name: stackName }} className="font-mono text-primary hover:underline">
                {stackName}
              </Link>
              <Badge variant="secondary">ID: {stackId}</Badge>
            </div>
            <Button
              variant="outline"
              size="sm"
              onClick={() => linkStack.mutate(null)}
              disabled={linkStack.isPending}
            >
              <Unlink className="size-4 mr-1" />
              Unlink
            </Button>
          </div>
        ) : (
          <div className="flex items-center gap-3">
            <Select value={selectedStackId} onValueChange={setSelectedStackId}>
              <SelectTrigger className="w-64">
                <SelectValue placeholder="Select a stack..." />
              </SelectTrigger>
              <SelectContent>
                {stacks?.map((stack) => (
                  <SelectItem key={stack.id} value={String(stack.id)}>
                    {stack.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            <Button
              size="sm"
              onClick={() => {
                if (selectedStackId) {
                  linkStack.mutate(Number(selectedStackId))
                  setSelectedStackId('')
                }
              }}
              disabled={!selectedStackId || linkStack.isPending}
            >
              Link Stack
            </Button>
          </div>
        )}
      </CardContent>
    </Card>
  )
}

function ServiceStatusBadge({ status }: { status: string }) {
  const variant = status === 'running' ? 'default'
    : status === 'exited' || status === 'stopped' ? 'secondary'
    : 'outline'

  const colors = status === 'running' ? 'bg-green-500/10 text-green-500 border-green-500/20'
    : status === 'exited' || status === 'stopped' ? 'bg-red-500/10 text-red-500 border-red-500/20'
    : ''

  return <Badge variant={variant} className={`text-xs ${colors}`}>{status}</Badge>
}

function Row({ label, value, mono }: { label: string; value: string; mono?: boolean }) {
  return (
    <div className="flex">
      <span className="text-muted-foreground w-40 shrink-0">{label}:</span>
      <span className={mono ? 'font-mono text-xs' : ''}>{value}</span>
    </div>
  )
}
