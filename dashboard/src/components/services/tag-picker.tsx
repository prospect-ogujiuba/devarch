import { useMemo } from 'react'
import { Input } from '@/components/ui/input'
import { Combobox, type ComboboxOption } from '@/components/ui/combobox'
import { useImageTags } from '@/features/registry/queries'
import { detectRegistry } from '@/lib/registry'
import { formatBytes } from '@/lib/format'

interface TagPickerProps {
  imageName: string
  value: string
  onValueChange: (value: string) => void
}

export function TagPicker({ imageName, value, onValueChange }: TagPickerProps) {
  const { registryName, repository } = detectRegistry(imageName)
  const { data: tags, isLoading, isError } = useImageTags(registryName, repository, {
    enabled: !!imageName,
  })

  const options: ComboboxOption[] = useMemo(() => {
    if (!tags) return []
    return tags.map((t) => {
      const parts: string[] = []
      if (t.pushed_at) {
        parts.push(new Date(t.pushed_at).toLocaleDateString())
      }
      if (t.size_bytes > 0) {
        parts.push(formatBytes(t.size_bytes))
      }
      return {
        value: t.name,
        label: t.name,
        detail: parts.join(' · ') || undefined,
      }
    })
  }, [tags])

  if (!imageName || isError) {
    return (
      <Input
        value={value}
        onChange={(e) => onValueChange(e.target.value)}
        placeholder="latest"
      />
    )
  }

  return (
    <Combobox
      value={value}
      onValueChange={onValueChange}
      options={options}
      placeholder="latest"
      loading={isLoading}
      allowCustomValue
    />
  )
}
