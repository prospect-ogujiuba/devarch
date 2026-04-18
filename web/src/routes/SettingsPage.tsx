import { FormEvent, useMemo, useState } from 'react'
import { Card } from '../components/Card'
import { resetUiPreferences, getApiBase, getDefaultLogTail, getSelectedWorkspace, setApiBase, setDefaultLogTail } from '../lib/settings'

export function SettingsPage() {
  const [apiBase, setApiBaseState] = useState(getApiBase())
  const [logTail, setLogTailState] = useState(String(getDefaultLogTail()))
  const [savedAt, setSavedAt] = useState<string | null>(null)
  const selectedWorkspace = getSelectedWorkspace()

  const apiPreview = useMemo(() => {
    if (typeof window === 'undefined') {
      return apiBase || '/api'
    }
    return new URL((apiBase || '/api').startsWith('http') ? apiBase : apiBase || '/api', window.location.origin).toString()
  }, [apiBase])

  const save = (event: FormEvent) => {
    event.preventDefault()
    setApiBase(apiBase)
    setDefaultLogTail(Number.parseInt(logTail || '200', 10))
    setSavedAt(new Date().toLocaleTimeString())
  }

  const reset = () => {
    resetUiPreferences()
    setApiBaseState(getApiBase())
    setLogTailState(String(getDefaultLogTail()))
    setSavedAt(new Date().toLocaleTimeString())
  }

  return (
    <div className="page stack-lg">
      <header className="page__header">
        <div>
          <div className="page__eyebrow">Phase 5 · Settings</div>
          <h1 className="page__title">Settings</h1>
          <p className="page__subtitle">
            Keep UI settings small and developer-tool oriented while the engine-owned behavior stays in the API and runtime layers.
          </p>
        </div>
      </header>

      <div className="detail-grid">
        <Card title="UI preferences" description="Stored locally in the browser for the current operator.">
          <form className="stack" onSubmit={save}>
            <label className="form-field">
              <span>API base</span>
              <input value={apiBase} onChange={(event) => setApiBaseState(event.target.value)} placeholder="/api" />
            </label>

            <label className="form-field">
              <span>Default log tail</span>
              <input
                type="number"
                min={1}
                value={logTail}
                onChange={(event) => setLogTailState(event.target.value)}
              />
            </label>

            <div className="button-group">
              <button className="button" type="submit">
                Save settings
              </button>
              <button className="button button--ghost" type="button" onClick={reset}>
                Reset defaults
              </button>
            </div>
          </form>
          {savedAt ? <div className="panel">Saved at {savedAt}</div> : null}
        </Card>

        <Card title="Current wiring" description="Thin and explicit integration details for the current UI slice.">
          <dl className="key-value-grid key-value-grid--single">
            <div>
              <dt>Resolved API base</dt>
              <dd>{apiPreview}</dd>
            </div>
            <div>
              <dt>Selected workspace</dt>
              <dd>{selectedWorkspace || 'None selected yet'}</dd>
            </div>
            <div>
              <dt>Navigation scope</dt>
              <dd>Workspaces, Catalog, Activity, Settings</dd>
            </div>
            <div>
              <dt>Manifest writes</dt>
              <dd>Not exposed by the current Phase 4 API</dd>
            </div>
          </dl>
        </Card>

        <Card title="Slice notes" description="Intentional limitations kept explicit instead of hidden in code.">
          <div className="stack-sm">
            <div className="panel panel--warning">
              Raw manifest editing is a validated local draft until the API exposes a manifest write/save surface.
            </div>
            <div className="panel">
              Activity is scoped to one workspace at a time because Phase 4 currently exposes workspace-scoped SSE endpoints.
            </div>
            <div className="panel">
              The V2 route map intentionally avoids the V1 dashboard category/service/instance route sprawl.
            </div>
          </div>
        </Card>
      </div>
    </div>
  )
}
