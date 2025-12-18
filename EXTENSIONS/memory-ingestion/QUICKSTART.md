# Memory Connector - Quick Start Guide

Get the Memory API ingestion connector running in 5 minutes.

## Prerequisites

- Go 1.21 or later
- Memory API URL and API key
- LightRAG instance running in API mode

## Step 1: Build the Connector

```bash
cd EXTENSIONS/memory-ingestion

# Install dependencies
make deps

# Build the binary
make build

# Verify installation
./bin/memory-connector version
```

## Step 2: Configure Your First Connector

Edit `configs/config.yaml`:

```yaml
memory_api:
  url: "https://your-memory-api.com"  # Your Memory API URL
  api_key: ""  # Leave empty, set via env var

lightrag:
  url: "http://localhost:9621"  # Your LightRAG API endpoint

connectors:
  - id: "my-first-connector"
    enabled: true
    context_id: "your-context-id-here"  # Replace with your context ID

    schedule:
      type: "interval"
      interval_hours: 1  # Sync every hour

    ingestion:
      query_range: "day"
      query_limit: 100
      max_concurrency: 5

    transform:
      strategy: "standard"  # or "rich" for enhanced context
      include_metadata: true
```

## Step 3: Set Your API Keys

**Memory API (Required):**

```bash
export MEMCON_MEMORY_API_API_KEY="your-memory-api-key"
```

**LightRAG API (Optional - only if authentication is enabled):**

First, check if your LightRAG server has authentication enabled:
```bash
curl http://localhost:9621/auth-status
```

If `"auth_configured": true`, set the API key:
```bash
export MEMCON_LIGHTRAG_API_KEY="your-lightrag-api-key"
```

If `"auth_configured": false`, skip this step - authentication is disabled.

**Alternative:** Set directly in config.yaml (not recommended for production):

```yaml
memory_api:
  api_key: "your-memory-api-key"
lightrag:
  api_key: "your-lightrag-api-key"  # Only if auth is enabled
```

## Step 3.5: Test API Endpoints (Optional)

Before running a full sync, verify your API endpoints are working:

**Test LightRAG Server:**

```bash
# Check authentication status
curl http://localhost:9621/auth-status
```

If `auth_configured` is `true`, you need to set the API key. If `false`, authentication is disabled and you can skip setting `MEMCON_LIGHTRAG_API_KEY`.

```bash
# Health check
curl http://localhost:9621/health

# Insert a test document (with API key if auth is enabled)
curl -X POST http://localhost:9621/documents/text \
  -H "Content-Type: application/json" \
  -H "X-API-Key: YOUR_API_KEY" \
  -d '{
    "text": "This is a test memory from my connector setup.",
    "metadata": {
      "source": "test",
      "timestamp": "2024-01-18T12:00:00Z"
    }
  }'

# Or without API key if auth is disabled
curl -X POST http://localhost:9621/documents/text \
  -H "Content-Type: application/json" \
  -d '{
    "text": "This is a test memory from my connector setup.",
    "metadata": {
      "source": "test",
      "timestamp": "2024-01-18T12:00:00Z"
    }
  }'
```

Expected response:
```json
{
  "status": "success",
  "doc_id": "some-document-id"
}
```

**Test Memory API:**

```bash
# Fetch memories for debugging
curl -X GET "https://your-memory-api.com/memory/YOUR_CONTEXT_ID?limit=10&range=day" \
  -H "X-API-KEY: your-memory-api-key"
```

Expected response:
```json
{
  "count": 10,
  "memories": [
    {
      "id": "...",
      "transcript": "...",
      "created_at": "2024-01-18T12:00:00Z",
      ...
    }
  ]
}
```

If either test fails, fix the issue before proceeding to the next step.

## Step 4: Run Your First Sync

**Manual one-time sync:**

```bash
./bin/memory-connector sync --connector my-first-connector
```

You should see output like:

```
=== Sync Report ===
Connector ID: my-first-connector
Status: success
Duration: 2.5s
Fetched: 50
Processed: 50
Skipped: 0
Failed: 0
Success Rate: 100.00%
```

**View sync status:**

```bash
./bin/memory-connector status --connector my-first-connector
```

**JSON output:**

```bash
./bin/memory-connector sync --connector my-first-connector --json
```

## Step 5: Run as a Service (Optional)

For continuous automated syncs:

```bash
./bin/memory-connector serve --config configs/config.yaml
```

This will:
- Start the HTTP management API on port 8080
- Schedule automatic syncs based on your connector schedules
- Keep running in the background

**Stop with:** `Ctrl+C`

## Common Configuration Patterns

### Hourly Sync (Most Common)

```yaml
schedule:
  type: "interval"
  interval_hours: 1
```

### Daily at Midnight

```yaml
schedule:
  type: "cron"
  cron_expr: "0 0 0 * * *"
```

