import { TabsList, TabsTrigger } from '@/components/ui/tabs'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'

interface TabItem {
  value: string
  label: string
}

interface ResponsiveTabsListProps {
  tabs: TabItem[]
  value: string
  onValueChange: (value: string) => void
}

export function ResponsiveTabsList({ tabs, value, onValueChange }: ResponsiveTabsListProps) {
  return (
    <>
      <div className="sm:hidden">
        <Select value={value} onValueChange={onValueChange}>
          <SelectTrigger className="w-full">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            {tabs.map((tab) => (
              <SelectItem key={tab.value} value={tab.value}>{tab.label}</SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>
      <TabsList className="hidden sm:inline-flex">
        {tabs.map((tab) => (
          <TabsTrigger key={tab.value} value={tab.value}>{tab.label}</TabsTrigger>
        ))}
      </TabsList>
    </>
  )
}
