import { createFileRoute, Link, useNavigate } from '@tanstack/react-router'
import { ArrowLeft, Star, Download, BadgeCheck, Loader2, Plus } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { useImageTags, useImageInfo } from '@/features/registry/queries'
import { formatBytes } from '@/lib/format'

export const Route = createFileRoute('/registries/$registry/$')({
  component: ImageDetailPage,
})

function formatCount(n: number): string {
  if (n >= 1_000_000_000) return `${(n / 1_000_000_000).toFixed(1)}B`
  if (n >= 1_000_000) return `${(n / 1_000_000).toFixed(1)}M`
  if (n >= 1_000) return `${(n / 1_000).toFixed(1)}K`
  return String(n)
}

function ImageDetailPage() {
  const { registry, _splat } = Route.useParams()
  const repository = _splat ?? ''
  const navigate = useNavigate()
  const { data: imageInfo, isLoading: infoLoading } = useImageInfo(registry, repository, { enabled: !!repository })
  const { data: tags, isLoading: tagsLoading } = useImageTags(registry, repository, { enabled: !!repository })

  const imageName = registry === 'dockerhub'
    ? (repository.startsWith('library/') ? repository.slice(8) : repository)
    : repository

  const handleUseImage = (tag?: string) => {
    navigate({
      to: '/services/new',
      search: {
        image_name: imageName,
        image_tag: tag ?? 'latest',
      },
    })
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-4">
        <Link to="/registries" className="text-muted-foreground hover:text-foreground">
          <ArrowLeft className="size-5" />
        </Link>
        <div className="min-w-0 flex-1">
          <div className="flex flex-wrap items-center gap-2">
            <h1 className="truncate text-xl font-bold sm:text-2xl">{repository}</h1>
            {imageInfo?.is_official && (
              <Badge variant="secondary">
                <BadgeCheck className="mr-1 size-3" />
                Official
              </Badge>
            )}
          </div>
          <p className="text-sm text-muted-foreground">{registry}</p>
        </div>
        <Button onClick={() => handleUseImage()} className="shrink-0">
          <Plus className="mr-1 size-4" />
          Use Image
        </Button>
      </div>

      {infoLoading ? (
        <Card>
          <CardContent className="flex items-center justify-center py-12">
            <Loader2 className="size-6 animate-spin text-muted-foreground" />
          </CardContent>
        </Card>
      ) : imageInfo ? (
        <Card>
          <CardHeader>
            <CardTitle className="text-base">Image Info</CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            {imageInfo.description && (
              <p className="text-sm text-muted-foreground">{imageInfo.description}</p>
            )}
            <div className="flex items-center gap-6 text-sm">
              <span className="flex items-center gap-1.5">
                <Star className="size-4 text-muted-foreground" />
                {formatCount(imageInfo.star_count ?? 0)} stars
              </span>
              <span className="flex items-center gap-1.5">
                <Download className="size-4 text-muted-foreground" />
                {formatCount(imageInfo.pull_count ?? 0)} pulls
              </span>
            </div>
          </CardContent>
        </Card>
      ) : null}

      <Card>
        <CardHeader>
          <CardTitle className="text-base">Tags</CardTitle>
        </CardHeader>
        <CardContent>
          {tagsLoading ? (
            <div className="flex items-center justify-center py-8">
              <Loader2 className="size-6 animate-spin text-muted-foreground" />
            </div>
          ) : !tags || tags.length === 0 ? (
            <p className="text-sm text-muted-foreground">No tags found</p>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b text-left text-muted-foreground">
                    <th className="pb-2 pr-4 font-medium">Tag</th>
                    <th className="pb-2 pr-4 font-medium">Size</th>
                    <th className="pb-2 pr-4 font-medium">Pushed</th>
                    <th className="pb-2 pr-4 font-medium">Architectures</th>
                    <th className="pb-2 font-medium"></th>
                  </tr>
                </thead>
                <tbody>
                  {tags.map((tag) => (
                    <tr key={tag.name} className="border-b last:border-0">
                      <td className="py-2 pr-4 font-mono">{tag.name}</td>
                      <td className="py-2 pr-4 text-muted-foreground">
                        {tag.size_bytes > 0 ? formatBytes(tag.size_bytes) : '-'}
                      </td>
                      <td className="py-2 pr-4 text-muted-foreground">
                        {tag.pushed_at ? new Date(tag.pushed_at).toLocaleDateString() : '-'}
                      </td>
                      <td className="py-2 pr-4">
                        <div className="flex flex-wrap gap-1">
                          {(tag.architectures ?? []).slice(0, 4).map((a, i) => (
                            <Badge key={i} variant="outline" className="text-xs">
                              {a.os}/{a.architecture}{a.variant ? `/${a.variant}` : ''}
                            </Badge>
                          ))}
                          {(tag.architectures?.length ?? 0) > 4 && (
                            <Badge variant="outline" className="text-xs">
                              +{tag.architectures!.length - 4}
                            </Badge>
                          )}
                        </div>
                      </td>
                      <td className="py-2 text-right">
                        <Button size="sm" variant="ghost" onClick={() => handleUseImage(tag.name)}>
                          Use
                        </Button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
