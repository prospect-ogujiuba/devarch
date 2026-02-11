import { useInstance, useUpdateInstance } from './queries'
import { useService } from '@/features/services/queries'

export function useInstanceDetailController(stackName: string, instanceId: string) {
  const { data: instance, isLoading } = useInstance(stackName, instanceId)
  const { data: templateService } = useService(instance?.template_name ?? '')
  const updateInstance = useUpdateInstance(stackName, instanceId)

  return {
    instance,
    templateService,
    isLoading,
    updateInstance,
  }
}
