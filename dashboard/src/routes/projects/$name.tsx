import { createFileRoute, Link } from '@tanstack/react-router'
import { z } from 'zod'
import { ArrowLeft, Loader2, GitBranch, Package, Code2, ExternalLink, Puzzle, Palette, Layers, FileCode, Diff, MoreVertical, Pencil, Trash2, Play } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Tabs, TabsContent } from '@/components/ui/tabs'
import { ResponsiveTabsList } from '@/components/ui/responsive-tabs-list'
import { LifecycleButtons } from '@/components/ui/entity-actions'
import { ProjectLogo } from '@/components/projects/project-logo'
import { Button } from '@/components/ui/button'
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger, DropdownMenuSeparator } from '@/components/ui/dropdown-menu'
import { useProject, useProjectStatus, useProjectControl } from '@/features/projects/queries'
import { useInstances } from '@/features/instances/queries'
import { useStackCompose, useGeneratePlan, useApplyPlan } from '@/features/stacks/queries'
import { useGenerateProjectProxyConfig } from '@/features/proxy/queries'
import { ProxyConfigPanel } from '@/components/proxy/proxy-config-panel'
import { EditProjectDialog } from '@/components/projects/edit-project-dialog'
import { DeleteProjectDialog } from '@/components/projects/delete-project-dialog'
import { useState } from 'react'
import type { StackPlan } from '@/types/api'

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
  const { data: statuses } = useProjectStatus(name, !!project)
  const { start, stop, restart } = useProjectControl(name)
  const generateProxyConfig = useGenerateProjectProxyConfig(name)

  const [editOpen, setEditOpen] = useState(false)
  const [deleteOpen, setDeleteOpen] = useState(false)

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

  const anyRunning = project.running_count > 0

  const tabItems = [
    { value: 'instances', label: `Instances (${project.instance_count})` },
    { value: 'compose', label: 'Compose' },
    { value: 'plan', label: 'Plan' },
    { value: 'info', label: 'Info' },
    { value: 'deps', label: `Dependencies (${flatDeps.length})` },
    ...(isWordPress && plugins.length > 0 ? [{ value: 'plugins', label: `Plugins (${plugins.length})` }] : []),
    ...(isWordPress && themes.length > 0 ? [{ value: 'themes', label: `Themes (${themes.length})` }] : []),
    { value: 'scripts', label: `Scripts (${scriptEntries.length})` },
    { value: 'git', label: 'Git' },
    ...(project.domain ? [{ value: 'proxy', label: 'Proxy' }] : []),
  ]
  const tabValues = tabItems.map((t) => t.value)
  const activeTab = routeSearch.tab && tabValues.includes(routeSearch.tab) ? routeSearch.tab : 'instances'

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
            <RunningBadge running={project.running_count} total={project.instance_count} />
          </div>
          {project.description && (
            <p className="text-muted-foreground">{project.description}</p>
          )}
        </div>
        <div className="grid w-full grid-cols-2 gap-2 sm:flex sm:w-auto sm:items-center">
          <LifecycleButtons
            isRunning={anyRunning}
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
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="outline" size="icon"><MoreVertical className="size-4" /></Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuItem onClick={() => setEditOpen(true)}>
                <Pencil className="size-4 mr-2" />Edit
              </DropdownMenuItem>
              {project.domain && (
                <DropdownMenuItem asChild>
                  <a href={`https://${project.domain}`} target="_blank" rel="noopener noreferrer">
                    <ExternalLink className="size-4 mr-2" />{project.domain}
                  </a>
                </DropdownMenuItem>
              )}
              <DropdownMenuItem asChild>
                <Link to="/stacks/$name" params={{ name: project.stack_name }}>
                  <Layers className="size-4 mr-2" />View Stack
                </Link>
              </DropdownMenuItem>
              <DropdownMenuSeparator />
              <DropdownMenuItem onClick={() => setDeleteOpen(true)} className="text-destructive">
                <Trash2 className="size-4 mr-2" />Delete
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      </div>

      <div className="grid gap-4 md:grid-cols-4">
        {project.framework && (
          <Card className="py-4">
            <CardHeader className="pb-2">
              <CardTitle className="text-sm font-medium text-muted-foreground flex items-center gap-2">
                <Code2 className="size-4" />Framework
              </CardTitle>
            </CardHeader>
            <CardContent><div className="text-lg font-semibold">{project.framework}</div></CardContent>
          </Card>
        )}
        {project.language && (
          <Card className="py-4">
            <CardHeader className="pb-2">
              <CardTitle className="text-sm font-medium text-muted-foreground flex items-center gap-2">
                <Code2 className="size-4" />Language
              </CardTitle>
            </CardHeader>
            <CardContent><div className="text-lg font-semibold capitalize">{project.language}</div></CardContent>
          </Card>
        )}
        {project.package_manager && (
          <Card className="py-4">
            <CardHeader className="pb-2">
              <CardTitle className="text-sm font-medium text-muted-foreground flex items-center gap-2">
                <Package className="size-4" />Package Manager
              </CardTitle>
            </CardHeader>
            <CardContent><div className="text-lg font-semibold">{project.package_manager}</div></CardContent>
          </Card>
        )}
        {project.git_branch && (
          <Card className="py-4">
            <CardHeader className="pb-2">
              <CardTitle className="text-sm font-medium text-muted-foreground flex items-center gap-2">
                <GitBranch className="size-4" />Branch
              </CardTitle>
            </CardHeader>
            <CardContent><div className="text-lg font-semibold">{project.git_branch}</div></CardContent>
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

        <TabsContent value="instances">
          <InstancesTab stackName={project.stack_name} statuses={statuses} />
        </TabsContent>

        <TabsContent value="compose">
          <ComposeTab stackName={project.stack_name} />
        </TabsContent>

        <TabsContent value="plan">
          <PlanTab stackName={project.stack_name} />
        </TabsContent>

        <TabsContent value="info" className="space-y-4">
          <Card>
            <CardHeader><CardTitle className="text-base">Details</CardTitle></CardHeader>
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
                <Row label="Stack" value={project.stack_name} />
                {isWordPress && Boolean(deps.db_name) && <Row label="Database" value={String(deps.db_name)} />}
                {isWordPress && Boolean(deps.table_prefix) && <Row label="Table Prefix" value={String(deps.table_prefix)} mono />}
                {project.last_scanned_at && (
                  <Row label="Last Scanned" value={new Date(project.last_scanned_at).toLocaleString()} />
                )}
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="deps">
          <Card>
            <CardHeader><CardTitle className="text-base">Dependencies</CardTitle></CardHeader>
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
                  <Puzzle className="size-4" />Plugins ({plugins.length})
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
                  <Palette className="size-4" />Themes ({themes.length})
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
            <CardHeader><CardTitle className="text-base">Scripts</CardTitle></CardHeader>
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
            <CardHeader><CardTitle className="text-base">Git Info</CardTitle></CardHeader>
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

        {project.domain && (
          <TabsContent value="proxy">
            <ProxyConfigPanel scope="project" name={project.name} generateMutation={generateProxyConfig} />
          </TabsContent>
        )}
      </Tabs>

      {project && <EditProjectDialog project={project} open={editOpen} onOpenChange={setEditOpen} />}
      <DeleteProjectDialog projectName={name} open={deleteOpen} onOpenChange={setDeleteOpen} />
    </div>
  )
}

