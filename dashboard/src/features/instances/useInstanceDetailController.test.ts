import { renderHook } from '@testing-library/react'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { useInstanceDetailController } from './useInstanceDetailController'
import { createWrapper } from '@/test/test-utils'

vi.mock('./queries')
vi.mock('@/features/services/queries')

import { useInstance, useUpdateInstance } from './queries'
import { useService } from '@/features/services/queries'

describe('useInstanceDetailController', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('data passthrough', () => {
    it('returns instance, templateService, and updateInstance when data is available', () => {
      const mockInstance = {
        id: 'inst-1',
        stack_name: 'test-stack',
        template_name: 'nginx',
        instance_name: 'web',
      }
      const mockService = {
        name: 'nginx',
        display_name: 'Nginx',
      }
      const mockUpdateMutation = {
        mutate: vi.fn(),
        isPending: false,
      }

      vi.mocked(useInstance).mockReturnValue({
        data: mockInstance,
        isLoading: false,
        error: null,
      } as any)

      vi.mocked(useService).mockReturnValue({
        data: mockService,
      } as any)

      vi.mocked(useUpdateInstance).mockReturnValue(mockUpdateMutation as any)

      const { result } = renderHook(() => useInstanceDetailController('test-stack', 'inst-1'), {
        wrapper: createWrapper(),
      })

      expect(result.current.instance).toBe(mockInstance)
      expect(result.current.templateService).toBe(mockService)
      expect(result.current.updateInstance).toBe(mockUpdateMutation)
    })
  })

  describe('loading states', () => {
    it('returns isLoading: true when instance query is loading', () => {
      vi.mocked(useInstance).mockReturnValue({
        data: undefined,
        isLoading: true,
        error: null,
      } as any)

      vi.mocked(useService).mockReturnValue({
        data: undefined,
      } as any)

      vi.mocked(useUpdateInstance).mockReturnValue({
        mutate: vi.fn(),
        isPending: false,
      } as any)

      const { result } = renderHook(() => useInstanceDetailController('test-stack', 'inst-1'), {
        wrapper: createWrapper(),
      })

      expect(result.current.isLoading).toBe(true)
    })

    it('returns isLoading: false when instance query completes', () => {
      vi.mocked(useInstance).mockReturnValue({
        data: { id: 'inst-1', template_name: 'nginx' },
        isLoading: false,
        error: null,
      } as any)

      vi.mocked(useService).mockReturnValue({
        data: undefined,
      } as any)

      vi.mocked(useUpdateInstance).mockReturnValue({
        mutate: vi.fn(),
        isPending: false,
      } as any)

      const { result } = renderHook(() => useInstanceDetailController('test-stack', 'inst-1'), {
        wrapper: createWrapper(),
      })

      expect(result.current.isLoading).toBe(false)
    })
  })

  describe('conditional query (template_name)', () => {
    it('calls useService with empty string when instance data is undefined', () => {
      vi.mocked(useInstance).mockReturnValue({
        data: undefined,
        isLoading: false,
        error: null,
      } as any)

      vi.mocked(useService).mockReturnValue({
        data: undefined,
      } as any)

      vi.mocked(useUpdateInstance).mockReturnValue({
        mutate: vi.fn(),
        isPending: false,
      } as any)

      renderHook(() => useInstanceDetailController('test-stack', 'inst-1'), {
        wrapper: createWrapper(),
      })

      expect(vi.mocked(useService)).toHaveBeenCalledWith('')
    })

    it('calls useService with template_name when instance has template_name', () => {
      vi.mocked(useInstance).mockReturnValue({
        data: {
          id: 'inst-1',
          template_name: 'nginx',
        },
        isLoading: false,
        error: null,
      } as any)

      vi.mocked(useService).mockReturnValue({
        data: undefined,
      } as any)

      vi.mocked(useUpdateInstance).mockReturnValue({
        mutate: vi.fn(),
        isPending: false,
      } as any)

      renderHook(() => useInstanceDetailController('test-stack', 'inst-1'), {
        wrapper: createWrapper(),
      })

      expect(vi.mocked(useService)).toHaveBeenCalledWith('nginx')
    })
  })

  describe('mutation exposure', () => {
    it('exposes updateInstance with mutate function', () => {
      vi.mocked(useInstance).mockReturnValue({
        data: undefined,
        isLoading: false,
        error: null,
      } as any)

      vi.mocked(useService).mockReturnValue({
        data: undefined,
      } as any)

      const mockMutate = vi.fn()
      vi.mocked(useUpdateInstance).mockReturnValue({
        mutate: mockMutate,
        isPending: false,
      } as any)

      const { result } = renderHook(() => useInstanceDetailController('test-stack', 'inst-1'), {
        wrapper: createWrapper(),
      })

      expect(typeof result.current.updateInstance.mutate).toBe('function')
    })
  })

  describe('edge case: undefined data', () => {
    it('returns undefined for instance and templateService when all queries return undefined', () => {
      vi.mocked(useInstance).mockReturnValue({
        data: undefined,
        isLoading: false,
        error: null,
      } as any)

      vi.mocked(useService).mockReturnValue({
        data: undefined,
      } as any)

      vi.mocked(useUpdateInstance).mockReturnValue({
        mutate: vi.fn(),
        isPending: false,
      } as any)

      const { result } = renderHook(() => useInstanceDetailController('test-stack', 'inst-1'), {
        wrapper: createWrapper(),
      })

      expect(result.current.instance).toBeUndefined()
      expect(result.current.templateService).toBeUndefined()
    })
  })
})
