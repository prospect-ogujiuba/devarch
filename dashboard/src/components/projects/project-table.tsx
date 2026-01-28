import { Link } from '@tanstack/react-router'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Badge } from '@/components/ui/badge'
import { ProjectLogo } from '@/components/projects/project-logo'
import type { Project } from '@/types/api'

const typeColors: Record<string, string> = {
  laravel: 'bg-red-500/10 text-red-500 border-red-500/20',
  wordpress: 'bg-blue-500/10 text-blue-500 border-blue-500/20',
  node: 'bg-green-500/10 text-green-500 border-green-500/20',
  go: 'bg-cyan-500/10 text-cyan-500 border-cyan-500/20',
  rust: 'bg-orange-500/10 text-orange-500 border-orange-500/20',
  python: 'bg-yellow-500/10 text-yellow-500 border-yellow-500/20',
  php: 'bg-purple-500/10 text-purple-500 border-purple-500/20',
}

interface ProjectTableProps {
  projects: Project[]
}

export function ProjectTable({ projects }: ProjectTableProps) {
  return (
    <div className="rounded-lg border">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Project</TableHead>
            <TableHead>Type</TableHead>
            <TableHead>Framework</TableHead>
            <TableHead>Language</TableHead>
            <TableHead>Domain</TableHead>
            <TableHead>Services</TableHead>
            <TableHead>Branch</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {projects.length === 0 ? (
            <TableRow>
              <TableCell colSpan={7} className="text-center text-muted-foreground py-8">
                No projects found
              </TableCell>
            </TableRow>
          ) : (
            projects.map((project) => (
              <TableRow key={project.id}>
                <TableCell>
                  <Link
                    to="/projects/$name"
                    params={{ name: project.name }}
                    className="font-medium hover:underline flex items-center gap-2"
                  >
                    <ProjectLogo
                      projectType={project.project_type}
                      framework={project.framework}
                      language={project.language}
                    />
                    {project.name}
                  </Link>
                </TableCell>
                <TableCell>
                  <Badge variant="outline" className={typeColors[project.project_type] ?? ''}>
                    {project.project_type}
                  </Badge>
                </TableCell>
                <TableCell className="text-muted-foreground">
                  {project.framework ?? '-'}
                </TableCell>
                <TableCell className="text-muted-foreground">
                  {project.language ?? '-'}
                </TableCell>
                <TableCell className="text-muted-foreground">
                  {project.domain ?? '-'}
                </TableCell>
                <TableCell>{project.service_count}</TableCell>
                <TableCell className="text-muted-foreground font-mono text-xs">
                  {project.git_branch ?? '-'}
                </TableCell>
              </TableRow>
            ))
          )}
        </TableBody>
      </Table>
    </div>
  )
}
