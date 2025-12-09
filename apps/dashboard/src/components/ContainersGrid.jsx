import { ContainerCard } from './ContainerCard'

export function ContainersGrid({ containers, onRefresh, selectedContainers, onToggleSelect, onSelectAll }) {
  console.log('[GRID] Rendering', containers.length, 'cards')

  const hasSelection = selectedContainers && onToggleSelect && onSelectAll
  const allSelected = hasSelection && containers.length > 0 && selectedContainers.size === containers.length

  return (
    <>
      {hasSelection && (
        <div className="mb-4">
          <label className="flex items-center gap-2 text-sm">
            <input
              type="checkbox"
              checked={allSelected}
              onChange={onSelectAll}
              className="w-4 h-4 rounded"
            />
            <span className="dark:text-white">
              Select All ({selectedContainers.size} of {containers.length} selected)
            </span>
          </label>
        </div>
      )}

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {containers.map((container) => (
          <div key={container.name} className="relative h-full">
            {hasSelection && (
              <div className="absolute top-3 right-3 z-10">
                <label className="flex items-center justify-center w-8 h-8 bg-white dark:bg-slate-700 rounded-lg shadow-md hover:shadow-lg transition-shadow cursor-pointer border-2 border-slate-200 dark:border-slate-600">
                  <input
                    type="checkbox"
                    checked={selectedContainers.has(container.name)}
                    onChange={() => onToggleSelect(container.name)}
                    className="w-5 h-5 rounded cursor-pointer accent-blue-500"
                    onClick={(e) => e.stopPropagation()}
                  />
                </label>
              </div>
            )}
            <ContainerCard container={container} onRefresh={onRefresh} />
          </div>
        ))}
      </div>
    </>
  )
}
