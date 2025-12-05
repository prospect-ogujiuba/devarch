import { useState, useEffect, useCallback } from 'react'

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || '/api'
const REFRESH_INTERVAL = import.meta.env.VITE_REFRESH_INTERVAL || 30000
const ENABLE_AUTO_REFRESH = import.meta.env.VITE_ENABLE_AUTO_REFRESH !== 'false'

/**
 * Custom hook for fetching and managing container data
 */
export function useContainers() {
  const [containers, setContainers] = useState([])
  const [stats, setStats] = useState({
    total: 0,
    running: 0,
    stopped: 0,
    database: 0,
    backend: 0,
    proxy: 0,
    analytics: 0,
    dbms: 0,
    exporters: 0,
    messaging: 0,
    search: 0,
    mail: 0,
    project: 0,
    management: 0,
    ai: 0,
    other: 0,
  })
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)
  const [lastUpdated, setLastUpdated] = useState(null)

  const fetchContainers = useCallback(async (filter = null, search = null, sort = 'name', order = 'asc') => {
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
      const url = `${API_BASE_URL}/containers.php${queryString ? '?' + queryString : ''}`

      const response = await fetch(url)

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`)
      }

      const result = await response.json()

      if (result.success) {
        setContainers(result.data.containers)
        setStats(result.data.stats)
        setLastUpdated(new Date())
      } else {
        throw new Error(result.message || 'Failed to fetch containers')
      }
    } catch (err) {
      setError(err.message)
      console.error('Error fetching containers:', err)
    } finally {
      setLoading(false)
    }
  }, [])

  // Auto-refresh functionality
  useEffect(() => {
    if (!ENABLE_AUTO_REFRESH) return

    const interval = setInterval(() => {
      fetchContainers()
    }, REFRESH_INTERVAL)

    return () => clearInterval(interval)
  }, [fetchContainers])

  return {
    containers,
    stats,
    loading,
    error,
    lastUpdated,
    fetchContainers,
    refetch: fetchContainers,
  }
}
