import { useState, useEffect, useMemo } from 'react'
import { ContainerStatsGrid } from '../components/StatCard'
import { SearchBar } from '../components/SearchBar'
import { FilterBar } from '../components/FilterBar'
import { SortControls } from '../components/SortControls'
import { ContainersGrid } from '../components/ContainersGrid'
import { BulkActionsToolbar } from '../components/BulkActionsToolbar'
import { LoadingSpinner } from '../components/LoadingSpinner'
import { ErrorMessage } from '../components/ErrorMessage'
import { EmptyState } from '../components/EmptyState'
import { useContainers } from '../hooks/useContainers'
import { useBulkControl } from '../hooks/useBulkControl'
import { useLocalStorage } from '../hooks/useLocalStorage'
import { useDebounce } from '../hooks/useDebounce'

export function Containers() {
  const { containers, stats, loading, error, fetchContainers } = useContainers()
  const { bulkAction } = useBulkControl()

  // UI State
  const [searchQuery, setSearchQuery] = useState('')
  const [activeFilter, setActiveFilter] = useState('all')
  const [sortBy, setSortBy] = useState('name')
  const [sortOrder, setSortOrder] = useState('asc')

  // Bulk selection state
  const [selectedContainers, setSelectedContainers] = useState(new Set())

  console.log('[RENDER] activeFilter:', activeFilter)

  // Debounce search to avoid excessive filtering
  const debouncedSearch = useDebounce(searchQuery, 300)

  // Filter and sort containers
  const filteredContainers = useMemo(() => {
    console.log('[FILTER] Active filter:', activeFilter, 'Total containers:', containers.length)

    const filtered = containers.filter((container) => {
      // Search filter
      if (debouncedSearch) {
        const search = debouncedSearch.toLowerCase()
        const matchesSearch =
          container.name.toLowerCase().includes(search) ||
          container.image.toLowerCase().includes(search) ||
          container.category.toLowerCase().includes(search)
        if (!matchesSearch) return false
      }

      // Status/Category filter
      if (activeFilter === 'all') return true
      if (activeFilter === 'running') return container.status === 'running'
      if (activeFilter === 'stopped') return container.status !== 'running' && container.status !== 'not-created'
      if (activeFilter === 'not-created') return container.status === 'not-created'

      // Category filter
      const matches = container.category === activeFilter
      if (matches) {
        console.log('[FILTER] Match:', container.name, 'category:', container.category)
      }
      return matches
    })

    console.log('[FILTER] Result:', filtered.length, 'containers')

    // Sort
    const sorted = filtered.sort((a, b) => {
      const aVal = a[sortBy] || ''
      const bVal = b[sortBy] || ''

      let result = 0
      if (sortBy === 'cpu') {
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

    return sorted
  }, [containers, activeFilter, debouncedSearch, sortBy, sortOrder])

  // Initial data fetch
  useEffect(() => {
    fetchContainers()
  }, [])

  // Handle filter change
  const handleFilterChange = (filter) => {
    setActiveFilter(filter)
  }

  // Handle sort change
  const handleSortChange = (newSortBy, newSortOrder) => {
    setSortBy(newSortBy)
    setSortOrder(newSortOrder)
  }

  // Bulk selection handlers
  const toggleSelect = (name) => {
    setSelectedContainers(prev => {
      const next = new Set(prev)
      next.has(name) ? next.delete(name) : next.add(name)
      return next
    })
  }

  const selectAll = () => {
    if (selectedContainers.size === filteredContainers.length) {
      // Deselect all
      setSelectedContainers(new Set())
    } else {
      // Select all filtered containers
      setSelectedContainers(new Set(filteredContainers.map(c => c.name)))
    }
  }

  const clearSelection = () => setSelectedContainers(new Set())

  const handleBulkComplete = (result) => {
    console.log('Bulk operation result:', result)
    clearSelection()
    // Refresh containers after bulk operation
    setTimeout(() => fetchContainers(), 1000)
  }

  // Container filter options with visual separators
  const containerFilters = [
    // Status Group
    { value: 'all', label: 'All', count: stats.total },
    { value: 'running', label: 'Running', count: stats.running },
    { value: 'stopped', label: 'Stopped', count: stats.stopped },
    { value: 'not-created', label: 'Not Created', count: stats.notCreated },
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

      {/* Bulk Actions Toolbar */}
      <BulkActionsToolbar
        selectedContainers={selectedContainers}
        onClear={clearSelection}
        onComplete={handleBulkComplete}
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
        <ContainersGrid
          containers={filteredContainers}
          onRefresh={fetchContainers}
          selectedContainers={selectedContainers}
          onToggleSelect={toggleSelect}
          onSelectAll={selectAll}
        />
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
