package handlers

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/priz/devarch-api/internal/api/respond"
	"github.com/priz/devarch-api/pkg/registry"
)

type RegistryHandler struct {
	db      *sql.DB
	manager *registry.Manager
	cache   *registry.Cache
}

func NewRegistryHandler(db *sql.DB, manager *registry.Manager) *RegistryHandler {
	return &RegistryHandler{
		db:      db,
		manager: manager,
		cache:   registry.NewCache(),
	}
}

type imageResponse struct {
	Repository  string `json:"repository"`
	Description string `json:"description,omitempty"`
	StarCount   *int   `json:"star_count,omitempty"`
	PullCount   *int64 `json:"pull_count,omitempty"`
	IsOfficial  bool   `json:"is_official"`
	Registry    string `json:"registry"`
}

// GetImage godoc
// @Summary      Get service image metadata
// @Description  Returns cached registry metadata for a service's container image
// @Tags         registries
// @Produce      json
// @Param        name path string true "Service name"
// @Success      200 {object} respond.SuccessEnvelope{data=imageResponse}
// @Failure      404 {object} respond.ErrorEnvelope
// @Failure      500 {object} respond.ErrorEnvelope
// @Router       /services/{name}/image [get]
// @Security     ApiKeyAuth
func (h *RegistryHandler) GetImage(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	var imageName string
	err := h.db.QueryRow("SELECT image_name FROM services WHERE name = $1", name).Scan(&imageName)
	if err == sql.ErrNoRows {
		respond.NotFound(w, r, "service", name)
		return
	}
	if err != nil {
		respond.InternalError(w, r, err)
		return
	}

	var img imageResponse
	var desc sql.NullString
	var stars sql.NullInt32
	var pulls sql.NullInt64
	var registryName string

	err = h.db.QueryRow(`
		SELECT i.repository, i.description, i.star_count, i.pull_count, i.is_official, r.name
		FROM images i
		JOIN registries r ON i.registry_id = r.id
		WHERE i.repository = $1
	`, imageName).Scan(&img.Repository, &desc, &stars, &pulls, &img.IsOfficial, &registryName)
	if err == sql.ErrNoRows {
		img.Repository = imageName
		img.Registry = "unknown"
		respond.JSON(w, r, http.StatusOK,img)
		return
	}
	if err != nil {
		respond.InternalError(w, r, err)
		return
	}

	if desc.Valid {
		img.Description = desc.String
	}
	if stars.Valid {
		s := int(stars.Int32)
		img.StarCount = &s
	}
	if pulls.Valid {
		img.PullCount = &pulls.Int64
	}
	img.Registry = registryName

	respond.JSON(w, r, http.StatusOK,img)
}

type tagResponse struct {
	Tag       string `json:"tag"`
	Digest    string `json:"digest,omitempty"`
	SizeBytes *int64 `json:"size_bytes,omitempty"`
	PushedAt  string `json:"pushed_at,omitempty"`
}

// GetTags godoc
// @Summary      Get service image tags
// @Description  Returns cached tags for a service's container image (up to 20 most recent)
// @Tags         registries
// @Produce      json
// @Param        name path string true "Service name"
// @Success      200 {object} respond.SuccessEnvelope{data=[]tagResponse}
// @Failure      404 {object} respond.ErrorEnvelope
// @Failure      500 {object} respond.ErrorEnvelope
// @Router       /services/{name}/tags [get]
// @Security     ApiKeyAuth
func (h *RegistryHandler) GetTags(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	var imageName string
	err := h.db.QueryRow("SELECT image_name FROM services WHERE name = $1", name).Scan(&imageName)
	if err == sql.ErrNoRows {
		respond.NotFound(w, r, "service", name)
		return
	}
	if err != nil {
		respond.InternalError(w, r, err)
		return
	}

	rows, err := h.db.Query(`
		SELECT t.tag, t.digest, t.size_bytes, t.pushed_at
		FROM image_tags t
		JOIN images i ON t.image_id = i.id
		WHERE i.repository = $1
		ORDER BY t.pushed_at DESC NULLS LAST
		LIMIT 20
	`, imageName)
	if err != nil {
		respond.InternalError(w, r, err)
		return
	}
	defer rows.Close()

	var tags []tagResponse
	for rows.Next() {
		var t tagResponse
		var digest sql.NullString
		var size sql.NullInt64
		var pushed sql.NullTime

		if err := rows.Scan(&t.Tag, &digest, &size, &pushed); err != nil {
			continue
		}
		if digest.Valid {
			t.Digest = digest.String
		}
		if size.Valid {
			t.SizeBytes = &size.Int64
		}
		if pushed.Valid {
			t.PushedAt = pushed.Time.Format("2006-01-02T15:04:05Z")
		}
		tags = append(tags, t)
	}

	respond.JSON(w, r, http.StatusOK,tags)
}

