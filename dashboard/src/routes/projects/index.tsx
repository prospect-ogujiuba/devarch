import { createFileRoute } from '@tanstack/react-router'
import { Loader2, RefreshCw, FolderOpen } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { useProjects, useScanProjects } from '@/features/projects/queries'
import { ProjectCard } from '@/components/projects/project-card'
import { useState } from 'react'

export const Route = createFileRoute('/projects/')({
  component: ProjectsPage,
})

function ProjectsPage() {
  const { data: projects, isLoading } = useProjects()
  const scanMutation = useScanProjects()
  const [typeFilter, setTypeFilter] = useState('')

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <Loader2 className="size-8 animate-spin text-muted-foreground" />
      </div>
    )
  }

  const types = [...new Set((projects ?? []).map(p => p.project_type))].sort()
  const filtered = typeFilter
    ? (projects ?? []).filter(p => p.project_type === typeFilter)
    : (projects ?? [])

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">Projects</h1>
          <p className="text-muted-foreground">
            {filtered.length} project{filtered.length !== 1 ? 's' : ''} in apps/
          </p>
        </div>
        <Button
          variant="outline"
          size="sm"
          onClick={() => scanMutation.mutate()}
          disabled={scanMutation.isPending}
        >
          {scanMutation.isPending ? (
            <Loader2 className="size-4 animate-spin" />
          ) : (
            <RefreshCw className="size-4" />
          )}
          Rescan
        </Button>
      </div>

      {types.length > 1 && (
        <div className="flex items-center gap-2 flex-wrap">
          <Button
            variant={typeFilter === '' ? 'default' : 'outline'}
            size="sm"
            onClick={() => setTypeFilter('')}
          >
            All
          </Button>
          {types.map(type_ => (
            <Button
              key={type_}
              variant={typeFilter === type_ ? 'default' : 'outline'}
              size="sm"
              onClick={() => setTypeFilter(type_)}
              className="capitalize"
            >
              {type_}
            </Button>
          ))}
        </div>
      )}

      {filtered.length === 0 ? (
        <div className="text-center py-16">
          <FolderOpen className="size-12 text-muted-foreground mx-auto mb-4" />
          <p className="text-muted-foreground">No projects found</p>
          <Button
            variant="outline"
            size="sm"
            className="mt-4"
            onClick={() => scanMutation.mutate()}
          >
            Scan apps folder
          </Button>
        </div>
      ) : (
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          {filtered.map(project => (
            <ProjectCard key={project.id} project={project} />
          ))}
        </div>
      )}
    </div>
  )
}
