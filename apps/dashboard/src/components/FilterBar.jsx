import { formatRuntime } from '../utils/formatters'

const FILTER_OPTIONS = [
  { value: 'all', label: 'All Apps' },
  { value: 'php', label: formatRuntime('php') },
  { value: 'node', label: formatRuntime('node') },
  { value: 'python', label: formatRuntime('python') },
  { value: 'go', label: formatRuntime('go') },
  { value: 'dotnet', label: formatRuntime('dotnet') },
]

export function FilterBar({ activeFilter, onFilterChange }) {
  return (
    <div className="flex flex-wrap gap-2">
      {FILTER_OPTIONS.map((option) => (
        <button
          key={option.value}
          onClick={() => onFilterChange(option.value)}
          className={`px-4 py-2 rounded-lg text-sm font-medium transition-all ${
            activeFilter === option.value
              ? 'bg-slate-900 dark:bg-slate-100 text-white dark:text-slate-900 border-2 border-slate-900 dark:border-slate-100'
              : 'bg-white dark:bg-slate-800 text-slate-700 dark:text-slate-300 border-2 border-slate-200 dark:border-slate-700 hover:border-slate-400 dark:hover:border-slate-500'
          }`}
        >
          {option.label}
        </button>
      ))}
    </div>
  )
}
