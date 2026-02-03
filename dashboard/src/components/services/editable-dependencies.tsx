import { useState } from 'react'
import { Pencil, X, Check } from 'lucide-react'
import { Link } from '@tanstack/react-router'
import { Card, CardContent, CardHeader, CardTitle, CardAction } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { useUpdateDependencies, useServices } from '@/features/services/queries'

interface Props {
  name: string
  dependencies: string[]
}

export function EditableDependencies({ name, dependencies }: Props) {
  const [editing, setEditing] = useState(false)
  const [selected, setSelected] = useState<string[]>([])
  const { data: servicesData } = useServices()
  const mutation = useUpdateDependencies()

  const allServices = (servicesData?.services ?? []).filter((s) => s.name !== name)

  const startEdit = () => {
    setSelected([...dependencies])
    setEditing(true)
  }

  const toggle = (svcName: string) => {
    setSelected((prev) =>
      prev.includes(svcName) ? prev.filter((d) => d !== svcName) : [...prev, svcName],
    )
  }

  const save = () => {
    mutation.mutate({ name, data: { dependencies: selected } }, { onSuccess: () => setEditing(false) })
  }

  if (editing) {
    return (
      <Card>
        <CardHeader className="flex flex-row items-center justify-between">
          <CardTitle className="text-base">Dependencies</CardTitle>
          <div className="flex gap-1">
            <Button variant="ghost" size="icon-sm" onClick={() => setEditing(false)}><X className="size-4" /></Button>
            <Button variant="default" size="icon-sm" onClick={save} disabled={mutation.isPending}><Check className="size-4" /></Button>
          </div>
        </CardHeader>
        <CardContent>
          <div className="flex flex-wrap gap-2">
            {allServices.map((s) => (
              <Button
                key={s.name}
                variant={selected.includes(s.name) ? 'default' : 'outline'}
                size="sm"
                onClick={() => toggle(s.name)}
              >
                {s.name}
              </Button>
            ))}
            {allServices.length === 0 && <p className="text-muted-foreground text-sm">No other services</p>}
          </div>
        </CardContent>
      </Card>
    )
  }

  if (dependencies.length === 0) return null

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base">Dependencies</CardTitle>
        <CardAction>
          <Button variant="ghost" size="icon-sm" onClick={startEdit}><Pencil className="size-4" /></Button>
        </CardAction>
      </CardHeader>
      <CardContent>
        <div className="flex flex-wrap gap-2">
          {dependencies.map((dep) => (
            <Badge key={dep} variant="outline">
              <Link to="/services/$name" params={{ name: dep }} className="hover:underline">
                {dep}
              </Link>
            </Badge>
          ))}
        </div>
      </CardContent>
    </Card>
  )
}
