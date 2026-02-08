import { useEffect, useMemo, useRef } from 'react'
import { createFileRoute } from '@tanstack/react-router'
import { z } from 'zod'
import { Loader2, RefreshCw, FolderOpen, Package, Code, Globe } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { useProjects, useScanProjects } from '@/features/projects/queries'
import { ProjectCard } from '@/components/projects/project-card'
import { ProjectTable } from '@/components/projects/project-table'
import { ListToolbar } from '@/components/ui/list-toolbar'
import { FilterBar, type FilterOption } from '@/components/ui/filter-bar'
import { StatCard } from '@/components/ui/stat-card'
import { EmptyState } from '@/components/ui/empty-state'
import { useListControls } from '@/hooks/use-list-controls'
import type { Project } from '@/types/api'

export const Route = createFileRoute('/projects/')({
  validateSearch: z.object({
    q: z.string().optional(),
    type: z.string().optional(),
    language: z.string().optional(),
    sort: z.enum(['name', 'type', 'services', 'updated']).optional(),
    dir: z.enum(['asc', 'desc']).optional(),
    view: z.enum(['table', 'grid']).optional(),
  }),
  component: ProjectsPage,
})

const searchFn = (p: Project, q: string) => {
  const lower = q.toLowerCase()
  return (
    p.name.toLowerCase().includes(lower) ||
    (p.framework?.toLowerCase().includes(lower) ?? false) ||
    (p.language?.toLowerCase().includes(lower) ?? false) ||
    (p.domain?.toLowerCase().includes(lower) ?? false)
  )
}

const filterFns = {
  type: (p: Project, value: string) => p.project_type === value,
  language: (p: Project, value: string) => p.language === value,
}

const sortFns: Record<string, (a: Project, b: Project) => number> = {
  name: (a, b) => a.name.localeCompare(b.name),
  type: (a, b) => a.project_type.localeCompare(b.project_type),
  services: (a, b) => a.service_count - b.service_count,
  updated: (a, b) => new Date(b.updated_at).getTime() - new Date(a.updated_at).getTime(),
}

const sortOptions = [
  { value: 'name', label: 'Name' },
  { value: 'type', label: 'Type' },
  { value: 'services', label: 'Services' },
  { value: 'updated', label: 'Updated' },
]

