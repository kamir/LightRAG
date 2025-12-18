# Memory Connector Testing Guide

Comprehensive testing guide for the Memory Connector system.

## Table of Contents

- [Overview](#overview)
- [Testing Levels](#testing-levels)
- [Test Setup](#test-setup)
- [Unit Testing](#unit-testing)
- [Integration Testing](#integration-testing)
- [End-to-End Testing](#end-to-end-testing)
- [Performance Testing](#performance-testing)
- [Test Data](#test-data)
- [Continuous Testing](#continuous-testing)

---

## Overview

The Memory Connector requires comprehensive testing at multiple levels:

1. **Unit Tests** - Individual components (transformers, clients, utils)
2. **Integration Tests** - Component interactions (sync engine, API)
3. **End-to-End Tests** - Complete workflows (fetch → transform → ingest)
4. **Performance Tests** - Throughput, latency, resource usage

---

## Testing Levels

### Unit Tests

**Scope:** Individual functions and methods

**Coverage Goals:**
- Transformers: 90%+
- Utilities (geohash, location): 95%+
- Models: 85%+
- Clients: 80%+ (mocked external APIs)

**Tools:**
- Go testing framework (`testing`)
- Table-driven tests
- Mock interfaces

### Integration Tests

**Scope:** Component interactions

**What to Test:**
- Sync engine with real state store
- API server with real clients
- Transformer with various memory types

**Tools:**
- Docker Compose for dependencies
- Test fixtures for sample data
- In-memory state stores

### End-to-End Tests

**Scope:** Complete user workflows

**Scenarios:**
- Manual sync command
- Scheduled sync with serve mode
- Reverse lookup API queries
- Error recovery and retry

**Tools:**
- Docker Compose with full stack
- Actual LightRAG instance
- Mock Memory API

### Performance Tests

**Scope:** System performance and scalability

**Metrics:**
- Ingestion rate (memories/second)
- API latency (p50, p95, p99)
- Memory usage
- Concurrency limits

**Tools:**
- Go benchmarks (`testing.B`)
- Apache Bench / wrk for API tests
- pprof for profiling

---

## Test Setup

### Prerequisites

```bash
# Install test dependencies
go get -u github.com/stretchr/testify/assert
go get -u github.com/stretchr/testify/mock

# Install Docker and Docker Compose (for integration tests)
# Install LightRAG (for E2E tests)
```

### Test Environment

```bash
# Create test data directory
mkdir -p test/fixtures
mkdir -p test/data

# Start test dependencies
docker-compose -f test/docker-compose.test.yml up -d

# Set test environment variables
export MEMCON_TEST_MODE=true
export MEMCON_MEMORY_API_URL="http://localhost:9999"  # Mock server
export MEMCON_LIGHTRAG_URL="http://localhost:9621"
```

---

## Unit Testing

### Transformer Tests

**File:** `pkg/transformer/strategies_test.go`

```go
package transformer

import (
	"testing"
	"time"

	"github.com/kamir/memory-connector/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestStandardStrategy_Transform(t *testing.T) {
	tests := []struct {
		name    string
		memory  *models.Memory
		config  TransformConfig
		wantErr bool
		checks  func(*testing.T, string, map[string]string)
	}{
		{
			name: "basic memory with transcript",
			memory: &models.Memory{
				ID:          "test123",
				ContextID:   "ctx1",
				Transcript:  "This is a test memory.",
				CreatedAt:   "2024-01-19T10:30:00Z",
				Audio:       true,
				Image:       false,
			},
			config: TransformConfig{
				Strategy:        "standard",
				IncludeMetadata: true,
				ConnectorID:     "test-connector",
			},
			wantErr: false,
			checks: func(t *testing.T, text string, metadata map[string]string) {
				assert.Contains(t, text, "This is a test memory")
				assert.Equal(t, "test123", metadata["memory_id"])
				assert.Equal(t, "ctx1", metadata["context_id"])
				assert.Equal(t, "memory://ctx1/test123", metadata["file_path"])
				assert.Equal(t, "true", metadata["has_audio"])
				assert.Equal(t, "false", metadata["has_image"])
			},
		},
		{
			name: "memory with location",
			memory: &models.Memory{
				ID:          "test456",
				ContextID:   "ctx1",
				Transcript:  "Meeting in Munich",
				CreatedAt:   "2024-01-19T10:30:00Z",
				LocationLat: ptrFloat64(48.1351),
				LocationLon: ptrFloat64(11.5820),
			},
			config: TransformConfig{
				Strategy:        "standard",
				IncludeMetadata: true,
				ConnectorID:     "test-connector",
			},
			wantErr: false,
			checks: func(t *testing.T, text string, metadata map[string]string) {
				assert.Equal(t, "48.135100", metadata["location_lat"])
				assert.Equal(t, "11.582000", metadata["location_lon"])
				assert.NotEmpty(t, metadata["location_geohash"])
			},
		},
		{
			name: "memory with tags",
			memory: &models.Memory{
				ID:         "test789",
				ContextID:  "ctx1",
				Transcript: "Important work meeting",
				CreatedAt:  "2024-01-19T10:30:00Z",
				Tags:       []string{"work", "meeting", "important"},
			},
			config: TransformConfig{
				Strategy:        "standard",
				IncludeMetadata: true,
				ConnectorID:     "test-connector",
			},
			wantErr: false,
			checks: func(t *testing.T, text string, metadata map[string]string) {
				assert.Equal(t, "work,meeting,important", metadata["tags"])
				assert.Equal(t, "3", metadata["tag_count"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			strategy := &StandardStrategy{}
			text, metadata, err := strategy.Transform(tt.memory, tt.config)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				tt.checks(t, text, metadata)
			}
		})
	}
}

func ptrFloat64(f float64) *float64 {
	return &f
}
```

### Geohash Tests

**File:** `pkg/utils/geohash_test.go`

```go
package utils

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncode(t *testing.T) {
	tests := []struct {
		name      string
		lat       float64
		lon       float64
		precision int
		want      string
		wantErr   bool
	}{
		{
			name:      "Munich coordinates",
			lat:       48.1351,
			lon:       11.5820,
			precision: 8,
			want:      "u0qj5v2k",
			wantErr:   false,
		},
		{
			name:      "New York coordinates",
			lat:       40.7128,
			lon:       -74.0060,
			precision: 8,
			want:      "dr5regw3",
			wantErr:   false,
		},
		{
			name:      "invalid latitude",
			lat:       91.0,
			lon:       0.0,
			precision: 8,
			wantErr:   true,
		},
		{
			name:      "invalid longitude",
			lat:       0.0,
			lon:       181.0,
			precision: 8,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Encode(tt.lat, tt.lon, tt.precision)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestDecode(t *testing.T) {
	tests := []struct {
		name    string
		geohash string
		wantLat float64
		wantLon float64
		wantErr bool
	}{
		{
			name:    "Munich geohash",
			geohash: "u0qj5v2k",
			wantLat: 48.1351,
			wantLon: 11.5820,
			wantErr: false,
		},
		{
			name:    "empty geohash",
			geohash: "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lat, lon, err := Decode(tt.geohash)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				// Allow small error due to geohash precision
				assert.InDelta(t, tt.wantLat, lat, 0.001)
				assert.InDelta(t, tt.wantLon, lon, 0.001)
			}
		})
	}
}

func TestDistance(t *testing.T) {
	// Munich to Berlin: ~504 km
	munich := "u0qj5v2k"
	berlin := "u33db3gz"

	dist, err := Distance(munich, berlin)
	assert.NoError(t, err)
	assert.InDelta(t, 504.0, dist, 10.0) // Allow 10km error
}

func BenchmarkEncode(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = Encode(48.1351, 11.5820, 8)
	}
}

func BenchmarkDecode(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _, _ = Decode("u0qj5v2k")
	}
}
```

### Memory URI Tests

**File:** `pkg/utils/memory_uri_test.go`

```go
package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseMemoryURI(t *testing.T) {
	tests := []struct {
		name           string
		uri            string
		wantContextID  string
		wantMemoryID   string
		wantErr        bool
	}{
		{
			name:          "valid URI",
			uri:           "memory://ctx1/mem123",
			wantContextID: "ctx1",
			wantMemoryID:  "mem123",
			wantErr:       false,
		},
		{
			name:    "invalid scheme",
			uri:     "http://ctx1/mem123",
			wantErr: true,
		},
		{
			name:    "missing memory ID",
			uri:     "memory://ctx1/",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseMemoryURI(tt.uri)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantContextID, result.ContextID)
				assert.Equal(t, tt.wantMemoryID, result.MemoryID)
			}
		})
	}
}
```

---

## Integration Testing

### Sync Engine Integration Test

**File:** `pkg/sync/engine_integration_test.go`

```go
// +build integration

package sync

import (
	"testing"

	"github.com/kamir/memory-connector/pkg/client"
	"github.com/kamir/memory-connector/pkg/config"
	"github.com/kamir/memory-connector/pkg/state"
	"github.com/stretchr/testify/assert"
)

func TestSyncEngine_FullWorkflow(t *testing.T) {
	// Setup
	cfg := &config.Config{
		MemoryAPI: config.MemoryAPIConfig{
			URL:    "http://localhost:9999", // Mock server
			APIKey: "test-key",
		},
		LightRAG: config.LightRAGConfig{
			URL: "http://localhost:9621",
		},
	}

	memoryClient := client.NewMemoryClient(cfg.MemoryAPI.URL, cfg.MemoryAPI.APIKey)
	lightragClient := client.NewLightRAGClient(cfg.LightRAG.URL, cfg.LightRAG.APIKey)
	stateStore := state.NewJSONStateStore("./test/data")

	// Create connector config
	connectorCfg := &models.ConnectorConfig{
		ID:        "test-connector",
		ContextID: "test-context",
		// ... other config
	}

	// Execute sync
	engine := NewSyncEngine(memoryClient, lightragClient, stateStore)
	report, err := engine.Sync(connectorCfg)

	// Verify
	assert.NoError(t, err)
	assert.NotNil(t, report)
	assert.Equal(t, "success", report.Status)
	assert.Greater(t, report.Processed, 0)
}
```

### API Server Integration Test

**File:** `pkg/api/server_integration_test.go`

```go
// +build integration

package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestServer_LookupEndpoints(t *testing.T) {
	// Setup test server
	server := setupTestServer(t)
	ts := httptest.NewServer(server.handler)
	defer ts.Close()

	tests := []struct {
		name       string
		endpoint   string
		wantStatus int
	}{
		{
			name:       "health check",
			endpoint:   "/health",
			wantStatus: http.StatusOK,
		},
		{
			name:       "lookup by entity",
			endpoint:   "/api/v1/lookup/by-entity?name=TestEntity",
			wantStatus: http.StatusOK,
		},
		{
			name:       "resolve URI",
			endpoint:   "/api/v1/lookup/resolve?uri=memory://ctx1/mem1",
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := http.Get(ts.URL + tt.endpoint)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantStatus, resp.StatusCode)
		})
	}
}
```

---

## End-to-End Testing

### Complete Workflow Test

**Script:** `test/e2e/test_complete_workflow.sh`

```bash
#!/bin/bash
set -e

echo "=== E2E Test: Complete Memory Ingestion Workflow ==="

# 1. Start services
echo "Starting dependencies..."
docker-compose -f test/docker-compose.test.yml up -d
sleep 5

# 2. Build connector
echo "Building connector..."
make build

# 3. Run sync
echo "Running sync..."
./bin/memory-connector sync --connector test-connector --config test/fixtures/test-config.yaml

# 4. Verify ingestion
echo "Verifying LightRAG ingestion..."
RESPONSE=$(curl -s http://localhost:9621/documents/count)
echo "Documents in LightRAG: $RESPONSE"

# 5. Test reverse lookup API
echo "Testing reverse lookup API..."
./bin/memory-connector serve --config test/fixtures/test-config.yaml &
SERVE_PID=$!
sleep 2

curl -s http://localhost:8080/health | jq .
curl -s "http://localhost:8080/api/v1/lookup/resolve?uri=memory://test-context/test-memory-1" | jq .

# 6. Cleanup
kill $SERVE_PID
docker-compose -f test/docker-compose.test.yml down

echo "✓ E2E test completed successfully"
```

---

## Performance Testing

### Ingestion Benchmark

**File:** `pkg/transformer/strategies_bench_test.go`

```go
package transformer

import (
	"testing"

	"github.com/kamir/memory-connector/pkg/models"
)

func BenchmarkStandardStrategy_Transform(b *testing.B) {
	strategy := &StandardStrategy{}
	memory := &models.Memory{
		ID:         "bench123",
		ContextID:  "ctx1",
		Transcript: "This is a benchmark test memory with some content.",
		CreatedAt:  "2024-01-19T10:30:00Z",
		Audio:      true,
		Image:      false,
	}
	config := TransformConfig{
		Strategy:        "standard",
		IncludeMetadata: true,
		ConnectorID:     "bench-connector",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = strategy.Transform(memory, config)
	}
}
```

### API Latency Test

```bash
# Install wrk
# brew install wrk (macOS)
# apt-get install wrk (Ubuntu)

# Test API latency
wrk -t4 -c100 -d30s --latency http://localhost:8080/health

# Test lookup endpoint
wrk -t4 -c100 -d30s --latency "http://localhost:8080/api/v1/lookup/resolve?uri=memory://ctx1/mem1"
```

### Expected Results

**Ingestion:**
- Standard transform: <1ms per memory
- Rich transform: <5ms per memory (without location enrichment)
- Rich transform: <200ms per memory (with location enrichment, cached)

**API:**
- Health check: <1ms
- Parse URI: <1ms
- Resolve URI (no fetch): <5ms
- Resolve URI (with fetch): <500ms

---

## Test Data

### Sample Memory Fixture

**File:** `test/fixtures/sample_memory.json`

```json
{
  "id": "test-memory-1",
  "context_id": "test-context",
  "transcript": "Had a great team meeting today to discuss the new project roadmap.",
  "created_at": "2024-01-19T10:30:00Z",
  "audio": true,
  "image": false,
  "gcs_uri": "gs://test-bucket/test-audio.mp3",
  "location_lat": 48.1351,
  "location_lon": 11.5820,
  "tags": ["work", "meeting", "roadmap"],
  "memory_type": "memory",
  "collection_name": "work_memories"
}
```

### Test Configuration

**File:** `test/fixtures/test-config.yaml`

```yaml
memory_api:
  url: "http://localhost:9999"
  api_key: "test-key"

lightrag:
  url: "http://localhost:9621"

connectors:
  - id: "test-connector"
    enabled: true
    context_id: "test-context"

    schedule:
      type: "manual"

    ingestion:
      query_range: "week"
      query_limit: 10
      max_concurrency: 2

    transform:
      strategy: "standard"
      include_metadata: true
      enrich_location: false

storage:
  type: "json"
  path: "./test/data"

logging:
  level: "debug"
  format: "console"
  output_path: "stdout"
```

---

## Continuous Testing

### GitHub Actions Workflow

**File:** `.github/workflows/test.yml`

```yaml
name: Tests

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

jobs:
  unit-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Run unit tests
        run: make test

      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          files: ./coverage.txt

  integration-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Start dependencies
        run: docker-compose -f test/docker-compose.test.yml up -d

      - name: Run integration tests
        run: make test-integration

  e2e-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Run E2E tests
        run: bash test/e2e/test_complete_workflow.sh
```

### Makefile Test Targets

```makefile
.PHONY: test test-unit test-integration test-e2e test-bench

test: test-unit

test-unit:
	go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...

test-integration:
	go test -v -tags=integration ./...

test-e2e:
	bash test/e2e/test_complete_workflow.sh

test-bench:
	go test -bench=. -benchmem ./...

test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
```

---

## Testing Checklist

Before releasing:

- [ ] All unit tests pass
- [ ] Integration tests pass
- [ ] E2E workflow test passes
- [ ] Benchmark results within targets
- [ ] Code coverage >80%
- [ ] No race conditions detected
- [ ] Error scenarios tested
- [ ] API contracts validated
- [ ] Documentation examples tested

---

## Related Documentation

- [ARCHITECTURE.md](ARCHITECTURE.md) - System architecture
- [API_REFERENCE.md](API_REFERENCE.md) - API documentation
- [QUICKSTART.md](../QUICKSTART.md) - Getting started guide

---

**Document Version:** 1.0.0
**Last Updated:** 2024-01-19
