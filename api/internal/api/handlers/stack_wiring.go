package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/priz/devarch-api/internal/wiring"
)

type wireResponse struct {
	ID               int               `json:"id"`
	ConsumerInstance string            `json:"consumer_instance"`
	ProviderInstance string            `json:"provider_instance"`
	ContractName     string            `json:"contract_name"`
	ContractType     string            `json:"contract_type"`
	Source           string            `json:"source"`
	InjectedEnvVars  map[string]string `json:"injected_env_vars"`
	CreatedAt        time.Time         `json:"created_at"`
}

type unresolvedImport struct {
	Instance     string `json:"instance"`
	ContractName string `json:"contract_name"`
	ContractType string `json:"contract_type"`
	Required     bool   `json:"required"`
	Reason       string `json:"reason"`
}

type listWiresResponse struct {
	Wires      []wireResponse     `json:"wires"`
	Unresolved []unresolvedImport `json:"unresolved"`
}

type createWireRequest struct {
	ConsumerInstance string `json:"consumer_instance_id"`
	ProviderInstance string `json:"provider_instance_id"`
	ImportContract   string `json:"import_contract_name"`
}

type resolveWiresResponse struct {
	Resolved int      `json:"resolved"`
	Warnings []string `json:"warnings"`
}

type cleanupWiresResponse struct {
	Deleted int `json:"deleted"`
}

