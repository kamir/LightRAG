# Memory Connector Architecture

Comprehensive architecture documentation for the Memory API to LightRAG connector.

## Table of Contents

- [System Overview](#system-overview)
- [Architecture Diagram](#architecture-diagram)
- [Component Details](#component-details)
- [Data Flow](#data-flow)
- [Traceability Chain](#traceability-chain)
- [Storage Layer](#storage-layer)
- [API Layer](#api-layer)
- [Configuration](#configuration)
- [Security](#security)
- [Performance](#performance)
- [Deployment](#deployment)

---

## System Overview

The Memory Connector bridges two systems:

1. **Memory API** - Source system storing user memories with audio, images, and metadata
2. **LightRAG** - Knowledge graph system that extracts entities and relationships

**Key Responsibilities:**
- Fetch memories from Memory API
- Transform memories into LightRAG documents
- Preserve complete metadata for traceability
- Provide reverse lookup from knowledge graph to source memories
- Schedule automatic synchronization

---

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                         Memory Connector                          │
│                                                                   │
│  ┌──────────────┐   ┌──────────────┐   ┌──────────────┐        │
│  │   CLI Layer  │   │  HTTP API    │   │  Scheduler   │        │
│  │              │   │              │   │              │        │
│  │ • sync       │   │ • /health    │   │ • cron       │        │
│  │ • serve      │   │ • /lookup/*  │   │ • interval   │        │
│  │ • status     │   │              │   │              │        │
│  │ • list       │   │              │   │              │        │
│  └──────┬───────┘   └──────┬───────┘   └──────┬───────┘        │
│         │                  │                  │                 │
│         └──────────────────┼──────────────────┘                 │
│                            │                                    │
│              ┌─────────────▼─────────────┐                      │
│              │   Core Sync Engine        │                      │
│              │                           │                      │
│              │  • Fetch memories         │                      │
│              │  • Transform data         │                      │
│              │  • Track state            │                      │
│              │  • Handle errors          │                      │
│              └─────┬───────────────┬─────┘                      │
│                    │               │                            │
│       ┌────────────▼────┐    ┌────▼──────────┐                 │
│       │  Transformer    │    │  State Store  │                 │
│       │                 │    │               │                 │
│       │  • Standard     │    │  • JSON       │                 │
│       │  • Rich         │    │  • SQLite     │                 │
│       │  • Metadata     │    │               │                 │
│       └────────┬────────┘    └───────────────┘                 │
│                │                                                │
└────────────────┼────────────────────────────────────────────────┘
                 │
        ┌────────┴────────┐
        │                 │
   ┌────▼────┐      ┌─────▼──────┐
   │ Memory  │      │  LightRAG  │
   │   API   │      │    API     │
   │         │      │            │
   │ • GET   │      │ • POST     │
   │ /memory │      │ /documents │
   │         │      │ • POST     │
   │         │      │ /query     │
   └─────────┘      └────────────┘
```

---

## Component Details

### 1. CLI Layer (`cmd/memory-connector/`)

**Purpose:** Command-line interface for manual operations and service management.

**Commands:**
- `sync` - Run one-time sync for a connector
- `serve` - Start as a service with scheduler and HTTP API
- `status` - Check sync status and statistics
- `list` - List configured connectors
- `version` - Show version information

**Key Files:**
- `main.go` - Entry point and command definitions
- Commands call into core sync engine

---

### 2. HTTP API Layer (`pkg/api/`)

**Purpose:** RESTful API for reverse lookup and management.

**Components:**

**`server.go`** - HTTP server implementation
```go
type Server struct {
    config         *config.Config
    memoryClient   *client.MemoryClient
    lightragClient *client.LightRAGClient
    logger         *zap.Logger
}
```

**Features:**
- CORS middleware for web clients
- Request/response logging
- Graceful shutdown
- Health checks

**`lookup_handlers.go`** - Reverse lookup endpoints
- by-entity: Find memories by entity name
- by-memory: Find entities by memory ID
- resolve-uri: Parse and resolve memory URIs
- parse-uris: Batch parse delimited URIs

**Key Design Decisions:**
- Stateless handlers (no session state)
- JSON responses for all endpoints
- Placeholder responses (actual LightRAG queries TBD)

---

### 3. Core Sync Engine (`pkg/sync/`)

**Purpose:** Orchestrates the synchronization process.

**Workflow:**
1. **Fetch** - Get memories from Memory API
2. **Filter** - Skip already-processed memories
3. **Transform** - Convert to LightRAG documents
4. **Ingest** - Send to LightRAG
5. **Track** - Update state store
6. **Report** - Generate sync report

**State Management:**
```go
type SyncState struct {
    ConnectorID      string
    LastSyncTime     time.Time
    LastMemoryID     string
    ProcessedCount   int
    FailedItems      []FailedItem
}
```

**Error Handling:**
- Retries with exponential backoff
- Dead letter queue for failed items
- Detailed error logging

---

### 4. Transformer (`pkg/transformer/`)

**Purpose:** Convert Memory API data to LightRAG documents.

**Strategies:**

**Standard Strategy:**
- Clean transcript text
- Basic temporal context
- Essential metadata only
- **Use case:** Most deployments

**Rich Strategy:**
- Structured headers with context
- Full metadata preservation
- Media attachment information
- **Use case:** Enhanced context needs

**Metadata Builder:**
```go
func buildCoreMetadata(memory *models.Memory, cfg *models.ConnectorConfig) map[string]string
```

**Fields:**
- **Identity:** memory_id, context_id, file_path (URI)
- **Temporal:** created_at, timestamp, year, month, date
- **Classification:** memory_type, collection, status
- **Media:** has_audio, has_image, audio_uri, image_uri
- **Location:** lat, lon, geohash
- **Tags:** tags, tag_count
- **System:** connector_id, strategy, ingestion_timestamp

---

### 5. Client Layer (`pkg/client/`)

**Purpose:** API clients for external services.

**MemoryClient:**
```go
type MemoryClient struct {
    baseURL string
    apiKey  string
    client  *http.Client
}

func (c *MemoryClient) FetchMemories(contextID, queryRange string, limit int) (*MemoriesResponse, error)
```

**LightRAGClient:**
```go
type LightRAGClient struct {
    baseURL     string
    apiKey      string
    accessToken string // Auto-fetched for guest access
    client      *http.Client
}

func (c *LightRAGClient) InsertDocument(doc *Document) error
func (c *LightRAGClient) Query(query, mode string) (*QueryResponse, error)
```

**Authentication:**
- Auto-detects LightRAG auth status
- Fetches guest token if auth disabled
- Uses API key if auth enabled

---

### 6. State Store (`pkg/state/`)

**Purpose:** Persist sync state for incremental processing.

**Backends:**

**JSON Backend:**
```go
type JSONStateStore struct {
    dataDir string
}
```
- Simple file-based storage
- Good for development
- One file per connector: `{connector-id}.json`

**SQLite Backend:**
```go
type SQLiteStateStore struct {
    db *sql.DB
}
```
- Production-ready
- Concurrent access support
- Single database file

**Schema:**
```sql
CREATE TABLE sync_states (
    connector_id TEXT PRIMARY KEY,
    last_sync_time TIMESTAMP,
    last_memory_id TEXT,
    processed_count INTEGER,
    failed_items TEXT -- JSON array
);
```

---

### 7. Configuration (`pkg/config/`)

**Purpose:** Load and validate configuration.

**Structure:**
```yaml
memory_api:
  url: "https://memory-api.example.com"
  api_key: "${MEMCON_MEMORY_API_API_KEY}"

lightrag:
  url: "http://localhost:9621"
  api_key: "${MEMCON_LIGHTRAG_API_KEY}"

connectors:
  - id: "connector-1"
    enabled: true
    context_id: "user-context-1"
    source_system: "https://memory-api.example.com"

    schedule:
      type: "interval"
      interval_hours: 1

    ingestion:
      query_range: "day"
      query_limit: 100
      max_concurrency: 5

    transform:
      strategy: "standard"
      include_metadata: true
      enrich_location: false
      media_context: "compact"

storage:
  type: "sqlite"
  path: "./data/state.db"

logging:
  level: "info"
  format: "json"
  output_path: "./logs/connector.log"
```

**Features:**
- Environment variable substitution
- Validation on load
- Hot-reload support (future)

---

## Data Flow

### Ingestion Flow

```
Memory API
    │
    │ GET /memory/{context_id}?range=day&limit=100
    ▼
┌────────────────┐
│ Fetch Memories │
└───────┬────────┘
        │
        │ []Memory
        ▼
┌────────────────┐
│ Filter Seen    │──► State Store
└───────┬────────┘      (check last_memory_id)
        │
        │ []Memory (new only)
        ▼
┌────────────────┐
│ Transform      │
│                │
│ • Extract text │
│ • Build URI    │
│ • Add metadata │
└───────┬────────┘
        │
        │ []Document
        ▼
┌────────────────┐
│ Ingest to RAG  │
│                │
│ • Chunk text   │──► LightRAG
│ • Extract NER  │    POST /documents/text
│ • Build graph  │
└───────┬────────┘
        │
        │ Success/Failure
        ▼
┌────────────────┐
│ Update State   │──► State Store
│                │    (save last_memory_id)
└────────────────┘
```

### Query Flow (Traceability)

```
User Query: "Who works in New York?"
    │
    ▼
┌─────────────┐
│  LightRAG   │
│  Query API  │
└──────┬──────┘
       │
       │ Response with sources:
       │ "memory://ctx1/abc<SEP>memory://ctx1/def"
       ▼
┌─────────────────┐
│ Parse URIs      │
│ (parse-uris)    │
└────────┬────────┘
         │
         │ [memory_id, context_id]
         ▼
┌─────────────────┐
│ Resolve URIs    │
│ (resolve)       │
└────────┬────────┘
         │
         │ With fetch=true
         ▼
┌─────────────────┐
│  Memory API     │
│  GET /memory/id │
└────────┬────────┘
         │
         │ Full memory context
         ▼
    User sees:
    - Entity answer
    - Source memories
    - Original transcripts
```

---

## Traceability Chain

Complete path from knowledge graph to source data:

```
┌─────────────────────────────────────────────────────────────┐
│                    LightRAG Knowledge Graph                  │
│                                                              │
│  Entity: "John Smith"                                        │
│  ├─ source_id: "chunk_abc123_0<SEP>chunk_abc123_1"         │
│  └─ Relationships: ["works_in New York"]                    │
└──────────────────────────┬──────────────────────────────────┘
                           │
                           │ Lookup by source_id
                           ▼
┌─────────────────────────────────────────────────────────────┐
│                      LightRAG Chunks                         │
│                                                              │
│  Chunk: "chunk_abc123_0"                                    │
│  ├─ full_doc_id: "doc_abc123"                              │
│  ├─ content: "John Smith works in New York..."              │
│  └─ chunk_order_index: 0                                    │
└──────────────────────────┬──────────────────────────────────┘
                           │
                           │ Lookup by full_doc_id
                           ▼
┌─────────────────────────────────────────────────────────────┐
│                    LightRAG Documents                        │
│                                                              │
│  Document: "doc_abc123"                                     │
│  ├─ file_path: "memory://user-context-1/abc123"            │
│  ├─ content: "Full transcript..."                           │
│  └─ metadata: {...}                                         │
└──────────────────────────┬──────────────────────────────────┘
                           │
                           │ Parse memory:// URI
                           ▼
┌─────────────────────────────────────────────────────────────┐
│                      Connector API                           │
│                                                              │
│  GET /api/v1/lookup/resolve?uri=memory://ctx1/abc123        │
│  ├─ context_id: "user-context-1"                           │
│  └─ memory_id: "abc123"                                     │
└──────────────────────────┬──────────────────────────────────┘
                           │
                           │ Fetch original
                           ▼
┌─────────────────────────────────────────────────────────────┐
│                       Memory API                             │
│                                                              │
│  GET /memory/abc123                                         │
│  ├─ transcript: "Original text..."                          │
│  ├─ audio: true, gcs_uri: "gs://..."                       │
│  ├─ created_at: "2024-01-19T10:30:00Z"                     │
│  └─ tags: ["work", "meeting"]                               │
└─────────────────────────────────────────────────────────────┘
```

**Key Points:**
- `file_path` metadata is the bridge
- `memory://` URI scheme enables parsing
- Complete chain preserves all context
- No data loss in transformation

---

## Storage Layer

### State Persistence

**Purpose:** Track sync progress to enable incremental processing.

**Stored Data:**
- Last sync timestamp
- Last processed memory ID
- Total processed count
- Failed items (DLQ)

**Design Decisions:**
1. **No duplicate processing** - State prevents re-processing
2. **Resume capability** - Sync can resume after interruption
3. **Failure tracking** - Failed items stored for retry

### Backend Comparison

| Feature | JSON | SQLite |
|---------|------|--------|
| Setup | None | Schema migration |
| Concurrency | Single process | Multi-process |
| Performance | Fast for <10 connectors | Fast at scale |
| Backup | Copy files | Single DB file |
| Query | Full scan | Indexed queries |
| Use case | Development | Production |

---

## API Layer

### Server Architecture

**HTTP Server:**
- Standard library `net/http`
- Middleware chain: CORS → Logging → Handler
- Graceful shutdown on signals

**Routing:**
```
/health                          → HealthHandler
/api/v1/lookup/by-entity         → LookupByEntityHandler
/api/v1/lookup/by-memory         → LookupByMemoryHandler
/api/v1/lookup/resolve           → ResolveURIHandler
/api/v1/lookup/parse-uris        → ParseURIsHandler
```

**Middleware:**
1. **CORS:** Allow web clients
2. **Logging:** Request/response logging with zap
3. **Recovery:** Panic recovery (future)
4. **Auth:** API key validation (future)

---

## Configuration

### Hierarchy

1. **Config file** - YAML file (required)
2. **Environment variables** - Override config values
3. **CLI flags** - Override env and config

**Environment Variables:**
```bash
# Memory API
MEMCON_MEMORY_API_URL="https://..."
MEMCON_MEMORY_API_API_KEY="key"

# LightRAG
MEMCON_LIGHTRAG_URL="http://..."
MEMCON_LIGHTRAG_API_KEY="key"  # Only if auth enabled

# Server
MEMCON_SERVER_PORT=8080
MEMCON_LOG_LEVEL=info
```

### Validation

Configuration is validated on load:
- Required fields present
- URLs well-formed
- Cron expressions valid
- File paths accessible

---

## Security

### Current State

- **Authentication:** None (internal use)
- **Encryption:** None (local network)
- **Input validation:** Basic parameter checking

### Future Enhancements

1. **API Authentication:**
   - API key header
   - JWT tokens
   - OAuth2 integration

2. **Data Encryption:**
   - TLS for HTTP traffic
   - Encrypted state storage
   - Credential encryption

3. **Access Control:**
   - Per-connector permissions
   - Rate limiting
   - IP whitelisting

4. **Privacy:**
   - PII detection
   - Data masking
   - GDPR compliance

---

## Performance

### Ingestion Performance

**Target Metrics:**
- **Throughput:** 10 memories/second
- **Latency:** <100ms per memory (transform)
- **LightRAG insert:** <500ms per document
- **Total:** ~600ms per memory

**Optimization:**
- Concurrent processing (configurable)
- Batch LightRAG inserts (future)
- Connection pooling

### Query Performance

**Target Metrics:**
- **API latency:** <50ms (parse only)
- **With fetch:** <500ms (includes Memory API call)
- **LightRAG query:** <3s (configurable timeout)

**Caching Strategy (Future):**
- Cache parsed URIs (TTL: 1 hour)
- Cache memory metadata (TTL: 5 minutes)
- Cache entity lookups (TTL: 30 seconds)

### Scalability

**Current Limits:**
- **Memories:** 100K per context (tested)
- **Connectors:** 10 per instance
- **Concurrency:** 5 workers per connector

**Scaling Options:**
1. **Vertical:** Increase workers, memory
2. **Horizontal:** Multiple connector instances
3. **Sharding:** Separate LightRAG per context

---

## Deployment

### Standalone Binary

```bash
# Build
make build

# Run
./bin/memory-connector serve --config config.yaml
```

### Docker

```dockerfile
FROM golang:1.21 AS builder
COPY . /app
RUN make build

FROM alpine:latest
COPY --from=builder /app/bin/memory-connector /usr/local/bin/
CMD ["memory-connector", "serve"]
```

### Systemd Service

```ini
[Unit]
Description=Memory Connector
After=network.target

[Service]
Type=simple
User=memcon
ExecStart=/usr/local/bin/memory-connector serve --config /etc/memcon/config.yaml
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

### Kubernetes

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: memory-connector
spec:
  replicas: 1
  template:
    spec:
      containers:
      - name: connector
        image: memory-connector:latest
        env:
        - name: MEMCON_MEMORY_API_API_KEY
          valueFrom:
            secretKeyRef:
              name: memcon-secrets
              key: memory-api-key
```

---

## Monitoring

### Metrics (Future)

**Ingestion Metrics:**
- `memcon_sync_total{connector, status}` - Total syncs
- `memcon_memories_processed{connector}` - Memories processed
- `memcon_sync_duration_seconds{connector}` - Sync duration
- `memcon_errors_total{connector, type}` - Error count

**API Metrics:**
- `memcon_api_requests_total{endpoint, status}` - Request count
- `memcon_api_latency_seconds{endpoint}` - Latency histogram
- `memcon_api_errors_total{endpoint, error}` - Error count

**System Metrics:**
- `memcon_goroutines` - Active goroutines
- `memcon_memory_bytes` - Memory usage
- `memcon_db_connections` - DB connection pool

### Logging

**Structured Logging (zap):**
```go
logger.Info("Sync completed",
    zap.String("connector_id", id),
    zap.Int("processed", count),
    zap.Duration("duration", d),
)
```

**Log Levels:**
- `DEBUG` - Verbose, development only
- `INFO` - Normal operations
- `WARN` - Recoverable errors
- `ERROR` - Failed operations

---

## Design Principles

1. **Simplicity:** Prefer simple solutions over complex abstractions
2. **Traceability:** Complete audit trail from graph to source
3. **Idempotency:** Re-running syncs is safe
4. **Observability:** Comprehensive logging and metrics
5. **Modularity:** Clean interfaces between components
6. **Configuration:** Flexible, validated configuration
7. **Error Handling:** Graceful degradation, detailed errors

---

## Future Architecture

### Planned Enhancements

1. **Streaming Ingestion:**
   - Webhook receiver for real-time updates
   - WebSocket connection to Memory API
   - Event-driven processing

2. **Advanced Queries:**
   - GraphQL API for flexible queries
   - Temporal queries (date ranges)
   - Spatial queries (geofencing)
   - Tag-based filtering

3. **Multi-modal Processing:**
   - Audio transcription pipeline
   - Image entity extraction
   - Video analysis support

4. **Distributed Architecture:**
   - Worker pool for parallel processing
   - Message queue for job distribution
   - Shared state with distributed locks

---

## References

- [MAPPING_PLAN.md](MAPPING_PLAN.md) - Complete mapping strategy
- [API_REFERENCE.md](API_REFERENCE.md) - API documentation
- [METADATA_REFERENCE.md](METADATA_REFERENCE.md) - Metadata fields
- [QUICKSTART.md](../QUICKSTART.md) - Getting started guide
- [README.md](../README.md) - Project overview

---

**Document Version:** 1.0.0
**Last Updated:** 2024-01-19
