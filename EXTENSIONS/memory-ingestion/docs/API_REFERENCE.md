# Memory Connector API Reference

Complete API documentation for the Memory Connector reverse lookup API.

## Table of Contents

- [Overview](#overview)
- [Base URL](#base-url)
- [Authentication](#authentication)
- [Endpoints](#endpoints)
  - [Health Check](#health-check)
  - [Lookup by Entity](#lookup-by-entity)
  - [Lookup by Memory](#lookup-by-memory)
  - [Resolve Memory URI](#resolve-memory-uri)
  - [Parse Multiple URIs](#parse-multiple-uris)
- [Response Format](#response-format)
- [Error Handling](#error-handling)
- [Rate Limiting](#rate-limiting)
- [Examples](#examples)

---

## Overview

The Memory Connector API provides reverse lookup capabilities for tracing knowledge graph entities back to source memories. It bridges LightRAG's knowledge graph with the original Memory API data.

**API Version:** v1
**Protocol:** HTTP/REST
**Response Format:** JSON

---

## Base URL

When running the connector in serve mode:

```
http://localhost:8080
```

Configurable via `--port` flag or config file.

---

## Authentication

Currently, the API is unauthenticated and intended for internal use. Future versions may add:
- API key authentication
- OAuth2 integration
- IP whitelisting

---

## Endpoints

### Health Check

Check if the API server is running.

**Endpoint:** `GET /health`

**Query Parameters:** None

**Response:**
```json
{
  "status": "healthy",
  "timestamp": "2024-01-19T10:30:00Z"
}
```

**Status Codes:**
- `200 OK` - Server is healthy

**Example:**
```bash
curl http://localhost:8080/health
```

---

### Lookup by Entity

Find all source memories that contributed to a specific knowledge graph entity.

**Endpoint:** `GET /api/v1/lookup/by-entity`

**Query Parameters:**
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `name` | string | Yes | Entity name (URL-encoded) |

**Response:**
```json
{
  "entity_name": "John Smith",
  "found": true,
  "memories": [
    {
      "memory_id": "abc123",
      "context_id": "user-context-1",
      "memory_uri": "memory://user-context-1/abc123",
      "source_system": "https://memory-api.example.com"
    },
    {
      "memory_id": "def456",
      "context_id": "user-context-1",
      "memory_uri": "memory://user-context-1/def456",
      "source_system": "https://memory-api.example.com"
    }
  ],
  "count": 2
}
```

**Status Codes:**
- `200 OK` - Query successful (may have 0 results)
- `400 Bad Request` - Missing or invalid `name` parameter
- `500 Internal Server Error` - Query failed

**Example:**
```bash
# Simple entity
curl "http://localhost:8080/api/v1/lookup/by-entity?name=John%20Smith"

# Entity with special characters
curl "http://localhost:8080/api/v1/lookup/by-entity?name=M%C3%BCnchen"
```

**Notes:**
- Entity names are case-sensitive
- URL encoding is required for special characters
- Returns empty array if entity not found (not an error)

---

### Lookup by Memory

Find all entities and relationships extracted from a specific memory.

**Endpoint:** `GET /api/v1/lookup/by-memory`

**Query Parameters:**
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `memory_id` | string | Yes | Memory identifier |
| `context_id` | string | Yes | Context identifier |

**Response:**
```json
{
  "memory_id": "abc123",
  "context_id": "user-context-1",
  "memory_uri": "memory://user-context-1/abc123",
  "found": true,
  "entities": [
    "John Smith",
    "New York",
    "Software Engineering"
  ],
  "relationships": [
    "John Smith works_in New York",
    "John Smith specializes_in Software Engineering"
  ],
  "entity_count": 3,
  "relationship_count": 2
}
```

**Status Codes:**
- `200 OK` - Query successful
- `400 Bad Request` - Missing required parameters
- `404 Not Found` - Memory not found in knowledge graph
- `500 Internal Server Error` - Query failed

**Example:**
```bash
curl "http://localhost:8080/api/v1/lookup/by-memory?memory_id=abc123&context_id=user-context-1"
```

**Notes:**
- Both parameters are required
- Returns entities extracted by LightRAG's NER
- Relationships show subject-predicate-object triples

---

### Resolve Memory URI

Parse and resolve a `memory://` URI to its components.

**Endpoint:** `GET /api/v1/lookup/resolve`

**Query Parameters:**
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `uri` | string | Yes | Memory URI (URL-encoded) |
| `fetch` | boolean | No | Fetch full memory from Memory API (default: false) |

**Response (without fetch):**
```json
{
  "memory_id": "abc123",
  "context_id": "user-context-1",
  "memory_uri": "memory://user-context-1/abc123",
  "valid": true,
  "source_system": "https://memory-api.example.com"
}
```

**Response (with fetch=true):**
```json
{
  "memory_id": "abc123",
  "context_id": "user-context-1",
  "memory_uri": "memory://user-context-1/abc123",
  "valid": true,
  "source_system": "https://memory-api.example.com",
  "memory": {
    "id": "abc123",
    "transcript": "Had a great meeting with the team...",
    "created_at": "2024-01-19T10:30:00Z",
    "audio": true,
    "image": false,
    "tags": ["work", "meeting"]
  }
}
```

**Status Codes:**
- `200 OK` - URI parsed successfully
- `400 Bad Request` - Invalid URI format
- `404 Not Found` - Memory not found (only with `fetch=true`)
- `500 Internal Server Error` - Fetch failed

**Example:**
```bash
# Parse only
curl "http://localhost:8080/api/v1/lookup/resolve?uri=memory://user-context-1/abc123"

# Parse and fetch
curl "http://localhost:8080/api/v1/lookup/resolve?uri=memory://user-context-1/abc123&fetch=true"
```

**Notes:**
- URI must use `memory://` scheme
- Format: `memory://{context_id}/{memory_id}`
- `fetch=true` makes additional call to Memory API

---

### Parse Multiple URIs

Parse a delimited string of memory URIs (as returned by LightRAG).

**Endpoint:** `POST /api/v1/lookup/parse-uris`

**Request Body:**
```json
{
  "uri_string": "memory://ctx1/mem1<SEP>memory://ctx1/mem2<SEP>memory://ctx2/mem3"
}
```

**Response:**
```json
{
  "count": 3,
  "memories": [
    {
      "memory_id": "mem1",
      "context_id": "ctx1",
      "memory_uri": "memory://ctx1/mem1"
    },
    {
      "memory_id": "mem2",
      "context_id": "ctx1",
      "memory_uri": "memory://ctx1/mem2"
    },
    {
      "memory_id": "mem3",
      "context_id": "ctx2",
      "memory_uri": "memory://ctx2/mem3"
    }
  ],
  "unique_ids": ["mem1", "mem2", "mem3"]
}
```

**Status Codes:**
- `200 OK` - URIs parsed successfully
- `400 Bad Request` - Invalid request body
- `500 Internal Server Error` - Parsing failed

**Example:**
```bash
curl -X POST http://localhost:8080/api/v1/lookup/parse-uris \
  -H "Content-Type: application/json" \
  -d '{
    "uri_string": "memory://ctx1/mem1<SEP>memory://ctx1/mem2<SEP>memory://ctx2/mem3"
  }'
```

**Notes:**
- LightRAG uses `<SEP>` as delimiter by default
- Invalid URIs are skipped (not returned)
- `unique_ids` contains deduplicated memory IDs

---

## Response Format

All API responses follow a consistent JSON structure:

### Success Response

```json
{
  // Endpoint-specific fields
  "memory_id": "abc123",
  "found": true,
  "count": 5
}
```

### Error Response

```json
{
  "error": "Invalid memory URI format",
  "details": "URI must start with 'memory://'",
  "timestamp": "2024-01-19T10:30:00Z"
}
```

---

## Error Handling

### HTTP Status Codes

| Code | Meaning | Description |
|------|---------|-------------|
| 200 | OK | Request successful |
| 400 | Bad Request | Invalid parameters or request body |
| 404 | Not Found | Resource not found |
| 500 | Internal Server Error | Server error, check logs |

### Error Response Fields

- `error` (string): Human-readable error message
- `details` (string, optional): Additional error context
- `timestamp` (string): ISO-8601 timestamp of error

### Common Errors

**Invalid URI Format:**
```json
{
  "error": "Invalid memory URI format",
  "details": "Expected format: memory://{context_id}/{memory_id}"
}
```

**Missing Parameters:**
```json
{
  "error": "Missing required parameter",
  "details": "Parameter 'name' is required"
}
```

**LightRAG Query Failed:**
```json
{
  "error": "Failed to query LightRAG",
  "details": "Connection refused: http://localhost:9621"
}
```

---

## Rate Limiting

Currently, no rate limiting is enforced. Future versions may add:
- Per-IP rate limits
- Per-endpoint quotas
- Burst protection

**Recommended Client Behavior:**
- Implement exponential backoff on 5xx errors
- Cache frequently accessed results
- Batch requests when possible

---

## Examples

### Complete Traceability Workflow

1. **Query LightRAG for entities:**
```bash
curl -X POST http://localhost:9621/query \
  -H "Content-Type: application/json" \
  -d '{"query": "Who works in New York?", "mode": "local"}'
```

2. **Extract memory URIs from sources:**
```json
{
  "response": "John Smith works in New York.",
  "sources": "memory://ctx1/abc123<SEP>memory://ctx1/def456"
}
```

3. **Parse URIs:**
```bash
curl -X POST http://localhost:8080/api/v1/lookup/parse-uris \
  -H "Content-Type: application/json" \
  -d '{"uri_string": "memory://ctx1/abc123<SEP>memory://ctx1/def456"}'
```

4. **Fetch original memories:**
```bash
curl "http://localhost:8080/api/v1/lookup/resolve?uri=memory://ctx1/abc123&fetch=true"
curl "http://localhost:8080/api/v1/lookup/resolve?uri=memory://ctx1/def456&fetch=true"
```

### Batch Processing

```python
import requests

# Collect URIs from multiple queries
uri_string = "<SEP>".join(all_uris)

# Parse in single request
response = requests.post(
    "http://localhost:8080/api/v1/lookup/parse-uris",
    json={"uri_string": uri_string}
)

memories = response.json()["memories"]
print(f"Found {len(memories)} unique memories")
```

### Entity Exploration

```bash
# Find all memories mentioning an entity
curl "http://localhost:8080/api/v1/lookup/by-entity?name=John%20Smith" | jq -r '.memories[].memory_uri'

# For each memory, fetch full context
for uri in $(curl -s "..." | jq -r '.memories[].memory_uri'); do
  curl "http://localhost:8080/api/v1/lookup/resolve?uri=$uri&fetch=true" | jq '.memory.transcript'
done
```

---

## Future Enhancements

Planned features for future API versions:

- **Temporal queries:** Filter memories by date range
- **Spatial queries:** Find memories near a location
- **Tag filtering:** Search by memory tags
- **Aggregations:** Count memories by entity, time, location
- **Webhooks:** Subscribe to new memory events
- **GraphQL API:** More flexible querying
- **WebSocket support:** Real-time updates

---

## Support

For issues or questions:
- **Documentation:** See `QUICKSTART.md` and `README.md`
- **Bug Reports:** Create an issue in the repository
- **Architecture:** See `docs/ARCHITECTURE.md`

---

**API Version:** 1.0.0
**Last Updated:** 2024-01-19