type vulnResponse struct {
	CVEID            string   `json:"cve_id"`
	Severity         string   `json:"severity"`
	Title            string   `json:"title,omitempty"`
	CVSSScore        *float64 `json:"cvss_score,omitempty"`
	PackageName      string   `json:"package_name"`
	InstalledVersion string   `json:"installed_version"`
	FixedVersion     string   `json:"fixed_version,omitempty"`
}

// GetVulnerabilities godoc
// @Summary      Get service image vulnerabilities
// @Description  Returns cached CVE vulnerabilities for a service's image tag
// @Tags         registries
// @Produce      json
// @Param        name path string true "Service name"
// @Success      200 {object} respond.SuccessEnvelope{data=[]vulnResponse}
// @Failure      404 {object} respond.ErrorEnvelope
// @Failure      500 {object} respond.ErrorEnvelope
// @Router       /services/{name}/vulnerabilities [get]
// @Security     ApiKeyAuth
func (h *RegistryHandler) GetVulnerabilities(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	var imageName, imageTag string
	err := h.db.QueryRow("SELECT image_name, image_tag FROM services WHERE name = $1", name).Scan(&imageName, &imageTag)
	if err == sql.ErrNoRows {
		respond.NotFound(w, r, "service", name)
		return
	}
	if err != nil {
		respond.InternalError(w, r, err)
		return
	}

	rows, err := h.db.Query(`
		SELECT v.cve_id, v.severity, v.title, v.cvss_score,
			tv.package_name, tv.installed_version, tv.fixed_version
		FROM image_tag_vulnerabilities tv
		JOIN vulnerabilities v ON tv.vulnerability_id = v.id
		JOIN image_tags t ON tv.tag_id = t.id
		JOIN images i ON t.image_id = i.id
		WHERE i.repository = $1 AND t.tag = $2
		ORDER BY v.cvss_score DESC NULLS LAST
	`, imageName, imageTag)
	if err != nil {
		respond.InternalError(w, r, err)
		return
	}
	defer rows.Close()

	var vulns []vulnResponse
	for rows.Next() {
		var vr vulnResponse
		var title sql.NullString
		var cvss sql.NullFloat64
		var fixed sql.NullString

		if err := rows.Scan(&vr.CVEID, &vr.Severity, &title, &cvss, &vr.PackageName, &vr.InstalledVersion, &fixed); err != nil {
			continue
		}
		if title.Valid {
			vr.Title = title.String
		}
		if cvss.Valid {
			vr.CVSSScore = &cvss.Float64
		}
		if fixed.Valid {
			vr.FixedVersion = fixed.String
		}
		vulns = append(vulns, vr)
	}

	respond.JSON(w, r, http.StatusOK,vulns)
}

type registryListItem struct {
	Name    string `json:"name"`
	BaseURL string `json:"base_url"`
	Enabled bool   `json:"enabled"`
}

