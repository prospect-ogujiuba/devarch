package registry

import (
	"context"
	"time"
)

type Registry interface {
	Name() string
	Authenticate(ctx context.Context) error
	GetImageInfo(ctx context.Context, repository string) (*ImageInfo, error)
	ListTags(ctx context.Context, repository string, opts ListTagsOptions) ([]TagInfo, error)
	GetTagManifest(ctx context.Context, repository, tag string) (*TagInfo, error)
	RateLimitRemaining() int
}

type ImageInfo struct {
	Repository   string
	Description  string
	StarCount    int
	PullCount    int64
	IsOfficial   bool
	LastUpdated  time.Time
}

type TagInfo struct {
	Name          string
	Digest        string
	SizeBytes     int64
	PushedAt      time.Time
	Architectures []ArchInfo
}

type ArchInfo struct {
	OS           string
	Architecture string
	Variant      string
	Digest       string
	SizeBytes    int64
}

type ListTagsOptions struct {
	PageSize int
	Page     int
}

type Manager struct {
	registries map[string]Registry
}

func NewManager() *Manager {
	return &Manager{
		registries: make(map[string]Registry),
	}
}

func (m *Manager) Register(r Registry) {
	m.registries[r.Name()] = r
}

func (m *Manager) Get(name string) Registry {
	return m.registries[name]
}

func (m *Manager) DetectRegistry(imageName string) string {
	if len(imageName) > 0 {
		if imageName[0] >= 'a' && imageName[0] <= 'z' {
			firstSlash := -1
			for i, c := range imageName {
				if c == '/' {
					firstSlash = i
					break
				}
			}
			if firstSlash == -1 {
				return "dockerhub"
			}
			prefix := imageName[:firstSlash]
			if prefix == "ghcr.io" {
				return "ghcr"
			}
			if prefix == "quay.io" {
				return "quay"
			}
			if prefix == "gcr.io" {
				return "gcr"
			}
			if !containsDot(prefix) {
				return "dockerhub"
			}
		}
	}
	return "dockerhub"
}

func containsDot(s string) bool {
	for _, c := range s {
		if c == '.' {
			return true
		}
	}
	return false
}

func (m *Manager) NormalizeRepository(imageName string) (registryName string, repository string) {
	registryName = m.DetectRegistry(imageName)

	switch registryName {
	case "ghcr":
		repository = trimPrefix(imageName, "ghcr.io/")
	case "quay":
		repository = trimPrefix(imageName, "quay.io/")
	case "gcr":
		repository = trimPrefix(imageName, "gcr.io/")
	default:
		if !containsSlash(imageName) {
			repository = "library/" + imageName
		} else {
			repository = imageName
		}
	}

	return
}

func trimPrefix(s, prefix string) string {
	if len(s) >= len(prefix) && s[:len(prefix)] == prefix {
		return s[len(prefix):]
	}
	return s
}

func containsSlash(s string) bool {
	for _, c := range s {
		if c == '/' {
			return true
		}
	}
	return false
}
