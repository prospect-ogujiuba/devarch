import { useState, useEffect, useMemo } from 'react'
import { createFileRoute, Link } from '@tanstack/react-router'
import { z } from 'zod'
import { Package, Star, Download, Search, BadgeCheck } from 'lucide-react'
import { Card, CardContent } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Badge } from '@/components/ui/badge'
import { PaginationControls } from '@/components/ui/pagination-controls'
import { useRegistries, useSearchImages } from '@/features/registry/queries'
import { DEFAULT_PAGE_SIZE, PAGE_SIZE_OPTIONS } from '@/lib/pagination'

export const Route = createFileRoute('/registries/')({
  validateSearch: z.object({
    q: z.string().optional(),
    registry: z.string().optional(),
  }),
  component: RegistriesPage,
})

function formatCount(n: number): string {
  if (n >= 1_000_000_000) return `${(n / 1_000_000_000).toFixed(1)}B`
  if (n >= 1_000_000) return `${(n / 1_000_000).toFixed(1)}M`
  if (n >= 1_000) return `${(n / 1_000).toFixed(1)}K`
  return String(n)
}

function RegistriesPage() {
  const routeSearch = Route.useSearch()
  const routeNavigate = Route.useNavigate()

  const { data: registries } = useRegistries()
  const [selectedRegistry, setSelectedRegistry] = useState(routeSearch.registry ?? 'dockerhub')
  const [searchInput, setSearchInput] = useState(routeSearch.q ?? '')
  const [debouncedQuery, setDebouncedQuery] = useState(routeSearch.q ?? '')

  useEffect(() => {
    const timer = setTimeout(() => setDebouncedQuery(searchInput), 300)
    return () => clearTimeout(timer)
  }, [searchInput])

  useEffect(() => {
    routeNavigate({
      search: (prev) => ({
        ...prev,
        q: debouncedQuery || undefined,
        registry: selectedRegistry !== 'dockerhub' ? selectedRegistry : undefined,
      }),
      replace: true,
    })
  }, [debouncedQuery, selectedRegistry, routeNavigate])

  const { data: results, isLoading } = useSearchImages(selectedRegistry, debouncedQuery)
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(DEFAULT_PAGE_SIZE)

  useEffect(() => {
    setPage(1) // eslint-disable-line react-hooks/set-state-in-effect
  }, [debouncedQuery, selectedRegistry])

  const totalItems = results?.length ?? 0
  const totalPages = Math.max(1, Math.ceil(totalItems / pageSize))
  const pagedResults = useMemo(
    () => (results ?? []).slice((page - 1) * pageSize, page * pageSize),
    [results, page, pageSize],
  )

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-3">
        <Package className="size-6 text-muted-foreground" />
        <h1 className="text-2xl font-bold">Registry Browser</h1>
      </div>

      <div className="flex flex-col gap-3 sm:flex-row">
        <Select value={selectedRegistry} onValueChange={setSelectedRegistry}>
          <SelectTrigger className="w-full sm:w-48">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            {(registries ?? []).map((r) => (
              <SelectItem key={r.name} value={r.name}>{r.name}</SelectItem>
            ))}
            {!registries?.length && <SelectItem value="dockerhub">dockerhub</SelectItem>}
          </SelectContent>
        </Select>
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            value={searchInput}
            onChange={(e) => setSearchInput(e.target.value)}
            placeholder="Search images..."
            className="pl-9"
          />
        </div>
      </div>

      {!debouncedQuery && results && results.length > 0 && (
        <h2 className="text-lg font-semibold">Popular Images</h2>
      )}

      {isLoading && (
        <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
          {Array.from({ length: 6 }).map((_, i) => (
            <Card key={i} className="animate-pulse">
              <CardContent className="space-y-3 pt-4">
                <div className="h-5 w-2/3 rounded bg-muted" />
                <div className="h-4 w-full rounded bg-muted" />
                <div className="h-4 w-1/2 rounded bg-muted" />
              </CardContent>
            </Card>
          ))}
        </div>
      )}

      {!isLoading && results && results.length === 0 && debouncedQuery && (
        <div className="flex flex-col items-center justify-center py-20 text-center">
          <Package className="mb-4 size-12 text-muted-foreground/50" />
          <p className="text-lg text-muted-foreground">No images found</p>
        </div>
      )}

      {results && results.length > 0 && (
        <>
          <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
            {pagedResults.map((img) => (
              <Link
                key={img.name}
                to="/registries/$registry/$"
                params={{ registry: selectedRegistry, _splat: img.name }}
                className="block"
              >
                <Card className="transition-colors hover:border-primary/50">
                  <CardContent className="space-y-2 pt-4">
                    <div className="flex items-start justify-between gap-2">
                      <h3 className="truncate font-semibold">{img.name}</h3>
                      {img.is_official && (
                        <Badge variant="secondary" className="shrink-0">
                          <BadgeCheck className="mr-1 size-3" />
                          Official
                        </Badge>
                      )}
                    </div>
                    <p className="line-clamp-2 text-sm text-muted-foreground">
                      {img.description || 'No description'}
                    </p>
                    <div className="flex items-center gap-4 text-xs text-muted-foreground">
                      <span className="flex items-center gap-1">
                        <Star className="size-3" /> {formatCount(img.star_count)}
                      </span>
                      <span className="flex items-center gap-1">
                        <Download className="size-3" /> {formatCount(img.pull_count)}
                      </span>
                    </div>
                  </CardContent>
                </Card>
              </Link>
            ))}
          </div>
          <PaginationControls
            page={page}
            totalPages={totalPages}
            totalItems={totalItems}
            pageSize={pageSize}
            pageSizeOptions={PAGE_SIZE_OPTIONS}
            onPageChange={setPage}
            onPageSizeChange={(size) => { setPageSize(size); setPage(1) }}
            itemLabel="images"
          />
        </>
      )}
    </div>
  )
}