### Manual Trigger Only

```yaml
schedule:
  type: "manual"
```

Then trigger via CLI: `./bin/memory-connector sync --connector my-connector`

## Transformation Strategies

### Standard (Recommended for most use cases)

Simple transcript extraction with basic metadata:

```yaml
transform:
  strategy: "standard"
  include_metadata: true
  enrich_location: false
```

### Rich (For enhanced context)

Includes temporal context, location enrichment, and media information:

```yaml
transform:
  strategy: "rich"
  include_metadata: true
  enrich_location: true
```

## Storage Backends

### JSON (Default, Simple)

Good for development and small-scale deployments:

```yaml
storage:
  type: "json"
  path: "./data"
```

State files are stored as: `./data/connector-id.json`

### SQLite (Recommended for Production)

Better performance and concurrent access:

```yaml
storage:
  type: "sqlite"
  path: "./data/state.db"
```

## Logging Configuration

### Console (Development)

```yaml
logging:
  level: "debug"  # debug, info, warn, error
  format: "console"
  output_path: "stdout"
```

### JSON (Production)

```yaml
logging:
  level: "info"
  format: "json"
  output_path: "./logs/connector.log"
```

## Troubleshooting

### "Failed to fetch memories"

- Check Memory API URL is correct
- Verify API key is set correctly
- Ensure context_id exists in Memory API
- **Test manually:**
  ```bash
  curl -X GET "https://your-memory-api.com/memory/YOUR_CONTEXT_ID?limit=10&range=day" \
    -H "X-API-KEY: your-memory-api-key"
  ```

### "Failed to insert document" or "API Key required"

- Check LightRAG is running: `curl http://localhost:9621/health`
- Verify LightRAG URL in config
- **Check if authentication is enabled:**
  ```bash
  curl http://localhost:9621/auth-status
  ```
  - If `"auth_configured": true` → Set `MEMCON_LIGHTRAG_API_KEY` environment variable
  - If `"auth_configured": false` → Leave `MEMCON_LIGHTRAG_API_KEY` unset (auth disabled)
- **Test manually with API key:**
  ```bash
  curl -X POST http://localhost:9621/documents/text \
    -H "Content-Type: application/json" \
    -H "X-API-Key: YOUR_API_KEY" \
    -d '{"text": "test", "metadata": {}}'
  ```
- **Or test without API key if auth is disabled:**
  ```bash
  curl -X POST http://localhost:9621/documents/text \
    -H "Content-Type: application/json" \
    -d '{"text": "test", "metadata": {}}'
  ```

### "No new memories found"

- All memories already processed (expected behavior)
- Check sync state: `./bin/memory-connector status --connector my-connector`
- Adjust `query_range` in config (try "week" or "month")

### View detailed logs

```bash
# Enable debug logging
./bin/memory-connector sync --connector my-connector --config <(
  yq eval '.logging.level = "debug"' configs/config.yaml
)
```

## Docker Quick Start

### Build and Run

```bash
# Build image
make docker-build

# Run manual sync
docker run --rm \
  -v $(pwd)/configs:/app/configs \
  -v $(pwd)/data:/app/data \
  -e MEMCON_MEMORY_API_API_KEY="your-api-key" \
  memory-connector:latest sync --connector my-first-connector

# Run as service
docker run -d \
  --name memory-connector \
  -v $(pwd)/configs:/app/configs \
  -v $(pwd)/data:/app/data \
  -e MEMCON_MEMORY_API_API_KEY="your-api-key" \
  -p 8080:8080 \
  memory-connector:latest serve
```

## Verify Everything Works

**1. List configured connectors:**
```bash
./bin/memory-connector list
```

**2. Check connector status:**
```bash
./bin/memory-connector status --connector my-first-connector
```

**3. Run a test sync:**
```bash
./bin/memory-connector sync --connector my-first-connector
```

**4. Check state was saved:**
```bash
# For JSON backend
cat data/my-first-connector.json | jq .

# For SQLite backend
sqlite3 data/state.db "SELECT * FROM sync_states;"
```

## Next Steps

- **Production Deployment:** See `deployments/` directory for systemd and Kubernetes manifests
- **Multiple Connectors:** Add more connector configs in `configs/config.yaml`
- **Advanced Configuration:** See full README.md for all options
- **Testing:** Run `make test` to verify your setup
- **Monitoring:** Check failed items DLQ: `./bin/memory-connector status --connector my-connector --json | jq '.failed_items'`

## Need Help?

- **Full Documentation:** `README.md`
- **Architecture Details:** `../../FEATURES/FR01-memory-ingestion/`
- **Configuration Reference:** `configs/config.yaml` (commented examples)
- **Issues:** Create an issue in the repository

---

**That's it!** You're now ingesting memories into LightRAG automatically. 🎉
