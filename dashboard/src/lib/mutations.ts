import { useMutation, useQueryClient, type UseMutationResult } from '@tanstack/react-query'
import { toast } from 'sonner'
import { getErrorMessage } from '@/lib/api'

interface MutationHelperConfig<TData, TVariables> {
  mutationFn: (vars: TVariables) => Promise<TData>
  successMessage?: string | ((vars: TVariables, data: TData) => string)
  errorMessage?: string | ((error: unknown, vars: TVariables) => string)
  invalidate?: string[][]
  onSuccess?: (data: TData, vars: TVariables) => void
}

export function useMutationHelper<TData = unknown, TVariables = void>(
  config: MutationHelperConfig<TData, TVariables>
): UseMutationResult<TData, unknown, TVariables> {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: config.mutationFn,
    onSuccess: (data, vars) => {
      if (config.successMessage) {
        const message = typeof config.successMessage === 'function'
          ? config.successMessage(vars, data)
          : config.successMessage
        toast.success(message)
      }

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
      toast.error(message)
    },
  })
}
