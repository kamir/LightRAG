package utils

import (
	"fmt"
	"strings"
)

// MemoryURI represents a parsed memory:// URI
type MemoryURI struct {
	ContextID string
	MemoryID  string
}

// ParseMemoryURI parses a memory:// URI into its components
// Format: memory://{context_id}/{memory_id}
// Example: memory://107677460544181387647___mft/fJQAOybZ4366sQHKd40O
func ParseMemoryURI(uri string) (*MemoryURI, error) {
	// Remove memory:// prefix
	if !strings.HasPrefix(uri, "memory://") {
		return nil, fmt.Errorf("invalid memory URI: must start with 'memory://' (got: %s)", uri)
	}

	// Remove prefix
	path := strings.TrimPrefix(uri, "memory://")

	// Split into context_id and memory_id
	parts := strings.SplitN(path, "/", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid memory URI format: expected 'memory://{context_id}/{memory_id}' (got: %s)", uri)
	}

	contextID := parts[0]
	memoryID := parts[1]

	if contextID == "" {
		return nil, fmt.Errorf("invalid memory URI: context_id cannot be empty")
	}
	if memoryID == "" {
		return nil, fmt.Errorf("invalid memory URI: memory_id cannot be empty")
	}

	return &MemoryURI{
		ContextID: contextID,
		MemoryID:  memoryID,
	}, nil
}

// ParseMemoryURIs parses multiple memory URIs from a delimited string
// Delimiter: <SEP> (as used by LightRAG in file_path fields)
func ParseMemoryURIs(uriString string) ([]*MemoryURI, error) {
	if uriString == "" {
		return nil, nil
	}

	// Split by <SEP> delimiter (LightRAG's GRAPH_FIELD_SEP)
	uriStrings := strings.Split(uriString, "<SEP>")

	uris := make([]*MemoryURI, 0, len(uriStrings))
	for _, uriStr := range uriStrings {
		uriStr = strings.TrimSpace(uriStr)
		if uriStr == "" {
			continue
		}

		uri, err := ParseMemoryURI(uriStr)
		if err != nil {
			// Skip invalid URIs, but log them
			continue
		}

		uris = append(uris, uri)
	}

	return uris, nil
}

// BuildMemoryURI constructs a memory:// URI from context_id and memory_id
func BuildMemoryURI(contextID, memoryID string) string {
	return fmt.Sprintf("memory://%s/%s", contextID, memoryID)
}

// ExtractMemoryIDs extracts just the memory IDs from a list of memory URIs
func ExtractMemoryIDs(uris []*MemoryURI) []string {
	ids := make([]string, len(uris))
	for i, uri := range uris {
		ids[i] = uri.MemoryID
	}
	return ids
}

// ExtractUniqueMemoryIDs extracts unique memory IDs from a delimited URI string
func ExtractUniqueMemoryIDs(uriString string) ([]string, error) {
	uris, err := ParseMemoryURIs(uriString)
	if err != nil {
		return nil, err
	}

	// Use map to deduplicate
	idMap := make(map[string]bool)
	for _, uri := range uris {
		idMap[uri.MemoryID] = true
	}

	// Convert to slice
	ids := make([]string, 0, len(idMap))
	for id := range idMap {
		ids = append(ids, id)
	}

	return ids, nil
}
