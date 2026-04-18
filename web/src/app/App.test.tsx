import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { AppRoutes } from './App'

function jsonResponse(body: unknown, status = 200) {
  return new Response(JSON.stringify(body), {
    status,
    headers: { 'Content-Type': 'application/json' },
  })
}

function installFetchMock() {
  const responses: Record<string, unknown> = {
    '/api/workspaces': [
      {
        name: 'shop-local',
        displayName: 'Shop Local',
        description: 'Template-only Phase 1 workspace example.',
        provider: 'docker',
        capabilities: { inspect: true, logs: true, exec: true },
        resourceCount: 4,
      },
    ],
    '/api/workspaces/shop-local': {
      name: 'shop-local',
      displayName: 'Shop Local',
      description: 'Template-only Phase 1 workspace example.',
      provider: 'docker',
      capabilities: { inspect: true, logs: true, exec: true },
      resourceCount: 4,
      manifestPath: '/examples/v2/workspaces/shop-local/devarch.workspace.yaml',
      resourceKeys: ['api', 'postgres', 'redis', 'web'],
    },
    '/api/workspaces/shop-local/manifest': {
      apiVersion: 'devarch.io/v2alpha1',
      kind: 'Workspace',
      metadata: {
        name: 'shop-local',
        displayName: 'Shop Local',
      },
      runtime: { provider: 'auto', isolatedNetwork: true },
      resources: {
        api: { template: 'node-api', ports: [{ host: 8200, container: 3000, protocol: 'tcp' }] },
      },
    },
    '/api/workspaces/shop-local/graph': {
      graph: {
        workspace: { name: 'shop-local', displayName: 'Shop Local' },
        resources: [
          {
            key: 'api',
            enabled: true,
            host: 'api',
            template: { name: 'node-api' },
            dependsOn: ['postgres', 'redis'],
            imports: [
              { contract: 'postgres', from: 'postgres' },
              { contract: 'redis', from: 'redis' },
            ],
            exports: [{ contract: 'http' }],
            ports: [{ host: 8200, container: 3000, protocol: 'tcp' }],
          },
        ],
      },
      contracts: {
        links: [
          {
            consumer: 'api',
            contract: 'postgres',
            provider: 'postgres',
            source: 'explicit',
          },
        ],
      },
    },
    '/api/workspaces/shop-local/status': {
      desired: {
        name: 'shop-local',
        provider: 'docker',
        capabilities: { inspect: true, logs: true, exec: true },
        resources: [
          {
            key: 'api',
            enabled: true,
            logicalHost: 'api',
            runtimeName: 'devarch-shop-local-api',
            templateName: 'node-api',
            dependsOn: ['postgres', 'redis'],
            spec: {
              image: 'node:22-alpine',
              env: { NODE_ENV: 'development' },
              ports: [{ container: 3000, published: 8200, protocol: 'tcp' }],
            },
          },
        ],
      },
      snapshot: {
        workspace: { name: 'shop-local', provider: 'docker' },
        resources: [
          {
            key: 'api',
            runtimeName: 'devarch-shop-local-api',
            state: { status: 'running', running: true },
            spec: {},
          },
        ],
      },
    },
    '/api/workspaces/shop-local/plan': {
      workspace: 'shop-local',
      provider: 'docker',
      actions: [
        {
          scope: 'resource',
          target: 'api',
          runtimeName: 'devarch-shop-local-api',
          kind: 'add',
          reasons: ['resource is not present in runtime snapshot'],
        },
      ],
      diagnostics: [
        {
          severity: 'error',
          code: 'secret-flatten',
          resource: 'api',
          message: 'export env DATABASE_URL cannot flatten secretRef into composite string',
        },
      ],
    },
  }

  vi.stubGlobal(
    'fetch',
    vi.fn(async (input: RequestInfo | URL) => {
      const url = typeof input === 'string' ? input : input instanceof URL ? input.toString() : input.url
      const parsed = new URL(url, 'http://localhost')
      const key = `${parsed.pathname}${parsed.search}`
      if (!(key in responses)) {
        throw new Error(`Unhandled fetch for ${key}`)
      }
      return jsonResponse(responses[key])
    }),
  )
}

describe('AppRoutes', () => {
  beforeEach(() => {
    localStorage.clear()
    installFetchMock()
  })

  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('renders the workspace plan view with constrained navigation', async () => {
    render(
      <MemoryRouter
        initialEntries={['/workspaces/shop-local?tab=plan']}
        future={{ v7_startTransition: true, v7_relativeSplatPath: true }}
      >
        <AppRoutes />
      </MemoryRouter>,
    )

    expect(screen.getByRole('link', { name: /Workspaces/i })).toBeInTheDocument()
    expect(screen.getByRole('link', { name: /Catalog/i })).toBeInTheDocument()
    expect(screen.getByRole('link', { name: /Activity/i })).toBeInTheDocument()
    expect(screen.getByRole('link', { name: /Settings/i })).toBeInTheDocument()
    expect(screen.queryByText('Services')).not.toBeInTheDocument()
    expect(screen.queryByText('Categories')).not.toBeInTheDocument()
    expect(screen.queryByText('Stacks')).not.toBeInTheDocument()

    expect(await screen.findByRole('heading', { name: 'Shop Local' })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /Apply workspace/i })).toBeInTheDocument()
    expect(screen.getByText(/resource is not present in runtime snapshot/i)).toBeInTheDocument()
    expect(screen.getByText(/secret-flatten/i)).toBeInTheDocument()
  })
})
