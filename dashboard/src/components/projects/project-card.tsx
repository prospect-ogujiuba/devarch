import { Link } from '@tanstack/react-router'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import {
  Code2, Globe, GitBranch, Package, ExternalLink,
  FileCode, Blocks, Hexagon, Gem, Flame,
} from 'lucide-react'
import type { Project } from '@/types/api'

const typeIcons: Record<string, React.ReactNode> = {
  laravel: <Flame className="size-5 text-red-500" />,
  wordpress: <Blocks className="size-5 text-blue-500" />,
  node: <Hexagon className="size-5 text-green-500" />,
  go: <Code2 className="size-5 text-cyan-500" />,
  rust: <Gem className="size-5 text-orange-500" />,
  python: <FileCode className="size-5 text-yellow-500" />,
  php: <FileCode className="size-5 text-purple-500" />,
  unknown: <Package className="size-5 text-muted-foreground" />,
}

const typeColors: Record<string, string> = {
  laravel: 'bg-red-500/10 text-red-500 border-red-500/20',
  wordpress: 'bg-blue-500/10 text-blue-500 border-blue-500/20',
  node: 'bg-green-500/10 text-green-500 border-green-500/20',
  go: 'bg-cyan-500/10 text-cyan-500 border-cyan-500/20',
  rust: 'bg-orange-500/10 text-orange-500 border-orange-500/20',
  python: 'bg-yellow-500/10 text-yellow-500 border-yellow-500/20',
  php: 'bg-purple-500/10 text-purple-500 border-purple-500/20',
}

export function ProjectCard({ project }: { project: Project }) {
  const depCount = Object.keys(project.dependencies ?? {}).length
  const icon = typeIcons[project.project_type] ?? typeIcons.unknown
  const colorClass = typeColors[project.project_type] ?? ''

  return (
    <Link to="/projects/$name" params={{ name: project.name }}>
      <Card className="py-4 hover:border-foreground/20 transition-colors cursor-pointer h-full">
        <CardHeader className="pb-2">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              {icon}
              <CardTitle className="text-base">{project.name}</CardTitle>
            </div>
            <Badge variant="outline" className={colorClass}>
              {project.project_type}
            </Badge>
          </div>
        </CardHeader>
        <CardContent className="space-y-3">
          {project.description && (
            <p className="text-sm text-muted-foreground line-clamp-2">{project.description}</p>
          )}

          <div className="flex flex-wrap gap-2">
            {project.framework && (
              <Badge variant="secondary" className="text-xs">
                {project.framework}
              </Badge>
            )}
            {project.language && (
              <Badge variant="secondary" className="text-xs">
                {project.language}
              </Badge>
            )}
            {project.has_frontend && project.frontend_framework && (
              <Badge variant="secondary" className="text-xs">
                {project.frontend_framework}
              </Badge>
            )}
          </div>

          <div className="flex items-center gap-4 text-xs text-muted-foreground">
            {project.domain && (
              <span className="flex items-center gap-1">
                <Globe className="size-3" />
                {project.domain}
              </span>
            )}
            {project.git_branch && (
              <span className="flex items-center gap-1">
                <GitBranch className="size-3" />
                {project.git_branch}
              </span>
            )}
            {depCount > 0 && (
              <span className="flex items-center gap-1">
                <Package className="size-3" />
                {depCount} deps
              </span>
            )}
            {project.proxy_port && (
              <span className="flex items-center gap-1">
                <ExternalLink className="size-3" />
                :{project.proxy_port}
              </span>
            )}
          </div>
        </CardContent>
      </Card>
    </Link>
  )
}
