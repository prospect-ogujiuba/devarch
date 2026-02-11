package orchestration

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/priz/devarch-api/internal/compose"
	"github.com/priz/devarch-api/internal/container"
	"github.com/priz/devarch-api/internal/export"
	"github.com/priz/devarch-api/internal/identity"
	"github.com/priz/devarch-api/internal/lock"
	"github.com/priz/devarch-api/internal/plan"
	"github.com/priz/devarch-api/internal/wiring"
)

type Service struct {
	db              *sql.DB
	containerClient *container.Client
}

func NewService(db *sql.DB, cc *container.Client) *Service {
	return &Service{
		db:              db,
		containerClient: cc,
	}
}

type ApplyResult struct {
	Output       string             `json:"output"`
	LockWarnings []lock.LockWarning `json:"lock_warnings,omitempty"`
}

type ResolveResult struct {
	ResolvedCount int      `json:"resolved_count"`
	Warnings      []string `json:"warnings,omitempty"`
}

// GeneratePlan generates a deployment plan for a stack
func (s *Service) GeneratePlan(stackName string) (*plan.Plan, error) {
	var stackID int
	var networkName sql.NullString
	var stackUpdatedAt time.Time
	var enabled bool
	err := s.db.QueryRow(`
		SELECT id, network_name, updated_at, enabled
		FROM stacks
		WHERE name = $1 AND deleted_at IS NULL
	`, stackName).Scan(&stackID, &networkName, &stackUpdatedAt, &enabled)

	if err == sql.ErrNoRows {
		return nil, ErrStackNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("query stack: %w", err)
	}

	if !enabled {
		return nil, ErrStackDisabled
	}

	rows, err := s.db.Query(`
		SELECT si.instance_id, s.name as template_name, si.container_name, si.enabled, si.updated_at
		FROM service_instances si
		JOIN services s ON s.id = si.template_service_id
		WHERE si.stack_id = $1 AND si.deleted_at IS NULL
		ORDER BY si.instance_id
	`, stackID)
	if err != nil {
		return nil, fmt.Errorf("query instances: %w", err)
	}
	defer rows.Close()

	var desired []plan.DesiredInstance
	var timestamps []plan.InstanceTimestamp
	for rows.Next() {
		var instanceID, templateName string
		var containerName sql.NullString
		var enabled bool
		var updatedAt time.Time

		if err := rows.Scan(&instanceID, &templateName, &containerName, &enabled, &updatedAt); err != nil {
			return nil, fmt.Errorf("scan instance: %w", err)
		}

		name := ""
		if containerName.Valid && containerName.String != "" {
			name = containerName.String
		} else {
			name = identity.ContainerName(stackName, instanceID)
		}

		desired = append(desired, plan.DesiredInstance{
			InstanceID:    instanceID,
			TemplateName:  templateName,
			ContainerName: name,
			Enabled:       enabled,
		})

		timestamps = append(timestamps, plan.InstanceTimestamp{
			InstanceID: instanceID,
			UpdatedAt:  updatedAt,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate instances: %w", err)
	}

	running, err := s.containerClient.ListContainersWithLabels(map[string]string{
		"devarch.stack_id": stackName,
	})
	if err != nil {
		// Log error but continue with empty slice - runtime may be down, plan should still work
		running = []string{}
	}

	changes := plan.ComputeDiff(desired, running)
	if changes == nil {
		changes = []plan.Change{}
	}
	token := plan.GenerateToken(stackUpdatedAt, timestamps)

	planResp := plan.Plan{
		StackName:   stackName,
		StackID:     stackID,
		Changes:     changes,
		Token:       token,
		GeneratedAt: time.Now(),
		Warnings:    []string{},
	}

	wiringSection, wiringWarnings, err := s.resolveAndBuildWiring(stackID, stackName)
	if err != nil {
		return nil, fmt.Errorf("resolve wiring: %w", err)
	}
	if wiringSection != nil {
		planResp.Wiring = wiringSection
		planResp.Warnings = append(planResp.Warnings, wiringWarnings...)
	}

	resourceLimits, resourceWarnings, err := s.loadResourceLimitsForStack(stackID)
	if err != nil {
		return nil, fmt.Errorf("load resource limits: %w", err)
	}
	if len(resourceLimits) > 0 {
		planResp.ResourceLimits = resourceLimits
		planResp.Warnings = append(planResp.Warnings, resourceWarnings...)
	}

	return &planResp, nil
}

// ApplyPlan applies a deployment plan to a stack
func (s *Service) ApplyPlan(ctx context.Context, stackName string, token string, lockFile *lock.LockFile) (*ApplyResult, error) {
	var stackID int
	var networkName sql.NullString
	var enabled bool
	err := s.db.QueryRowContext(ctx, `
		SELECT id, network_name, enabled
		FROM stacks
		WHERE name = $1 AND deleted_at IS NULL
	`, stackName).Scan(&stackID, &networkName, &enabled)

	if err == sql.ErrNoRows {
		return nil, ErrStackNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("query stack: %w", err)
	}

	if !enabled {
		return nil, ErrStackDisabled
	}

	var acquired bool
	err = s.db.QueryRowContext(ctx, "SELECT pg_try_advisory_lock($1)", stackID).Scan(&acquired)
	if err != nil || !acquired {
		return nil, ErrLockConflict
	}
	defer func() {
		unlockCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		s.db.ExecContext(unlockCtx, "SELECT pg_advisory_unlock($1)", stackID)
	}()

	if err := plan.ValidateToken(s.db, stackID, token); err != nil {
		if errors.Is(err, plan.ErrStalePlan) {
			return nil, ErrStalePlan
		}
		return nil, fmt.Errorf("validate token: %w", err)
	}

	netName := ""
	if networkName.Valid && networkName.String != "" {
		netName = networkName.String
	} else {
		netName = identity.NetworkName(stackName)
	}

	labels := map[string]string{
		"devarch.managed_by": "devarch",
		"devarch.stack":      stackName,
	}
	if err := s.containerClient.CreateNetwork(netName, labels); err != nil {
		return nil, fmt.Errorf("create network: %w", err)
	}

	projectRoot := os.Getenv("PROJECT_ROOT")
	if projectRoot == "" {
		return nil, ErrProjectRoot
	}

	gen := compose.NewGenerator(s.db, netName)
	if root := os.Getenv("PROJECT_ROOT"); root != "" {
		gen.SetProjectRoot(root)
	}
	if hostRoot := os.Getenv("HOST_PROJECT_ROOT"); hostRoot != "" {
		gen.SetHostProjectRoot(hostRoot)
	}
	if ws := os.Getenv("WORKSPACE_ROOT"); ws != "" {
		gen.SetWorkspaceRoot(ws)
	}

	if err := gen.MaterializeStackConfigs(stackName, projectRoot); err != nil {
		configDir := filepath.Join(projectRoot, ".runtime", "compose", "stacks", stackName)
		os.RemoveAll(configDir)
		return nil, fmt.Errorf("materialize configs: %w", err)
	}

	yamlBytes, _, err := gen.GenerateStack(stackName)
	if err != nil {
		configDir := filepath.Join(projectRoot, ".runtime", "compose", "stacks", stackName)
		os.RemoveAll(configDir)
		return nil, fmt.Errorf("generate stack: %w", err)
	}

	tmpFile, err := os.CreateTemp("", "devarch-apply-*.yml")
	if err != nil {
		configDir := filepath.Join(projectRoot, ".runtime", "compose", "stacks", stackName)
		os.RemoveAll(configDir)
		return nil, fmt.Errorf("create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write(yamlBytes); err != nil {
		tmpFile.Close()
		configDir := filepath.Join(projectRoot, ".runtime", "compose", "stacks", stackName)
		os.RemoveAll(configDir)
		return nil, fmt.Errorf("write temp file: %w", err)
	}
	tmpFile.Close()

	output, err := s.containerClient.RunCompose(tmpFile.Name(), "--project-name", "devarch-"+stackName, "up", "-d")
	if err != nil {
		return nil, fmt.Errorf("compose up failed: %v\n%s", err, output)
	}

	result := &ApplyResult{
		Output: output,
	}

	if lockFile != nil {
		validator := lock.NewValidator(s.db, s.containerClient)
		validationResult, err := validator.Validate(lockFile, stackName)
		if err == nil && len(validationResult.Warnings) > 0 {
			result.LockWarnings = validationResult.Warnings
		}
	}

	return result, nil
}

// ResolveWiring resolves and persists auto-wires for a stack
func (s *Service) ResolveWiring(stackName string) (*ResolveResult, error) {
	var stackID int
	err := s.db.QueryRow(`
		SELECT id FROM stacks WHERE name = $1 AND deleted_at IS NULL
	`, stackName).Scan(&stackID)
	if err == sql.ErrNoRows {
		return nil, ErrStackNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("query stack: %w", err)
	}

	providers, err := s.loadAllProviders(stackID)
	if err != nil {
		return nil, fmt.Errorf("load providers: %w", err)
	}

	consumers, err := s.loadAllConsumers(stackID)
	if err != nil {
		return nil, fmt.Errorf("load consumers: %w", err)
	}

	existingWires, err := s.loadExistingWires(stackID)
	if err != nil {
		return nil, fmt.Errorf("load existing wires: %w", err)
	}

	candidates, warnings := wiring.ResolveAutoWires(stackName, providers, consumers, existingWires)

	validationWarnings, err := wiring.ValidateWiring(candidates, existingWires)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrValidation, err.Error())
	}
	warnings = append(warnings, validationWarnings...)

	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`DELETE FROM service_instance_wires WHERE stack_id = $1 AND source = 'auto'`, stackID); err != nil {
		return nil, fmt.Errorf("delete auto wires: %w", err)
	}

	for _, candidate := range candidates {
		_, err := tx.Exec(`
			INSERT INTO service_instance_wires (stack_id, consumer_instance_id, provider_instance_id, import_contract_id, export_contract_id, source)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, stackID, candidate.ConsumerInstanceID, candidate.ProviderInstanceID, candidate.ImportContractID, candidate.ExportContractID, candidate.Source)
		if err != nil {
			return nil, fmt.Errorf("insert wire: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit transaction: %w", err)
	}

	return &ResolveResult{
		ResolvedCount: len(candidates),
		Warnings:      warnings,
	}, nil
}

// loadResourceLimitsForStack loads resource limits for all instances in a stack
func (s *Service) loadResourceLimitsForStack(stackID int) (map[string]plan.ResourceLimitEntry, []string, error) {
	rows, err := s.db.Query(`
		SELECT si.instance_id, irl.cpu_limit, irl.cpu_reservation, irl.memory_limit, irl.memory_reservation
		FROM instance_resource_limits irl
		JOIN service_instances si ON si.id = irl.instance_id
		WHERE si.stack_id = $1 AND si.deleted_at IS NULL
		ORDER BY si.instance_id
	`, stackID)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	limits := make(map[string]plan.ResourceLimitEntry)
	warnings := []string{}

	for rows.Next() {
		var instanceID string
		var cpuLimit, cpuReservation, memoryLimit, memoryReservation sql.NullString
		if err := rows.Scan(&instanceID, &cpuLimit, &cpuReservation, &memoryLimit, &memoryReservation); err != nil {
			return nil, nil, err
		}

		entry := plan.ResourceLimitEntry{}
		if cpuLimit.Valid && cpuLimit.String != "" {
			entry.CPULimit = cpuLimit.String
		}
		if cpuReservation.Valid && cpuReservation.String != "" {
			entry.CPUReservation = cpuReservation.String
		}
		if memoryLimit.Valid && memoryLimit.String != "" {
			entry.MemoryLimit = memoryLimit.String
		}
		if memoryReservation.Valid && memoryReservation.String != "" {
			entry.MemoryReservation = memoryReservation.String
		}

		limits[instanceID] = entry
	}

	return limits, warnings, rows.Err()
}

// resolveAndBuildWiring resolves wiring and builds the wiring section for a plan
func (s *Service) resolveAndBuildWiring(stackID int, stackName string) (*plan.WiringSection, []string, error) {
	providers, err := s.loadProviders(stackID)
	if err != nil {
		return nil, nil, fmt.Errorf("load providers: %w", err)
	}

	consumers, err := s.loadConsumers(stackID)
	if err != nil {
		return nil, nil, fmt.Errorf("load consumers: %w", err)
	}

	existingWires, err := s.loadExistingWires(stackID)
	if err != nil {
		return nil, nil, fmt.Errorf("load existing wires: %w", err)
	}

	candidates, warnings := wiring.ResolveAutoWires(stackName, providers, consumers, existingWires)

	tx, err := s.db.Begin()
	if err != nil {
		return nil, nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	_, err = tx.Exec(`DELETE FROM service_instance_wires WHERE stack_id = $1 AND source = 'auto'`, stackID)
	if err != nil {
		return nil, nil, fmt.Errorf("delete old auto wires: %w", err)
	}

	for _, candidate := range candidates {
		_, err = tx.Exec(`
			INSERT INTO service_instance_wires (
				stack_id, consumer_instance_id, provider_instance_id,
				import_contract_id, export_contract_id, source,
				consumer_contract_type, provider_contract_type
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		`, stackID, candidate.ConsumerInstanceID, candidate.ProviderInstanceID,
			candidate.ImportContractID, candidate.ExportContractID, candidate.Source,
			candidate.ConsumerType, candidate.ProviderType)
		if err != nil {
			return nil, nil, fmt.Errorf("insert auto wire: %w", err)
		}
	}

	_, err = tx.Exec(`
		DELETE FROM service_instance_wires
		WHERE stack_id = $1 AND (
			consumer_instance_id IN (SELECT id FROM service_instances WHERE deleted_at IS NOT NULL) OR
			provider_instance_id IN (SELECT id FROM service_instances WHERE deleted_at IS NOT NULL)
		)
	`, stackID)
	if err != nil {
		return nil, nil, fmt.Errorf("cleanup orphaned wires: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, nil, fmt.Errorf("commit wiring: %w", err)
	}

	allWires, err := s.loadAllWires(stackID, stackName)
	if err != nil {
		return nil, nil, fmt.Errorf("load all wires: %w", err)
	}

	section := &plan.WiringSection{
		ActiveWires: allWires,
		Warnings:    []plan.WiringWarning{},
	}

	for _, msg := range warnings {
		section.Warnings = append(section.Warnings, plan.WiringWarning{
			Severity: "warning",
			Message:  msg,
		})
	}

	return section, warnings, nil
}

// loadProviders loads providers for enabled instances (used by GeneratePlan)
func (s *Service) loadProviders(stackID int) ([]wiring.Provider, error) {
	rows, err := s.db.Query(`
		SELECT si.id, si.instance_id, se.id, se.name, se.type, se.port, se.protocol
		FROM service_instances si
		JOIN service_exports se ON se.service_id = si.template_service_id
		WHERE si.stack_id = $1 AND si.deleted_at IS NULL AND si.enabled = true
		ORDER BY si.id, se.name
	`, stackID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var providers []wiring.Provider
	for rows.Next() {
		var p wiring.Provider
		if err := rows.Scan(&p.InstanceID, &p.InstanceName, &p.ExportContractID, &p.ContractName, &p.ContractType, &p.Port, &p.Protocol); err != nil {
			return nil, err
		}
		providers = append(providers, p)
	}
	return providers, rows.Err()
}

// loadAllProviders loads providers from all instances (used by ResolveWiring, no enabled filter)
func (s *Service) loadAllProviders(stackID int) ([]wiring.Provider, error) {
	rows, err := s.db.Query(`
		SELECT si.id, si.instance_id, se.id, se.name, se.type, se.port, se.protocol
		FROM service_instances si
		JOIN service_exports se ON se.service_id = si.service_id
		WHERE si.stack_id = $1 AND si.deleted_at IS NULL
		ORDER BY si.instance_id, se.name
	`, stackID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var providers []wiring.Provider
	for rows.Next() {
		var p wiring.Provider
		if err := rows.Scan(&p.InstanceID, &p.InstanceName, &p.ExportContractID, &p.ContractName, &p.ContractType, &p.Port, &p.Protocol); err != nil {
			return nil, err
		}
		providers = append(providers, p)
	}
	return providers, rows.Err()
}

// loadConsumers loads consumers for enabled instances (used by GeneratePlan)
func (s *Service) loadConsumers(stackID int) ([]wiring.Consumer, error) {
	rows, err := s.db.Query(`
		SELECT si.id, si.instance_id, sic.id, sic.name, sic.type, sic.required, COALESCE(sic.env_vars, '{}')
		FROM service_instances si
		JOIN service_import_contracts sic ON sic.service_id = si.template_service_id
		WHERE si.stack_id = $1 AND si.deleted_at IS NULL AND si.enabled = true
		ORDER BY si.id, sic.name
	`, stackID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var consumers []wiring.Consumer
	for rows.Next() {
		var c wiring.Consumer
		var envVarsJSON []byte
		if err := rows.Scan(&c.InstanceID, &c.InstanceName, &c.ImportContractID, &c.ContractName, &c.ContractType, &c.Required, &envVarsJSON); err != nil {
			return nil, err
		}
		c.EnvVars = make(map[string]string)
		if len(envVarsJSON) > 0 {
			if err := json.Unmarshal(envVarsJSON, &c.EnvVars); err != nil {
				return nil, fmt.Errorf("unmarshal env_vars: %w", err)
			}
		}
		consumers = append(consumers, c)
	}
	return consumers, rows.Err()
}

// loadAllConsumers loads consumers from all instances (used by ResolveWiring, no enabled filter)
func (s *Service) loadAllConsumers(stackID int) ([]wiring.Consumer, error) {
	rows, err := s.db.Query(`
		SELECT si.id, si.instance_id, ic.id, ic.name, ic.type, ic.required, COALESCE(ic.env_vars, '{}')
		FROM service_instances si
		JOIN service_import_contracts ic ON ic.service_id = si.service_id
		WHERE si.stack_id = $1 AND si.deleted_at IS NULL
		ORDER BY si.instance_id, ic.name
	`, stackID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var consumers []wiring.Consumer
	for rows.Next() {
		var c wiring.Consumer
		var envVarsJSON []byte
		if err := rows.Scan(&c.InstanceID, &c.InstanceName, &c.ImportContractID, &c.ContractName, &c.ContractType, &c.Required, &envVarsJSON); err != nil {
			return nil, err
		}
		c.EnvVars = make(map[string]string)
		if len(envVarsJSON) > 0 {
			if err := json.Unmarshal(envVarsJSON, &c.EnvVars); err != nil {
				return nil, fmt.Errorf("unmarshal env_vars: %w", err)
			}
		}
		consumers = append(consumers, c)
	}
	return consumers, rows.Err()
}

// loadExistingWires loads all existing wires for a stack
func (s *Service) loadExistingWires(stackID int) ([]wiring.ExistingWire, error) {
	rows, err := s.db.Query(`
		SELECT id, consumer_instance_id, provider_instance_id, import_contract_id, source
		FROM service_instance_wires
		WHERE stack_id = $1
	`, stackID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var wires []wiring.ExistingWire
	for rows.Next() {
		var w wiring.ExistingWire
		if err := rows.Scan(&w.ID, &w.ConsumerInstanceID, &w.ProviderInstanceID, &w.ImportContractID, &w.Source); err != nil {
			return nil, err
		}
		wires = append(wires, w)
	}
	return wires, rows.Err()
}

// loadAllWires loads all wires with their injected env vars for a stack
func (s *Service) loadAllWires(stackID int, stackName string) ([]plan.WirePlanEntry, error) {
	rows, err := s.db.Query(`
		SELECT
			si_consumer.instance_id,
			si_provider.instance_id,
			sic.name,
			sic.type,
			siw.source,
			se.port,
			se.protocol,
			COALESCE(sic.env_vars, '{}')
		FROM service_instance_wires siw
		JOIN service_instances si_consumer ON si_consumer.id = siw.consumer_instance_id
		JOIN service_instances si_provider ON si_provider.id = siw.provider_instance_id
		JOIN service_import_contracts sic ON sic.id = siw.import_contract_id
		JOIN service_exports se ON se.id = siw.export_contract_id
		WHERE siw.stack_id = $1
		ORDER BY si_consumer.instance_id, sic.name
	`, stackID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []plan.WirePlanEntry
	for rows.Next() {
		var entry plan.WirePlanEntry
		var port int
		var protocol string
		var envVarsJSON []byte
		if err := rows.Scan(&entry.ConsumerInstance, &entry.ProviderInstance, &entry.ContractName, &entry.ContractType, &entry.Source, &port, &protocol, &envVarsJSON); err != nil {
			return nil, err
		}

		envVars := make(map[string]string)
		if len(envVarsJSON) > 0 {
			if err := json.Unmarshal(envVarsJSON, &envVars); err != nil {
				return nil, fmt.Errorf("unmarshal env_vars: %w", err)
			}
		}

		provider := wiring.Provider{
			InstanceName: entry.ProviderInstance,
			ContractName: entry.ContractName,
			Port:         port,
			Protocol:     protocol,
		}
		consumer := wiring.Consumer{
			EnvVars: envVars,
		}

		injected := wiring.InjectEnvVars(stackName, provider, consumer)
		redacted := make(map[string]string, len(injected))
		for k, v := range injected {
			if export.IsSecretKey(k) {
				redacted[k] = "***"
			} else {
				redacted[k] = v
			}
		}
		entry.InjectedEnvVars = redacted
		entries = append(entries, entry)
	}
	return entries, rows.Err()
}
