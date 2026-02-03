package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	_ "github.com/lib/pq"
)

func main() {
	var (
		dbURL  = flag.String("db", "", "Database URL (or set DATABASE_URL env)")
		outFile = flag.String("out", "", "Output file path (default: stdout)")
	)
	flag.Parse()

	if *dbURL == "" {
		*dbURL = os.Getenv("DATABASE_URL")
	}
	if *dbURL == "" {
		*dbURL = "postgres://devarch:devarch@localhost:5432/devarch?sslmode=disable"
	}

	db, err := sql.Open("postgres", *dbURL)
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("failed to ping: %v", err)
	}

	var b strings.Builder
	b.WriteString("BEGIN;\n\n")

	if err := exportCategories(db, &b); err != nil {
		log.Fatalf("export categories: %v", err)
	}
	if err := exportServices(db, &b); err != nil {
		log.Fatalf("export services: %v", err)
	}
	if err := exportPorts(db, &b); err != nil {
		log.Fatalf("export ports: %v", err)
	}
	if err := exportVolumes(db, &b); err != nil {
		log.Fatalf("export volumes: %v", err)
	}
	if err := exportEnvVars(db, &b); err != nil {
		log.Fatalf("export env_vars: %v", err)
	}
	if err := exportLabels(db, &b); err != nil {
		log.Fatalf("export labels: %v", err)
	}
	if err := exportHealthchecks(db, &b); err != nil {
		log.Fatalf("export healthchecks: %v", err)
	}
	if err := exportDomains(db, &b); err != nil {
		log.Fatalf("export domains: %v", err)
	}
	if err := exportDependencies(db, &b); err != nil {
		log.Fatalf("export dependencies: %v", err)
	}
	if err := exportConfigFiles(db, &b); err != nil {
		log.Fatalf("export config_files: %v", err)
	}

	b.WriteString("COMMIT;\n")

	if *outFile != "" {
		if err := os.WriteFile(*outFile, []byte(b.String()), 0644); err != nil {
			log.Fatalf("write file: %v", err)
		}
		log.Printf("wrote seed to %s", *outFile)
	} else {
		fmt.Print(b.String())
	}
}

func esc(s string) string {
	return strings.ReplaceAll(s, "'", "''")
}

func nullStr(ns sql.NullString) string {
	if !ns.Valid {
		return "NULL"
	}
	return fmt.Sprintf("'%s'", esc(ns.String))
}

func nullInt(ni sql.NullInt64) string {
	if !ni.Valid {
		return "NULL"
	}
	return fmt.Sprintf("%d", ni.Int64)
}

func exportCategories(db *sql.DB, b *strings.Builder) error {
	rows, err := db.Query(`SELECT name, display_name, color, startup_order FROM categories ORDER BY id`)
	if err != nil {
		return err
	}
	defer rows.Close()

	b.WriteString("-- Categories\n")
	for rows.Next() {
		var name string
		var displayName, color sql.NullString
		var startupOrder sql.NullInt64
		if err := rows.Scan(&name, &displayName, &color, &startupOrder); err != nil {
			return err
		}
		b.WriteString(fmt.Sprintf(
			"INSERT INTO categories (name, display_name, color, startup_order) VALUES ('%s', %s, %s, %s) ON CONFLICT (name) DO UPDATE SET display_name = EXCLUDED.display_name, color = EXCLUDED.color, startup_order = EXCLUDED.startup_order, updated_at = NOW();\n",
			esc(name), nullStr(displayName), nullStr(color), nullInt(startupOrder),
		))
	}
	b.WriteString("\n")
	return rows.Err()
}

