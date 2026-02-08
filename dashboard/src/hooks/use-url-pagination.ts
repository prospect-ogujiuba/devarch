import { useCallback, useEffect, useMemo } from 'react'

interface UrlPaginationConfig<T> {
  items: T[]
  routeSearch: Record<string, string | undefined>
  navigate: (opts: { search: (prev: Record<string, unknown>) => Record<string, unknown>; replace: boolean }) => void
  defaultPageSize: number
  pageSizeOptions: number[]
}

function parsePositiveInt(value: string | undefined, fallback: number): number {
  if (!value) return fallback
  const parsed = Number.parseInt(value, 10)
  if (!Number.isFinite(parsed) || parsed < 1) return fallback
  return parsed
}

export function useUrlPagination<T>({
  items,
  routeSearch,
  navigate,
  defaultPageSize,
  pageSizeOptions,
}: UrlPaginationConfig<T>) {
  const totalItems = items.length
  const pageSize = pageSizeOptions.includes(parsePositiveInt(routeSearch.size, defaultPageSize))
    ? parsePositiveInt(routeSearch.size, defaultPageSize)
    : defaultPageSize
  const totalPages = Math.max(1, Math.ceil(totalItems / pageSize))
  const rawPage = parsePositiveInt(routeSearch.page, 1)
  const page = Math.min(rawPage, totalPages)

  const pagedItems = useMemo(() => {
    const start = (page - 1) * pageSize
    return items.slice(start, start + pageSize)
  }, [items, page, pageSize])

  const writeSearch = useCallback((nextPage: number, nextPageSize: number) => {
    const normalizedPage = Math.max(1, nextPage)
    const normalizedSize = pageSizeOptions.includes(nextPageSize) ? nextPageSize : defaultPageSize
    const pageParam = normalizedPage === 1 ? undefined : String(normalizedPage)
    const sizeParam = normalizedSize === defaultPageSize ? undefined : String(normalizedSize)

    if (routeSearch.page === pageParam && routeSearch.size === sizeParam) {
      return
    }

    navigate({
      search: (prev) => ({
        ...prev,
        page: pageParam,
        size: sizeParam,
      }),
      replace: true,
    })
  }, [defaultPageSize, navigate, pageSizeOptions, routeSearch.page, routeSearch.size])

  const setPage = useCallback((nextPage: number) => {
    writeSearch(Math.min(Math.max(nextPage, 1), totalPages), pageSize)
  }, [pageSize, totalPages, writeSearch])

  const setPageSize = useCallback((nextPageSize: number) => {
    writeSearch(1, nextPageSize)
  }, [writeSearch])

  const resetPage = useCallback(() => {
    writeSearch(1, pageSize)
  }, [pageSize, writeSearch])

  useEffect(() => {
    const normalizedPage = rawPage === page ? routeSearch.page : (page === 1 ? undefined : String(page))
    const normalizedSize = routeSearch.size === undefined && pageSize === defaultPageSize
      ? undefined
      : (pageSize === defaultPageSize ? undefined : String(pageSize))

    if (routeSearch.page !== normalizedPage || routeSearch.size !== normalizedSize) {
      navigate({
        search: (prev) => ({
          ...prev,
          page: normalizedPage,
          size: normalizedSize,
        }),
        replace: true,
      })
    }
  }, [defaultPageSize, navigate, page, pageSize, rawPage, routeSearch.page, routeSearch.size])

  return {
    page,
    pageSize,
    totalPages,
    totalItems,
    pagedItems,
    pageSizeOptions,
    setPage,
    setPageSize,
    resetPage,
  }
}
