package sync

import (
	"context"
	"database/sql"
	"log/slog"
	"time"

	"github.com/priz/devarch-api/pkg/registry"
)

type RegistrySync struct {
	db      *sql.DB
	manager *registry.Manager
	logger  *slog.Logger
}

func NewRegistrySync(db *sql.DB, mgr *registry.Manager, logger *slog.Logger) *RegistrySync {
	return &RegistrySync{
		db:      db,
		manager: mgr,
		logger:  logger,
	}
}

func (rs *RegistrySync) SyncAll(ctx context.Context) error {
	rows, err := rs.db.Query("SELECT DISTINCT image_name FROM services WHERE enabled = true")
	if err != nil {
		return err
	}
	defer rows.Close()

	var images []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			continue
		}
		images = append(images, name)
	}

	for _, imageName := range images {
		if err := rs.SyncImage(ctx, imageName); err != nil {
			rs.logger.Error("failed to sync image", "image", imageName, "error", err)
		}
	}

	return nil
}

func (rs *RegistrySync) SyncImage(ctx context.Context, imageName string) error {
	registryName, repository := rs.manager.NormalizeRepository(imageName)

	reg := rs.manager.Get(registryName)
	if reg == nil {
		return nil
	}

	var registryID int
	err := rs.db.QueryRow("SELECT id FROM registries WHERE name = $1", registryName).Scan(&registryID)
	if err != nil {
		return err
	}

	info, err := reg.GetImageInfo(ctx, repository)
	if err != nil {
		rs.logger.Error("failed to get image info", "image", imageName, "error", err)
		info = &registry.ImageInfo{Repository: imageName}
	}

	var imageID int
	err = rs.db.QueryRow(`
		INSERT INTO images (registry_id, repository, description, star_count, pull_count, is_official, last_synced_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
		ON CONFLICT (registry_id, repository) DO UPDATE SET
			description = $3,
			star_count = $4,
			pull_count = $5,
			is_official = $6,
			last_synced_at = NOW()
		RETURNING id
	`, registryID, imageName, info.Description, info.StarCount, info.PullCount, info.IsOfficial).Scan(&imageID)
	if err != nil {
		return err
	}

	tags, err := reg.ListTags(ctx, repository, registry.ListTagsOptions{PageSize: 20})
	if err != nil {
		rs.logger.Error("failed to list tags", "image", imageName, "error", err)
		return nil
	}

	for _, tag := range tags {
		var pushedAt *time.Time
		if !tag.PushedAt.IsZero() {
			pushedAt = &tag.PushedAt
		}

		var tagID int
		err = rs.db.QueryRow(`
			INSERT INTO image_tags (image_id, tag, digest, size_bytes, pushed_at, last_synced_at)
			VALUES ($1, $2, $3, $4, $5, NOW())
			ON CONFLICT (image_id, tag) DO UPDATE SET
				digest = $3,
				size_bytes = $4,
				pushed_at = $5,
				last_synced_at = NOW()
			RETURNING id
		`, imageID, tag.Name, tag.Digest, tag.SizeBytes, pushedAt).Scan(&tagID)
		if err != nil {
			rs.logger.Error("failed to upsert tag", "tag", tag.Name, "error", err)
			continue
		}

		rs.db.Exec("DELETE FROM image_architectures WHERE tag_id = $1", tagID)
		for _, arch := range tag.Architectures {
			rs.db.Exec(`
				INSERT INTO image_architectures (tag_id, os, architecture, variant, digest, size_bytes)
				VALUES ($1, $2, $3, $4, $5, $6)
			`, tagID, arch.OS, arch.Architecture, arch.Variant, arch.Digest, arch.SizeBytes)
		}
	}

	return nil
}