func (h *StackHandler) ListWires(w http.ResponseWriter, r *http.Request) {
	stackName := chi.URLParam(r, "name")

	var stackID int
	err := h.db.QueryRow(`
		SELECT id FROM stacks WHERE name = $1 AND deleted_at IS NULL
	`, stackName).Scan(&stackID)
	if err == sql.ErrNoRows {
		http.Error(w, "stack not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rows, err := h.db.Query(`
		SELECT w.id, w.source, w.created_at,
		       ci.instance_id as consumer_instance,
		       pi.instance_id as provider_instance,
		       ic.name as contract_name,
		       ic.type as contract_type,
		       COALESCE(ic.env_vars, '{}') as env_vars,
		       se.port, se.protocol
		FROM service_instance_wires w
		JOIN service_instances ci ON w.consumer_instance_id = ci.id
		JOIN service_instances pi ON w.provider_instance_id = pi.id
		JOIN service_import_contracts ic ON w.import_contract_id = ic.id
		JOIN service_exports se ON w.export_contract_id = se.id
		WHERE w.stack_id = $1
		ORDER BY ci.instance_id, ic.name
	`, stackID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var wires []wireResponse
	for rows.Next() {
		var wr wireResponse
		var envVarsJSON []byte
		var port int
		var protocol string
		var consumerInstance, providerInstance string

		if err := rows.Scan(
			&wr.ID, &wr.Source, &wr.CreatedAt,
			&consumerInstance, &providerInstance,
			&wr.ContractName, &wr.ContractType,
			&envVarsJSON, &port, &protocol,
		); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		wr.ConsumerInstance = consumerInstance
		wr.ProviderInstance = providerInstance

		var envVars map[string]string
		if err := json.Unmarshal(envVarsJSON, &envVars); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		provider := wiring.Provider{
			InstanceName: providerInstance,
			ContractName: wr.ContractName,
			Port:         port,
			Protocol:     protocol,
		}
		consumer := wiring.Consumer{
			InstanceName: consumerInstance,
			EnvVars:      envVars,
		}

		wr.InjectedEnvVars = wiring.InjectEnvVars(stackName, provider, consumer)

		wires = append(wires, wr)
	}

	if wires == nil {
		wires = []wireResponse{}
	}

	unresolvedRows, err := h.db.Query(`
		SELECT DISTINCT
		       si.instance_id,
		       ic.name,
		       ic.type,
		       ic.required
		FROM service_instances si
		JOIN service_import_contracts ic ON ic.service_id = si.service_id
		LEFT JOIN service_instance_wires w ON (
		    w.consumer_instance_id = si.id
		    AND w.import_contract_id = ic.id
		    AND w.stack_id = $1
		)
		WHERE si.stack_id = $1
		  AND si.deleted_at IS NULL
		  AND w.id IS NULL
		ORDER BY si.instance_id, ic.name
	`, stackID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer unresolvedRows.Close()

	var unresolved []unresolvedImport
	for unresolvedRows.Next() {
		var ui unresolvedImport
		if err := unresolvedRows.Scan(&ui.Instance, &ui.ContractName, &ui.ContractType, &ui.Required); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		ui.Reason = "no matching provider found or ambiguous"
		unresolved = append(unresolved, ui)
	}

	if unresolved == nil {
		unresolved = []unresolvedImport{}
	}

	resp := listWiresResponse{
		Wires:      wires,
		Unresolved: unresolved,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *StackHandler) ResolveWires(w http.ResponseWriter, r *http.Request) {
	stackName := chi.URLParam(r, "name")

	var stackID int
	err := h.db.QueryRow(`
		SELECT id FROM stacks WHERE name = $1 AND deleted_at IS NULL
	`, stackName).Scan(&stackID)
	if err == sql.ErrNoRows {
		http.Error(w, "stack not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	providerRows, err := h.db.Query(`
		SELECT si.id, si.instance_id, se.id, se.name, se.type, se.port, se.protocol
		FROM service_instances si
		JOIN service_exports se ON se.service_id = si.service_id
		WHERE si.stack_id = $1 AND si.deleted_at IS NULL
		ORDER BY si.instance_id, se.name
	`, stackID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer providerRows.Close()

	var providers []wiring.Provider
	for providerRows.Next() {
		var p wiring.Provider
		if err := providerRows.Scan(&p.InstanceID, &p.InstanceName, &p.ExportContractID, &p.ContractName, &p.ContractType, &p.Port, &p.Protocol); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		providers = append(providers, p)
	}

	consumerRows, err := h.db.Query(`
		SELECT si.id, si.instance_id, ic.id, ic.name, ic.type, ic.required, COALESCE(ic.env_vars, '{}')
		FROM service_instances si
		JOIN service_import_contracts ic ON ic.service_id = si.service_id
		WHERE si.stack_id = $1 AND si.deleted_at IS NULL
		ORDER BY si.instance_id, ic.name
	`, stackID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer consumerRows.Close()

	var consumers []wiring.Consumer
	for consumerRows.Next() {
		var c wiring.Consumer
		var envVarsJSON []byte
		if err := consumerRows.Scan(&c.InstanceID, &c.InstanceName, &c.ImportContractID, &c.ContractName, &c.ContractType, &c.Required, &envVarsJSON); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err := json.Unmarshal(envVarsJSON, &c.EnvVars); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		consumers = append(consumers, c)
	}

	wireRows, err := h.db.Query(`
		SELECT id, consumer_instance_id, provider_instance_id, import_contract_id, source
		FROM service_instance_wires
		WHERE stack_id = $1
	`, stackID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer wireRows.Close()

	var existingWires []wiring.ExistingWire
	for wireRows.Next() {
		var ew wiring.ExistingWire
		if err := wireRows.Scan(&ew.ID, &ew.ConsumerInstanceID, &ew.ProviderInstanceID, &ew.ImportContractID, &ew.Source); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		existingWires = append(existingWires, ew)
	}

	candidates, warnings := wiring.ResolveAutoWires(stackName, providers, consumers, existingWires)

	validationWarnings, err := wiring.ValidateWiring(candidates, existingWires)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	warnings = append(warnings, validationWarnings...)

	tx, err := h.db.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`DELETE FROM service_instance_wires WHERE stack_id = $1 AND source = 'auto'`, stackID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for _, candidate := range candidates {
		_, err := tx.Exec(`
			INSERT INTO service_instance_wires (stack_id, consumer_instance_id, provider_instance_id, import_contract_id, export_contract_id, source)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, stackID, candidate.ConsumerInstanceID, candidate.ProviderInstanceID, candidate.ImportContractID, candidate.ExportContractID, candidate.Source)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := resolveWiresResponse{
		Resolved: len(candidates),
		Warnings: warnings,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *StackHandler) CreateWire(w http.ResponseWriter, r *http.Request) {
	stackName := chi.URLParam(r, "name")

	var req createWireRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var stackID int
	err := h.db.QueryRow(`
		SELECT id FROM stacks WHERE name = $1 AND deleted_at IS NULL
	`, stackName).Scan(&stackID)
	if err == sql.ErrNoRows {
		http.Error(w, "stack not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var consumerInstanceID, consumerServiceID int
	err = h.db.QueryRow(`
		SELECT id, service_id FROM service_instances
		WHERE instance_id = $1 AND stack_id = $2 AND deleted_at IS NULL
	`, req.ConsumerInstance, stackID).Scan(&consumerInstanceID, &consumerServiceID)
	if err == sql.ErrNoRows {
		http.Error(w, "consumer instance not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var providerInstanceID, providerServiceID int
	err = h.db.QueryRow(`
		SELECT id, service_id FROM service_instances
		WHERE instance_id = $1 AND stack_id = $2 AND deleted_at IS NULL
	`, req.ProviderInstance, stackID).Scan(&providerInstanceID, &providerServiceID)
	if err == sql.ErrNoRows {
		http.Error(w, "provider instance not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var importContractID int
	var importType string
	err = h.db.QueryRow(`
		SELECT id, type FROM service_import_contracts
		WHERE name = $1 AND service_id = $2
	`, req.ImportContract, consumerServiceID).Scan(&importContractID, &importType)
	if err == sql.ErrNoRows {
		http.Error(w, "import contract not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var exportContractID int
	err = h.db.QueryRow(`
		SELECT id FROM service_exports
		WHERE type = $1 AND service_id = $2
	`, importType, providerServiceID).Scan(&exportContractID)
	if err == sql.ErrNoRows {
		http.Error(w, fmt.Sprintf("provider has no matching export of type %s", importType), http.StatusBadRequest)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if consumerInstanceID == providerInstanceID {
		http.Error(w, "cannot wire instance to itself", http.StatusBadRequest)
		return
	}

	tx, err := h.db.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	_, err = tx.Exec(`
		DELETE FROM service_instance_wires
		WHERE consumer_instance_id = $1 AND import_contract_id = $2
	`, consumerInstanceID, importContractID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var wireID int
	err = tx.QueryRow(`
		INSERT INTO service_instance_wires (stack_id, consumer_instance_id, provider_instance_id, import_contract_id, export_contract_id, source)
		VALUES ($1, $2, $3, $4, $5, 'explicit')
		RETURNING id
	`, stackID, consumerInstanceID, providerInstanceID, importContractID, exportContractID).Scan(&wireID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var createdAt time.Time
	err = h.db.QueryRow(`SELECT created_at FROM service_instance_wires WHERE id = $1`, wireID).Scan(&createdAt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := wireResponse{
		ID:               wireID,
		ConsumerInstance: req.ConsumerInstance,
		ProviderInstance: req.ProviderInstance,
		ContractName:     req.ImportContract,
		ContractType:     importType,
		Source:           "explicit",
		InjectedEnvVars:  make(map[string]string),
		CreatedAt:        createdAt,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

func (h *StackHandler) DeleteWire(w http.ResponseWriter, r *http.Request) {
	stackName := chi.URLParam(r, "name")
	wireIDStr := chi.URLParam(r, "wireId")

	wireID, err := strconv.Atoi(wireIDStr)
	if err != nil {
		http.Error(w, "invalid wire ID", http.StatusBadRequest)
		return
	}

	var stackID int
	err = h.db.QueryRow(`
		SELECT id FROM stacks WHERE name = $1 AND deleted_at IS NULL
	`, stackName).Scan(&stackID)
	if err == sql.ErrNoRows {
		http.Error(w, "stack not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	result, err := h.db.Exec(`
		DELETE FROM service_instance_wires WHERE id = $1 AND stack_id = $2
	`, wireID, stackID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rows, err := result.RowsAffected()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if rows == 0 {
		http.Error(w, "wire not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *StackHandler) CleanupOrphanedWires(w http.ResponseWriter, r *http.Request) {
	stackName := chi.URLParam(r, "name")

	var stackID int
	err := h.db.QueryRow(`
		SELECT id FROM stacks WHERE name = $1 AND deleted_at IS NULL
	`, stackName).Scan(&stackID)
	if err == sql.ErrNoRows {
		http.Error(w, "stack not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	orphanedIDs, err := wiring.FindOrphanedWires(h.db, stackID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(orphanedIDs) == 0 {
		resp := cleanupWiresResponse{Deleted: 0}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}

	placeholders := ""
	args := []interface{}{stackID}
	for i, id := range orphanedIDs {
		if i > 0 {
			placeholders += ","
		}
		placeholders += fmt.Sprintf("$%d", i+2)
		args = append(args, id)
	}

	query := fmt.Sprintf("DELETE FROM service_instance_wires WHERE stack_id = $1 AND id IN (%s)", placeholders)
	result, err := h.db.Exec(query, args...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	deleted, _ := result.RowsAffected()

	resp := cleanupWiresResponse{Deleted: int(deleted)}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