function ProjectsPage() {
  const { data: projects, isLoading } = useProjects()
  const routeSearch = Route.useSearch()
  const navigate = Route.useNavigate()
  const scanMutation = useScanProjects()
  const items = useMemo(() => projects ?? [], [projects])

  const controls = useListControls({
    storageKey: 'projects',
    items,
    searchFn,
    filterFns,
    sortFns,
    defaultSort: 'name',
    defaultView: 'grid',
  })

  const {
    search,
    setSearch,
    filters,
    setFilter,
    sortBy,
    setSortBy,
    sortDir,
    setSortDir,
    viewMode,
    setViewMode,
  } = controls
  const syncingFromUrlRef = useRef(false)

  useEffect(() => {
    syncingFromUrlRef.current = true
    setSearch(routeSearch.q ?? '')
    setFilter('type', routeSearch.type ?? 'all')
    setFilter('language', routeSearch.language ?? 'all')
    setSortBy(routeSearch.sort ?? 'name')
    setSortDir(routeSearch.dir ?? 'asc')
    setViewMode(routeSearch.view ?? 'grid')
  }, [
    routeSearch.q,
    routeSearch.type,
    routeSearch.language,
    routeSearch.sort,
    routeSearch.dir,
    routeSearch.view,
    setSearch,
    setFilter,
    setSortBy,
    setSortDir,
    setViewMode,
  ])

  useEffect(() => {
    if (syncingFromUrlRef.current) {
      syncingFromUrlRef.current = false
      return
    }

    const nextQ = search || undefined
    const nextType = filters.type && filters.type !== 'all' ? filters.type : undefined
    const nextLanguage = filters.language && filters.language !== 'all' ? filters.language : undefined
    const nextSort =
      sortBy !== 'name' && sortOptions.some((option) => option.value === sortBy)
        ? (sortBy as typeof routeSearch.sort)
        : undefined
    const nextDir = sortDir === 'asc' ? undefined : sortDir
    const nextView = viewMode === 'grid' ? undefined : viewMode

    if (
      routeSearch.q === nextQ
      && routeSearch.type === nextType
      && routeSearch.language === nextLanguage
      && routeSearch.sort === nextSort
      && routeSearch.dir === nextDir
      && routeSearch.view === nextView
    ) {
      return
    }

    navigate({
      search: (prev) => ({
        ...prev,
        q: nextQ,
        type: nextType,
        language: nextLanguage,
        sort: nextSort,
        dir: nextDir,
        view: nextView,
      }),
      replace: true,
    })
  }, [
    search,
    filters.type,
    filters.language,
    sortBy,
    sortDir,
    viewMode,
    routeSearch.q,
    routeSearch.type,
    routeSearch.language,
    routeSearch.sort,
    routeSearch.dir,
    routeSearch.view,
    navigate,
  ])

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <Loader2 className="size-8 animate-spin text-muted-foreground" />
      </div>
    )
  }

  const types = [...new Set(items.map((p) => p.project_type))].sort()
  const languages = [...new Set(items.map((p) => p.language).filter(Boolean))].sort() as string[]
  const totalServices = items.reduce((acc, p) => acc + p.service_count, 0)
  const withDomains = items.filter((p) => p.domain).length

  const typeOptions: FilterOption[] = [
    { value: 'all', label: 'All Types', count: items.length },
    ...types.map((t) => ({
      value: t,
      label: t,
      count: items.filter((p) => p.project_type === t).length,
    })),
  ]

  const languageOptions: FilterOption[] = [
    { value: 'all', label: 'All Languages' },
    ...languages.map((l) => ({
      value: l,
      label: l,
      count: items.filter((p) => p.language === l).length,
    })),
  ]

  return (
    <div className="space-y-5 sm:space-y-6">
      <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 className="text-xl font-bold sm:text-2xl">Projects</h1>
          <p className="text-sm text-muted-foreground sm:text-base">
            {controls.filtered.length} project{controls.filtered.length !== 1 ? 's' : ''} in apps/
          </p>
        </div>
        <Button
          variant="outline"
          size="sm"
          className="w-full sm:w-auto"
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

      <div className="grid grid-cols-2 gap-3 sm:grid-cols-2 lg:grid-cols-4">
        <StatCard icon={FolderOpen} label="Projects" value={items.length} />
        <StatCard icon={Code} label="Types" value={types.length} />
        <StatCard icon={Package} label="Total Services" value={totalServices} />
        <StatCard icon={Globe} label="With Domains" value={withDomains} />
      </div>

      <ListToolbar
        search={controls.search}
        onSearchChange={controls.setSearch}
        searchPlaceholder="Search projects..."
        sortOptions={sortOptions}
        sortBy={controls.sortBy}
        sortDir={controls.sortDir}
        onSortByChange={controls.setSortBy}
        onSortDirChange={controls.setSortDir}
        viewMode={controls.viewMode}
        onViewModeChange={controls.setViewMode}
      >
        <FilterBar
          options={typeOptions}
          value={controls.filters.type ?? 'all'}
          onChange={(v) => controls.setFilter('type', v)}
        />
        {languages.length > 1 && (
          <FilterBar
            options={languageOptions}
            value={controls.filters.language ?? 'all'}
            onChange={(v) => controls.setFilter('language', v)}
          />
        )}
      </ListToolbar>

      {controls.filtered.length === 0 ? (
        <EmptyState
          icon={FolderOpen}
          message="No projects found"
          action={{ label: 'Scan apps folder', onClick: () => scanMutation.mutate() }}
        />
      ) : controls.viewMode === 'table' ? (
        <ProjectTable projects={controls.filtered} />
      ) : (
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          {controls.filtered.map((project) => (
            <ProjectCard key={project.id} project={project} />
          ))}
        </div>
      )}

      <div className="text-sm text-muted-foreground">
        Showing {controls.filtered.length} of {items.length} projects
      </div>
    </div>
  )
}
