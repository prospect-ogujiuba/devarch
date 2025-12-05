/**
 * Runtime color mappings
 */
export const RUNTIME_COLORS = {
  php: '#8892BF',
  node: '#68A063',
  python: '#3776AB',
  go: '#00ADD8',
  dotnet: '#512BD4',
  unknown: '#94a3b8',
}

/**
 * Status color mappings
 */
export const STATUS_COLORS = {
  active: '#eab308',
  ready: '#22c55e',
  stopped: '#94a3b8',
  unknown: '#cbd5e1',
}

/**
 * Container status color mappings
 */
export const CONTAINER_STATUS_COLORS = {
  running: '#22c55e',
  stopped: '#6b7280',
  exited: '#ef4444',
  paused: '#eab308',
  unknown: '#cbd5e1',
}

/**
 * Get Tailwind color class for runtime badge
 */
export function getRuntimeBgClass(runtime) {
  const classMap = {
    php: 'bg-[#8892BF]',
    node: 'bg-[#68A063]',
    python: 'bg-[#3776AB]',
    go: 'bg-[#00ADD8]',
    dotnet: 'bg-[#512BD4]',
    unknown: 'bg-slate-400',
  }
  return classMap[runtime] || classMap.unknown
}

/**
 * Get Tailwind color class for status indicator
 */
export function getStatusBgClass(status) {
  const classMap = {
    active: 'bg-yellow-400',
    ready: 'bg-green-500',
    stopped: 'bg-slate-400',
    unknown: 'bg-slate-300 dark:bg-slate-600',
  }
  return classMap[status] || classMap.unknown
}

/**
 * Get Tailwind text color class for stats
 */
export function getRuntimeTextClass(runtime) {
  const classMap = {
    php: 'text-[#8892BF]',
    node: 'text-[#68A063]',
    python: 'text-[#3776AB]',
    go: 'text-[#00ADD8]',
    dotnet: 'text-[#512BD4]',
    unknown: 'text-slate-400',
  }
  return classMap[runtime] || classMap.unknown
}
