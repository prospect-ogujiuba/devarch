import { renderHook } from '@testing-library/react'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { useServiceDetailController } from './useServiceDetailController'
import { createWrapper } from '@/test/test-utils'
import type { Service } from '@/types/api'

vi.mock('./queries')
vi.mock('@/features/proxy/queries')
vi.mock('@/lib/format', () => ({ computeUptime: vi.fn() }))

import { useService, useServiceCompose, useDeleteService, useUpdateService } from './queries'
import { useGenerateServiceProxyConfig } from '@/features/proxy/queries'
import { computeUptime } from '@/lib/format'

function mockService(overrides?: Partial<Service>): Service {
  return {
    id: 1,
    name: 'nginx',
    category_id: 1,
    image_name: 'nginx',
    image_tag: 'latest',
    restart_policy: 'unless-stopped',
    enabled: true,
    created_at: '2026-01-01T00:00:00Z',
    updated_at: '2026-01-01T00:00:00Z',
    status: { status: 'running', restart_count: 0, started_at: '2026-01-01T00:00:00Z' },
    metrics: {
      cpu_percentage: 45.2,
      memory_used_mb: 128,
      memory_limit_mb: 512,
      memory_percentage: 25,
      network_rx_bytes: 1024,
      network_tx_bytes: 2048,
    },
    ...overrides,
  }
}

