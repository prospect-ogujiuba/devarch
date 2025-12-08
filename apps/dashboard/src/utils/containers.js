/**
 * Container utility functions
 */

/**
 * Format container status for display
 */
export function formatContainerStatus(status) {
  if (!status) return 'Unknown'
  return status.charAt(0).toUpperCase() + status.slice(1)
}

/**
 * Format uptime for display
 */
export function formatUptime(uptime) {
  if (!uptime) return 'N/A'
  return uptime
}

/**
 * Get Tailwind color class for container status indicator
 */
export function getContainerStatusBgClass(status) {
  const classMap = {
    running: 'bg-green-500',
    exited: 'bg-red-500',
    stopped: 'bg-slate-400',
    paused: 'bg-yellow-400',
    'not-created': 'bg-slate-300 dark:bg-slate-600',
  }
  return classMap[status?.toLowerCase()] || 'bg-slate-300 dark:bg-slate-600'
}

/**
 * Get Tailwind color class for category badge
 */
export function getCategoryBgClass(category) {
  const classMap = {
    database: 'bg-[#3b82f6]',
    dbms: 'bg-[#8b5cf6]',
    proxy: 'bg-[#10b981]',
    management: 'bg-[#f59e0b]',
    backend: 'bg-[#ef4444]',
    project: 'bg-[#ec4899]',
    mail: 'bg-[#6366f1]',
    exporters: 'bg-[#14b8a6]',
    analytics: 'bg-[#06b6d4]',
    messaging: 'bg-[#f97316]',
    search: 'bg-[#84cc16]',
    ai: 'bg-[#a855f7]',
    other: 'bg-slate-400',
  }
  return classMap[category] || classMap.other
}

/**
 * Format category for display
 */
export function formatCategory(category) {
  if (!category) return 'Other'
  return category.charAt(0).toUpperCase() + category.slice(1)
}

/**
 * Get background color class for health status indicator
 */
export function getHealthStatusBgClass(status) {
  switch (status) {
    case 'healthy':
      return 'bg-green-200'
    case 'unhealthy':
      return 'bg-red-200'
    case 'starting':
      return 'bg-yellow-200'
    default:
      return 'bg-slate-200'
  }
}

/**
 * Format health status for display
 */
export function formatHealthStatus(status) {
  if (!status) return 'N/A'
  return status.charAt(0).toUpperCase() + status.slice(1)
}

/**
 * Get color class for resource usage bars based on percentage
 * Green <50%, Yellow 50-80%, Red >80%
 */
export function getResourceBarColor(percentage) {
  if (percentage >= 80) return 'bg-red-600'
  if (percentage >= 50) return 'bg-yellow-500'
  return 'bg-green-600'
}
