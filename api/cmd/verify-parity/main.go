package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	_ "github.com/lib/pq"
	"github.com/priz/devarch-api/internal/compose"
	"github.com/priz/devarch-api/pkg/models"
)

func main() {
	var (
		dbURL        = flag.String("db", "", "Database URL (or set DATABASE_URL env)")
		composeDir   = flag.String("compose-dir", "", "Path to compose directory")
		projectRoot  = flag.String("project-root", "", "Project root for resolving relative paths")
		service      = flag.String("service", "", "Verify single service by name (optional)")
		verbose      = flag.Bool("verbose", false, "Print details for passing services too")
		jsonOut      = flag.Bool("json", false, "Print machine-readable JSON output")
		strict       = flag.Bool("strict", false, "Fail when any whitelist exception is used")
		whitelist    = flag.String("whitelist", "", "Path to parity whitelist JSON file")
		goldenSource = flag.String("golden-services", "", "Path to golden services JSON file")
	)
	flag.Parse()

	if *dbURL == "" {
		*dbURL = os.Getenv("DATABASE_URL")
	}
	if *dbURL == "" {
		*dbURL = "postgres://devarch:devarch@localhost:5432/devarch?sslmode=disable"
	}

	if *composeDir == "" {
		log.Fatal("compose-dir is required")
	}

	if *projectRoot == "" {
		log.Fatal("project-root is required")
	}

	if *whitelist == "" {
		*whitelist = filepath.Join(*projectRoot, "api", "cmd", "verify-parity", "whitelist.json")
	}
	if *goldenSource == "" {
		*goldenSource = filepath.Join(*projectRoot, "api", "cmd", "verify-parity", "golden-services.json")
	}

	whitelistEntries, err := loadWhitelist(*whitelist)
	if err != nil {
		log.Fatalf("failed to load whitelist: %v", err)
	}

	goldenServices, err := loadGoldenServices(*goldenSource)
	if err != nil {
		log.Fatalf("failed to load golden services: %v", err)
	}

	if err := validateWhitelistAgainstGolden(whitelistEntries, goldenServices); err != nil {
		log.Fatalf("invalid whitelist: %v", err)
	}

	db, err := sql.Open("postgres", *dbURL)
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("failed to ping: %v", err)
	}

	// Run fresh import
	log.Println("importing compose files...")
	importer := compose.NewImporterWithRoot(db, *composeDir, *projectRoot)

	if err := importer.ImportAll(); err != nil {
		log.Fatalf("import failed: %v", err)
	}

	log.Println("import complete")
	log.Println("")

	// Get services to verify
	var services []serviceRecord
	if *service != "" {
		var svc serviceRecord
		err := db.QueryRow(`
			SELECT s.id, s.name, c.name
			FROM services s
			JOIN categories c ON s.category_id = c.id
			WHERE s.name = $1
		`, *service).Scan(&svc.ID, &svc.Name, &svc.Category)
		if err != nil {
			log.Fatalf("service not found: %s", *service)
		}
		services = append(services, svc)
	} else {
		rows, err := db.Query(`
			SELECT s.id, s.name, c.name
			FROM services s
			JOIN categories c ON s.category_id = c.id
			ORDER BY c.startup_order, s.name
		`)
		if err != nil {
			log.Fatalf("failed to query services: %v", err)
		}
		defer rows.Close()

		for rows.Next() {
			var svc serviceRecord
			if err := rows.Scan(&svc.ID, &svc.Name, &svc.Category); err != nil {
				log.Fatalf("scan error: %v", err)
			}
			services = append(services, svc)
		}
	}

	// Run verification for each service
	gen := compose.NewGenerator(db, "", nil)
	gen.SetProjectRoot(*projectRoot)
	gen.SetHostProjectRoot(*projectRoot)

	passed := 0
	failed := 0
	whitelisted := 0
	serviceReports := make([]serviceParityReport, 0, len(services))

	for _, svc := range services {
		result := verifyService(db, gen, svc, *composeDir, *projectRoot, *verbose)
		audit := applyWhitelistGovernance(svc.Name, result.Messages, whitelistEntries, goldenServices)

		report := serviceParityReport{
			Name:        svc.Name,
			Category:    svc.Category,
			Pass:        len(audit.Failures) == 0,
			Golden:      goldenServices[strings.ToLower(svc.Name)],
			Failures:    audit.Failures,
			Whitelisted: audit.Whitelisted,
		}
		serviceReports = append(serviceReports, report)

		if len(audit.Failures) == 0 && len(audit.Whitelisted) == 0 {
			passed++
			if *verbose && !*jsonOut {
				fmt.Printf("PASS  %s (%s)\n", svc.Name, svc.Category)
			}
		} else if len(audit.Failures) == 0 {
			whitelisted++
			if *verbose && !*jsonOut {
				fmt.Printf("ALLOW %s (%s)\n", svc.Name, svc.Category)
				for _, msg := range audit.Whitelisted {
					fmt.Printf("  - %s\n", msg)
				}
				fmt.Println()
			}
		} else {
			failed++
			if !*jsonOut {
				fmt.Printf("FAIL  %s (%s)\n", svc.Name, svc.Category)
				for _, msg := range audit.Failures {
					fmt.Printf("  - %s\n", msg)
				}
				for _, msg := range audit.Whitelisted {
					fmt.Printf("  - [whitelisted] %s\n", msg)
				}
				fmt.Println()
			}
		}
	}

	exitCode := 0
	if failed > 0 || (*strict && whitelisted > 0) {
		exitCode = 1
	}

	if *jsonOut {
		summary := paritySummary{
			Passed:      passed,
			Failed:      failed,
			Whitelisted: whitelisted,
			Total:       len(services),
			Strict:      *strict,
			ExitCode:    exitCode,
		}
		goldenList := make([]string, 0, len(goldenServices))
		for svc := range goldenServices {
			goldenList = append(goldenList, svc)
		}
		sort.Strings(goldenList)

		output := parityJSONOutput{
			Summary:         summary,
			Services:        serviceReports,
			WhitelistPath:   *whitelist,
			GoldenListPath:  *goldenSource,
			GoldenServices:  goldenList,
			ComposeDir:      *composeDir,
			ProjectRoot:     *projectRoot,
			ServiceSelector: *service,
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(output); err != nil {
			log.Fatalf("failed to write JSON output: %v", err)
		}
	} else {
		fmt.Printf("Summary: %d/%d passed, %d failed, %d whitelisted\n", passed, len(services), failed, whitelisted)
		if *strict {
			fmt.Printf("Strict mode: %t\n", *strict)
		}
	}

	if exitCode != 0 {
		os.Exit(exitCode)
	}
}

