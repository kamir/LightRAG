package api

import (
	"encoding/json"
	"net/http"

	"github.com/kamir/memory-connector/pkg/client"
	"github.com/kamir/memory-connector/pkg/config"
	"github.com/kamir/memory-connector/pkg/utils"
	"go.uber.org/zap"
)

// LookupHandler handles reverse lookup API requests
type LookupHandler struct {
	config         *config.Config
	memoryClient   *client.MemoryClient
	lightragClient *client.LightRAGClient
	logger         *zap.Logger
}

// NewLookupHandler creates a new lookup handler
func NewLookupHandler(
	cfg *config.Config,
	memoryClient *client.MemoryClient,
	lightragClient *client.LightRAGClient,
	logger *zap.Logger,
) *LookupHandler {
	return &LookupHandler{
		config:         cfg,
		memoryClient:   memoryClient,
		lightragClient: lightragClient,
		logger:         logger,
	}
}

// MemoryReference represents a reference to a source memory
type MemoryReference struct {
	MemoryID   string `json:"memory_id"`
	ContextID  string `json:"context_id"`
	MemoryURI  string `json:"memory_uri"`
	SourceSystem string `json:"source_system,omitempty"`
}

// ByEntityResponse represents the response for /lookup/by-entity
type ByEntityResponse struct {
	EntityName string            `json:"entity_name"`
	Found      bool              `json:"found"`
	Memories   []MemoryReference `json:"memories"`
	Count      int               `json:"count"`
}

