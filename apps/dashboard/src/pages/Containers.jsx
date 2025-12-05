import { useState, useEffect } from 'react'
import { ContainerStatsGrid } from '../components/StatCard'
import { SearchBar } from '../components/SearchBar'
import { FilterBar } from '../components/FilterBar'
import { SortControls } from '../components/SortControls'
import { ContainersGrid } from '../components/ContainersGrid'
import { LoadingSpinner } from '../components/LoadingSpinner'
import { ErrorMessage } from '../components/ErrorMessage'
import { EmptyState } from '../components/EmptyState'
import { useContainers } from '../hooks/useContainers'
import { useLocalStorage } from '../hooks/useLocalStorage'
import { useDebounce } from '../hooks/useDebounce'

export function Containers() {
  const { containers, stats, loading, error, fetchContainers } = useContainers()

  // UI State
  const [searchQuery, setSearchQuery] = useState('')
  const [activeFilter, setActiveFilter] = useLocalStorage('devarch-container-filter', 'all')
  const [sortBy, setSortBy] = useLocalStorage('devarch-container-sortBy', 'name')
  const [sortOrder, setSortOrder] = useLocalStorage('devarch-container-sortOrder', 'asc')

  // Debounce search to avoid excessive filtering
  const debouncedSearch = useDebounce(searchQuery, 300)

  // Filter and sort containers locally for instant feedback
  const filteredContainers = containers
    .filter((container) => {
      // Apply status/category filter
      if (activeFilter !== 'all') {
        if (activeFilter === 'running' && container.status !== 'running') {
          return false
        }
        if (activeFilter === 'stopped' && container.status === 'running') {
          return false
        }
        if (!['all', 'running', 'stopped'].includes(activeFilter) && container.category !== activeFilter) {
          return false
        }
      }
      // Apply search filter
      if (debouncedSearch) {
        const searchLower = debouncedSearch.toLowerCase()
        return (
          container.name.toLowerCase().includes(searchLower) ||
          container.image.toLowerCase().includes(searchLower) ||
          container.category.toLowerCase().includes(searchLower)
        )
      }
      return true
    })
    .sort((a, b) => {
      const aVal = a[sortBy] || ''
      const bVal = b[sortBy] || ''

      let result = 0
      if (sortBy === 'cpu') {
        // Parse CPU percentages
        const aCpu = parseFloat(String(aVal).replace('%', '')) || 0
        const bCpu = parseFloat(String(bVal).replace('%', '')) || 0
        result = aCpu - bCpu
      } else if (typeof aVal === 'string') {
        result = aVal.localeCompare(bVal)
      } else {
        result = aVal - bVal
      }

      return sortOrder === 'desc' ? -result : result
    })

  // Initial data fetch
  useEffect(() => {
    fetchContainers()
  }, [fetchContainers])

  // Handle filter change
  const handleFilterChange = (filter) => {
    setActiveFilter(filter)
  }

  // Handle sort change
  const handleSortChange = (newSortBy, newSortOrder) => {
    setSortBy(newSortBy)
    setSortOrder(newSortOrder)
  }

  // Container filter options with visual separators
  const containerFilters = [
    // Status Group
    { value: 'all', label: 'All', count: stats.total },
    { value: 'running', label: 'Running', count: stats.running },
    { value: 'stopped', label: 'Stopped', count: stats.stopped },
    { separator: true },
    // Infrastructure Group
    { value: 'database', label: 'Database', count: stats.database },
    { value: 'backend', label: 'Backend', count: stats.backend },
    { value: 'proxy', label: 'Proxy', count: stats.proxy },
    { value: 'messaging', label: 'Messaging', count: stats.messaging },
    { value: 'search', label: 'Search', count: stats.search },
    { separator: true },
    // Tools Group
    { value: 'dbms', label: 'DBMS', count: stats.dbms },
    { value: 'management', label: 'Management', count: stats.management },
    { value: 'project', label: 'Project', count: stats.project },
    { value: 'mail', label: 'Mail', count: stats.mail },
    { value: 'ai', label: 'AI', count: stats.ai },
    { separator: true },
    // Observability Group
    { value: 'analytics', label: 'Analytics', count: stats.analytics },
    { value: 'exporters', label: 'Exporters', count: stats.exporters },
  ]

  return (
    <div className="space-y-6">
      {/* Statistics */}
      <ContainerStatsGrid stats={stats} />

      {/* Controls */}
      <div className="flex flex-col lg:flex-row gap-4 lg:items-center lg:justify-between">
        <div className="flex-1 max-w-xl">
          <SearchBar
            value={searchQuery}
            onChange={setSearchQuery}
            placeholder="Search containers..."
          />
        </div>
        <div className="flex flex-col sm:flex-row gap-4 sm:items-center">
          <SortControls
            sortBy={sortBy}
            sortOrder={sortOrder}
            onSortChange={handleSortChange}
            options={[
              { value: 'name', label: 'Name' },
              { value: 'status', label: 'Status' },
              { value: 'cpu', label: 'CPU' },
              { value: 'memory', label: 'Memory' },
            ]}
          />
        </div>
      </div>

      {/* Filters */}
      <FilterBar
        activeFilter={activeFilter}
        onFilterChange={handleFilterChange}
        filters={containerFilters}
      />

      {/* Content */}
      {loading && containers.length === 0 ? (
        <div className="py-16">
          <LoadingSpinner size="lg" />
        </div>
      ) : error ? (
        <ErrorMessage error={error} onRetry={fetchContainers} />
      ) : filteredContainers.length === 0 ? (
        <EmptyState filter={activeFilter} searchQuery={debouncedSearch} />
      ) : (
        <ContainersGrid containers={filteredContainers} />
      )}

      {/* Results count */}
      {filteredContainers.length > 0 && (
        <div className="text-center text-sm text-slate-600 dark:text-slate-400 pt-4">
          Showing {filteredContainers.length} of {stats.total} containers
        </div>
      )}
    </div>
  )
}
