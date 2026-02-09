import { useState, useRef, useEffect } from 'react'
import { Command } from 'cmdk'
import { Loader2, ChevronsUpDown, Check } from 'lucide-react'
import { cn } from '@/lib/utils'

export interface ComboboxOption {
  value: string
  label: string
  detail?: string
}

interface ComboboxProps {
  value: string
  onValueChange: (value: string) => void
  options: ComboboxOption[]
  placeholder?: string
  loading?: boolean
  allowCustomValue?: boolean
  className?: string
}

export function Combobox({
  value,
  onValueChange,
  options,
  placeholder = 'Select...',
  loading,
  allowCustomValue,
  className,
}: ComboboxProps) {
  const [open, setOpen] = useState(false)
  const [search, setSearch] = useState('')
  const containerRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    function handleClickOutside(e: MouseEvent) {
      if (containerRef.current && !containerRef.current.contains(e.target as Node)) {
        setOpen(false)
      }
    }
    document.addEventListener('mousedown', handleClickOutside)
    return () => document.removeEventListener('mousedown', handleClickOutside)
  }, [])

  const handleSelect = (selected: string) => {
    onValueChange(selected)
    setSearch('')
    setOpen(false)
  }

  const handleInputChange = (v: string) => {
    setSearch(v)
    if (allowCustomValue) {
      onValueChange(v)
    }
    if (!open) setOpen(true)
  }

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Escape') {
      setOpen(false)
    }
  }

  return (
    <div ref={containerRef} className={cn('relative', className)}>
      <Command shouldFilter={false} onKeyDown={handleKeyDown}>
        <div className="flex items-center rounded-md border border-input bg-background">
          <Command.Input
            value={open ? search : value}
            onValueChange={handleInputChange}
            onFocus={() => setOpen(true)}
            placeholder={placeholder}
            className="h-9 w-full bg-transparent px-3 py-1 text-sm outline-none placeholder:text-muted-foreground"
          />
          <div className="flex items-center gap-1 pr-2">
            {loading && <Loader2 className="size-4 animate-spin text-muted-foreground" />}
            <ChevronsUpDown className="size-4 text-muted-foreground" />
          </div>
        </div>

        {open && (
          <Command.List className="absolute z-50 mt-1 max-h-60 w-full overflow-auto rounded-md border bg-popover p-1 shadow-md">
            {!loading && options.length === 0 && (
              <Command.Empty className="px-2 py-6 text-center text-sm text-muted-foreground">
                No results found
              </Command.Empty>
            )}
            {options.map((opt) => (
              <Command.Item
                key={opt.value}
                value={opt.value}
                onSelect={handleSelect}
                className="flex cursor-pointer items-center gap-2 rounded-sm px-2 py-1.5 text-sm outline-none hover:bg-accent hover:text-accent-foreground data-[selected=true]:bg-accent data-[selected=true]:text-accent-foreground"
              >
                <Check className={cn('size-4', value === opt.value ? 'opacity-100' : 'opacity-0')} />
                <div className="flex-1 truncate">
                  <span>{opt.label}</span>
                  {opt.detail && (
                    <span className="ml-2 text-xs text-muted-foreground">{opt.detail}</span>
                  )}
                </div>
              </Command.Item>
            ))}
          </Command.List>
        )}
      </Command>
    </div>
  )
}
