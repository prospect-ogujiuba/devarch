package sync

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"time"
)

type TrivyScanner struct {
	db *sql.DB
}

func NewTrivyScanner(db *sql.DB) *TrivyScanner {
	return &TrivyScanner{db: db}
}

type trivyReport struct {
	Results []trivyResult `json:"Results"`
}

type trivyResult struct {
	Target          string             `json:"Target"`
	Vulnerabilities []trivyVulnerability `json:"Vulnerabilities"`
}

type trivyVulnerability struct {
	VulnerabilityID  string  `json:"VulnerabilityID"`
	Severity         string  `json:"Severity"`
	Title            string  `json:"Title"`
	Description      string  `json:"Description"`
	PkgName          string  `json:"PkgName"`
	InstalledVersion string  `json:"InstalledVersion"`
	FixedVersion     string  `json:"FixedVersion"`
	PublishedDate    string  `json:"PublishedDate"`
	CVSS             map[string]struct {
		V3Score float64 `json:"V3Score"`
	} `json:"CVSS"`
}

func (ts *TrivyScanner) ScanAll(ctx context.Context) error {
	rows, err := ts.db.Query(`
		SELECT s.image_name, s.image_tag
		FROM services s
		WHERE s.enabled = true
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	type imageTag struct {
		name string
		tag  string
	}
	var images []imageTag
	for rows.Next() {
		var it imageTag
		if err := rows.Scan(&it.name, &it.tag); err != nil {
			continue
		}
		images = append(images, it)
	}

	for _, img := range images {
		if err := ts.ScanImage(ctx, img.name, img.tag); err != nil {
			log.Printf("failed to scan %s:%s: %v", img.name, img.tag, err)
		}
	}

	return nil
}

func (ts *TrivyScanner) ScanImage(ctx context.Context, imageName, tag string) error {
	image := fmt.Sprintf("%s:%s", imageName, tag)

	cmd := exec.CommandContext(ctx, "trivy", "image", "--format", "json", "--severity", "HIGH,CRITICAL", image)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("trivy scan failed: %s", stderr.String())
	}

	var report trivyReport
	if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
		return fmt.Errorf("parse trivy output: %w", err)
	}

	var tagID int
	err := ts.db.QueryRow(`
		SELECT t.id FROM image_tags t
		JOIN images i ON t.image_id = i.id
		WHERE i.repository = $1 AND t.tag = $2
	`, imageName, tag).Scan(&tagID)
	if err == sql.ErrNoRows {
		return nil
	}
	if err != nil {
		return err
	}

	for _, result := range report.Results {
		for _, vuln := range result.Vulnerabilities {
			var vulnID int
			var cvss *float64
			if len(vuln.CVSS) > 0 {
				for _, c := range vuln.CVSS {
					score := c.V3Score
					cvss = &score
					break
				}
			}

			var publishedAt *time.Time
			if vuln.PublishedDate != "" {
				if t, err := time.Parse(time.RFC3339, vuln.PublishedDate); err == nil {
					publishedAt = &t
				}
			}

			err = ts.db.QueryRow(`
				INSERT INTO vulnerabilities (cve_id, severity, title, description, cvss_score, published_at)
				VALUES ($1, $2, $3, $4, $5, $6)
				ON CONFLICT (cve_id) DO UPDATE SET
					severity = $2,
					title = $3,
					description = $4,
					cvss_score = $5,
					published_at = $6
				RETURNING id
			`, vuln.VulnerabilityID, vuln.Severity, vuln.Title, vuln.Description, cvss, publishedAt).Scan(&vulnID)
			if err != nil {
				log.Printf("failed to upsert vulnerability %s: %v", vuln.VulnerabilityID, err)
				continue
			}

			ts.db.Exec(`
				INSERT INTO image_tag_vulnerabilities (tag_id, vulnerability_id, package_name, installed_version, fixed_version)
				VALUES ($1, $2, $3, $4, $5)
				ON CONFLICT (tag_id, vulnerability_id, package_name) DO UPDATE SET
					installed_version = $4,
					fixed_version = $5,
					scanned_at = NOW()
			`, tagID, vulnID, vuln.PkgName, vuln.InstalledVersion, vuln.FixedVersion)
		}
	}

	return nil
}