type serviceRecord struct {
	ID       int
	Name     string
	Category string
}

type verifyResult struct {
	Pass     bool
	Messages []string
}

type whitelistEntry struct {
	Service string `json:"service"`
	Field   string `json:"field"`
	Reason  string `json:"reason"`
}

type whitelistFile struct {
	Entries []whitelistEntry `json:"entries"`
}

type goldenServiceFile struct {
	Services []string `json:"services"`
}

type auditedMessages struct {
	Failures    []string
	Whitelisted []string
}

type serviceParityReport struct {
	Name        string   `json:"name"`
	Category    string   `json:"category"`
	Pass        bool     `json:"pass"`
	Golden      bool     `json:"golden"`
	Failures    []string `json:"failures,omitempty"`
	Whitelisted []string `json:"whitelisted,omitempty"`
}

type paritySummary struct {
	Passed      int  `json:"passed"`
	Failed      int  `json:"failed"`
	Whitelisted int  `json:"whitelisted"`
	Total       int  `json:"total"`
	Strict      bool `json:"strict"`
	ExitCode    int  `json:"exit_code"`
}

type parityJSONOutput struct {
	Summary         paritySummary         `json:"summary"`
	Services        []serviceParityReport `json:"services"`
	WhitelistPath   string                `json:"whitelist_path"`
	GoldenListPath  string                `json:"golden_services_path"`
	GoldenServices  []string              `json:"golden_services"`
	ComposeDir      string                `json:"compose_dir"`
	ProjectRoot     string                `json:"project_root"`
	ServiceSelector string                `json:"service_selector,omitempty"`
}

func loadWhitelist(path string) (map[string]whitelistEntry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var file whitelistFile
	if err := json.Unmarshal(data, &file); err != nil {
		return nil, err
	}

	entries := make(map[string]whitelistEntry, len(file.Entries))
	for _, entry := range file.Entries {
		service := strings.TrimSpace(strings.ToLower(entry.Service))
		field := strings.TrimSpace(strings.ToLower(entry.Field))
		reason := strings.TrimSpace(entry.Reason)
		if service == "" || field == "" || reason == "" {
			return nil, fmt.Errorf("invalid whitelist entry: service, field, and reason are required")
		}
		entry.Service = service
		entry.Field = field
		entry.Reason = reason
		entries[whitelistKey(service, field)] = entry
	}

	return entries, nil
}