function RunningBadge({ running, total }: { running: number; total: number }) {
  if (total === 0) return null
  const color = running > 0
    ? 'bg-green-500/10 text-green-500 border-green-500/20'
    : 'bg-muted text-muted-foreground border-border'
  return <Badge variant="outline" className={color}>{running}/{total} running</Badge>
}

function InstancesTab({ stackName, statuses }: { stackName: string; statuses?: { name: string; status: string }[] }) {
  const { data: instances, isLoading } = useInstances(stackName)
  const statusMap = new Map(statuses?.map(s => [s.name, s]) ?? [])

  if (isLoading) return <div className="flex justify-center py-8"><Loader2 className="size-6 animate-spin text-muted-foreground" /></div>

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base flex items-center gap-2">
          <Layers className="size-4" />
          Instances
          <Link to="/stacks/$name" params={{ name: stackName }} className="text-sm font-normal text-primary hover:underline ml-2">
            {stackName}
          </Link>
        </CardTitle>
      </CardHeader>
      <CardContent>
        {instances && instances.length > 0 ? (
          <div className="grid gap-2">
            {instances.map((inst) => {
              const st = statusMap.get(inst.container_name ?? '')
              return (
                <div key={inst.id} className="flex items-center justify-between py-2 border-b border-border/50 last:border-0">
                  <div className="flex items-center gap-3">
                    <Link
                      to="/stacks/$name/instances/$instance"
                      params={{ name: stackName, instance: inst.instance_id }}
                      className="font-mono font-medium text-primary hover:underline"
                    >
                      {inst.instance_id}
                    </Link>
                    <Badge variant="outline" className="text-xs">{inst.template_name}</Badge>
                    {!inst.enabled && <Badge variant="secondary" className="text-xs">disabled</Badge>}
                  </div>
                  <div className="flex items-center gap-2">
                    {st && <ServiceStatusBadge status={st.status} />}
                  </div>
                </div>
              )
            })}
          </div>
        ) : (
          <p className="text-muted-foreground">No instances yet. Start the project to create instances from compose.</p>
        )}
      </CardContent>
    </Card>
  )
}

