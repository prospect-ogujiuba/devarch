export function EmptyState({ filter, searchQuery }) {
  let message = 'No applications found'
  let description = 'No applications matching the selected filter were detected.'

  if (searchQuery) {
    message = 'No matching applications'
    description = `No applications match your search for "${searchQuery}".`
  } else if (filter && filter !== 'all') {
    description = `No ${filter.toUpperCase()} applications were detected.`
  }

  return (
    <div className="text-center py-16 px-4">
      <div className="text-6xl mb-4">📦</div>
      <h3 className="text-xl font-semibold text-slate-900 dark:text-slate-100 mb-2">
        {message}
      </h3>
      <p className="text-slate-600 dark:text-slate-400">
        {description}
      </p>
    </div>
  )
}