describe('useServiceDetailController', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('data passthrough', () => {
    it('returns service, composeYaml, and loading states from underlying queries', () => {
      const mockSvc = mockService()
      const mockCompose = { yaml: 'version: "3.8"', warnings: [], instance_count: 1 }

      vi.mocked(useService).mockReturnValue({
        data: mockSvc,
        isLoading: false,
        error: null,
      } as any)

      vi.mocked(useServiceCompose).mockReturnValue({
        data: mockCompose,
        isLoading: false,
      } as any)

      vi.mocked(useDeleteService).mockReturnValue({ mutate: vi.fn() } as any)
      vi.mocked(useUpdateService).mockReturnValue({ mutate: vi.fn() } as any)
      vi.mocked(useGenerateServiceProxyConfig).mockReturnValue({ mutate: vi.fn() } as any)

      const { result } = renderHook(() => useServiceDetailController('nginx'), {
        wrapper: createWrapper(),
      })

      expect(result.current.service).toBe(mockSvc)
      expect(result.current.composeYaml).toBe(mockCompose)
      expect(result.current.isLoading).toBe(false)
      expect(result.current.composeLoading).toBe(false)
    })
  })

  describe('derived state: status', () => {
    it('returns status from service.status.status when present', () => {
      const mockSvc = mockService({ status: { status: 'running', restart_count: 0 } })

      vi.mocked(useService).mockReturnValue({ data: mockSvc, isLoading: false } as any)
      vi.mocked(useServiceCompose).mockReturnValue({ data: undefined, isLoading: false } as any)
      vi.mocked(useDeleteService).mockReturnValue({ mutate: vi.fn() } as any)
      vi.mocked(useUpdateService).mockReturnValue({ mutate: vi.fn() } as any)
      vi.mocked(useGenerateServiceProxyConfig).mockReturnValue({ mutate: vi.fn() } as any)

      const { result } = renderHook(() => useServiceDetailController('nginx'), {
        wrapper: createWrapper(),
      })

      expect(result.current.status).toBe('running')
    })

    it('falls back to "stopped" when status object is undefined', () => {
      const mockSvc = mockService({ status: undefined })

      vi.mocked(useService).mockReturnValue({ data: mockSvc, isLoading: false } as any)
      vi.mocked(useServiceCompose).mockReturnValue({ data: undefined, isLoading: false } as any)
      vi.mocked(useDeleteService).mockReturnValue({ mutate: vi.fn() } as any)
      vi.mocked(useUpdateService).mockReturnValue({ mutate: vi.fn() } as any)
      vi.mocked(useGenerateServiceProxyConfig).mockReturnValue({ mutate: vi.fn() } as any)

      const { result } = renderHook(() => useServiceDetailController('nginx'), {
        wrapper: createWrapper(),
      })

      expect(result.current.status).toBe('stopped')
    })
  })

  describe('derived state: image', () => {
    it('returns formatted image string when service is present', () => {
      const mockSvc = mockService({ image_name: 'nginx', image_tag: 'alpine' })

      vi.mocked(useService).mockReturnValue({ data: mockSvc, isLoading: false } as any)
      vi.mocked(useServiceCompose).mockReturnValue({ data: undefined, isLoading: false } as any)
      vi.mocked(useDeleteService).mockReturnValue({ mutate: vi.fn() } as any)
      vi.mocked(useUpdateService).mockReturnValue({ mutate: vi.fn() } as any)
      vi.mocked(useGenerateServiceProxyConfig).mockReturnValue({ mutate: vi.fn() } as any)

      const { result } = renderHook(() => useServiceDetailController('nginx'), {
        wrapper: createWrapper(),
      })

      expect(result.current.image).toBe('nginx:alpine')
    })

    it('returns empty string when service is undefined', () => {
      vi.mocked(useService).mockReturnValue({ data: undefined, isLoading: false } as any)
      vi.mocked(useServiceCompose).mockReturnValue({ data: undefined, isLoading: false } as any)
      vi.mocked(useDeleteService).mockReturnValue({ mutate: vi.fn() } as any)
      vi.mocked(useUpdateService).mockReturnValue({ mutate: vi.fn() } as any)
      vi.mocked(useGenerateServiceProxyConfig).mockReturnValue({ mutate: vi.fn() } as any)

      const { result } = renderHook(() => useServiceDetailController('nginx'), {
        wrapper: createWrapper(),
      })

      expect(result.current.image).toBe('')
    })
  })

  describe('derived state: healthStatus', () => {
    it('returns health_status when present in service.status', () => {
      const mockSvc = mockService({
        status: { status: 'running', restart_count: 0, health_status: 'healthy' },
      })

      vi.mocked(useService).mockReturnValue({ data: mockSvc, isLoading: false } as any)
      vi.mocked(useServiceCompose).mockReturnValue({ data: undefined, isLoading: false } as any)
      vi.mocked(useDeleteService).mockReturnValue({ mutate: vi.fn() } as any)
      vi.mocked(useUpdateService).mockReturnValue({ mutate: vi.fn() } as any)
      vi.mocked(useGenerateServiceProxyConfig).mockReturnValue({ mutate: vi.fn() } as any)

      const { result } = renderHook(() => useServiceDetailController('nginx'), {
        wrapper: createWrapper(),
      })

      expect(result.current.healthStatus).toBe('healthy')
    })

    it('returns "configured" when no health_status but healthcheck exists', () => {
      const mockSvc = mockService({
        status: { status: 'running', restart_count: 0 },
        healthcheck: {
          id: 1,
          service_id: 1,
          test: 'curl -f http://localhost || exit 1',
          interval_seconds: 30,
          timeout_seconds: 10,
          retries: 3,
          start_period_seconds: 0,
        },
      })

      vi.mocked(useService).mockReturnValue({ data: mockSvc, isLoading: false } as any)
      vi.mocked(useServiceCompose).mockReturnValue({ data: undefined, isLoading: false } as any)
      vi.mocked(useDeleteService).mockReturnValue({ mutate: vi.fn() } as any)
      vi.mocked(useUpdateService).mockReturnValue({ mutate: vi.fn() } as any)
      vi.mocked(useGenerateServiceProxyConfig).mockReturnValue({ mutate: vi.fn() } as any)

      const { result } = renderHook(() => useServiceDetailController('nginx'), {
        wrapper: createWrapper(),
      })

      expect(result.current.healthStatus).toBe('configured')
    })

    it('returns "none" when no health_status and no healthcheck', () => {
      const mockSvc = mockService({
        status: { status: 'running', restart_count: 0 },
        healthcheck: undefined,
      })

      vi.mocked(useService).mockReturnValue({ data: mockSvc, isLoading: false } as any)
      vi.mocked(useServiceCompose).mockReturnValue({ data: undefined, isLoading: false } as any)
      vi.mocked(useDeleteService).mockReturnValue({ mutate: vi.fn() } as any)
      vi.mocked(useUpdateService).mockReturnValue({ mutate: vi.fn() } as any)
      vi.mocked(useGenerateServiceProxyConfig).mockReturnValue({ mutate: vi.fn() } as any)

      const { result } = renderHook(() => useServiceDetailController('nginx'), {
        wrapper: createWrapper(),
      })

      expect(result.current.healthStatus).toBe('none')
    })
  })

  describe('derived state: uptime', () => {
    it('calls computeUptime with started_at and returns result', () => {
      const mockSvc = mockService({
        status: { status: 'running', restart_count: 0, started_at: '2026-01-01T00:00:00Z' },
      })

      vi.mocked(computeUptime).mockReturnValue(3600)
      vi.mocked(useService).mockReturnValue({ data: mockSvc, isLoading: false } as any)
      vi.mocked(useServiceCompose).mockReturnValue({ data: undefined, isLoading: false } as any)
      vi.mocked(useDeleteService).mockReturnValue({ mutate: vi.fn() } as any)
      vi.mocked(useUpdateService).mockReturnValue({ mutate: vi.fn() } as any)
      vi.mocked(useGenerateServiceProxyConfig).mockReturnValue({ mutate: vi.fn() } as any)

      const { result } = renderHook(() => useServiceDetailController('nginx'), {
        wrapper: createWrapper(),
      })

      expect(result.current.uptime).toBe(3600)
      expect(computeUptime).toHaveBeenCalledWith('2026-01-01T00:00:00Z')
    })
  })

  describe('derived state: metrics', () => {
    it('returns all metric values when metrics are present', () => {
      const mockSvc = mockService({
        metrics: {
          cpu_percentage: 45.2,
          memory_used_mb: 128,
          memory_limit_mb: 512,
          memory_percentage: 25,
          network_rx_bytes: 1024,
          network_tx_bytes: 2048,
        },
      })

      vi.mocked(useService).mockReturnValue({ data: mockSvc, isLoading: false } as any)
      vi.mocked(useServiceCompose).mockReturnValue({ data: undefined, isLoading: false } as any)
      vi.mocked(useDeleteService).mockReturnValue({ mutate: vi.fn() } as any)
      vi.mocked(useUpdateService).mockReturnValue({ mutate: vi.fn() } as any)
      vi.mocked(useGenerateServiceProxyConfig).mockReturnValue({ mutate: vi.fn() } as any)

      const { result } = renderHook(() => useServiceDetailController('nginx'), {
        wrapper: createWrapper(),
      })

      expect(result.current.cpuPct).toBe(45.2)
      expect(result.current.memUsed).toBe(128)
      expect(result.current.memLimit).toBe(512)
      expect(result.current.rxBytes).toBe(1024)
      expect(result.current.txBytes).toBe(2048)
    })
  })

  describe('metrics defaults when undefined', () => {
    it('returns 0 for all metric values when metrics is undefined', () => {
      const mockSvc = mockService({ metrics: undefined })

      vi.mocked(useService).mockReturnValue({ data: mockSvc, isLoading: false } as any)
      vi.mocked(useServiceCompose).mockReturnValue({ data: undefined, isLoading: false } as any)
      vi.mocked(useDeleteService).mockReturnValue({ mutate: vi.fn() } as any)
      vi.mocked(useUpdateService).mockReturnValue({ mutate: vi.fn() } as any)
      vi.mocked(useGenerateServiceProxyConfig).mockReturnValue({ mutate: vi.fn() } as any)

      const { result } = renderHook(() => useServiceDetailController('nginx'), {
        wrapper: createWrapper(),
      })

      expect(result.current.cpuPct).toBe(0)
      expect(result.current.memUsed).toBe(0)
      expect(result.current.memLimit).toBe(0)
      expect(result.current.rxBytes).toBe(0)
      expect(result.current.txBytes).toBe(0)
    })
  })

  describe('loading states', () => {
    it('returns isLoading true when useService is loading', () => {
      vi.mocked(useService).mockReturnValue({ data: undefined, isLoading: true } as any)
      vi.mocked(useServiceCompose).mockReturnValue({ data: undefined, isLoading: false } as any)
      vi.mocked(useDeleteService).mockReturnValue({ mutate: vi.fn() } as any)
      vi.mocked(useUpdateService).mockReturnValue({ mutate: vi.fn() } as any)
      vi.mocked(useGenerateServiceProxyConfig).mockReturnValue({ mutate: vi.fn() } as any)

      const { result } = renderHook(() => useServiceDetailController('nginx'), {
        wrapper: createWrapper(),
      })

      expect(result.current.isLoading).toBe(true)
    })

    it('returns composeLoading true when useServiceCompose is loading', () => {
      vi.mocked(useService).mockReturnValue({ data: undefined, isLoading: false } as any)
      vi.mocked(useServiceCompose).mockReturnValue({ data: undefined, isLoading: true } as any)
      vi.mocked(useDeleteService).mockReturnValue({ mutate: vi.fn() } as any)
      vi.mocked(useUpdateService).mockReturnValue({ mutate: vi.fn() } as any)
      vi.mocked(useGenerateServiceProxyConfig).mockReturnValue({ mutate: vi.fn() } as any)

      const { result } = renderHook(() => useServiceDetailController('nginx'), {
        wrapper: createWrapper(),
      })

      expect(result.current.composeLoading).toBe(true)
    })
  })

  describe('mutation exposure', () => {
    it('exposes deleteService, updateService, and generateProxyConfig with mutate functions', () => {
      const mockDeleteMutation = { mutate: vi.fn() }
      const mockUpdateMutation = { mutate: vi.fn() }
      const mockGenerateMutation = { mutate: vi.fn() }

      vi.mocked(useService).mockReturnValue({ data: undefined, isLoading: false } as any)
      vi.mocked(useServiceCompose).mockReturnValue({ data: undefined, isLoading: false } as any)
      vi.mocked(useDeleteService).mockReturnValue(mockDeleteMutation as any)
      vi.mocked(useUpdateService).mockReturnValue(mockUpdateMutation as any)
      vi.mocked(useGenerateServiceProxyConfig).mockReturnValue(mockGenerateMutation as any)

      const { result } = renderHook(() => useServiceDetailController('nginx'), {
        wrapper: createWrapper(),
      })

      expect(result.current.deleteService).toBe(mockDeleteMutation)
      expect(result.current.updateService).toBe(mockUpdateMutation)
      expect(result.current.generateProxyConfig).toBe(mockGenerateMutation)
      expect(result.current.deleteService.mutate).toBeInstanceOf(Function)
      expect(result.current.updateService.mutate).toBeInstanceOf(Function)
      expect(result.current.generateProxyConfig.mutate).toBeInstanceOf(Function)
    })
  })
})
