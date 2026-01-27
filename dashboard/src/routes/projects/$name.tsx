import { createFileRoute, Link } from '@tanstack/react-router'
import { ArrowLeft, Loader2, Globe, GitBranch, Package, Code2, ExternalLink, Puzzle, Palette } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { ProjectLogo } from '@/components/projects/project-logo'
import { useProject } from '@/features/projects/queries'

export const Route = createFileRoute('/projects/$name')({
  component: ProjectDetailPage,
})

function ProjectDetailPage() {
  const { name } = Route.useParams()
  const { data: project, isLoading } = useProject(name)

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

  // For non-WordPress: flat deps as [name, version] pairs
  const flatDeps = isWordPress
    ? phpDeps
    : Object.entries(deps).filter(([, v]) => typeof v === 'string') as [string, string][]

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-4">
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
          </div>
          {project.description && (
            <p className="text-muted-foreground">{project.description}</p>
          )}
        </div>
        {project.domain && (
          <a
            href={`http://${project.domain}`}
            target="_blank"
            rel="noopener noreferrer"
            className="inline-flex items-center gap-1 text-sm text-primary hover:underline"
          >
            <ExternalLink className="size-4" />
            {project.domain}
          </a>
        )}
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

      <Tabs defaultValue="info" className="space-y-4">
        <TabsList>
          <TabsTrigger value="info">Info</TabsTrigger>
          <TabsTrigger value="deps">Dependencies ({flatDeps.length})</TabsTrigger>
          {isWordPress && plugins.length > 0 && (
            <TabsTrigger value="plugins" className="flex items-center gap-1.5">
              <Puzzle className="size-3.5" />
              Plugins ({plugins.length})
            </TabsTrigger>
          )}
          {isWordPress && themes.length > 0 && (
            <TabsTrigger value="themes" className="flex items-center gap-1.5">
              <Palette className="size-3.5" />
              Themes ({themes.length})
            </TabsTrigger>
          )}
          <TabsTrigger value="scripts">Scripts ({scriptEntries.length})</TabsTrigger>
          <TabsTrigger value="git">Git</TabsTrigger>
        </TabsList>

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
                {isWordPress && deps.db_name && <Row label="Database" value={String(deps.db_name)} />}
                {isWordPress && deps.table_prefix && <Row label="Table Prefix" value={String(deps.table_prefix)} mono />}
                {project.last_scanned_at && (
                  <Row label="Last Scanned" value={new Date(project.last_scanned_at).toLocaleString()} />
                )}
              </div>
            </CardContent>
          </Card>
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

function Row({ label, value, mono }: { label: string; value: string; mono?: boolean }) {
  return (
    <div className="flex">
      <span className="text-muted-foreground w-40 shrink-0">{label}:</span>
      <span className={mono ? 'font-mono text-xs' : ''}>{value}</span>
    </div>
  )
}