func loadGoldenServices(path string) (map[string]bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var file goldenServiceFile
	if err := json.Unmarshal(data, &file); err != nil {
		return nil, err
	}

	services := make(map[string]bool, len(file.Services))
	for _, name := range file.Services {
		normalized := strings.TrimSpace(strings.ToLower(name))
		if normalized == "" {
			continue
		}
		services[normalized] = true
	}

	return services, nil
}

func validateWhitelistAgainstGolden(whitelist map[string]whitelistEntry, golden map[string]bool) error {
	for _, entry := range whitelist {
		if golden[entry.Service] {
			return fmt.Errorf("service %q is golden and cannot be whitelisted", entry.Service)
		}
	}
	return nil
}

func applyWhitelistGovernance(serviceName string, messages []string, whitelist map[string]whitelistEntry, golden map[string]bool) auditedMessages {
	audit := auditedMessages{Failures: make([]string, 0), Whitelisted: make([]string, 0)}
	service := strings.ToLower(serviceName)

	for _, msg := range messages {
		field := parityField(msg)
		if _, ok := whitelist[whitelistKey(service, field)]; ok {
			if golden[service] {
				audit.Failures = append(audit.Failures, fmt.Sprintf("%s (whitelist denied for golden service)", msg))
				continue
			}
			audit.Whitelisted = append(audit.Whitelisted, msg)
			continue
		}
		audit.Failures = append(audit.Failures, msg)
	}

	return audit
}

func parityField(msg string) string {
	parts := strings.SplitN(msg, ":", 2)
	return strings.TrimSpace(strings.ToLower(parts[0]))
}

func whitelistKey(service, field string) string {
	return service + "::" + field
}

func verifyService(db *sql.DB, gen *compose.Generator, svc serviceRecord, composeDir, projectRoot string, verbose bool) verifyResult {
	result := verifyResult{Pass: true}

	// Load original compose file — try exact name match first, then scan category dir
	original, err := findOriginalService(composeDir, svc)
	if err != nil {
		result.Pass = false
		result.Messages = append(result.Messages, fmt.Sprintf("parse original: %v", err))
		return result
	}

	// Load service from DB
	var dbService models.Service
	err = db.QueryRow(`
		SELECT id, name, category_id, image_name, image_tag, restart_policy, command, user_spec, compose_overrides, container_name_template
		FROM services WHERE id = $1
	`, svc.ID).Scan(
		&dbService.ID, &dbService.Name, &dbService.CategoryID,
		&dbService.ImageName, &dbService.ImageTag, &dbService.RestartPolicy,
		&dbService.Command, &dbService.UserSpec, &dbService.ComposeOverrides,
		&dbService.ContainerNameTemplate,
	)
	if err != nil {
		result.Pass = false
		result.Messages = append(result.Messages, fmt.Sprintf("load service from DB: %v", err))
		return result
	}

	// Materialize config files
	tempDir := filepath.Join(projectRoot, "api")
	if err := gen.MaterializeConfigFiles(&dbService, tempDir); err != nil {
		result.Pass = false
		result.Messages = append(result.Messages, fmt.Sprintf("materialize config files: %v", err))
		return result
	}

	// Generate compose
	generatedYAML, err := gen.Generate(&dbService)
	if err != nil {
		result.Pass = false
		result.Messages = append(result.Messages, fmt.Sprintf("generate: %v", err))
		return result
	}

	// Parse generated compose to get comparable structure
	tempFile := filepath.Join(os.TempDir(), fmt.Sprintf("generated-%s.yml", svc.Name))
	if err := os.WriteFile(tempFile, generatedYAML, 0644); err != nil {
		result.Pass = false
		result.Messages = append(result.Messages, fmt.Sprintf("write temp file: %v", err))
		return result
	}
	defer os.Remove(tempFile)

	generatedServices, err := compose.ParseFileAll(tempFile)
	if err != nil {
		result.Pass = false
		result.Messages = append(result.Messages, fmt.Sprintf("parse generated: %v", err))
		return result
	}

	var generated *compose.ParsedService
	for _, s := range generatedServices {
		if s.Name == svc.Name {
			generated = s
			break
		}
	}
	if generated == nil {
		result.Pass = false
		result.Messages = append(result.Messages, "service not found in generated compose")
		return result
	}

	// Compare fields
	compareImage(original, generated, &result)
	compareContainerName(original, generated, &result)
	compareRestartPolicy(original, generated, &result)
	compareCommand(original, generated, &result)
	compareUser(original, generated, &result)
	comparePorts(original, generated, &result)
	compareVolumes(original, generated, &result)
	compareEnvVars(original, generated, &result)
	compareEnvFiles(original, generated, &result)
	compareDependencies(original, generated, &result)
	compareHealthcheck(original, generated, &result)
	compareLabels(original, generated, &result)
	compareNetworks(original, generated, &result)

	return result
}

