import { useEffect, useRef } from 'react'
import { useListControls, type ListControlsOptions } from './use-list-controls'
import type { ViewMode } from '@/components/ui/view-switcher'
import type { SortOption } from '@/components/ui/sort-controls'

interface UrlSyncConfig {
  routeSearch: Record<string, string | undefined>
  navigate: (opts: { search: (prev: Record<string, unknown>) => Record<string, unknown>; replace: boolean }) => void
  sortOptions: SortOption[]
  filterKeys?: string[]
}

export function useUrlSyncedListControls<T>(
  listOptions: ListControlsOptions<T>,
  urlConfig: UrlSyncConfig,
) {
  const controls = useListControls(listOptions)
  const { routeSearch, navigate, sortOptions, filterKeys = [] } = urlConfig
  const syncingFromUrlRef = useRef(false)

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

  const defaultSort = listOptions.defaultSort ?? Object.keys(listOptions.sortFns)[0] ?? 'name'
  const defaultView = listOptions.defaultView ?? 'grid'

  useEffect(() => {
    syncingFromUrlRef.current = true
    setSearch(routeSearch.q ?? '')
    setSortBy(routeSearch.sort ?? defaultSort)
    setSortDir((routeSearch.dir as 'asc' | 'desc') ?? 'asc')
    setViewMode((routeSearch.view as ViewMode) ?? defaultView)
    for (const key of filterKeys) {
      setFilter(key, routeSearch[key] ?? 'all')
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [
    routeSearch.q,
    routeSearch.sort,
    routeSearch.dir,
    routeSearch.view,
    // eslint-disable-next-line react-hooks/exhaustive-deps
    ...filterKeys.map((k) => routeSearch[k]),
  ])

  useEffect(() => {
    if (syncingFromUrlRef.current) {
      syncingFromUrlRef.current = false
      return
    }

    const nextSearch: Record<string, string | undefined> = {}

    nextSearch.q = search || undefined
    nextSearch.sort =
      sortBy !== defaultSort && sortOptions.some((o) => o.value === sortBy)
        ? sortBy
        : undefined
    nextSearch.dir = sortDir === 'asc' ? undefined : sortDir
    nextSearch.view = viewMode === defaultView ? undefined : viewMode

    for (const key of filterKeys) {
      const val = filters[key]
      nextSearch[key] = val && val !== 'all' ? val : undefined
    }

    let changed = false
    for (const key of ['q', 'sort', 'dir', 'view', ...filterKeys]) {
      if (routeSearch[key] !== nextSearch[key]) {
        changed = true
        break
      }
    }
    if (!changed) return

    navigate({
      search: (prev) => ({ ...prev, ...nextSearch }),
      replace: true,
    })
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [
    search,
    sortBy,
    sortDir,
    viewMode,
    // eslint-disable-next-line react-hooks/exhaustive-deps
    ...filterKeys.map((k) => filters[k]),
    // eslint-disable-next-line react-hooks/exhaustive-deps
    ...['q', 'sort', 'dir', 'view', ...filterKeys].map((k) => routeSearch[k]),
    navigate,
  ])

  return controls
}
