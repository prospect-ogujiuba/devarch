package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type RegistryHandler struct {
	db *sql.DB
}

func NewRegistryHandler(db *sql.DB) *RegistryHandler {
	return &RegistryHandler{db: db}
}

type imageResponse struct {
	Repository  string  `json:"repository"`
	Description string  `json:"description,omitempty"`
	StarCount   *int    `json:"star_count,omitempty"`
	PullCount   *int64  `json:"pull_count,omitempty"`
	IsOfficial  bool    `json:"is_official"`
	Registry    string  `json:"registry"`
}

func (h *RegistryHandler) GetImage(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	var imageName string
	err := h.db.QueryRow("SELECT image_name FROM services WHERE name = $1", name).Scan(&imageName)
	if err == sql.ErrNoRows {
		http.Error(w, "service not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(img)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(img)
}

type tagResponse struct {
	Tag       string `json:"tag"`
	Digest    string `json:"digest,omitempty"`
	SizeBytes *int64 `json:"size_bytes,omitempty"`
	PushedAt  string `json:"pushed_at,omitempty"`
}

func (h *RegistryHandler) GetTags(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	var imageName string
	err := h.db.QueryRow("SELECT image_name FROM services WHERE name = $1", name).Scan(&imageName)
	if err == sql.ErrNoRows {
		http.Error(w, "service not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tags)
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

func (h *RegistryHandler) GetVulnerabilities(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	var imageName, imageTag string
	err := h.db.QueryRow("SELECT image_name, image_tag FROM services WHERE name = $1", name).Scan(&imageName, &imageTag)
	if err == sql.ErrNoRows {
		http.Error(w, "service not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(vulns)
}
