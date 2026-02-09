import { useMemo, useCallback } from 'react'
import { createFileRoute } from '@tanstack/react-router'
import { z } from 'zod'
import { Loader2, RefreshCw, FolderOpen, Package, Code, Globe } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { useProjects, useScanProjects } from '@/features/projects/queries'
import { ProjectCard } from '@/components/projects/project-card'
import { ProjectTable } from '@/components/projects/project-table'
import { FilterBar, type FilterOption } from '@/components/ui/filter-bar'
import { ListPageScaffold } from '@/components/ui/list-page-scaffold'
import { PaginationControls } from '@/components/ui/pagination-controls'
import { useUrlSyncedListControls } from '@/hooks/use-url-synced-list-controls'
import { useUrlPagination } from '@/hooks/use-url-pagination'
import { DEFAULT_PAGE_SIZE, PAGE_SIZE_OPTIONS } from '@/lib/pagination'
import type { Project } from '@/types/api'

export const Route = createFileRoute('/projects/')({
  validateSearch: z.object({
    q: z.string().optional(),
    type: z.string().optional(),
    language: z.string().optional(),
    sort: z.enum(['name', 'type', 'services', 'updated']).optional(),
    dir: z.enum(['asc', 'desc']).optional(),
    view: z.enum(['table', 'grid']).optional(),
    page: z.string().optional(),
    size: z.string().optional(),
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

  const controls = useUrlSyncedListControls(
    { storageKey: 'projects', items, searchFn, filterFns, sortFns, defaultSort: 'name', defaultView: 'grid' },
    { routeSearch, navigate, sortOptions, filterKeys: ['type', 'language'] },
  )
  const pagination = useUrlPagination({
    items: controls.filtered,
    routeSearch,
    navigate,
    defaultPageSize: DEFAULT_PAGE_SIZE,
    pageSizeOptions: PAGE_SIZE_OPTIONS,
  })

  const handleSearchChange = useCallback((value: string) => {
    pagination.resetPage()
    controls.setSearch(value)
  }, [controls, pagination])

  const handleSortByChange = useCallback((value: string) => {
    pagination.resetPage()
    controls.setSortBy(value)
  }, [controls, pagination])

  const handleSortDirChange = useCallback((dir: 'asc' | 'desc') => {
    pagination.resetPage()
    controls.setSortDir(dir)
  }, [controls, pagination])

  const handleViewModeChange = useCallback((mode: 'table' | 'grid') => {
    pagination.resetPage()
    controls.setViewMode(mode)
  }, [controls, pagination])

  const handleFilterChange = useCallback((key: 'type' | 'language', value: string) => {
    pagination.resetPage()
    controls.setFilter(key, value)
  }, [controls, pagination])

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
    <ListPageScaffold
      title="Projects"
      subtitle={`${controls.filtered.length} project${controls.filtered.length !== 1 ? 's' : ''} in apps/`}
      isLoading={isLoading}
      statCards={[
        { icon: FolderOpen, label: 'Projects', value: items.length },
        { icon: Code, label: 'Types', value: types.length },
        { icon: Package, label: 'Total Services', value: totalServices },
        { icon: Globe, label: 'With Domains', value: withDomains },
      ]}
      statGridClassName="grid grid-cols-2 gap-3 sm:grid-cols-2 lg:grid-cols-4"
      controls={{
        ...controls,
        setSearch: handleSearchChange,
        setSortBy: handleSortByChange,
        setSortDir: handleSortDirChange,
        setViewMode: handleViewModeChange,
      }}
      sortOptions={sortOptions}
      searchPlaceholder="Search projects..."
      actionButton={
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
      }
      emptyIcon={FolderOpen}
      emptyMessage="No projects found"
      emptyAction={{ label: 'Scan apps folder', onClick: () => scanMutation.mutate() }}
      items={controls.filtered}
      filterChildren={
        <>
          <FilterBar
            options={typeOptions}
            value={controls.filters.type ?? 'all'}
            onChange={(v) => handleFilterChange('type', v)}
          />
          {languages.length > 1 && (
            <FilterBar
              options={languageOptions}
              value={controls.filters.language ?? 'all'}
              onChange={(v) => handleFilterChange('language', v)}
            />
          )}
        </>
      }
      tableView={() => <ProjectTable projects={pagination.pagedItems} />}
      gridView={() => (
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          {pagination.pagedItems.map((project) => (
            <ProjectCard key={project.id} project={project} />
          ))}
        </div>
      )}
    >
      {pagination.totalItems > 0 && (
        <PaginationControls
          page={pagination.page}
          totalPages={pagination.totalPages}
          totalItems={pagination.totalItems}
          pageSize={pagination.pageSize}
          pageSizeOptions={PAGE_SIZE_OPTIONS}
          onPageChange={pagination.setPage}
          onPageSizeChange={pagination.setPageSize}
          itemLabel="projects"
        />
      )}
    </ListPageScaffold>
  )
}
