import { useState } from 'react'
import { Loader2 } from 'lucide-react'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Button } from '@/components/ui/button'
import { useCreateWire } from '@/features/stacks/queries'

interface UnresolvedContract {
  instance: string
  contract_name: string
  contract_type: string
  required: boolean
  reason: string
  available_providers?: { instance_id: string; instance_name: string }[]
}

interface CreateWireDialogProps {
  stackName: string
  open: boolean
  onOpenChange: (open: boolean) => void
  unresolved: UnresolvedContract[]
}

export function CreateWireDialog({
  stackName,
  open,
  onOpenChange,
  unresolved,
}: CreateWireDialogProps) {
  const createWire = useCreateWire(stackName)
  const [selectedContract, setSelectedContract] = useState<string>('')
  const [selectedProvider, setSelectedProvider] = useState<string>('')

  const unresolvedWithProviders = unresolved.filter(
    (u) => u.available_providers && u.available_providers.length > 0
  )

  const selectedUnresolved = unresolvedWithProviders.find(
    (u) => `${u.instance}:${u.contract_name}` === selectedContract
  )

  const handleSubmit = () => {
    if (!selectedUnresolved || !selectedProvider) return

    createWire.mutate(
      {
        consumer_instance_id: selectedUnresolved.instance,
        provider_instance_id: selectedProvider,
        import_contract_name: selectedUnresolved.contract_name,
      },
      {
        onSuccess: () => {
          setSelectedContract('')
          setSelectedProvider('')
          onOpenChange(false)
        },
      }
    )
  }

  const handleOpenChange = (newOpen: boolean) => {
    if (!newOpen) {
      setSelectedContract('')
      setSelectedProvider('')
    }
    onOpenChange(newOpen)
  }

  const handleContractChange = (value: string) => {
    setSelectedContract(value)
    setSelectedProvider('')
  }

  const canSubmit = selectedContract && selectedProvider && !createWire.isPending

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Create Wire</DialogTitle>
          <DialogDescription>
            Create an explicit wire between a consumer and provider
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4 py-4">
          <div className="space-y-2">
            <label htmlFor="contract" className="text-sm font-medium">Consumer & Contract</label>
            <Select value={selectedContract} onValueChange={handleContractChange}>
              <SelectTrigger id="contract">
                <SelectValue placeholder="Select import contract" />
              </SelectTrigger>
              <SelectContent>
                {unresolvedWithProviders.length === 0 ? (
                  <div className="px-2 py-3 text-sm text-muted-foreground text-center">
                    No unresolved contracts with available providers
                  </div>
                ) : (
                  unresolvedWithProviders.map((u) => (
                    <SelectItem
                      key={`${u.instance}:${u.contract_name}`}
                      value={`${u.instance}:${u.contract_name}`}
                    >
                      <div className="flex flex-col">
                        <span className="font-medium">
                          {u.instance} → {u.contract_name}
                        </span>
                        <span className="text-xs text-muted-foreground">
                          {u.contract_type}
                          {u.available_providers && ` • ${u.available_providers.length} provider${u.available_providers.length === 1 ? '' : 's'}`}
                        </span>
                      </div>
                    </SelectItem>
                  ))
                )}
              </SelectContent>
            </Select>
          </div>

          {selectedUnresolved && (
            <div className="space-y-2">
              <label htmlFor="provider" className="text-sm font-medium">Provider Instance</label>
              <Select value={selectedProvider} onValueChange={setSelectedProvider}>
                <SelectTrigger id="provider">
                  <SelectValue placeholder="Select provider" />
                </SelectTrigger>
                <SelectContent>
                  {selectedUnresolved.available_providers?.map((provider) => (
                    <SelectItem
                      key={provider.instance_id}
                      value={provider.instance_id}
                    >
                      {provider.instance_name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          )}
        </div>

        <DialogFooter>
          <Button
            type="button"
            variant="outline"
            onClick={() => handleOpenChange(false)}
            disabled={createWire.isPending}
          >
            Cancel
          </Button>
          <Button
            type="submit"
            onClick={handleSubmit}
            disabled={!canSubmit}
          >
            {createWire.isPending && <Loader2 className="size-4 animate-spin" />}
            Create Wire
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
