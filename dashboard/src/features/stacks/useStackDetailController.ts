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
} from '@/features/stacks/queries'
import { useInstances } from '@/features/instances/queries'
import { useGenerateStackProxyConfig } from '@/features/proxy/queries'

export function useStackDetailController(stackName: string) {
  const stackQuery = useStack(stackName)
  const instancesQuery = useInstances(stackName)
  const networkQuery = useStackNetwork(stackName)
  const composeQuery = useStackCompose(stackName)

  const enableStack = useEnableStack()
  const disableStack = useDisableStack()
  const stopStack = useStopStack()
  const startStack = useStartStack()
  const restartStack = useRestartStack()
  const generatePlan = useGeneratePlan(stackName)
  const applyPlan = useApplyPlan(stackName)
  const createNetwork = useCreateNetwork()
  const removeNetwork = useRemoveNetwork()
  const exportStack = useExportStack(stackName)
  const importStack = useImportStack()
  const generateProxyConfig = useGenerateStackProxyConfig(stackName)

  const connectedContainers = networkQuery.data?.containers ?? []
  const runningContainerNames = new Set(connectedContainers)

  return {
    stack: stackQuery.data,
    instances: instancesQuery.data ?? [],
    networkStatus: networkQuery.data,
    composeData: composeQuery.data,
    composeLoading: composeQuery.isLoading,
    isLoading: stackQuery.isLoading,
    connectedContainers,
    runningContainerNames,
    enableStack,
    disableStack,
    stopStack,
    startStack,
    restartStack,
    generatePlan,
    applyPlan,
    createNetwork,
    removeNetwork,
    exportStack,
    importStack,
    generateProxyConfig,
  }
}