// ByMemoryResponse represents the response for /lookup/by-memory
type ByMemoryResponse struct {
	MemoryID     string   `json:"memory_id"`
	ContextID    string   `json:"context_id"`
	MemoryURI    string   `json:"memory_uri"`
	Found        bool     `json:"found"`
	Entities     []string `json:"entities,omitempty"`
	Relationships []string `json:"relationships,omitempty"`
	Message      string   `json:"message,omitempty"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// HandleByEntity handles GET /api/v1/lookup/by-entity?name={entity_name}
//
// This endpoint queries LightRAG for an entity and returns all source memories
// that contributed to that entity's creation.
func (h *LookupHandler) HandleByEntity(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.sendError(w, http.StatusMethodNotAllowed, "method not allowed", "")
		return
	}

	entityName := r.URL.Query().Get("name")
	if entityName == "" {
		h.sendError(w, http.StatusBadRequest, "missing required parameter", "name parameter is required")
		return
	}

	h.logger.Info("Lookup by entity request",
		zap.String("entity_name", entityName),
	)

	// TODO: Query LightRAG for entity
	// For now, return a placeholder response indicating the feature needs LightRAG query implementation

	response := ByEntityResponse{
		EntityName: entityName,
		Found:      false,
		Memories:   []MemoryReference{},
		Count:      0,
	}

	h.sendJSON(w, http.StatusOK, response)
}

// HandleByMemory handles GET /api/v1/lookup/by-memory?memory_id={id}&context_id={ctx}
//
// This endpoint queries LightRAG for all entities and relationships that reference
// a specific memory, effectively showing what knowledge was extracted from that memory.
func (h *LookupHandler) HandleByMemory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.sendError(w, http.StatusMethodNotAllowed, "method not allowed", "")
		return
	}

	memoryID := r.URL.Query().Get("memory_id")
	contextID := r.URL.Query().Get("context_id")

	if memoryID == "" {
		h.sendError(w, http.StatusBadRequest, "missing required parameter", "memory_id parameter is required")
		return
	}
	if contextID == "" {
		h.sendError(w, http.StatusBadRequest, "missing required parameter", "context_id parameter is required")
		return
	}

	memoryURI := utils.BuildMemoryURI(contextID, memoryID)

	h.logger.Info("Lookup by memory request",
		zap.String("memory_id", memoryID),
		zap.String("context_id", contextID),
		zap.String("memory_uri", memoryURI),
	)

	// TODO: Query LightRAG for entities/relationships containing this memory URI
	// For now, return a placeholder response

	response := ByMemoryResponse{
		MemoryID:     memoryID,
		ContextID:    contextID,
		MemoryURI:    memoryURI,
		Found:        false,
		Entities:     []string{},
		Relationships: []string{},
		Message:      "Querying LightRAG knowledge graph - implementation pending",
	}

	h.sendJSON(w, http.StatusOK, response)
}

// HandleResolveURI handles GET /api/v1/lookup/resolve?uri={memory_uri}
//
// This endpoint parses a memory:// URI and optionally fetches the full memory
// from the Memory API.
func (h *LookupHandler) HandleResolveURI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.sendError(w, http.StatusMethodNotAllowed, "method not allowed", "")
		return
	}

	uriString := r.URL.Query().Get("uri")
	fetchFull := r.URL.Query().Get("fetch") == "true"

	if uriString == "" {
		h.sendError(w, http.StatusBadRequest, "missing required parameter", "uri parameter is required")
		return
	}

	// Parse URI
	uri, err := utils.ParseMemoryURI(uriString)
	if err != nil {
		h.sendError(w, http.StatusBadRequest, "invalid memory URI", err.Error())
		return
	}

	h.logger.Info("Resolve URI request",
		zap.String("uri", uriString),
		zap.Bool("fetch_full", fetchFull),
	)

	response := map[string]interface{}{
		"memory_id":  uri.MemoryID,
		"context_id": uri.ContextID,
		"memory_uri": uriString,
		"valid":      true,
	}

	// If fetch=true, get the full memory from Memory API
	if fetchFull {
		// Find connector config for this context_id
		var sourceSystem string
		for _, connector := range h.config.Connectors {
			if connector.ContextID == uri.ContextID {
				sourceSystem = connector.SourceSystem
				break
			}
		}

		if sourceSystem != "" {
			response["source_system"] = sourceSystem
		}

		response["note"] = "Full memory fetch from Memory API not yet implemented"
	}

	h.sendJSON(w, http.StatusOK, response)
}

// HandleParseURIs handles POST /api/v1/lookup/parse-uris
//
// This endpoint parses a delimited string of memory URIs (as returned by LightRAG)
// and returns the parsed components.
func (h *LookupHandler) HandleParseURIs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.sendError(w, http.StatusMethodNotAllowed, "method not allowed", "")
		return
	}

	var request struct {
		URIString string `json:"uri_string"`
		Delimiter string `json:"delimiter,omitempty"` // Default: <SEP>
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.sendError(w, http.StatusBadRequest, "invalid request body", err.Error())
		return
	}

	if request.URIString == "" {
		h.sendError(w, http.StatusBadRequest, "missing required field", "uri_string is required")
		return
	}

	h.logger.Info("Parse URIs request",
		zap.String("uri_string_prefix", request.URIString[:min(50, len(request.URIString))]),
	)

	// Parse URIs
	uris, err := utils.ParseMemoryURIs(request.URIString)
	if err != nil {
		h.sendError(w, http.StatusBadRequest, "failed to parse URIs", err.Error())
		return
	}

	// Convert to response format
	references := make([]MemoryReference, len(uris))
	for i, uri := range uris {
		references[i] = MemoryReference{
			MemoryID:  uri.MemoryID,
			ContextID: uri.ContextID,
			MemoryURI: utils.BuildMemoryURI(uri.ContextID, uri.MemoryID),
		}
	}

	response := map[string]interface{}{
		"count":     len(references),
		"memories":  references,
		"unique_ids": utils.ExtractMemoryIDs(uris),
	}

	h.sendJSON(w, http.StatusOK, response)
}

// Helper methods

func (h *LookupHandler) sendJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

func (h *LookupHandler) sendError(w http.ResponseWriter, statusCode int, error string, message string) {
	h.sendJSON(w, statusCode, ErrorResponse{
		Error:   error,
		Message: message,
	})
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
