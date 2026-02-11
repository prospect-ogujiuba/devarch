package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/priz/devarch-api/internal/api/respond"
	"github.com/priz/devarch-api/internal/orchestration"
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

// ListWires godoc
// @Summary      List stack wires
// @Description  List service wiring connections and unresolved imports
// @Tags         stacks
// @Produce      json
// @Param        name path string true "Stack name"
// @Success      200 {object} respond.SuccessEnvelope{data=listWiresResponse}
// @Failure      404 {object} respond.ErrorEnvelope
// @Failure      500 {object} respond.ErrorEnvelope
// @Router       /stacks/{name}/wires [get]
// @Security     ApiKeyAuth
func (h *StackHandler) ListWires(w http.ResponseWriter, r *http.Request) {
	stackName := chi.URLParam(r, "name")

	var stackID int
	err := h.db.QueryRow(`
		SELECT id FROM stacks WHERE name = $1 AND deleted_at IS NULL
	`, stackName).Scan(&stackID)
	if err == sql.ErrNoRows {
		respond.NotFound(w, r, "stack", stackName)
		return
	}
	if err != nil {
		respond.InternalError(w, r, err)
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
		respond.InternalError(w, r, err)
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
			respond.InternalError(w, r, err)
			return
		}

		wr.ConsumerInstance = consumerInstance
		wr.ProviderInstance = providerInstance

		var envVars map[string]string
		if err := json.Unmarshal(envVarsJSON, &envVars); err != nil {
			respond.InternalError(w, r, err)
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
		respond.InternalError(w, r, err)
		return
	}
	defer unresolvedRows.Close()

	var unresolved []unresolvedImport
	for unresolvedRows.Next() {
		var ui unresolvedImport
		if err := unresolvedRows.Scan(&ui.Instance, &ui.ContractName, &ui.ContractType, &ui.Required); err != nil {
			respond.InternalError(w, r, err)
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

	respond.JSON(w, r, http.StatusOK, resp)
}

// ResolveWires godoc
// @Summary      Auto-resolve stack wires
// @Description  Automatically create wires for unresolved imports
// @Tags         stacks
// @Produce      json
// @Param        name path string true "Stack name"
// @Success      200 {object} respond.SuccessEnvelope{data=object}
// @Failure      404 {object} respond.ErrorEnvelope
// @Failure      500 {object} respond.ErrorEnvelope
// @Router       /stacks/{name}/wires/resolve [post]
// @Security     ApiKeyAuth
func (h *StackHandler) ResolveWires(w http.ResponseWriter, r *http.Request) {
	stackName := chi.URLParam(r, "name")

	result, err := h.orchestrationService.ResolveWiring(stackName)
	if err != nil {
		switch {
		case errors.Is(err, orchestration.ErrStackNotFound):
			respond.NotFound(w, r, "stack", stackName)
		case errors.Is(err, orchestration.ErrValidation):
			respond.BadRequest(w, r, err.Error())
		default:
			respond.InternalError(w, r, err)
		}
		return
	}

	respond.Action(w, r, http.StatusOK, "resolved",
		respond.WithMetadata("resolved_count", result.ResolvedCount),
		respond.WithWarnings(result.Warnings),
	)
}

// CreateWire godoc
// @Summary      Create a wire
// @Description  Manually create a wiring connection between instances
// @Tags         stacks
// @Accept       json
// @Produce      json
// @Param        name path string true "Stack name"
// @Param        wire body createWireRequest true "Wire details"
// @Success      201 {object} respond.SuccessEnvelope{data=object}
// @Failure      400 {object} respond.ErrorEnvelope
// @Failure      404 {object} respond.ErrorEnvelope
// @Failure      500 {object} respond.ErrorEnvelope
// @Router       /stacks/{name}/wires [post]
// @Security     ApiKeyAuth
func (h *StackHandler) CreateWire(w http.ResponseWriter, r *http.Request) {
	stackName := chi.URLParam(r, "name")

	var req createWireRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.BadRequest(w, r, err.Error())
		return
	}

	var stackID int
	err := h.db.QueryRow(`
		SELECT id FROM stacks WHERE name = $1 AND deleted_at IS NULL
	`, stackName).Scan(&stackID)
	if err == sql.ErrNoRows {
		respond.NotFound(w, r, "stack", stackName)
		return
	}
	if err != nil {
		respond.InternalError(w, r, err)
		return
	}

	var consumerInstanceID, consumerServiceID int
	err = h.db.QueryRow(`
		SELECT id, service_id FROM service_instances
		WHERE instance_id = $1 AND stack_id = $2 AND deleted_at IS NULL
	`, req.ConsumerInstance, stackID).Scan(&consumerInstanceID, &consumerServiceID)
	if err == sql.ErrNoRows {
		respond.NotFound(w, r, "consumer instance", req.ConsumerInstance)
		return
	}
	if err != nil {
		respond.InternalError(w, r, err)
		return
	}

	var providerInstanceID, providerServiceID int
	err = h.db.QueryRow(`
		SELECT id, service_id FROM service_instances
		WHERE instance_id = $1 AND stack_id = $2 AND deleted_at IS NULL
	`, req.ProviderInstance, stackID).Scan(&providerInstanceID, &providerServiceID)
	if err == sql.ErrNoRows {
		respond.NotFound(w, r, "provider instance", req.ProviderInstance)
		return
	}
	if err != nil {
		respond.InternalError(w, r, err)
		return
	}

	var importContractID int
	var importType string
	err = h.db.QueryRow(`
		SELECT id, type FROM service_import_contracts
		WHERE name = $1 AND service_id = $2
	`, req.ImportContract, consumerServiceID).Scan(&importContractID, &importType)
	if err == sql.ErrNoRows {
		respond.NotFound(w, r, "import contract", req.ImportContract)
		return
	}
	if err != nil {
		respond.InternalError(w, r, err)
		return
	}

	var exportContractID int
	err = h.db.QueryRow(`
		SELECT id FROM service_exports
		WHERE type = $1 AND service_id = $2
	`, importType, providerServiceID).Scan(&exportContractID)
	if err == sql.ErrNoRows {
		respond.BadRequest(w, r, fmt.Sprintf("provider has no matching export of type %s", importType))
		return
	}
	if err != nil {
		respond.InternalError(w, r, err)
		return
	}

	if consumerInstanceID == providerInstanceID {
		respond.BadRequest(w, r, "cannot wire instance to itself")
		return
	}

	tx, err := h.db.Begin()
	if err != nil {
		respond.InternalError(w, r, err)
		return
	}
	defer tx.Rollback()

	_, err = tx.Exec(`
		DELETE FROM service_instance_wires
		WHERE consumer_instance_id = $1 AND import_contract_id = $2
	`, consumerInstanceID, importContractID)
	if err != nil {
		respond.InternalError(w, r, err)
		return
	}

	var wireID int
	err = tx.QueryRow(`
		INSERT INTO service_instance_wires (stack_id, consumer_instance_id, provider_instance_id, import_contract_id, export_contract_id, source)
		VALUES ($1, $2, $3, $4, $5, 'explicit')
		RETURNING id
	`, stackID, consumerInstanceID, providerInstanceID, importContractID, exportContractID).Scan(&wireID)
	if err != nil {
		respond.InternalError(w, r, err)
		return
	}

	if err := tx.Commit(); err != nil {
		respond.InternalError(w, r, err)
		return
	}

	var createdAt time.Time
	err = h.db.QueryRow(`SELECT created_at FROM service_instance_wires WHERE id = $1`, wireID).Scan(&createdAt)
	if err != nil {
		respond.InternalError(w, r, err)
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

	respond.JSON(w, r, http.StatusCreated, resp)
}

// DeleteWire godoc
// @Summary      Delete a wire
// @Tags         stacks
// @Param        name path string true "Stack name"
// @Param        wireId path int true "Wire ID"
// @Success      204
// @Failure      400 {object} respond.ErrorEnvelope
// @Failure      404 {object} respond.ErrorEnvelope
// @Failure      500 {object} respond.ErrorEnvelope
// @Router       /stacks/{name}/wires/{wireId} [delete]
// @Security     ApiKeyAuth
func (h *StackHandler) DeleteWire(w http.ResponseWriter, r *http.Request) {
	stackName := chi.URLParam(r, "name")
	wireIDStr := chi.URLParam(r, "wireId")

	wireID, err := strconv.Atoi(wireIDStr)
	if err != nil {
		respond.BadRequest(w, r, "invalid wire ID")
		return
	}

	var stackID int
	err = h.db.QueryRow(`
		SELECT id FROM stacks WHERE name = $1 AND deleted_at IS NULL
	`, stackName).Scan(&stackID)
	if err == sql.ErrNoRows {
		respond.NotFound(w, r, "stack", stackName)
		return
	}
	if err != nil {
		respond.InternalError(w, r, err)
		return
	}

	result, err := h.db.Exec(`
		DELETE FROM service_instance_wires WHERE id = $1 AND stack_id = $2
	`, wireID, stackID)
	if err != nil {
		respond.InternalError(w, r, err)
		return
	}

	rows, err := result.RowsAffected()
	if err != nil {
		respond.InternalError(w, r, err)
		return
	}

	if rows == 0 {
		respond.NotFound(w, r, "wire", wireIDStr)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// CleanupOrphanedWires godoc
// @Summary      Cleanup orphaned wires
// @Description  Remove wires pointing to non-existent instances
// @Tags         stacks
// @Produce      json
// @Param        name path string true "Stack name"
// @Success      200 {object} respond.SuccessEnvelope{data=respond.ActionResponse}
// @Failure      404 {object} respond.ErrorEnvelope
// @Failure      500 {object} respond.ErrorEnvelope
// @Router       /stacks/{name}/wires/cleanup [post]
// @Security     ApiKeyAuth
func (h *StackHandler) CleanupOrphanedWires(w http.ResponseWriter, r *http.Request) {
	stackName := chi.URLParam(r, "name")

	var stackID int
	err := h.db.QueryRow(`
		SELECT id FROM stacks WHERE name = $1 AND deleted_at IS NULL
	`, stackName).Scan(&stackID)
	if err == sql.ErrNoRows {
		respond.NotFound(w, r, "stack", stackName)
		return
	}
	if err != nil {
		respond.InternalError(w, r, err)
		return
	}

	orphanedIDs, err := wiring.FindOrphanedWires(h.db, stackID)
	if err != nil {
		respond.InternalError(w, r, err)
		return
	}

	if len(orphanedIDs) == 0 {
		respond.Action(w, r, http.StatusOK, "cleaned",
			respond.WithMetadata("deleted_count", 0),
		)
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
		respond.InternalError(w, r, err)
		return
	}

	deleted, _ := result.RowsAffected()

	respond.Action(w, r, http.StatusOK, "cleaned",
		respond.WithMetadata("deleted_count", int(deleted)),
	)
}
