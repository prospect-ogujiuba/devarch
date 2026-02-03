import { useState } from 'react'
import { Plus, Trash2, Save, Loader2, FileText } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { ConfirmDialog } from '@/components/ui/confirm-dialog'
import { useServiceConfigFiles, useServiceConfigFile, useSaveConfigFile, useDeleteConfigFile } from '@/features/services/queries'
import { CodeEditor, detectLang } from '@/components/services/code-editor'

interface Props {
  serviceName: string
}

export function ConfigFilesPanel({ serviceName }: Props) {
  const { data: files = [], isLoading } = useServiceConfigFiles(serviceName)
  const saveFile = useSaveConfigFile()
  const deleteFile = useDeleteConfigFile()

  const [selectedPath, setSelectedPath] = useState<string>('')
  const [newFilePath, setNewFilePath] = useState('')
  const [showNewInput, setShowNewInput] = useState(false)
  const [deleteTarget, setDeleteTarget] = useState<string | null>(null)

  const createFile = () => {
    if (!newFilePath) return
    saveFile.mutate(
      { name: serviceName, filePath: newFilePath, content: '' },
      {
        onSuccess: () => {
          setSelectedPath(newFilePath)
          setNewFilePath('')
          setShowNewInput(false)
        },
      },
    )
  }

  const handleDelete = () => {
    if (!deleteTarget) return
    deleteFile.mutate(
      { name: serviceName, filePath: deleteTarget },
      {
        onSuccess: () => {
          if (selectedPath === deleteTarget) setSelectedPath('')
          setDeleteTarget(null)
        },
      },
    )
  }

  if (isLoading) {
    return (
      <Card>
        <CardContent className="flex items-center justify-center py-8">
          <Loader2 className="size-6 animate-spin text-muted-foreground" />
        </CardContent>
      </Card>
    )
  }

  return (
    <div className="grid gap-4 md:grid-cols-[250px_1fr]">
      <Card>
        <CardHeader className="flex flex-row items-center justify-between pb-2">
          <CardTitle className="text-base">Files</CardTitle>
          <Button variant="outline" size="icon-sm" onClick={() => setShowNewInput(true)}>
            <Plus className="size-4" />
          </Button>
        </CardHeader>
        <CardContent className="space-y-1">
          {showNewInput && (
            <div className="flex gap-1 mb-2">
              <Input
                className="text-xs"
                value={newFilePath}
                onChange={(e) => setNewFilePath(e.target.value)}
                placeholder="path/to/file"
                onKeyDown={(e) => e.key === 'Enter' && createFile()}
              />
              <Button size="icon-sm" onClick={createFile} disabled={saveFile.isPending}>
                <Plus className="size-4" />
              </Button>
            </div>
          )}
          {files.length === 0 && !showNewInput && (
            <p className="text-muted-foreground text-sm">No config files</p>
          )}
          {files.map((f) => (
            <div
              key={f.file_path}
              className={`flex items-center gap-2 px-2 py-1.5 rounded-md cursor-pointer text-sm ${
                selectedPath === f.file_path ? 'bg-accent text-accent-foreground' : 'hover:bg-muted'
              }`}
            >
              <button className="flex-1 text-left flex items-center gap-2 truncate" onClick={() => setSelectedPath(f.file_path)}>
                <FileText className="size-3.5 shrink-0" />
                <span className="truncate">{f.file_path}</span>
              </button>
              <Button
                variant="ghost"
                size="icon-sm"
                className="size-6 shrink-0"
                onClick={(e) => { e.stopPropagation(); setDeleteTarget(f.file_path) }}
              >
                <Trash2 className="size-3 text-destructive" />
              </Button>
            </div>
          ))}
        </CardContent>
      </Card>

      <Card>
        <CardContent className="p-0">
          {selectedPath ? (
            <FileEditor serviceName={serviceName} filePath={selectedPath} />
          ) : (
            <div className="flex items-center justify-center h-[400px] text-muted-foreground">
              Select a file to edit
            </div>
          )}
        </CardContent>
      </Card>

      <ConfirmDialog
        open={!!deleteTarget}
        onOpenChange={(open) => !open && setDeleteTarget(null)}
        title="Delete file?"
        description={`Delete ${deleteTarget}? This cannot be undone.`}
        confirmLabel="Delete"
        onConfirm={handleDelete}
        variant="destructive"
      />
    </div>
  )
}

function FileEditor({ serviceName, filePath }: { serviceName: string; filePath: string }) {
  const { data: file, isLoading } = useServiceConfigFile(serviceName, filePath)
  const saveFile = useSaveConfigFile()
  const [content, setContent] = useState<string | null>(null)
  const lang = detectLang(filePath)

  const currentContent = content ?? file?.content ?? ''
  const isDirty = content !== null && content !== (file?.content ?? '')

  const handleSave = () => {
    saveFile.mutate(
      { name: serviceName, filePath, content: currentContent, fileMode: file?.file_mode },
      { onSuccess: () => setContent(null) },
    )
  }

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-[400px]">
        <Loader2 className="size-6 animate-spin text-muted-foreground" />
      </div>
    )
  }

  return (
    <div>
      <div className="flex items-center justify-between px-4 py-2 border-b">
        <span className="text-sm font-mono truncate">{filePath}</span>
        <Button size="sm" disabled={!isDirty || saveFile.isPending} onClick={handleSave}>
          {saveFile.isPending ? <Loader2 className="size-4 animate-spin" /> : <Save className="size-4" />}
          Save
        </Button>
      </div>
      <CodeEditor
        value={currentContent}
        onChange={setContent}
        language={lang}
      />
    </div>
  )
}
