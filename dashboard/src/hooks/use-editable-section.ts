import { useState } from 'react'

export function useEditableSection<TDraft>(toDrafts: () => TDraft[]) {
  const [editing, setEditing] = useState(false)
  const [drafts, setDrafts] = useState<TDraft[]>([])

  const startEdit = () => {
    setDrafts(toDrafts())
    setEditing(true)
  }

  const cancel = () => setEditing(false)

  const update = (index: number, patch: Partial<TDraft>) => {
    setDrafts((prev) => prev.map((d, i) => (i === index ? { ...d, ...patch } : d)))
  }

  const add = (template: TDraft) => {
    setDrafts((prev) => [...prev, template])
  }

  const remove = (index: number) => {
    setDrafts((prev) => prev.filter((_, i) => i !== index))
  }

  return { editing, drafts, setDrafts, setEditing, startEdit, cancel, update, add, remove }
}
