import { useState, useEffect } from 'react'
import { Header } from './components/Header'
import { StatsGrid } from './components/StatCard'
import { SearchBar } from './components/SearchBar'
import { FilterBar } from './components/FilterBar'
import { SortControls } from './components/SortControls'
import { AppsGrid } from './components/AppsGrid'
import { LoadingSpinner } from './components/LoadingSpinner'
import { ErrorMessage } from './components/ErrorMessage'
import { EmptyState } from './components/EmptyState'
import { useApps } from './hooks/useApps'
import { useLocalStorage } from './hooks/useLocalStorage'
import { useDebounce } from './hooks/useDebounce'
import { Containers } from './pages/Containers'

function App() {
  const { apps, stats, loading, error, lastUpdated, fetchApps } = useApps()

  // Tab navigation state
  const [activeTab, setActiveTab] = useLocalStorage('devarch-activeTab', 'apps')

  // UI State
  const [searchQuery, setSearchQuery] = useState('')
  const [activeFilter, setActiveFilter] = useLocalStorage('devarch-filter', 'all')
  const [sortBy, setSortBy] = useLocalStorage('devarch-sortBy', 'name')
  const [sortOrder, setSortOrder] = useLocalStorage('devarch-sortOrder', 'asc')

  // Debounce search to avoid excessive filtering
  const debouncedSearch = useDebounce(searchQuery, 300)

  // Filter and sort apps locally for instant feedback
  const filteredApps = apps
    .filter((app) => {
      // Apply runtime filter
      if (activeFilter !== 'all' && app.runtime !== activeFilter) {
        return false
      }
      // Apply search filter
      if (debouncedSearch) {
        const searchLower = debouncedSearch.toLowerCase()
        return (
          app.name.toLowerCase().includes(searchLower) ||
          app.framework.toLowerCase().includes(searchLower) ||
          app.runtime.toLowerCase().includes(searchLower)
        )
      }
      return true
    })
    .sort((a, b) => {
      const aVal = a[sortBy] || ''
      const bVal = b[sortBy] || ''

      let result = 0
      if (typeof aVal === 'string') {
        result = aVal.localeCompare(bVal)
      } else {
        result = aVal - bVal
      }

      return sortOrder === 'desc' ? -result : result
    })

  // Initial data fetch
  useEffect(() => {
    fetchApps()
  }, [fetchApps])

  // Handle filter change
  const handleFilterChange = (filter) => {
    setActiveFilter(filter)
  }

  // Handle sort change
  const handleSortChange = (newSortBy, newSortOrder) => {
    setSortBy(newSortBy)
    setSortOrder(newSortOrder)
  }

  // Handle manual refresh
  const handleRefresh = () => {
    fetchApps()
  }

  return (
    <div className="min-h-screen bg-slate-50 dark:bg-slate-900">
      <Header
        activeTab={activeTab}
        onTabChange={setActiveTab}
        lastUpdated={lastUpdated}
        onRefresh={handleRefresh}
        isRefreshing={loading && apps.length > 0}
      />

      <main className="max-w-[1400px] mx-auto px-4 sm:px-6 lg:px-8 py-6 sm:py-8 space-y-6">
        {activeTab === 'apps' ? (
          <>
            {/* Statistics */}
            <StatsGrid stats={stats} />

            {/* Controls */}
            <div className="flex flex-col lg:flex-row gap-4 lg:items-center lg:justify-between">
              <div className="flex-1 max-w-xl">
                <SearchBar
                  value={searchQuery}
                  onChange={setSearchQuery}
                  placeholder="Search applications..."
                />
              </div>
              <div className="flex flex-col sm:flex-row gap-4 sm:items-center">
                <SortControls
                  sortBy={sortBy}
                  sortOrder={sortOrder}
                  onSortChange={handleSortChange}
                />
              </div>
            </div>

            {/* Filters */}
            <FilterBar
              activeFilter={activeFilter}
              onFilterChange={handleFilterChange}
            />

            {/* Content */}
            {loading && apps.length === 0 ? (
              <div className="py-16">
                <LoadingSpinner size="lg" />
              </div>
            ) : error ? (
              <ErrorMessage error={error} onRetry={handleRefresh} />
            ) : filteredApps.length === 0 ? (
              <EmptyState filter={activeFilter} searchQuery={debouncedSearch} />
            ) : (
              <AppsGrid apps={filteredApps} />
            )}

            {/* Results count */}
            {filteredApps.length > 0 && (
              <div className="text-center text-sm text-slate-600 dark:text-slate-400 pt-4">
                Showing {filteredApps.length} of {stats.total} applications
              </div>
            )}
          </>
        ) : (
          <Containers />
        )}
      </main>
    </div>
  )
}

export default App
