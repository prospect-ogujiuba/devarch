import { createFileRoute } from '@tanstack/react-router'
import { useState, useMemo, useRef } from 'react'
import { HardDrive, Download, Trash2, Eraser, Loader2 } from 'lucide-react'
import { toast } from 'sonner'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { useQueryClient } from '@tanstack/react-query'
import { useImages, useRemoveImage, usePruneImages, pullImageWithProgress } from '@/features/images/queries'
import { formatBytes } from '@/lib/format'
import type { ImagePullProgress } from '@/types/api'

export const Route = createFileRoute('/images/')({
  component: ImagesPage,
})

function formatAge(created: number): string {
  const seconds = Math.floor(Date.now() / 1000 - created)
  if (seconds < 60) return `${seconds}s ago`
  if (seconds < 3600) return `${Math.floor(seconds / 60)}m ago`
  if (seconds < 86400) return `${Math.floor(seconds / 3600)}h ago`
  const days = Math.floor(seconds / 86400)
  if (days < 30) return `${days}d ago`
  if (days < 365) return `${Math.floor(days / 30)}mo ago`
  return `${Math.floor(days / 365)}y ago`
}

function ImagesPage() {
  const queryClient = useQueryClient()
  const { data, isLoading } = useImages(false)
  const removeMutation = useRemoveImage()
  const pruneMutation = usePruneImages()

  const [pullOpen, setPullOpen] = useState(false)
  const [pullReference, setPullReference] = useState('')
  const [pulling, setPulling] = useState(false)
  const [pullProgress, setPullProgress] = useState<ImagePullProgress[]>([])
  const progressRef = useRef<HTMLDivElement>(null)

  const [removeTarget, setRemoveTarget] = useState<string | null>(null)
  const [pruneOpen, setPruneOpen] = useState(false)

  const images = useMemo(() => data ?? [], [data])

  const handlePullStart = async () => {
    if (!pullReference.trim()) return
    setPulling(true)
    setPullProgress([])
    try {
      await pullImageWithProgress(pullReference, (report) => {
        setPullProgress((prev) => [...prev.slice(-50), report])
        requestAnimationFrame(() => {
          progressRef.current?.scrollTo({ top: progressRef.current.scrollHeight })
        })
      })
      setPullProgress((prev) => [...prev, { stream: `✓ Successfully pulled ${pullReference}` }])
      toast.success(`Pulled ${pullReference}`)
      queryClient.invalidateQueries({ queryKey: ['images'] })
    } catch (error) {
      setPullProgress((prev) => [
        ...prev,
        { error: error instanceof Error ? error.message : 'Pull failed' },
      ])
      toast.error(`Failed to pull: ${error instanceof Error ? error.message : 'Unknown error'}`)
    } finally {
      setPulling(false)
    }
  }

  const handleRemove = async () => {
    if (!removeTarget) return
    await removeMutation.mutateAsync({ name: removeTarget })
    setRemoveTarget(null)
  }

  const handlePrune = async () => {
    await pruneMutation.mutateAsync(true)
    setPruneOpen(false)
  }

  return (
    <div className="space-y-4 p-4 md:p-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">Container Images</h1>
          <p className="text-sm text-muted-foreground">
            Manage container images on this host
          </p>
        </div>
        <div className="flex gap-2">
          <Button variant="outline" size="sm" onClick={() => setPruneOpen(true)}>
            <Eraser className="size-4" />
            Prune
          </Button>
          <Button size="sm" onClick={() => setPullOpen(true)}>
            <Download className="size-4" />
            Pull Image
          </Button>
        </div>
      </div>

      {isLoading ? (
        <div className="flex items-center justify-center py-12">
          <Loader2 className="size-8 animate-spin text-muted-foreground" />
        </div>
      ) : images.length === 0 ? (
        <div className="flex flex-col items-center justify-center rounded-lg border border-dashed py-12">
          <HardDrive className="size-12 text-muted-foreground/50" />
          <p className="mt-2 text-sm text-muted-foreground">No images found</p>
        </div>
      ) : (
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Image</TableHead>
              <TableHead>Tag</TableHead>
              <TableHead>ID</TableHead>
              <TableHead>Size</TableHead>
              <TableHead>Created</TableHead>
              <TableHead className="w-[80px]"></TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {images.map((image) => {
              const firstTag = image.RepoTags?.[0] || '<none>'
              const [repo, tag] = firstTag.includes(':') ? firstTag.split(':') : [firstTag, '<none>']
              const shortId = image.Id.replace(/^sha256:/, '').slice(0, 12)
              return (
                <TableRow key={image.Id}>
                  <TableCell className="font-medium">{repo}</TableCell>
                  <TableCell className="text-muted-foreground">{tag}</TableCell>
                  <TableCell className="font-mono text-xs text-muted-foreground">{shortId}</TableCell>
                  <TableCell>{formatBytes(image.Size)}</TableCell>
                  <TableCell className="text-muted-foreground">{formatAge(image.Created)}</TableCell>
                  <TableCell>
                    <Button
                      variant="ghost"
                      size="icon-sm"
                      onClick={() => setRemoveTarget(firstTag)}
                      disabled={removeMutation.isPending}
                    >
                      <Trash2 className="size-4 text-muted-foreground hover:text-destructive" />
                    </Button>
                  </TableCell>
                </TableRow>
              )
            })}
          </TableBody>
        </Table>
      )}

      <Dialog open={pullOpen} onOpenChange={(open) => {
        setPullOpen(open)
        if (!open) {
          setPullReference('')
          setPullProgress([])
        }
      }}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Pull Container Image</DialogTitle>
            <DialogDescription>
              Enter the image reference (e.g., nginx:latest, postgres:16)
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4">
            <Input
              placeholder="image:tag"
              value={pullReference}
              onChange={(e) => setPullReference(e.target.value)}
              disabled={pulling || pullProgress.length > 0}
            />
            {pullProgress.length > 0 && (
              <div ref={progressRef} className="max-h-64 space-y-1 overflow-y-auto rounded-md border bg-muted/50 p-3 font-mono text-xs">
                {pullProgress.map((p, i) => (
                  <div key={i} className={p.error ? "text-destructive" : undefined}>
                    {p.error ?? p.stream ?? p.id}
                  </div>
                ))}
              </div>
            )}
          </div>
          <DialogFooter>
            {!pulling && pullProgress.length > 0 ? (
              <Button onClick={() => setPullOpen(false)}>
                Done
              </Button>
            ) : (
              <>
                <Button variant="outline" onClick={() => setPullOpen(false)} disabled={pulling}>
                  Cancel
                </Button>
                <Button onClick={handlePullStart} disabled={!pullReference.trim() || pulling}>
                  {pulling && <Loader2 className="size-4 animate-spin" />}
                  Pull
                </Button>
              </>
            )}
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog open={!!removeTarget} onOpenChange={(open) => !open && setRemoveTarget(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Remove Image</DialogTitle>
            <DialogDescription>
              Are you sure you want to remove <span className="font-medium">{removeTarget}</span>?
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setRemoveTarget(null)}>
              Cancel
            </Button>
            <Button variant="destructive" onClick={handleRemove} disabled={removeMutation.isPending}>
              {removeMutation.isPending && <Loader2 className="size-4 animate-spin" />}
              Remove
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog open={pruneOpen} onOpenChange={setPruneOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Prune Dangling Images</DialogTitle>
            <DialogDescription>
              This will remove all dangling images (untagged and not used by any container).
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setPruneOpen(false)}>
              Cancel
            </Button>
            <Button onClick={handlePrune} disabled={pruneMutation.isPending}>
              {pruneMutation.isPending && <Loader2 className="size-4 animate-spin" />}
              Prune
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
