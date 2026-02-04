import { useState } from 'react'
import { Plus, Trash2, X, Check, RotateCcw, FileText } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle, CardAction } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog'
import { CodeEditor, detectLang } from '@/components/services/code-editor'
import { api } from '@/lib/api'
import { toast } from 'sonner'
import type { InstanceDetail, Service, InstanceConfigFile } from '@/types/api'

interface Props {
  instance: InstanceDetail
  templateData: Service
  stackName: string
  instanceId: string
}

interface TemplateConfigFile {
  file_path: string
  content: string
}

export function OverrideConfigFiles({ instance, stackName, instanceId }: Props) {
  const [selectedFile, setSelectedFile] = useState<string | null>(null)
  const [fileContent, setFileContent] = useState('')
  const [fileMode, setFileMode] = useState('0644')
  const [newFileName, setNewFileName] = useState('')
  const [newFileOpen, setNewFileOpen] = useState(false)
  const [saving, setSaving] = useState(false)

  const templateConfigFiles: TemplateConfigFile[] = []
  const overrideConfigFiles = instance.config_files ?? []

  const allFilePaths = [...new Set([
    ...templateConfigFiles.map((f: TemplateConfigFile) => f.file_path),
    ...overrideConfigFiles.map((f: InstanceConfigFile) => f.file_path),
  ])]

  const getFileInfo = (path: string): { hasTemplate: boolean; hasOverride: boolean; templateContent?: string; overrideFile?: InstanceConfigFile } => {
    const templateFile = templateConfigFiles.find((f: TemplateConfigFile) => f.file_path === path)
    const overrideFile = overrideConfigFiles.find((f: InstanceConfigFile) => f.file_path === path)
    return {
      hasTemplate: !!templateFile,
      hasOverride: !!overrideFile,
      templateContent: templateFile?.content,
      overrideFile,
    }
  }

  const openFile = async (path: string) => {
    const info = getFileInfo(path)
    if (info.hasOverride) {
      setFileContent(info.overrideFile?.content ?? '')
      setFileMode(info.overrideFile?.file_mode ?? '0644')
    } else if (info.hasTemplate) {
      setFileContent(info.templateContent ?? '')
      setFileMode('0644')
    } else {
      setFileContent('')
      setFileMode('0644')
    }
    setSelectedFile(path)
  }

  const closeFile = () => {
    setSelectedFile(null)
    setFileContent('')
  }

  const saveFile = async () => {
    if (!selectedFile) return
    setSaving(true)
    try {
      await api.put(`/stacks/${stackName}/instances/${instanceId}/files/${selectedFile}`, {
        content: fileContent,
        file_mode: fileMode,
      })
      toast.success('File saved')
      closeFile()
      window.location.reload()
    } catch (error: any) {
      toast.error(error.response?.data || 'Failed to save file')
    } finally {
      setSaving(false)
    }
  }

  const deleteFile = async (path: string) => {
    try {
      await api.delete(`/stacks/${stackName}/instances/${instanceId}/files/${path}`)
      toast.success('File deleted')
      window.location.reload()
    } catch (error: any) {
      toast.error(error.response?.data || 'Failed to delete file')
    }
  }

  const resetFile = async (path: string) => {
    try {
      await api.delete(`/stacks/${stackName}/instances/${instanceId}/files/${path}`)
      toast.success('File reset to template')
      window.location.reload()
    } catch (error: any) {
      toast.error(error.response?.data || 'Failed to reset file')
    }
  }

  const createNewFile = async () => {
    if (!newFileName) return
    setSaving(true)
    try {
      await api.put(`/stacks/${stackName}/instances/${instanceId}/files/${newFileName}`, {
        content: '',
        file_mode: '0644',
      })
      toast.success('File created')
      setNewFileOpen(false)
      setNewFileName('')
      window.location.reload()
    } catch (error: any) {
      toast.error(error.response?.data || 'Failed to create file')
    } finally {
      setSaving(false)
    }
  }

  return (
    <>
      <Card>
        <CardHeader>
          <CardTitle className="text-base">Config File Overrides</CardTitle>
          <CardAction>
            <Button variant="outline" size="sm" onClick={() => setNewFileOpen(true)}>
              <Plus className="size-4" /> Add File
            </Button>
          </CardAction>
        </CardHeader>
        <CardContent className="space-y-4">
          {templateConfigFiles.length > 0 && (
            <div>
              <div className="text-xs font-medium text-muted-foreground mb-2">Template Files</div>
              <div className="space-y-2">
                {templateConfigFiles.map((file: TemplateConfigFile) => {
                  const info = getFileInfo(file.file_path)
                  return (
                    <div key={file.file_path} className="flex items-center justify-between p-2 border rounded">
                      <div className="flex items-center gap-2">
                        <FileText className="size-4 text-muted-foreground" />
                        <code className="text-sm">{file.file_path}</code>
                        {info.hasOverride && (
                          <Badge variant="default" className="text-xs">overridden</Badge>
                        )}
                      </div>
                      <Button variant="outline" size="sm" onClick={() => openFile(file.file_path)}>
                        {info.hasOverride ? 'Edit Override' : 'View'}
                      </Button>
                    </div>
                  )
                })}
              </div>
            </div>
          )}

          {overrideConfigFiles.filter((f) => !templateConfigFiles.some((t) => t.file_path === f.file_path)).length > 0 && (
            <div>
              <div className="text-xs font-medium mb-2">Custom Files</div>
              <div className="space-y-2">
                {overrideConfigFiles
                  .filter((f: InstanceConfigFile) => !templateConfigFiles.some((t: TemplateConfigFile) => t.file_path === f.file_path))
                  .map((file) => (
                    <div key={file.file_path} className="flex items-center justify-between p-2 border rounded border-l-2 border-blue-500">
                      <div className="flex items-center gap-2">
                        <FileText className="size-4" />
                        <code className="text-sm">{file.file_path}</code>
                        <Badge variant="default" className="text-xs">custom</Badge>
                      </div>
                      <div className="flex gap-2">
                        <Button variant="outline" size="sm" onClick={() => openFile(file.file_path)}>
                          Edit
                        </Button>
                        <Button variant="destructive" size="sm" onClick={() => deleteFile(file.file_path)}>
                          <Trash2 className="size-4" />
                        </Button>
                      </div>
                    </div>
                  ))}
              </div>
            </div>
          )}

          {allFilePaths.length === 0 && (
            <p className="text-muted-foreground text-sm">No config files</p>
          )}
        </CardContent>
      </Card>

      {selectedFile && (
        <Dialog open={!!selectedFile} onOpenChange={(open) => !open && closeFile()}>
          <DialogContent className="max-w-4xl">
            <DialogHeader>
              <DialogTitle>{selectedFile}</DialogTitle>
            </DialogHeader>
            <div className="space-y-4">
              {getFileInfo(selectedFile).templateContent && (
                <div>
                  <div className="text-xs font-medium text-muted-foreground mb-2">Template Content (read-only)</div>
                  <CodeEditor
                    value={getFileInfo(selectedFile).templateContent ?? ''}
                    onChange={() => {}}
                    language={detectLang(selectedFile)}
                    readOnly
                  />
                </div>
              )}
              <div>
                <div className="text-xs font-medium mb-2">Override Content</div>
                <CodeEditor
                  value={fileContent}
                  onChange={setFileContent}
                  language={detectLang(selectedFile)}
                />
              </div>
              <div className="flex items-center gap-2">
                <label className="text-sm font-medium">File Mode:</label>
                <Input
                  className="w-24"
                  value={fileMode}
                  onChange={(e) => setFileMode(e.target.value)}
                />
              </div>
            </div>
            <DialogFooter>
              {getFileInfo(selectedFile).hasOverride && (
                <Button variant="outline" onClick={() => resetFile(selectedFile)}>
                  <RotateCcw className="size-4" /> Reset to Template
                </Button>
              )}
              <Button variant="outline" onClick={closeFile}>Cancel</Button>
              <Button onClick={saveFile} disabled={saving}>
                {saving ? <X className="size-4 animate-spin" /> : <Check className="size-4" />}
                Save
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      )}

      <Dialog open={newFileOpen} onOpenChange={setNewFileOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Add Config File</DialogTitle>
          </DialogHeader>
          <div className="grid gap-4 py-4">
            <div className="grid gap-2">
              <label className="text-sm font-medium">File Path</label>
              <Input
                value={newFileName}
                onChange={(e) => setNewFileName(e.target.value)}
                placeholder="/etc/config/app.conf"
              />
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setNewFileOpen(false)}>Cancel</Button>
            <Button onClick={createNewFile} disabled={saving || !newFileName}>
              {saving ? <X className="size-4 animate-spin" /> : 'Create'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  )
}
