import { useState, useMemo, useCallback } from 'react'
import { useLocalStorage } from './use-local-storage'
import { useDebounce } from './use-debounce'
import type { ViewMode } from '@/components/ui/view-switcher'

export interface ListControlsOptions<T> {
  storageKey: string
  items: T[]
  searchFn: (item: T, query: string) => boolean
  filterFns?: Record<string, (item: T, value: string) => boolean>
  sortFns: Record<string, (a: T, b: T) => number>
  defaultSort?: string
  defaultView?: ViewMode
}

export function useListControls<T>({
  storageKey,
  items,
  searchFn,
  filterFns = {},
  sortFns,
  defaultSort = Object.keys(sortFns)[0] ?? 'name',
  defaultView = 'grid',
}: ListControlsOptions<T>) {
  const [viewMode, setViewMode] = useLocalStorage<ViewMode>(`${storageKey}-view`, defaultView)
  const [sortBy, setSortBy] = useLocalStorage(`${storageKey}-sort`, defaultSort)
  const [sortDir, setSortDir] = useLocalStorage<'asc' | 'desc'>(`${storageKey}-sort-dir`, 'asc')
  const [searchRaw, setSearch] = useState('')
  const search = useDebounce(searchRaw, 200)
  const [filters, setFilters] = useState<Record<string, string>>({})
  const [selected, setSelected] = useState<Set<string>>(new Set())

  const setFilter = useCallback((key: string, value: string) => {
    setFilters((prev) => ({ ...prev, [key]: value }))
  }, [])

  const toggleSelect = useCallback((id: string) => {
    setSelected((prev) => {
      const next = new Set(prev)
      if (next.has(id)) next.delete(id)
      else next.add(id)
      return next
    })
  }, [])

  const selectAll = useCallback((ids: string[]) => {
    setSelected(new Set(ids))
  }, [])

  const clearSelection = useCallback(() => {
    setSelected(new Set())
  }, [])

  const filtered = useMemo(() => {
    let result = items

    if (search) {
      result = result.filter((item) => searchFn(item, search))
    }

    for (const [key, filterFn] of Object.entries(filterFns)) {
      const value = filters[key]
      if (value && value !== 'all') {
        result = result.filter((item) => filterFn(item, value))
      }
    }

    const sortFn = sortFns[sortBy]
    if (sortFn) {
      result = [...result].sort((a, b) => {
        const cmp = sortFn(a, b)
        return sortDir === 'desc' ? -cmp : cmp
      })
    }

    return result
  }, [items, search, filters, sortBy, sortDir, searchFn, filterFns, sortFns])

  return {
    viewMode,
    setViewMode,
    sortBy,
    setSortBy,
    sortDir,
    setSortDir,
    search: searchRaw,
    setSearch,
    filters,
    setFilter,
    selected,
    toggleSelect,
    selectAll,
    clearSelection,
    filtered,
    total: items.length,
  }
}