// ListRegistries godoc
// @Summary      List container registries
// @Description  Returns all configured container registries with base URLs and enabled status
// @Tags         registries
// @Produce      json
// @Success      200 {object} respond.SuccessEnvelope{data=[]registryListItem}
// @Failure      500 {object} respond.ErrorEnvelope
// @Router       /registries [get]
// @Security     ApiKeyAuth
func (h *RegistryHandler) ListRegistries(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.Query("SELECT name, base_url, enabled FROM registries ORDER BY name")
	if err != nil {
		respond.InternalError(w, r, err)
		return
	}
	defer rows.Close()

	var registries []registryListItem
	for rows.Next() {
		var item registryListItem
		if err := rows.Scan(&item.Name, &item.BaseURL, &item.Enabled); err != nil {
			continue
		}
		registries = append(registries, item)
	}

	if registries == nil {
		registries = []registryListItem{}
	}

	respond.JSON(w, r, http.StatusOK,registries)
}

// SearchImages godoc
// @Summary      Search registry for images
// @Description  Searches a container registry for images matching a query string with pagination
// @Tags         registries
// @Produce      json
// @Param        registry path string true "Registry name"
// @Param        q query string true "Search query"
// @Param        page_size query int false "Page size"
// @Param        page query int false "Page number"
// @Success      200 {object} respond.SuccessEnvelope{data=[]registry.SearchResult}
// @Failure      400 {object} respond.ErrorEnvelope
// @Failure      404 {object} respond.ErrorEnvelope
// @Failure      501 {object} respond.ErrorEnvelope
// @Failure      500 {object} respond.ErrorEnvelope
// @Router       /registries/{registry}/search [get]
// @Security     ApiKeyAuth
func (h *RegistryHandler) SearchImages(w http.ResponseWriter, r *http.Request) {
	registryName := chi.URLParam(r, "registry")
	query := r.URL.Query().Get("q")

	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))

	var cacheKey string
	if query == "" {
		cacheKey = fmt.Sprintf("search:%s:popular:%d:%d", registryName, pageSize, page)
	} else {
		cacheKey = fmt.Sprintf("search:%s:%s:%d:%d", registryName, query, pageSize, page)
	}
	if cached, ok := h.cache.Get(cacheKey); ok {
		respond.JSON(w, r, http.StatusOK,cached)
		return
	}

	reg := h.manager.Get(registryName)
	if reg == nil {
		respond.NotFound(w, r, "registry", registryName)
		return
	}

	results, err := reg.SearchImages(r.Context(), query, registry.SearchOptions{
		PageSize: pageSize,
		Page:     page,
	})
	if errors.Is(err, registry.ErrSearchNotSupported) {
		respond.Error(w, r, http.StatusNotImplemented, "not_implemented", "search not supported by this registry", nil)
		return
	}
	if err != nil {
		respond.InternalError(w, r, err)
		return
	}

	if results == nil {
		results = []registry.SearchResult{}
	}

	h.cache.Set(cacheKey, results, 5*time.Minute)

	respond.JSON(w, r, http.StatusOK,results)
}

func (h *RegistryHandler) GetImageInfoLive(w http.ResponseWriter, r *http.Request) {
	registryName := chi.URLParam(r, "registry")
	repository := chi.URLParam(r, "*")
	if repository == "" {
		respond.BadRequest(w, r, "repository path required")
		return
	}

	cacheKey := fmt.Sprintf("imageinfo:%s:%s", registryName, repository)
	if cached, ok := h.cache.Get(cacheKey); ok {
		respond.JSON(w, r, http.StatusOK,cached)
		return
	}

	reg := h.manager.Get(registryName)
	if reg == nil {
		respond.NotFound(w, r, "registry", registryName)
		return
	}

	info, err := reg.GetImageInfo(r.Context(), repository)
	if err != nil {
		respond.InternalError(w, r, err)
		return
	}

	h.cache.Set(cacheKey, info, 10*time.Minute)

	respond.JSON(w, r, http.StatusOK,info)
}

type liveTagResponse struct {
	Name          string              `json:"name"`
	Digest        string              `json:"digest,omitempty"`
	SizeBytes     int64               `json:"size_bytes"`
	PushedAt      string              `json:"pushed_at,omitempty"`
	Architectures []registry.ArchInfo `json:"architectures,omitempty"`
}