func findOriginalService(composeDir string, svc serviceRecord) (*compose.ParsedService, error) {
	exactPath := filepath.Join(composeDir, svc.Category, svc.Name, svc.Name+".yml")
	if services, err := compose.ParseFileAll(exactPath); err == nil {
		for _, s := range services {
			if s.Name == svc.Name {
				return s, nil
			}
		}
	}

	catDir := filepath.Join(composeDir, svc.Category)
	entries, err := os.ReadDir(catDir)
	if err != nil {
		return nil, fmt.Errorf("read category dir %s: %w", svc.Category, err)
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		ymlPath := filepath.Join(catDir, entry.Name(), entry.Name()+".yml")
		services, err := compose.ParseFileAll(ymlPath)
		if err != nil {
			continue
		}
		for _, s := range services {
			if s.Name == svc.Name {
				return s, nil
			}
		}
	}
	return nil, fmt.Errorf("service %s not found in any compose file under %s/", svc.Name, svc.Category)
}

func compareImage(original, generated *compose.ParsedService, result *verifyResult) {
	origImage := fmt.Sprintf("%s:%s", original.ImageName, original.ImageTag)
	genImage := fmt.Sprintf("%s:%s", generated.ImageName, generated.ImageTag)
	if origImage != genImage {
		result.Pass = false
		result.Messages = append(result.Messages, fmt.Sprintf("image: expected [%s] got [%s]", origImage, genImage))
	}
}

func compareContainerName(original, generated *compose.ParsedService, result *verifyResult) {
	if original.ContainerName != generated.ContainerName {
		result.Pass = false
		result.Messages = append(result.Messages, fmt.Sprintf("container_name: expected [%s] got [%s]", original.ContainerName, generated.ContainerName))
	}
}

func compareRestartPolicy(original, generated *compose.ParsedService, result *verifyResult) {
	if original.RestartPolicy != generated.RestartPolicy {
		result.Pass = false
		result.Messages = append(result.Messages, fmt.Sprintf("restart: expected [%s] got [%s]", original.RestartPolicy, generated.RestartPolicy))
	}
}

func compareCommand(original, generated *compose.ParsedService, result *verifyResult) {
	if original.Command != generated.Command {
		result.Pass = false
		result.Messages = append(result.Messages, fmt.Sprintf("command: expected [%s] got [%s]", original.Command, generated.Command))
	}
}

func compareUser(original, generated *compose.ParsedService, result *verifyResult) {
	if original.UserSpec != generated.UserSpec {
		result.Pass = false
		result.Messages = append(result.Messages, fmt.Sprintf("user: expected [%s] got [%s]", original.UserSpec, generated.UserSpec))
	}
}

func comparePorts(original, generated *compose.ParsedService, result *verifyResult) {
	// Sort for order-independent comparison
	origPorts := sortPorts(original.Ports)
	genPorts := sortPorts(generated.Ports)

	if len(origPorts) != len(genPorts) {
		result.Pass = false
		result.Messages = append(result.Messages, fmt.Sprintf("ports: count mismatch (expected %d, got %d)", len(origPorts), len(genPorts)))
		return
	}

	for i, op := range origPorts {
		gp := genPorts[i]
		if op.HostIP != gp.HostIP || op.HostPort != gp.HostPort || op.ContainerPort != gp.ContainerPort || op.Protocol != gp.Protocol {
			result.Pass = false
			result.Messages = append(result.Messages, fmt.Sprintf("ports: mismatch at index %d: expected [%s:%d:%d/%s] got [%s:%d:%d/%s]",
				i, op.HostIP, op.HostPort, op.ContainerPort, op.Protocol,
				gp.HostIP, gp.HostPort, gp.ContainerPort, gp.Protocol))
		}
	}
}

