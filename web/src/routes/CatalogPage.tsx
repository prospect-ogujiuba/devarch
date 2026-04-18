import { useEffect } from 'react'
import { useSearchParams } from 'react-router-dom'
import { Card } from '../components/Card'
import { EmptyState } from '../components/EmptyState'
import { ErrorPanel } from '../components/ErrorPanel'
import { LoadingBlock } from '../components/LoadingBlock'
import { StatusBadge } from '../components/StatusBadge'
import { useCatalogTemplate, useCatalogTemplates } from '../features/catalog/hooks'
import { compactJson, formatEnvValue, formatPort, formatVolume } from '../lib/format'
import { classNames } from '../lib/utils'

export function CatalogPage() {
  const [searchParams, setSearchParams] = useSearchParams()
  const selectedTemplate = searchParams.get('template') ?? ''
  const templates = useCatalogTemplates()
  const template = useCatalogTemplate(selectedTemplate)

  useEffect(() => {
    if (!selectedTemplate && templates.data.length > 0) {
      setSearchParams({ template: templates.data[0].name }, { replace: true })
    }
  }, [selectedTemplate, setSearchParams, templates.data])

  return (
    <div className="page stack-lg">
      <header className="page__header">
        <div>
          <div className="page__eyebrow">Phase 5 · Catalog browser</div>
          <h1 className="page__title">Catalog</h1>
          <p className="page__subtitle">
            Browse reusable templates without reintroducing CRUD-heavy admin surfaces.
          </p>
        </div>
      </header>

      <div className="workspace-layout">
        <aside className="workspace-layout__sidebar">
          <Card
            title="Template list"
            description="Builtins and loaded templates exposed by the thin Phase 4 API."
            actions={
              <button className="button button--ghost" type="button" onClick={() => templates.reload()}>
                Refresh
              </button>
            }
          >
            {templates.loading ? <LoadingBlock label="Loading templates…" /> : null}
            {templates.error ? <ErrorPanel title="Template list unavailable" message={templates.error} /> : null}
            {!templates.loading && templates.data.length === 0 ? (
              <EmptyState title="No templates found" description="Add catalog roots to devarchd to populate the catalog browser." />
            ) : null}
            <div className="workspace-list">
              {templates.data.map((item) => (
                <button
                  key={item.name}
                  className={classNames('workspace-list__item', selectedTemplate === item.name && 'workspace-list__item--active')}
                  type="button"
                  onClick={() => setSearchParams({ template: item.name })}
                >
                  <div className="workspace-list__title-row">
                    <strong>{item.name}</strong>
                    <StatusBadge tone="info">template</StatusBadge>
                  </div>
                  <div className="chip-row">
                    {(item.tags || []).map((tag) => (
                      <span key={tag} className="chip">
                        {tag}
                      </span>
                    ))}
                  </div>
                  <p className="workspace-list__description">{item.description || 'No description provided.'}</p>
                </button>
              ))}
            </div>
          </Card>
        </aside>

        <section className="workspace-layout__detail">
          {template.loading && !template.data ? <LoadingBlock label="Loading template detail…" /> : null}
          {template.error ? <ErrorPanel title="Template detail unavailable" message={template.error} /> : null}
          {!template.loading && !template.error && !template.data ? (
            <EmptyState title="Select a template" description="Choose a template from the catalog list to inspect its detail." />
          ) : null}

          {template.data ? (
            <div className="detail-grid">
              <Card title={template.data.name} description={template.data.description || 'Reusable catalog template.'}>
                <dl className="key-value-grid">
                  <div>
                    <dt>Kind</dt>
                    <dd>{template.data.kind || 'Template'}</dd>
                  </div>
                  <div>
                    <dt>API version</dt>
                    <dd>{template.data.apiVersion || '—'}</dd>
                  </div>
                  <div>
                    <dt>Ports</dt>
                    <dd>{template.data.ports?.map(formatPort).join(', ') || '—'}</dd>
                  </div>
                  <div>
                    <dt>Volumes</dt>
                    <dd>{template.data.volumes?.map(formatVolume).join(', ') || '—'}</dd>
                  </div>
                  <div>
                    <dt>Imports</dt>
                    <dd>{template.data.imports?.map((item) => item.contract).join(', ') || '—'}</dd>
                  </div>
                  <div>
                    <dt>Exports</dt>
                    <dd>{template.data.exports?.map((item) => item.contract).join(', ') || '—'}</dd>
                  </div>
                </dl>
              </Card>

              <Card title="Runtime defaults" description="Template runtime defaults are exposed as plain data.">
                <pre className="log-view">{compactJson(template.data.runtime || {})}</pre>
              </Card>

              <Card title="Environment defaults" description="Template env defaults stay visible and compact.">
                {!(template.data.env && Object.keys(template.data.env).length > 0) ? (
                  <EmptyState title="No env defaults" description="This template does not define env defaults." />
                ) : (
                  <dl className="key-value-grid key-value-grid--single">
                    {Object.entries(template.data.env).map(([key, value]) => (
                      <div key={key}>
                        <dt>{key}</dt>
                        <dd>{formatEnvValue(value)}</dd>
                      </div>
                    ))}
                  </dl>
                )}
              </Card>

              <Card title="Develop + health" description="Developer-oriented template behavior in one place.">
                <div className="stack">
                  <div>
                    <div className="field-label">Health check</div>
                    <pre className="log-view log-view--compact">{compactJson(template.data.health || {})}</pre>
                  </div>
                  <div>
                    <div className="field-label">Develop</div>
                    <pre className="log-view log-view--compact">{compactJson(template.data.develop || {})}</pre>
                  </div>
                </div>
              </Card>
            </div>
          ) : null}
        </section>
      </div>
    </div>
  )
}