func (h *RegistryHandler) ListImageTags(w http.ResponseWriter, r *http.Request) {
	registryName := chi.URLParam(r, "registry")
	repository := chi.URLParam(r, "*")
	if repository == "" {
		respond.BadRequest(w, r, "repository path required")
		return
	}

	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))

	cacheKey := fmt.Sprintf("tags:%s:%s:%d:%d", registryName, repository, pageSize, page)
	if cached, ok := h.cache.Get(cacheKey); ok {
		respond.JSON(w, r, http.StatusOK,cached)
		return
	}

	reg := h.manager.Get(registryName)
	if reg == nil {
		respond.NotFound(w, r, "registry", registryName)
		return
	}

	tags, err := reg.ListTags(r.Context(), repository, registry.ListTagsOptions{
		PageSize: pageSize,
		Page:     page,
	})
	if err != nil {
		respond.InternalError(w, r, err)
		return
	}

	resp := make([]liveTagResponse, len(tags))
	for i, t := range tags {
		resp[i] = liveTagResponse{
			Name:          t.Name,
			Digest:        t.Digest,
			SizeBytes:     t.SizeBytes,
			Architectures: t.Architectures,
		}
		if !t.PushedAt.IsZero() {
			resp[i].PushedAt = t.PushedAt.Format(time.RFC3339)
		}
	}

	h.cache.Set(cacheKey, resp, 2*time.Minute)

	respond.JSON(w, r, http.StatusOK,resp)
}

// ImageRoute godoc
// @Summary      Get live image info or tags
// @Description  Routes to image info or tags list based on path suffix (live registry data, not cached)
// @Tags         registries
// @Produce      json
// @Param        registry path string true "Registry name"
// @Param        path path string true "Image path (e.g., library/nginx or library/nginx/tags)"
// @Success      200 {object} respond.SuccessEnvelope
// @Failure      400 {object} respond.ErrorEnvelope
// @Failure      404 {object} respond.ErrorEnvelope
// @Failure      500 {object} respond.ErrorEnvelope
// @Router       /registries/{registry}/images/{path} [get]
// @Security     ApiKeyAuth
func (h *RegistryHandler) ImageRoute(w http.ResponseWriter, r *http.Request) {
	wildcard := chi.URLParam(r, "*")
	if strings.HasSuffix(wildcard, "/tags") {
		repo := strings.TrimSuffix(wildcard, "/tags")
		h.handleListImageTags(w, r, chi.URLParam(r, "registry"), repo)
		return
	}
	h.GetImageInfoLive(w, r)
}

func (h *RegistryHandler) handleListImageTags(w http.ResponseWriter, r *http.Request, registryName, repository string) {
	if repository == "" {
		respond.BadRequest(w, r, "repository path required")
		return
	}

	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))

	cacheKey := fmt.Sprintf("tags:%s:%s:%d:%d", registryName, repository, pageSize, page)
	if cached, ok := h.cache.Get(cacheKey); ok {
		respond.JSON(w, r, http.StatusOK,cached)
		return
	}

	reg := h.manager.Get(registryName)
	if reg == nil {
		respond.NotFound(w, r, "registry", registryName)
		return
	}

	tags, err := reg.ListTags(r.Context(), repository, registry.ListTagsOptions{
		PageSize: pageSize,
		Page:     page,
	})
	if err != nil {
		respond.InternalError(w, r, err)
		return
	}

	resp := make([]liveTagResponse, len(tags))
	for i, t := range tags {
		resp[i] = liveTagResponse{
			Name:          t.Name,
			Digest:        t.Digest,
			SizeBytes:     t.SizeBytes,
			Architectures: t.Architectures,
		}
		if !t.PushedAt.IsZero() {
			resp[i].PushedAt = t.PushedAt.Format(time.RFC3339)
		}
	}

	h.cache.Set(cacheKey, resp, 2*time.Minute)

	respond.JSON(w, r, http.StatusOK,resp)
}
