import { useState, useEffect, useCallback } from 'react'

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || '/api'
const REFRESH_INTERVAL = import.meta.env.VITE_REFRESH_INTERVAL || 30000
const ENABLE_AUTO_REFRESH = import.meta.env.VITE_ENABLE_AUTO_REFRESH !== 'false'

/**
 * Custom hook for fetching and managing application data
 */
export function useApps() {
  const [apps, setApps] = useState([])
  const [stats, setStats] = useState({
    total: 0,
    php: 0,
    node: 0,
    python: 0,
    go: 0,
    dotnet: 0,
    unknown: 0,
  })
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)
  const [lastUpdated, setLastUpdated] = useState(null)

  const fetchApps = useCallback(async (filter = null, search = null, sort = 'name', order = 'asc') => {
    try {
      setLoading(true)
      setError(null)

      // Build query parameters
      const params = new URLSearchParams()
      if (filter && filter !== 'all') params.append('filter', filter)
      if (search) params.append('search', search)
      if (sort) params.append('sort', sort)
      if (order) params.append('order', order)

      const queryString = params.toString()
      const url = `${API_BASE_URL}/apps.php${queryString ? '?' + queryString : ''}`

      const response = await fetch(url)

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`)
      }

      const result = await response.json()

      if (result.success) {
        setApps(result.data.apps)
        setStats(result.data.stats)
        setLastUpdated(new Date())
      } else {
        throw new Error(result.message || 'Failed to fetch apps')
      }
    } catch (err) {
      setError(err.message)
      console.error('Error fetching apps:', err)
    } finally {
      setLoading(false)
    }
  }, [])

  // Auto-refresh functionality
  useEffect(() => {
    if (!ENABLE_AUTO_REFRESH) return

    const interval = setInterval(() => {
      fetchApps()
    }, REFRESH_INTERVAL)

    return () => clearInterval(interval)
  }, [fetchApps])

  return {
    apps,
    stats,
    loading,
    error,
    lastUpdated,
    fetchApps,
    refetch: fetchApps,
  }
}
