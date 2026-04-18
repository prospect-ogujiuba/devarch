import { useCallback, useEffect, useState } from 'react'
import type { TemplateDetail, TemplateSummary } from '../../generated/api'
import { api } from '../../lib/api'

function toErrorMessage(error: unknown) {
  return error instanceof Error ? error.message : 'Request failed'
}

export function useCatalogTemplates() {
  const [data, setData] = useState<TemplateSummary[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const load = useCallback(async () => {
    setLoading(true)
    setError(null)
    try {
      setData(await api.listTemplates())
    } catch (nextError) {
      setError(toErrorMessage(nextError))
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    void load()
  }, [load])

  return { data, loading, error, reload: load }
}

export function useCatalogTemplate(name?: string) {
  const [data, setData] = useState<TemplateDetail | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const load = useCallback(async () => {
    if (!name) {
      setData(null)
      setError(null)
      setLoading(false)
      return
    }

    setLoading(true)
    setError(null)
    try {
      setData(await api.getTemplate(name))
    } catch (nextError) {
      setError(toErrorMessage(nextError))
      setData(null)
    } finally {
      setLoading(false)
    }
  }, [name])

  useEffect(() => {
    void load()
  }, [load])

  return { data, loading, error, reload: load }
}