func sortPorts(ports []compose.ParsedPort) []compose.ParsedPort {
	sorted := make([]compose.ParsedPort, len(ports))
	copy(sorted, ports)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].HostIP != sorted[j].HostIP {
			return sorted[i].HostIP < sorted[j].HostIP
		}
		if sorted[i].HostPort != sorted[j].HostPort {
			return sorted[i].HostPort < sorted[j].HostPort
		}
		if sorted[i].ContainerPort != sorted[j].ContainerPort {
			return sorted[i].ContainerPort < sorted[j].ContainerPort
		}
		return sorted[i].Protocol < sorted[j].Protocol
	})
	return sorted
}

func compareVolumes(original, generated *compose.ParsedService, result *verifyResult) {
	// For volumes, we compare target + read_only only (skip source path comparison due to resolution differences)
	// Config mounts from original should appear in generated volumes

	// Build maps by target for easy lookup
	origMap := make(map[string]volumeComparable)
	for _, v := range original.Volumes {
		origMap[v.Target] = volumeComparable{Target: v.Target, ReadOnly: v.ReadOnly}
	}
	// Add config mounts from original
	for _, cm := range original.ConfigMounts {
		origMap[cm.TargetPath] = volumeComparable{Target: cm.TargetPath, ReadOnly: cm.ReadOnly}
	}

	genMap := make(map[string]volumeComparable)
	for _, v := range generated.Volumes {
		genMap[v.Target] = volumeComparable{Target: v.Target, ReadOnly: v.ReadOnly}
	}

	// Check all original volumes are in generated
	for target, origVol := range origMap {
		genVol, ok := genMap[target]
		if !ok {
			result.Pass = false
			result.Messages = append(result.Messages, fmt.Sprintf("volumes: missing target [%s]", target))
			continue
		}
		if origVol.ReadOnly != genVol.ReadOnly {
			result.Pass = false
			result.Messages = append(result.Messages, fmt.Sprintf("volumes: target [%s] readonly mismatch (expected %v, got %v)", target, origVol.ReadOnly, genVol.ReadOnly))
		}
	}

	// Check no extra volumes in generated
	for target := range genMap {
		if _, ok := origMap[target]; !ok {
			result.Pass = false
			result.Messages = append(result.Messages, fmt.Sprintf("volumes: unexpected target [%s]", target))
		}
	}
}

type volumeComparable struct {
	Target   string
	ReadOnly bool
}

func compareEnvVars(original, generated *compose.ParsedService, result *verifyResult) {
	origMap := make(map[string]string)
	for _, e := range original.EnvVars {
		origMap[e.Key] = e.Value
	}

	genMap := make(map[string]string)
	for _, e := range generated.EnvVars {
		genMap[e.Key] = e.Value
	}

	for key, origVal := range origMap {
		genVal, ok := genMap[key]
		if !ok {
			result.Pass = false
			result.Messages = append(result.Messages, fmt.Sprintf("env_vars: missing key [%s]", key))
			continue
		}
		if origVal != genVal {
			result.Pass = false
			result.Messages = append(result.Messages, fmt.Sprintf("env_vars: key [%s] value mismatch (expected [%s], got [%s])", key, origVal, genVal))
		}
	}

	for key := range genMap {
		if _, ok := origMap[key]; !ok {
			result.Pass = false
			result.Messages = append(result.Messages, fmt.Sprintf("env_vars: unexpected key [%s]", key))
		}
	}
}

func compareEnvFiles(original, generated *compose.ParsedService, result *verifyResult) {
	// Order-dependent comparison (sort_order matters)
	if len(original.EnvFiles) != len(generated.EnvFiles) {
		result.Pass = false
		result.Messages = append(result.Messages, fmt.Sprintf("env_files: count mismatch (expected %d, got %d)", len(original.EnvFiles), len(generated.EnvFiles)))
		return
	}

	for i := range original.EnvFiles {
		if original.EnvFiles[i] != generated.EnvFiles[i] {
			result.Pass = false
			result.Messages = append(result.Messages, fmt.Sprintf("env_files: mismatch at index %d (expected [%s], got [%s])", i, original.EnvFiles[i], generated.EnvFiles[i]))
		}
	}
}

