import { useRef } from 'react'
import { useMutation, useQueryClient, type UseMutationResult } from '@tanstack/react-query'
import { toast } from 'sonner'
import { getErrorMessage } from '@/lib/api'

interface MutationHelperConfig<TData, TVariables> {
  mutationKey?: string[]
  mutationFn: (vars: TVariables) => Promise<TData>
  loadingMessage?: string | ((vars: TVariables) => string)
  successMessage?: string | ((vars: TVariables, data: TData) => string)
  errorMessage?: string | ((error: unknown, vars: TVariables) => string)
  invalidate?: string[][]
  onSuccess?: (data: TData, vars: TVariables) => void
}

export function useMutationHelper<TData = unknown, TVariables = void>(
  config: MutationHelperConfig<TData, TVariables>
): UseMutationResult<TData, unknown, TVariables> {
  const queryClient = useQueryClient()
  const toastIdRef = useRef<string | number | undefined>(undefined)

  return useMutation({
    mutationKey: config.mutationKey,
    mutationFn: config.mutationFn,
    onMutate: (vars) => {
      if (config.loadingMessage) {
        const message = typeof config.loadingMessage === 'function'
          ? config.loadingMessage(vars)
          : config.loadingMessage
        toastIdRef.current = toast.loading(message)
      }
    },
    onSuccess: (data, vars) => {
      if (config.successMessage) {
        const message = typeof config.successMessage === 'function'
          ? config.successMessage(vars, data)
          : config.successMessage
        toast.success(message, { id: toastIdRef.current })
      } else if (toastIdRef.current) {
        toast.dismiss(toastIdRef.current)
      }
      toastIdRef.current = undefined

      if (config.invalidate) {
        for (const queryKey of config.invalidate) {
          queryClient.invalidateQueries({ queryKey })
        }
      }

      if (config.onSuccess) {
        config.onSuccess(data, vars)
      }
    },
    onError: (error, vars) => {
      const message = config.errorMessage
        ? typeof config.errorMessage === 'function'
          ? config.errorMessage(error, vars)
          : config.errorMessage
        : getErrorMessage(error, 'Operation failed')
      toast.error(message, { id: toastIdRef.current })
      toastIdRef.current = undefined
    },
  })
}
