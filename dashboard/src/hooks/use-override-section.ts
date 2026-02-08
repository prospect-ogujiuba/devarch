import { useEditableSection } from './use-editable-section'

export function useOverrideSection<TDraft>(toDrafts: () => TDraft[]) {
  const section = useEditableSection(toDrafts)

  const isDirty = JSON.stringify(section.drafts) !== JSON.stringify(toDrafts())

  return { ...section, isDirty }
}
