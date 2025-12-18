# Metadata Field Reference

Complete reference for all metadata fields added to LightRAG documents during transformation.

## Table of Contents

- [Overview](#overview)
- [Metadata Categories](#metadata-categories)
- [Field Definitions](#field-definitions)
- [Transformation Strategy Differences](#transformation-strategy-differences)
- [Query Examples](#query-examples)
- [Best Practices](#best-practices)

---

## Overview

Every memory ingested into LightRAG includes comprehensive metadata fields that:

1. **Enable traceability** - Link entities back to source memories
2. **Support filtering** - Query by time, location, type, tags
3. **Preserve context** - Maintain original memory characteristics
4. **Audit trail** - Track when and how data was ingested

All metadata fields are stored as strings in LightRAG for maximum compatibility.

---

## Metadata Categories

### 1. Identity & Traceability
Fields that uniquely identify the source memory and enable reverse lookup.

### 2. Temporal
Fields for time-based queries and chronological analysis.

### 3. Classification
Fields that categorize and organize memories.

### 4. Media
Fields describing attached audio, images, and files.

### 5. Location
Fields for spatial queries and geographic analysis.

### 6. Tags
User-defined labels and categories.

### 7. System
Internal fields for tracking ingestion and processing.

---

## Field Definitions

### Identity & Traceability

#### `memory_id`
- **Type:** string
- **Required:** Yes
- **Example:** `"fJQAOybZ4366sQHKd40O"`
- **Description:** Original memory identifier from Memory API
- **Use case:** Reverse lookup from LightRAG to Memory API
- **Added by:** All strategies

#### `context_id`
- **Type:** string
- **Required:** Yes
- **Example:** `"107677460544181387647___mft"`
- **Description:** User or context identifier (multi-tenancy)
- **Use case:** Filter memories by user/context
- **Added by:** All strategies

#### `file_path`
- **Type:** string (URI)
- **Required:** Yes
- **Format:** `memory://{context_id}/{memory_id}`
- **Example:** `"memory://107677460544181387647___mft/fJQAOybZ4366sQHKd40O"`
- **Description:** Canonical URI for this memory (LightRAG's citation field)
- **Use case:** Primary traceability mechanism
- **Added by:** All strategies
- **Important:** This is THE key field that enables reverse lookup. LightRAG uses `file_path` for source citations.

#### `source_system`
- **Type:** string (URL)
- **Required:** No
- **Example:** `"https://memory-api.example.com"`
- **Description:** URL of the source Memory API instance
- **Use case:** Multi-source deployments
- **Added by:** All strategies (from connector config)

---

### Temporal

#### `created_at`
- **Type:** string (ISO-8601)
- **Required:** Yes
- **Format:** `YYYY-MM-DDTHH:MM:SSZ`
- **Example:** `"2024-01-19T07:28:08Z"`
- **Description:** Original creation timestamp from Memory API
- **Use case:** Preserve exact creation time
- **Added by:** All strategies

#### `created_timestamp`
- **Type:** string (Unix timestamp)
- **Required:** Yes
- **Example:** `"1705651688"`
- **Description:** Unix timestamp (seconds since epoch)
- **Use case:** Sortable, numeric time queries
- **Added by:** All strategies

#### `created_year`
- **Type:** string
- **Required:** Yes
- **Format:** `YYYY`
- **Example:** `"2024"`
- **Description:** Year extracted from created_at
- **Use case:** Annual aggregations, year-based filtering
- **Added by:** All strategies

#### `created_month`
- **Type:** string
- **Required:** Yes
- **Format:** `MM` (zero-padded)
- **Example:** `"01"` (January)
- **Description:** Month extracted from created_at
- **Use case:** Monthly reports, seasonal analysis
- **Added by:** All strategies

#### `created_date`
- **Type:** string
- **Required:** Yes
- **Format:** `YYYY-MM-DD`
- **Example:** `"2024-01-19"`
- **Description:** Date extracted from created_at
- **Use case:** Daily filtering, calendar views
- **Added by:** All strategies

#### `created_time`
- **Type:** string
- **Required:** No (Rich strategy only)
- **Format:** `HH:MM:SS`
- **Example:** `"07:28:08"`
- **Description:** Time of day extracted from created_at
- **Use case:** Time-of-day analysis
- **Added by:** Rich strategy only

#### `created_weekday`
- **Type:** string
- **Required:** No (Rich strategy only)
- **Example:** `"Friday"`
- **Description:** Day of week extracted from created_at
- **Use case:** Weekly patterns analysis
- **Added by:** Rich strategy only

---

### Classification

#### `memory_type`
- **Type:** string
- **Required:** No
- **Example:** `"memory"`, `"note"`, `"conversation"`
- **Description:** Classification of memory type
- **Use case:** Filter by memory type
- **Added by:** All strategies (if present in Memory API)

#### `collection_name`
- **Type:** string
- **Required:** No
- **Example:** `"default_collection"`, `"work_memories"`
- **Description:** Collection or category name
- **Use case:** Organize memories into groups
- **Added by:** All strategies (if present in Memory API)

#### `transcript_status`
- **Type:** string
- **Required:** No
- **Example:** `"processed"`, `"pending"`, `"failed"`
- **Description:** Status of transcript processing
- **Use case:** Filter by processing status
- **Added by:** All strategies (if present in Memory API)

---

### Media

#### `has_audio`
- **Type:** string (boolean)
- **Required:** Yes
- **Values:** `"true"` or `"false"`
- **Example:** `"true"`
- **Description:** Whether memory has audio attachment
- **Use case:** Filter memories with audio
- **Added by:** All strategies

#### `has_image`
- **Type:** string (boolean)
- **Required:** Yes
- **Values:** `"true"` or `"false"`
- **Example:** `"false"`
- **Description:** Whether memory has image attachment
- **Use case:** Filter memories with images
- **Added by:** All strategies

#### `audio_uri`
- **Type:** string (URI)
- **Required:** No (only if has_audio=true)
- **Example:** `"gs://project_p4te/107677...memory.mp3"`
- **Description:** GCS URI of audio file
- **Use case:** Retrieve original audio file
- **Added by:** All strategies (if audio present)

#### `audio_filename`
- **Type:** string
- **Required:** No (only if has_audio=true)
- **Example:** `"107677460544181387647___mft-memory.mp3"`
- **Description:** Original audio filename
- **Use case:** Display friendly filename
- **Added by:** All strategies (if audio present)

#### `image_uri`
- **Type:** string (URI)
- **Required:** No (only if has_image=true)
- **Example:** `"gs://project_p4te/image_107677...jpeg"`
- **Description:** GCS URI of image file
- **Use case:** Retrieve original image file
- **Added by:** All strategies (if image present)

#### `image_filename`
- **Type:** string
- **Required:** No (only if has_image=true)
- **Example:** `"image_107677460544181387647___mft.jpeg"`
- **Description:** Original image filename
- **Use case:** Display friendly filename
- **Added by:** All strategies (if image present)

#### `media_context`
- **Type:** string
- **Required:** No (Rich strategy only)
- **Values:** `"audio"`, `"image"`, `"audio+image"`, `"none"`
- **Example:** `"audio+image"`
- **Description:** Summary of available media
- **Use case:** Quick media availability check
- **Added by:** Rich strategy only

---

### Location

#### `location_lat`
- **Type:** string (decimal degrees)
- **Required:** No (only if location present)
- **Format:** Decimal degrees with up to 6 decimal places
- **Example:** `"48.123456"`
- **Range:** -90 to +90
- **Description:** Latitude coordinate
- **Use case:** Spatial queries, mapping
- **Added by:** All strategies (if location present in Memory API)

#### `location_lon`
- **Type:** string (decimal degrees)
- **Required:** No (only if location present)
- **Format:** Decimal degrees with up to 6 decimal places
- **Example:** `"11.567890"`
- **Range:** -180 to +180
- **Description:** Longitude coordinate
- **Use case:** Spatial queries, mapping
- **Added by:** All strategies (if location present in Memory API)

#### `location_geohash`
- **Type:** string
- **Required:** No (only if enrich_location=true)
- **Format:** Geohash string (configurable precision)
- **Example:** `"u0qj5v2k"`
- **Description:** Geohash encoding of lat/lon
- **Use case:** Proximity queries, spatial indexing
- **Added by:** Only if `enrich_location: true` in connector config
- **Implementation:** Phase 4 (advanced features)

#### `location_place_name`
- **Type:** string
- **Required:** No (only if enrich_location=true)
- **Example:** `"Munich, Bavaria, Germany"`
- **Description:** Reverse-geocoded place name
- **Use case:** Human-readable location display
- **Added by:** Only if `enrich_location: true` in connector config
- **Implementation:** Phase 4 (advanced features)

#### `location_country`
- **Type:** string
- **Required:** No (only if enrich_location=true)
- **Format:** ISO 3166-1 alpha-2
- **Example:** `"DE"` (Germany)
- **Description:** Country code from reverse geocoding
- **Use case:** Country-level aggregations
- **Added by:** Only if `enrich_location: true` in connector config
- **Implementation:** Phase 4 (advanced features)

#### `location_timezone`
- **Type:** string
- **Required:** No (only if enrich_location=true)
- **Format:** IANA timezone identifier
- **Example:** `"Europe/Berlin"`
- **Description:** Timezone at memory location
- **Use case:** Local time conversion
- **Added by:** Only if `enrich_location: true` in connector config
- **Implementation:** Phase 4 (advanced features)

---

### Tags

#### `tags`
- **Type:** string (comma-separated)
- **Required:** No
- **Format:** Comma-separated list
- **Example:** `"work,meeting,important"`
- **Description:** User-defined tags from Memory API
- **Use case:** Tag-based filtering, categorization
- **Added by:** All strategies (if tags present in Memory API)

#### `tag_count`
- **Type:** string (integer)
- **Required:** No (only if tags present)
- **Example:** `"3"`
- **Description:** Number of tags
- **Use case:** Analytics, validation
- **Added by:** All strategies (if tags present)

---

### System

#### `transformation_strategy`
- **Type:** string
- **Required:** Yes
- **Values:** `"standard"` or `"rich"`
- **Example:** `"standard"`
- **Description:** Which transformation strategy was used
- **Use case:** Debug, audit, strategy comparison
- **Added by:** All strategies

#### `connector_id`
- **Type:** string
- **Required:** Yes
- **Example:** `"connector-3"`
- **Description:** ID of connector that ingested this memory
- **Use case:** Track ingestion source, multi-connector setups
- **Added by:** All strategies

#### `ingestion_timestamp`
- **Type:** string (ISO-8601)
- **Required:** Yes
- **Format:** `YYYY-MM-DDTHH:MM:SSZ`
- **Example:** `"2025-12-18T18:30:00Z"`
- **Description:** When this memory was ingested into LightRAG
- **Use case:** Audit trail, sync debugging
- **Added by:** All strategies

#### `description`
- **Type:** string
- **Required:** No
- **Example:** `"Summary of the meeting"`
- **Description:** Optional description from Memory API
- **Use case:** Additional context, search
- **Added by:** All strategies (if present in Memory API)

---

## Transformation Strategy Differences

### Standard Strategy

**Goal:** Clean, focused content with essential metadata

**Includes:**
- All Identity & Traceability fields
- All Temporal fields (basic: year, month, date)
- All Classification fields
- All Media fields (flags and URIs)
- All Location fields (if present)
- All Tags fields
- All System fields

**Text Format:**
```
{transcript text only}
```

**Metadata Count:** ~20-25 fields

---

### Rich Strategy

**Goal:** Enhanced context with structured content

**Includes:**
- Everything in Standard
- Extended temporal fields (time, weekday)
- Media context summary
- Structured text headers

**Text Format:**
```
=== Memory from 2024-01-19 ===

{description if present}

Transcript:
{transcript}

Context:
- Created: Friday, 2024-01-19 at 07:28:08
- Media: audio + image
- Tags: work, meeting, important
- Location: Munich, Bavaria, Germany

[Audio available: gs://...]
[Image available: gs://...]
```

**Metadata Count:** ~25-30 fields

---

## Query Examples

### LightRAG Queries Using Metadata

**Note:** LightRAG stores metadata but querying is done through its Python API. These examples show conceptual queries.

#### Find Memories by Date

```python
# Query memories from specific date
query_params = {
    "created_date": "2024-01-19"
}
```

#### Find Memories with Audio

```python
# Query memories that have audio
query_params = {
    "has_audio": "true"
}
```

#### Find Memories by Tag

```python
# Query memories tagged "work"
query_params = {
    "tags": "*work*"  # Contains "work"
}
```

#### Find Memories by Location (if enriched)

```python
# Query memories in Munich
query_params = {
    "location_place_name": "*Munich*"
}
```

#### Find Memories from Specific Connector

```python
# Query memories from connector-1
query_params = {
    "connector_id": "connector-1"
}
```

---

### Reverse Lookup Queries

Using the Memory Connector API:

#### Trace Entity to Source Memories

```bash
# Find all memories mentioning "John Smith"
curl "http://localhost:8080/api/v1/lookup/by-entity?name=John%20Smith"
```

Response includes `memory_id`, `context_id`, and `memory_uri`.

#### Find Entities from Memory

```bash
# Find all entities extracted from specific memory
curl "http://localhost:8080/api/v1/lookup/by-memory?memory_id=abc123&context_id=ctx1"
```

#### Resolve URI to Full Memory

```bash
# Get full memory context from URI
curl "http://localhost:8080/api/v1/lookup/resolve?uri=memory://ctx1/abc123&fetch=true"
```

---

## Best Practices

### 1. Always Include `file_path`

The `file_path` (memory URI) is **critical** for traceability. LightRAG uses this field for source citations.

```go
metadata["file_path"] = fmt.Sprintf("memory://%s/%s", contextID, memoryID)
```

### 2. Use Consistent Formats

- **Timestamps:** Always ISO-8601 with timezone
- **Booleans:** Always `"true"` or `"false"` (lowercase strings)
- **Numbers:** Store as strings for LightRAG compatibility

### 3. Index Important Fields

Configure LightRAG to index frequently-queried fields:
- `created_date`
- `has_audio` / `has_image`
- `tags`
- `context_id`

### 4. Choose Strategy Appropriately

- **Standard:** Most use cases, cleaner entity extraction
- **Rich:** When context is critical, exploratory analysis

### 5. Enrich Location Sparingly

Location enrichment (geohashing, reverse geocoding) adds latency and API costs. Enable only when needed:

```yaml
transform:
  enrich_location: true  # Only if location queries are important
```

### 6. Document Custom Fields

If adding custom metadata fields, document them here for your team.

---

## Metadata Evolution

### Adding New Fields

When adding new metadata fields:

1. Update `pkg/transformer/strategies.go`
2. Add field definition to this document
3. Update `MAPPING_PLAN.md`
4. Consider backward compatibility
5. Test with existing data

### Deprecating Fields

When deprecating fields:

1. Mark as deprecated in this document
2. Keep field for 2-3 versions (grace period)
3. Log warnings when used
4. Remove in major version update

---

## Validation

### Required Fields (must always be present)

- `memory_id`
- `context_id`
- `file_path`
- `created_at`
- `created_timestamp`
- `created_year`
- `created_month`
- `created_date`
- `has_audio`
- `has_image`
- `transformation_strategy`
- `connector_id`
- `ingestion_timestamp`

### Conditional Fields (present if condition met)

- `audio_uri`, `audio_filename` → if `has_audio="true"`
- `image_uri`, `image_filename` → if `has_image="true"`
- `location_lat`, `location_lon` → if location in Memory API
- `location_geohash`, `location_place_name` → if `enrich_location=true`
- `tags`, `tag_count` → if tags in Memory API

### Optional Fields (may or may not be present)

- `source_system`
- `memory_type`
- `collection_name`
- `transcript_status`
- `description`

---

## Performance Considerations

### Metadata Size Impact

Each metadata field adds to document size:
- **Standard:** ~2-3 KB metadata per memory
- **Rich:** ~3-5 KB metadata per memory

**Impact on 100K memories:**
- Standard: ~300 MB metadata
- Rich: ~500 MB metadata

### Query Performance

Fields indexed by LightRAG:
- Fast: `file_path`, `memory_id`, `context_id`
- Medium: `created_date`, `has_audio`, `has_image`
- Slow: Full-text search on `tags`, `description`

---

## Examples

### Complete Metadata Example (Standard)

```json
{
  "memory_id": "fJQAOybZ4366sQHKd40O",
  "context_id": "107677460544181387647___mft",
  "file_path": "memory://107677460544181387647___mft/fJQAOybZ4366sQHKd40O",
  "source_system": "https://memory-api.example.com",
  "created_at": "2024-01-19T07:28:08Z",
  "created_timestamp": "1705651688",
  "created_year": "2024",
  "created_month": "01",
  "created_date": "2024-01-19",
  "memory_type": "memory",
  "collection_name": "default_collection",
  "transcript_status": "processed",
  "has_audio": "true",
  "has_image": "false",
  "audio_uri": "gs://project_p4te/107677460544181387647___mft-memory.mp3",
  "audio_filename": "107677460544181387647___mft-memory.mp3",
  "location_lat": "48.123456",
  "location_lon": "11.567890",
  "tags": "work,meeting,important",
  "tag_count": "3",
  "transformation_strategy": "standard",
  "connector_id": "connector-3",
  "ingestion_timestamp": "2025-12-18T18:30:00Z",
  "description": "Team standup meeting"
}
```

### Complete Metadata Example (Rich with Location Enrichment)

All fields from Standard, plus:

```json
{
  "created_time": "07:28:08",
  "created_weekday": "Friday",
  "media_context": "audio",
  "location_geohash": "u0qj5v2k",
  "location_place_name": "Munich, Bavaria, Germany",
  "location_country": "DE",
  "location_timezone": "Europe/Berlin"
}
```

---

## Related Documentation

- [MAPPING_PLAN.md](MAPPING_PLAN.md) - Complete mapping strategy
- [ARCHITECTURE.md](ARCHITECTURE.md) - System architecture
- [API_REFERENCE.md](API_REFERENCE.md) - API documentation
- [QUICKSTART.md](../QUICKSTART.md) - Getting started

---

**Document Version:** 1.0.0
**Last Updated:** 2024-01-19