func compareDependencies(original, generated *compose.ParsedService, result *verifyResult) {
	origSet := make(map[string]bool)
	for _, d := range original.Dependencies {
		origSet[d] = true
	}

	genSet := make(map[string]bool)
	for _, d := range generated.Dependencies {
		genSet[d] = true
	}

	for dep := range origSet {
		if !genSet[dep] {
			result.Pass = false
			result.Messages = append(result.Messages, fmt.Sprintf("dependencies: missing [%s]", dep))
		}
	}

	for dep := range genSet {
		if !origSet[dep] {
			result.Pass = false
			result.Messages = append(result.Messages, fmt.Sprintf("dependencies: unexpected [%s]", dep))
		}
	}
}

func compareHealthcheck(original, generated *compose.ParsedService, result *verifyResult) {
	if (original.Healthcheck == nil) != (generated.Healthcheck == nil) {
		result.Pass = false
		result.Messages = append(result.Messages, "healthcheck: presence mismatch")
		return
	}

	if original.Healthcheck == nil {
		return
	}

	ohc := original.Healthcheck
	ghc := generated.Healthcheck

	if ohc.Test != ghc.Test {
		result.Pass = false
		result.Messages = append(result.Messages, fmt.Sprintf("healthcheck.test: expected [%s] got [%s]", ohc.Test, ghc.Test))
	}
	if ohc.IntervalSeconds != ghc.IntervalSeconds {
		result.Pass = false
		result.Messages = append(result.Messages, fmt.Sprintf("healthcheck.interval: expected [%d] got [%d]", ohc.IntervalSeconds, ghc.IntervalSeconds))
	}
	if ohc.TimeoutSeconds != ghc.TimeoutSeconds {
		result.Pass = false
		result.Messages = append(result.Messages, fmt.Sprintf("healthcheck.timeout: expected [%d] got [%d]", ohc.TimeoutSeconds, ghc.TimeoutSeconds))
	}
	if ohc.Retries != ghc.Retries {
		result.Pass = false
		result.Messages = append(result.Messages, fmt.Sprintf("healthcheck.retries: expected [%d] got [%d]", ohc.Retries, ghc.Retries))
	}
	if ohc.StartPeriodSeconds != ghc.StartPeriodSeconds {
		result.Pass = false
		result.Messages = append(result.Messages, fmt.Sprintf("healthcheck.start_period: expected [%d] got [%d]", ohc.StartPeriodSeconds, ghc.StartPeriodSeconds))
	}
}

func compareLabels(original, generated *compose.ParsedService, result *verifyResult) {
	origMap := make(map[string]string)
	for _, l := range original.Labels {
		origMap[l.Key] = l.Value
	}

	genMap := make(map[string]string)
	for _, l := range generated.Labels {
		genMap[l.Key] = l.Value
	}

	for key, origVal := range origMap {
		genVal, ok := genMap[key]
		if !ok {
			result.Pass = false
			result.Messages = append(result.Messages, fmt.Sprintf("labels: missing key [%s]", key))
			continue
		}
		if origVal != genVal {
			result.Pass = false
			result.Messages = append(result.Messages, fmt.Sprintf("labels: key [%s] value mismatch (expected [%s], got [%s])", key, origVal, genVal))
		}
	}

	for key := range genMap {
		if _, ok := origMap[key]; !ok {
			result.Pass = false
			result.Messages = append(result.Messages, fmt.Sprintf("labels: unexpected key [%s]", key))
		}
	}
}

func compareNetworks(original, generated *compose.ParsedService, result *verifyResult) {
	origSet := make(map[string]bool)
	for _, n := range original.Networks {
		origSet[n] = true
	}

	genSet := make(map[string]bool)
	for _, n := range generated.Networks {
		genSet[n] = true
	}

	var missing []string
	for net := range origSet {
		if !genSet[net] {
			missing = append(missing, net)
		}
	}

	var extra []string
	for net := range genSet {
		if !origSet[net] {
			extra = append(extra, net)
		}
	}

	if len(missing) > 0 {
		result.Pass = false
		result.Messages = append(result.Messages, fmt.Sprintf("networks: missing [%s]", strings.Join(missing, ", ")))
	}
	if len(extra) > 0 {
		result.Pass = false
		result.Messages = append(result.Messages, fmt.Sprintf("networks: unexpected [%s]", strings.Join(extra, ", ")))
	}
}
