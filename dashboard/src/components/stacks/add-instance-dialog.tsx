import { useState, useMemo } from 'react'
import { Loader2, Search, Package } from 'lucide-react'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { Badge } from '@/components/ui/badge'
import { useServices } from '@/features/services/queries'
import { useCreateInstance, useInstances } from '@/features/instances/queries'
import { cn } from '@/lib/utils'

interface AddInstanceDialogProps {
  stackName: string
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function AddInstanceDialog({ stackName, open, onOpenChange }: AddInstanceDialogProps) {
  const { data: servicesResult } = useServices()
  const { data: instances = [] } = useInstances(stackName)
  const createInstance = useCreateInstance(stackName)

  const [searchQuery, setSearchQuery] = useState('')
  const [selectedServiceId, setSelectedServiceId] = useState<number | null>(null)
  const [instanceName, setInstanceName] = useState('')
  const [description, setDescription] = useState('')
  const [nameError, setNameError] = useState('')

  const services = servicesResult?.services ?? []

  const filteredServices = useMemo(() => {
    if (!searchQuery) return services
    const query = searchQuery.toLowerCase()
    return services.filter(
      (svc) =>
        svc.name.toLowerCase().includes(query) ||
        svc.image_name.toLowerCase().includes(query) ||
        svc.category?.name.toLowerCase().includes(query)
    )
  }, [services, searchQuery])

  const getInstanceCount = (serviceId: number) => {
    return instances.filter((inst) => inst.template_service_id === serviceId).length
  }

  const generateInstanceName = (serviceName: string) => {
    const existing = instances.filter((inst) => inst.instance_id.startsWith(serviceName))
    if (existing.length === 0) return serviceName
    let counter = 2
    while (instances.some((inst) => inst.instance_id === `${serviceName}-${counter}`)) {
      counter++
    }
    return `${serviceName}-${counter}`
  }

  const handleServiceSelect = (serviceId: number, serviceName: string) => {
    setSelectedServiceId(serviceId)
    setInstanceName(generateInstanceName(serviceName))
    setNameError('')
  }

  const validateName = (value: string) => {
    if (!value) {
      setNameError('Name is required')
      return false
    }
    if (!/^[a-z0-9-]+$/.test(value)) {
      setNameError('Name must contain only lowercase letters, numbers, and hyphens')
      return false
    }
    if (value.length > 63) {
      setNameError('Name must be 63 characters or less')
      return false
    }
    if (value.startsWith('-') || value.endsWith('-')) {
      setNameError('Name cannot start or end with a hyphen')
      return false
    }
    if (instances.some((inst) => inst.instance_id === value)) {
      setNameError('Instance name already exists in this stack')
      return false
    }
    setNameError('')
    return true
  }

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (!selectedServiceId || !validateName(instanceName)) return

    createInstance.mutate(
      {
        instance_id: instanceName,
        template_service_id: selectedServiceId,
        description,
      },
      {
        onSuccess: () => {
          onOpenChange(false)
          setSelectedServiceId(null)
          setInstanceName('')
          setDescription('')
          setSearchQuery('')
        },
      }
    )
  }

  const handleOpenChange = (newOpen: boolean) => {
    if (!newOpen) {
      setSelectedServiceId(null)
      setInstanceName('')
      setDescription('')
      setSearchQuery('')
      setNameError('')
    }
    onOpenChange(newOpen)
  }

  const selectedService = services.find((svc) => svc.id === selectedServiceId)

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="max-w-3xl max-h-[80vh] flex flex-col">
        <DialogHeader>
          <DialogTitle>Add Service Instance</DialogTitle>
        </DialogHeader>

        <div className="flex-1 overflow-hidden flex flex-col gap-4">
          <div className="relative">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 size-4 text-muted-foreground" />
            <Input
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              placeholder="Search templates..."
              className="pl-9"
              autoFocus
            />
          </div>

          <div className="flex-1 overflow-y-auto border rounded-lg">
            {filteredServices.length === 0 ? (
              <div className="flex flex-col items-center justify-center py-12 text-muted-foreground">
                <Package className="size-12 mb-2 opacity-50" />
                <p>No templates found</p>
              </div>
            ) : (
              <div className="grid gap-2 p-2 sm:grid-cols-2">
                {filteredServices.map((service) => {
                  const count = getInstanceCount(service.id)
                  const isSelected = selectedServiceId === service.id
                  return (
                    <button
                      key={service.id}
                      onClick={() => handleServiceSelect(service.id, service.name)}
                      className={cn(
                        'text-left p-3 rounded-lg border transition-colors',
                        isSelected
                          ? 'border-primary bg-primary/5'
                          : 'border-border hover:border-primary/50 hover:bg-muted/50'
                      )}
                    >
                      <div className="flex items-start justify-between gap-2">
                        <div className="flex-1 min-w-0">
                          <div className="font-medium truncate">{service.name}</div>
                          <div className="text-xs text-muted-foreground truncate">
                            {service.image_name}
                          </div>
                          {service.category && (
                            <div className="text-xs text-muted-foreground mt-1">
                              {service.category.display_name || service.category.name}
                            </div>
                          )}
                        </div>
                        {count > 0 && (
                          <Badge variant="secondary" className="text-xs shrink-0">
                            {count} {count === 1 ? 'instance' : 'instances'}
                          </Badge>
                        )}
                      </div>
                    </button>
                  )
                })}
              </div>
            )}
          </div>

          {selectedService && (
            <form onSubmit={handleSubmit} className="space-y-4 border-t pt-4">
              <div className="grid gap-2">
                <label className="text-sm font-medium">Instance Name</label>
                <Input
                  value={instanceName}
                  onChange={(e) => {
                    setInstanceName(e.target.value)
                    setNameError('')
                  }}
                  onBlur={(e) => validateName(e.target.value)}
                  placeholder={selectedService.name}
                />
                {nameError && (
                  <p className="text-sm text-destructive">{nameError}</p>
                )}
              </div>
              <div className="grid gap-2">
                <label className="text-sm font-medium">Description (optional)</label>
                <Textarea
                  value={description}
                  onChange={(e) => setDescription(e.target.value)}
                  placeholder="Optional description"
                  rows={2}
                />
              </div>
              <DialogFooter>
                <Button type="button" variant="outline" onClick={() => handleOpenChange(false)}>
                  Cancel
                </Button>
                <Button type="submit" disabled={createInstance.isPending || !instanceName || !!nameError}>
                  {createInstance.isPending ? <Loader2 className="size-4 animate-spin" /> : 'Add'}
                </Button>
              </DialogFooter>
            </form>
          )}
        </div>
      </DialogContent>
    </Dialog>
  )
}
