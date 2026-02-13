import { renderHook } from '@testing-library/react'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { useStackDetailController } from './useStackDetailController'
import { createWrapper } from '@/test/test-utils'

vi.mock('./queries')
vi.mock('@/features/instances/queries')
vi.mock('@/features/proxy/queries')

import {
  useStack,
  useStackNetwork,
  useStackCompose,
  useEnableStack,
  useDisableStack,
  useStopStack,
  useStartStack,
  useRestartStack,
  useGeneratePlan,
  useApplyPlan,
  useCreateNetwork,
  useRemoveNetwork,
  useExportStack,
  useImportStack,
} from './queries'
import { useInstances } from '@/features/instances/queries'
import { useGenerateStackProxyConfig } from '@/features/proxy/queries'

function mockQueryResult<T>(data: T | undefined, isLoading = false) {
  return { data, isLoading, error: null, isError: false }
}

function mockMutation() {
  return { mutate: vi.fn(), mutateAsync: vi.fn(), isPending: false, isError: false }
}

describe('useStackDetailController', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('data passthrough', () => {
    it('returns stack, instances, networkStatus, and composeData when available', () => {
      const mockStack = { name: 'test-stack', enabled: true }
      const mockInstances = [
        { id: 'inst-1', instance_name: 'web' },
        { id: 'inst-2', instance_name: 'db' },
      ]
      const mockNetwork = { name: 'test-stack_network', containers: ['web-1', 'db-1'] }
      const mockCompose = { version: '3.8', services: {} }

      vi.mocked(useStack).mockReturnValue(mockQueryResult(mockStack, false) as any)
      vi.mocked(useInstances).mockReturnValue(mockQueryResult(mockInstances, false) as any)
      vi.mocked(useStackNetwork).mockReturnValue(mockQueryResult(mockNetwork, false) as any)
      vi.mocked(useStackCompose).mockReturnValue(mockQueryResult(mockCompose, false) as any)

      // Mock all mutations
      vi.mocked(useEnableStack).mockReturnValue(mockMutation() as any)
      vi.mocked(useDisableStack).mockReturnValue(mockMutation() as any)
      vi.mocked(useStopStack).mockReturnValue(mockMutation() as any)
      vi.mocked(useStartStack).mockReturnValue(mockMutation() as any)
      vi.mocked(useRestartStack).mockReturnValue(mockMutation() as any)
      vi.mocked(useGeneratePlan).mockReturnValue(mockMutation() as any)
      vi.mocked(useApplyPlan).mockReturnValue(mockMutation() as any)
      vi.mocked(useCreateNetwork).mockReturnValue(mockMutation() as any)
      vi.mocked(useRemoveNetwork).mockReturnValue(mockMutation() as any)
      vi.mocked(useExportStack).mockReturnValue(mockMutation() as any)
      vi.mocked(useImportStack).mockReturnValue(mockMutation() as any)
      vi.mocked(useGenerateStackProxyConfig).mockReturnValue(mockMutation() as any)

      const { result } = renderHook(() => useStackDetailController('test-stack'), {
        wrapper: createWrapper(),
      })

      expect(result.current.stack).toBe(mockStack)
      expect(result.current.instances).toBe(mockInstances)
      expect(result.current.networkStatus).toBe(mockNetwork)
      expect(result.current.composeData).toBe(mockCompose)
      expect(result.current.composeLoading).toBe(false)
    })
  })

  describe('derived state: connectedContainers', () => {
    it('returns connectedContainers array from network data', () => {
      const mockNetwork = { containers: ['web-1', 'db-1'] }

      vi.mocked(useStack).mockReturnValue(mockQueryResult(undefined) as any)
      vi.mocked(useInstances).mockReturnValue(mockQueryResult(undefined) as any)
      vi.mocked(useStackNetwork).mockReturnValue(mockQueryResult(mockNetwork) as any)
      vi.mocked(useStackCompose).mockReturnValue(mockQueryResult(undefined) as any)

      vi.mocked(useEnableStack).mockReturnValue(mockMutation() as any)
      vi.mocked(useDisableStack).mockReturnValue(mockMutation() as any)
      vi.mocked(useStopStack).mockReturnValue(mockMutation() as any)
      vi.mocked(useStartStack).mockReturnValue(mockMutation() as any)
      vi.mocked(useRestartStack).mockReturnValue(mockMutation() as any)
      vi.mocked(useGeneratePlan).mockReturnValue(mockMutation() as any)
      vi.mocked(useApplyPlan).mockReturnValue(mockMutation() as any)
      vi.mocked(useCreateNetwork).mockReturnValue(mockMutation() as any)
      vi.mocked(useRemoveNetwork).mockReturnValue(mockMutation() as any)
      vi.mocked(useExportStack).mockReturnValue(mockMutation() as any)
      vi.mocked(useImportStack).mockReturnValue(mockMutation() as any)
      vi.mocked(useGenerateStackProxyConfig).mockReturnValue(mockMutation() as any)

      const { result } = renderHook(() => useStackDetailController('test-stack'), {
        wrapper: createWrapper(),
      })

      expect(result.current.connectedContainers).toEqual(['web-1', 'db-1'])
    })

    it('returns empty array when network containers is empty', () => {
      const mockNetwork = { containers: [] }

      vi.mocked(useStack).mockReturnValue(mockQueryResult(undefined) as any)
      vi.mocked(useInstances).mockReturnValue(mockQueryResult(undefined) as any)
      vi.mocked(useStackNetwork).mockReturnValue(mockQueryResult(mockNetwork) as any)
      vi.mocked(useStackCompose).mockReturnValue(mockQueryResult(undefined) as any)

      vi.mocked(useEnableStack).mockReturnValue(mockMutation() as any)
      vi.mocked(useDisableStack).mockReturnValue(mockMutation() as any)
      vi.mocked(useStopStack).mockReturnValue(mockMutation() as any)
      vi.mocked(useStartStack).mockReturnValue(mockMutation() as any)
      vi.mocked(useRestartStack).mockReturnValue(mockMutation() as any)
      vi.mocked(useGeneratePlan).mockReturnValue(mockMutation() as any)
      vi.mocked(useApplyPlan).mockReturnValue(mockMutation() as any)
      vi.mocked(useCreateNetwork).mockReturnValue(mockMutation() as any)
      vi.mocked(useRemoveNetwork).mockReturnValue(mockMutation() as any)
      vi.mocked(useExportStack).mockReturnValue(mockMutation() as any)
      vi.mocked(useImportStack).mockReturnValue(mockMutation() as any)
      vi.mocked(useGenerateStackProxyConfig).mockReturnValue(mockMutation() as any)

      const { result } = renderHook(() => useStackDetailController('test-stack'), {
        wrapper: createWrapper(),
      })

      expect(result.current.connectedContainers).toEqual([])
    })
  })

  describe('derived state: null network data', () => {
    it('returns empty array when network data is undefined', () => {
      vi.mocked(useStack).mockReturnValue(mockQueryResult(undefined) as any)
      vi.mocked(useInstances).mockReturnValue(mockQueryResult(undefined) as any)
      vi.mocked(useStackNetwork).mockReturnValue(mockQueryResult(undefined) as any)
      vi.mocked(useStackCompose).mockReturnValue(mockQueryResult(undefined) as any)

      vi.mocked(useEnableStack).mockReturnValue(mockMutation() as any)
      vi.mocked(useDisableStack).mockReturnValue(mockMutation() as any)
      vi.mocked(useStopStack).mockReturnValue(mockMutation() as any)
      vi.mocked(useStartStack).mockReturnValue(mockMutation() as any)
      vi.mocked(useRestartStack).mockReturnValue(mockMutation() as any)
      vi.mocked(useGeneratePlan).mockReturnValue(mockMutation() as any)
      vi.mocked(useApplyPlan).mockReturnValue(mockMutation() as any)
      vi.mocked(useCreateNetwork).mockReturnValue(mockMutation() as any)
      vi.mocked(useRemoveNetwork).mockReturnValue(mockMutation() as any)
      vi.mocked(useExportStack).mockReturnValue(mockMutation() as any)
      vi.mocked(useImportStack).mockReturnValue(mockMutation() as any)
      vi.mocked(useGenerateStackProxyConfig).mockReturnValue(mockMutation() as any)

      const { result } = renderHook(() => useStackDetailController('test-stack'), {
        wrapper: createWrapper(),
      })

      expect(result.current.connectedContainers).toEqual([])
    })
  })

  describe('loading states', () => {
    it('returns isLoading: true when stack query is loading', () => {
      vi.mocked(useStack).mockReturnValue(mockQueryResult(undefined, true) as any)
      vi.mocked(useInstances).mockReturnValue(mockQueryResult(undefined) as any)
      vi.mocked(useStackNetwork).mockReturnValue(mockQueryResult(undefined) as any)
      vi.mocked(useStackCompose).mockReturnValue(mockQueryResult(undefined) as any)

      vi.mocked(useEnableStack).mockReturnValue(mockMutation() as any)
      vi.mocked(useDisableStack).mockReturnValue(mockMutation() as any)
      vi.mocked(useStopStack).mockReturnValue(mockMutation() as any)
      vi.mocked(useStartStack).mockReturnValue(mockMutation() as any)
      vi.mocked(useRestartStack).mockReturnValue(mockMutation() as any)
      vi.mocked(useGeneratePlan).mockReturnValue(mockMutation() as any)
      vi.mocked(useApplyPlan).mockReturnValue(mockMutation() as any)
      vi.mocked(useCreateNetwork).mockReturnValue(mockMutation() as any)
      vi.mocked(useRemoveNetwork).mockReturnValue(mockMutation() as any)
      vi.mocked(useExportStack).mockReturnValue(mockMutation() as any)
      vi.mocked(useImportStack).mockReturnValue(mockMutation() as any)
      vi.mocked(useGenerateStackProxyConfig).mockReturnValue(mockMutation() as any)

      const { result } = renderHook(() => useStackDetailController('test-stack'), {
        wrapper: createWrapper(),
      })

      expect(result.current.isLoading).toBe(true)
    })

    it('returns isLoading: false when stack query completes', () => {
      vi.mocked(useStack).mockReturnValue(mockQueryResult({ name: 'test-stack' }, false) as any)
      vi.mocked(useInstances).mockReturnValue(mockQueryResult(undefined) as any)
      vi.mocked(useStackNetwork).mockReturnValue(mockQueryResult(undefined) as any)
      vi.mocked(useStackCompose).mockReturnValue(mockQueryResult(undefined) as any)

      vi.mocked(useEnableStack).mockReturnValue(mockMutation() as any)
      vi.mocked(useDisableStack).mockReturnValue(mockMutation() as any)
      vi.mocked(useStopStack).mockReturnValue(mockMutation() as any)
      vi.mocked(useStartStack).mockReturnValue(mockMutation() as any)
      vi.mocked(useRestartStack).mockReturnValue(mockMutation() as any)
      vi.mocked(useGeneratePlan).mockReturnValue(mockMutation() as any)
      vi.mocked(useApplyPlan).mockReturnValue(mockMutation() as any)
      vi.mocked(useCreateNetwork).mockReturnValue(mockMutation() as any)
      vi.mocked(useRemoveNetwork).mockReturnValue(mockMutation() as any)
      vi.mocked(useExportStack).mockReturnValue(mockMutation() as any)
      vi.mocked(useImportStack).mockReturnValue(mockMutation() as any)
      vi.mocked(useGenerateStackProxyConfig).mockReturnValue(mockMutation() as any)

      const { result } = renderHook(() => useStackDetailController('test-stack'), {
        wrapper: createWrapper(),
      })

      expect(result.current.isLoading).toBe(false)
    })
  })

  describe('mutation exposure', () => {
    it('exposes all 12 mutation objects with mutate functions', () => {
      vi.mocked(useStack).mockReturnValue(mockQueryResult(undefined) as any)
      vi.mocked(useInstances).mockReturnValue(mockQueryResult(undefined) as any)
      vi.mocked(useStackNetwork).mockReturnValue(mockQueryResult(undefined) as any)
      vi.mocked(useStackCompose).mockReturnValue(mockQueryResult(undefined) as any)

      vi.mocked(useEnableStack).mockReturnValue(mockMutation() as any)
      vi.mocked(useDisableStack).mockReturnValue(mockMutation() as any)
      vi.mocked(useStopStack).mockReturnValue(mockMutation() as any)
      vi.mocked(useStartStack).mockReturnValue(mockMutation() as any)
      vi.mocked(useRestartStack).mockReturnValue(mockMutation() as any)
      vi.mocked(useGeneratePlan).mockReturnValue(mockMutation() as any)
      vi.mocked(useApplyPlan).mockReturnValue(mockMutation() as any)
      vi.mocked(useCreateNetwork).mockReturnValue(mockMutation() as any)
      vi.mocked(useRemoveNetwork).mockReturnValue(mockMutation() as any)
      vi.mocked(useExportStack).mockReturnValue(mockMutation() as any)
      vi.mocked(useImportStack).mockReturnValue(mockMutation() as any)
      vi.mocked(useGenerateStackProxyConfig).mockReturnValue(mockMutation() as any)

      const { result } = renderHook(() => useStackDetailController('test-stack'), {
        wrapper: createWrapper(),
      })

      // Verify all 12 mutations exist with mutate functions
      expect(typeof result.current.enableStack.mutate).toBe('function')
      expect(typeof result.current.disableStack.mutate).toBe('function')
      expect(typeof result.current.stopStack.mutate).toBe('function')
      expect(typeof result.current.startStack.mutate).toBe('function')
      expect(typeof result.current.restartStack.mutate).toBe('function')
      expect(typeof result.current.generatePlan.mutate).toBe('function')
      expect(typeof result.current.applyPlan.mutate).toBe('function')
      expect(typeof result.current.createNetwork.mutate).toBe('function')
      expect(typeof result.current.removeNetwork.mutate).toBe('function')
      expect(typeof result.current.exportStack.mutate).toBe('function')
      expect(typeof result.current.importStack.mutate).toBe('function')
      expect(typeof result.current.generateProxyConfig.mutate).toBe('function')
    })
  })

  describe('instances default', () => {
    it('returns empty array when instances data is undefined', () => {
      vi.mocked(useStack).mockReturnValue(mockQueryResult(undefined) as any)
      vi.mocked(useInstances).mockReturnValue(mockQueryResult(undefined) as any)
      vi.mocked(useStackNetwork).mockReturnValue(mockQueryResult(undefined) as any)
      vi.mocked(useStackCompose).mockReturnValue(mockQueryResult(undefined) as any)

      vi.mocked(useEnableStack).mockReturnValue(mockMutation() as any)
      vi.mocked(useDisableStack).mockReturnValue(mockMutation() as any)
      vi.mocked(useStopStack).mockReturnValue(mockMutation() as any)
      vi.mocked(useStartStack).mockReturnValue(mockMutation() as any)
      vi.mocked(useRestartStack).mockReturnValue(mockMutation() as any)
      vi.mocked(useGeneratePlan).mockReturnValue(mockMutation() as any)
      vi.mocked(useApplyPlan).mockReturnValue(mockMutation() as any)
      vi.mocked(useCreateNetwork).mockReturnValue(mockMutation() as any)
      vi.mocked(useRemoveNetwork).mockReturnValue(mockMutation() as any)
      vi.mocked(useExportStack).mockReturnValue(mockMutation() as any)
      vi.mocked(useImportStack).mockReturnValue(mockMutation() as any)
      vi.mocked(useGenerateStackProxyConfig).mockReturnValue(mockMutation() as any)

      const { result } = renderHook(() => useStackDetailController('test-stack'), {
        wrapper: createWrapper(),
      })

      expect(result.current.instances).toEqual([])
    })
  })
})