function ComposeTab({ stackName }: { stackName: string }) {
  const { data: compose, isLoading, error } = useStackCompose(stackName)

  if (isLoading) return <div className="flex justify-center py-8"><Loader2 className="size-6 animate-spin text-muted-foreground" /></div>
  if (error) return <Card><CardContent className="py-6 text-muted-foreground">No compose output available. Start the project first.</CardContent></Card>

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base flex items-center gap-2">
          <FileCode className="size-4" />Compose YAML
        </CardTitle>
      </CardHeader>
      <CardContent>
        <pre className="text-xs font-mono bg-muted p-4 rounded-md overflow-x-auto whitespace-pre max-h-[600px] overflow-y-auto">
          {compose?.yaml ?? 'No compose output'}
        </pre>
        {compose?.warnings && compose.warnings.length > 0 && (
          <div className="mt-4 space-y-1">
            {compose.warnings.map((w, i) => (
              <p key={i} className="text-sm text-yellow-500">{w}</p>
            ))}
          </div>
        )}
      </CardContent>
    </Card>
  )
}

function PlanTab({ stackName }: { stackName: string }) {
  const generatePlan = useGeneratePlan(stackName)
  const applyPlan = useApplyPlan(stackName)
  const [plan, setPlan] = useState<StackPlan | null>(null)

  const handleGenerate = () => {
    generatePlan.mutate(undefined, {
      onSuccess: (response) => {
        setPlan(response.data)
      },
    })
  }

  const handleApply = () => {
    if (!plan) return
    applyPlan.mutate({ token: plan.token }, {
      onSuccess: () => setPlan(null),
    })
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base flex items-center gap-2">
          <Diff className="size-4" />Deployment Plan
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        <Button onClick={handleGenerate} disabled={generatePlan.isPending} variant="outline" size="sm">
          {generatePlan.isPending ? <Loader2 className="size-4 animate-spin mr-2" /> : <Diff className="size-4 mr-2" />}
          Generate Plan
        </Button>

        {plan && (
          <div className="space-y-4">
            {plan.changes.length === 0 ? (
              <p className="text-sm text-muted-foreground">No changes needed. Stack is up to date.</p>
            ) : (
              <>
                <div className="grid gap-2">
                  {plan.changes.map((change, i) => (
                    <div key={i} className="flex items-center gap-3 text-sm py-1.5 border-b border-border/50 last:border-0">
                      <Badge variant={change.action === 'add' ? 'default' : change.action === 'remove' ? 'destructive' : 'secondary'} className="text-xs w-16 justify-center">
                        {change.action}
                      </Badge>
                      <span className="font-mono">{change.instance_id}</span>
                      <span className="text-muted-foreground">({change.template_name})</span>
                    </div>
                  ))}
                </div>
                <Button onClick={handleApply} disabled={applyPlan.isPending} size="sm">
                  {applyPlan.isPending ? <Loader2 className="size-4 animate-spin mr-2" /> : <Play className="size-4 mr-2" />}
                  Apply ({plan.changes.length} change{plan.changes.length !== 1 ? 's' : ''})
                </Button>
              </>
            )}
            {plan.warnings && plan.warnings.length > 0 && (
              <div className="space-y-1">
                {plan.warnings.map((w, i) => (
                  <p key={i} className="text-sm text-yellow-500">{w}</p>
                ))}
              </div>
            )}
          </div>
        )}
      </CardContent>
    </Card>
  )
}

function ServiceStatusBadge({ status }: { status: string }) {
  const colors = status === 'running' ? 'bg-green-500/10 text-green-500 border-green-500/20'
    : status === 'exited' || status === 'stopped' ? 'bg-red-500/10 text-red-500 border-red-500/20'
    : status === 'not-created' ? 'bg-muted text-muted-foreground border-border'
    : ''

  return <Badge variant="outline" className={`text-xs ${colors}`}>{status}</Badge>
}

function Row({ label, value, mono }: { label: string; value: string; mono?: boolean }) {
  return (
    <div className="flex">
      <span className="text-muted-foreground w-40 shrink-0">{label}:</span>
      <span className={mono ? 'font-mono text-xs' : ''}>{value}</span>
    </div>
  )
}