func exportServices(db *sql.DB, b *strings.Builder) error {
	rows, err := db.Query(`
		SELECT s.name, c.name, s.image_name, s.image_tag, s.restart_policy,
		       s.command, s.user_spec, s.enabled, s.compose_overrides
		FROM services s
		LEFT JOIN categories c ON c.id = s.category_id
		ORDER BY s.id
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	b.WriteString("-- Services\n")
	for rows.Next() {
		var name, imageName string
		var catName, imageTag, restartPolicy, command, userSpec sql.NullString
		var enabled bool
		var composeOverrides sql.NullString
		if err := rows.Scan(&name, &catName, &imageName, &imageTag, &restartPolicy, &command, &userSpec, &enabled, &composeOverrides); err != nil {
			return err
		}

		catExpr := "NULL"
		if catName.Valid {
			catExpr = fmt.Sprintf("(SELECT id FROM categories WHERE name = '%s')", esc(catName.String))
		}

		coVal := "'{}'::jsonb"
		if composeOverrides.Valid && composeOverrides.String != "" && composeOverrides.String != "{}" {
			coVal = fmt.Sprintf("'%s'::jsonb", esc(composeOverrides.String))
		}

		b.WriteString(fmt.Sprintf(
			"INSERT INTO services (name, category_id, image_name, image_tag, restart_policy, command, user_spec, enabled, compose_overrides) VALUES ('%s', %s, '%s', %s, %s, %s, %s, %t, %s) ON CONFLICT (name) DO UPDATE SET category_id = EXCLUDED.category_id, image_name = EXCLUDED.image_name, image_tag = EXCLUDED.image_tag, restart_policy = EXCLUDED.restart_policy, command = EXCLUDED.command, user_spec = EXCLUDED.user_spec, compose_overrides = EXCLUDED.compose_overrides, updated_at = NOW();\n",
			esc(name), catExpr, esc(imageName), nullStr(imageTag), nullStr(restartPolicy),
			nullStr(command), nullStr(userSpec), enabled, coVal,
		))
	}
	b.WriteString("\n")
	return rows.Err()
}

func exportPorts(db *sql.DB, b *strings.Builder) error {
	rows, err := db.Query(`
		SELECT s.name, p.host_ip, p.host_port, p.container_port, p.protocol
		FROM service_ports p
		JOIN services s ON s.id = p.service_id
		ORDER BY s.name, p.host_port
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	b.WriteString("-- Ports\n")
	for rows.Next() {
		var svcName, hostIP, protocol string
		var hostPort, containerPort int
		if err := rows.Scan(&svcName, &hostIP, &hostPort, &containerPort, &protocol); err != nil {
			return err
		}
		b.WriteString(fmt.Sprintf(
			"INSERT INTO service_ports (service_id, host_ip, host_port, container_port, protocol) VALUES ((SELECT id FROM services WHERE name = '%s'), '%s', %d, %d, '%s') ON CONFLICT (host_ip, host_port) DO NOTHING;\n",
			esc(svcName), esc(hostIP), hostPort, containerPort, esc(protocol),
		))
	}
	b.WriteString("\n")
	return rows.Err()
}

func exportVolumes(db *sql.DB, b *strings.Builder) error {
	rows, err := db.Query(`
		SELECT s.name, v.volume_type, v.source, v.target, v.read_only, v.is_external
		FROM service_volumes v
		JOIN services s ON s.id = v.service_id
		ORDER BY s.name, v.id
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	b.WriteString("-- Volumes\n")
	for rows.Next() {
		var svcName, volType, source, target string
		var readOnly, isExternal bool
		if err := rows.Scan(&svcName, &volType, &source, &target, &readOnly, &isExternal); err != nil {
			return err
		}
		b.WriteString(fmt.Sprintf(
			"INSERT INTO service_volumes (service_id, volume_type, source, target, read_only, is_external) SELECT (SELECT id FROM services WHERE name = '%s'), '%s', '%s', '%s', %t, %t WHERE NOT EXISTS (SELECT 1 FROM service_volumes WHERE service_id = (SELECT id FROM services WHERE name = '%s') AND source = '%s' AND target = '%s');\n",
			esc(svcName), esc(volType), esc(source), esc(target), readOnly, isExternal,
			esc(svcName), esc(source), esc(target),
		))
	}
	b.WriteString("\n")
	return rows.Err()
}

func exportEnvVars(db *sql.DB, b *strings.Builder) error {
	rows, err := db.Query(`
		SELECT s.name, e.key, e.value, e.is_secret
		FROM service_env_vars e
		JOIN services s ON s.id = e.service_id
		ORDER BY s.name, e.key
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	b.WriteString("-- Environment Variables\n")
	for rows.Next() {
		var svcName, key string
		var value sql.NullString
		var isSecret bool
		if err := rows.Scan(&svcName, &key, &value, &isSecret); err != nil {
			return err
		}
		b.WriteString(fmt.Sprintf(
			"INSERT INTO service_env_vars (service_id, key, value, is_secret) VALUES ((SELECT id FROM services WHERE name = '%s'), '%s', %s, %t) ON CONFLICT (service_id, key) DO UPDATE SET value = EXCLUDED.value, is_secret = EXCLUDED.is_secret;\n",
			esc(svcName), esc(key), nullStr(value), isSecret,
		))
	}
	b.WriteString("\n")
	return rows.Err()
}

func exportLabels(db *sql.DB, b *strings.Builder) error {
	rows, err := db.Query(`
		SELECT s.name, l.key, l.value
		FROM service_labels l
		JOIN services s ON s.id = l.service_id
		ORDER BY s.name, l.key
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	b.WriteString("-- Labels\n")
	for rows.Next() {
		var svcName, key string
		var value sql.NullString
		if err := rows.Scan(&svcName, &key, &value); err != nil {
			return err
		}
		b.WriteString(fmt.Sprintf(
			"INSERT INTO service_labels (service_id, key, value) VALUES ((SELECT id FROM services WHERE name = '%s'), '%s', %s) ON CONFLICT (service_id, key) DO UPDATE SET value = EXCLUDED.value;\n",
			esc(svcName), esc(key), nullStr(value),
		))
	}
	b.WriteString("\n")
	return rows.Err()
}

func exportHealthchecks(db *sql.DB, b *strings.Builder) error {
	rows, err := db.Query(`
		SELECT s.name, h.test, h.interval_seconds, h.timeout_seconds, h.retries, h.start_period_seconds
		FROM service_healthchecks h
		JOIN services s ON s.id = h.service_id
		ORDER BY s.name
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	b.WriteString("-- Healthchecks\n")
	for rows.Next() {
		var svcName, test string
		var interval, timeout, retries, startPeriod int
		if err := rows.Scan(&svcName, &test, &interval, &timeout, &retries, &startPeriod); err != nil {
			return err
		}
		b.WriteString(fmt.Sprintf(
			"INSERT INTO service_healthchecks (service_id, test, interval_seconds, timeout_seconds, retries, start_period_seconds) VALUES ((SELECT id FROM services WHERE name = '%s'), '%s', %d, %d, %d, %d) ON CONFLICT (service_id) DO UPDATE SET test = EXCLUDED.test, interval_seconds = EXCLUDED.interval_seconds, timeout_seconds = EXCLUDED.timeout_seconds, retries = EXCLUDED.retries, start_period_seconds = EXCLUDED.start_period_seconds;\n",
			esc(svcName), esc(test), interval, timeout, retries, startPeriod,
		))
	}
	b.WriteString("\n")
	return rows.Err()
}

func exportDomains(db *sql.DB, b *strings.Builder) error {
	rows, err := db.Query(`
		SELECT s.name, d.domain, d.proxy_port
		FROM service_domains d
		JOIN services s ON s.id = d.service_id
		ORDER BY s.name, d.domain
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	b.WriteString("-- Domains\n")
	for rows.Next() {
		var svcName, domain string
		var proxyPort sql.NullInt64
		if err := rows.Scan(&svcName, &domain, &proxyPort); err != nil {
			return err
		}
		b.WriteString(fmt.Sprintf(
			"INSERT INTO service_domains (service_id, domain, proxy_port) VALUES ((SELECT id FROM services WHERE name = '%s'), '%s', %s) ON CONFLICT (domain) DO UPDATE SET proxy_port = EXCLUDED.proxy_port;\n",
			esc(svcName), esc(domain), nullInt(proxyPort),
		))
	}
	b.WriteString("\n")
	return rows.Err()
}

func exportDependencies(db *sql.DB, b *strings.Builder) error {
	rows, err := db.Query(`
		SELECT s.name, dep.name, d.condition
		FROM service_dependencies d
		JOIN services s ON s.id = d.service_id
		JOIN services dep ON dep.id = d.depends_on_service_id
		ORDER BY s.name, dep.name
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	b.WriteString("-- Dependencies\n")
	for rows.Next() {
		var svcName, depName, condition string
		if err := rows.Scan(&svcName, &depName, &condition); err != nil {
			return err
		}
		b.WriteString(fmt.Sprintf(
			"INSERT INTO service_dependencies (service_id, depends_on_service_id, condition) VALUES ((SELECT id FROM services WHERE name = '%s'), (SELECT id FROM services WHERE name = '%s'), '%s') ON CONFLICT (service_id, depends_on_service_id) DO NOTHING;\n",
			esc(svcName), esc(depName), esc(condition),
		))
	}
	b.WriteString("\n")
	return rows.Err()
}

func exportConfigFiles(db *sql.DB, b *strings.Builder) error {
	rows, err := db.Query(`
		SELECT s.name, cf.file_path, cf.content, cf.file_mode, cf.is_template
		FROM service_config_files cf
		JOIN services s ON s.id = cf.service_id
		ORDER BY s.name, cf.file_path
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	b.WriteString("-- Config Files\n")
	for rows.Next() {
		var svcName, filePath, content, fileMode string
		var isTemplate bool
		if err := rows.Scan(&svcName, &filePath, &content, &fileMode, &isTemplate); err != nil {
			return err
		}
		b.WriteString(fmt.Sprintf(
			"INSERT INTO service_config_files (service_id, file_path, content, file_mode, is_template) VALUES ((SELECT id FROM services WHERE name = '%s'), '%s', '%s', '%s', %t) ON CONFLICT (service_id, file_path) DO UPDATE SET content = EXCLUDED.content, file_mode = EXCLUDED.file_mode, is_template = EXCLUDED.is_template, updated_at = NOW();\n",
			esc(svcName), esc(filePath), esc(content), esc(fileMode), isTemplate,
		))
	}
	b.WriteString("\n")
	return rows.Err()
}
