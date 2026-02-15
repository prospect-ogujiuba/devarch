import { useInstance, useUpdateInstance, useStopInstance, useStartInstance, useRestartInstance, useInstanceCompose } from './queries'
import { useService } from '@/features/services/queries'
import { useGenerateInstanceProxyConfig } from '@/features/proxy/queries'

export function useInstanceDetailController(stackName: string, instanceId: string) {
  const { data: instance, isLoading } = useInstance(stackName, instanceId)
  const { data: templateService } = useService(instance?.template_name ?? '')
  const { data: composeYaml } = useInstanceCompose(stackName, instanceId)
  const updateInstance = useUpdateInstance(stackName, instanceId)
  const stopInstance = useStopInstance(stackName, instanceId)
  const startInstance = useStartInstance(stackName, instanceId)
  const restartInstance = useRestartInstance(stackName, instanceId)
  const generateProxyConfig = useGenerateInstanceProxyConfig(stackName, instanceId)

  return {
    instance,
    templateService,
    composeYaml,
    isLoading,
    updateInstance,
    stopInstance,
    startInstance,
    restartInstance,
    generateProxyConfig,
  }
}
