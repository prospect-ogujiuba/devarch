/**
 * Format runtime name for display
 */
export function formatRuntime(runtime) {
  const nameMap = {
    php: 'PHP',
    node: 'Node.js',
    python: 'Python',
    go: 'Go',
    dotnet: '.NET',
    unknown: 'Unknown',
  }
  return nameMap[runtime] || runtime
}

/**
 * Format status for display
 */
export function formatStatus(status) {
  return status.charAt(0).toUpperCase() + status.slice(1)
}

/**
 * Format timestamp to relative time
 */
export function formatRelativeTime(date) {
  if (!date) return 'Never'

  const now = new Date()
  const diffMs = now - date
  const diffSecs = Math.floor(diffMs / 1000)
  const diffMins = Math.floor(diffSecs / 60)
  const diffHours = Math.floor(diffMins / 60)

  if (diffSecs < 60) {
    return 'Just now'
  } else if (diffMins < 60) {
    return `${diffMins} minute${diffMins !== 1 ? 's' : ''} ago`
  } else if (diffHours < 24) {
    return `${diffHours} hour${diffHours !== 1 ? 's' : ''} ago`
  } else {
    return date.toLocaleString()
  }
}

/**
 * Format time for display
 */
export function formatTime(date) {
  if (!date) return '--:--'
  return date.toLocaleTimeString('en-US', {
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  })
}
