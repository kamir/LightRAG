package transformer

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/kamir/memory-connector/pkg/models"
	"github.com/kamir/memory-connector/pkg/utils"
)

// StandardStrategy provides basic transformation of memory to text
type StandardStrategy struct{}

// Name returns the strategy name
func (s *StandardStrategy) Name() string {
	return "standard"
}

// Transform converts a memory to a simple text format
//
// Standard transformation produces clean, focused content optimized for entity extraction.
// Content format:
//   [Memory from 2024-01-19 07:28:08 UTC]
//   {transcript}
//   [Audio recording available]
func (s *StandardStrategy) Transform(memory *models.Memory, config TransformConfig) (string, map[string]string, error) {
	if memory.Transcript == "" {
		return "", nil, fmt.Errorf("memory %s has no transcript", memory.ID)
	}

	// Build content
	var builder strings.Builder

	// 1. Temporal context (helps with entity extraction)
	parsedTime, err := memory.ParseCreatedAt()
	if err == nil {
		builder.WriteString(fmt.Sprintf("[Memory from %s]\n", parsedTime.UTC().Format("2006-01-02 15:04:05 MST")))
	}

	// 2. Primary content
	builder.WriteString(memory.Transcript)

	// 3. Media context (based on config)
	mediaContext := config.MediaContext
	if mediaContext == "" {
		mediaContext = "compact" // Default
	}

	if mediaContext != "none" {
		mediaInfo := []string{}
		if memory.HasAudio() {
			mediaInfo = append(mediaInfo, "Audio recording available")
		}
		if memory.HasImage() {
			mediaInfo = append(mediaInfo, "Image attachment available")
		}
		if len(mediaInfo) > 0 {
			builder.WriteString("\n[" + strings.Join(mediaInfo, ", ") + "]")
		}
	}

	// 4. Location context (if enabled and available)
	if memory.HasLocation() && config.EnrichLocation {
		builder.WriteString(fmt.Sprintf("\n[Location: %.4f, %.4f]", *memory.LocationLat, *memory.LocationLon))
	}

	// Build metadata
	metadata := buildCoreMetadata(memory, config)

	return builder.String(), metadata, nil
}

// RichStrategy provides enriched transformation with contextual information
type RichStrategy struct{}

// Name returns the strategy name
func (s *RichStrategy) Name() string {
	return "rich"
}

// Transform converts a memory to a rich, context-enhanced format
//
// Rich transformation includes comprehensive context and metadata.
// Content format:
//   ============================================================
//   MEMORY: {id}
//   Date: 2024-01-19 07:28:08 UTC
//   Type: memory
//   Collection: default_collection
//   ============================================================
//
//   Summary: {description}
//
//   Content:
//   {transcript}
//
//   Attachments:
//   Audio: gs://...
//   Image: gs://...
func (s *RichStrategy) Transform(memory *models.Memory, config TransformConfig) (string, map[string]string, error) {
	if memory.Transcript == "" {
		return "", nil, fmt.Errorf("memory %s has no transcript", memory.ID)
	}

	var builder strings.Builder

	// 1. Header with metadata
	parsedTime, err := memory.ParseCreatedAt()
	if err == nil {
		builder.WriteString("============================================================\n")
		builder.WriteString(fmt.Sprintf("MEMORY: %s\n", memory.ID))
		builder.WriteString(fmt.Sprintf("Date: %s\n", parsedTime.UTC().Format("2006-01-02 15:04:05 MST")))
		if memory.Type != "" {
			builder.WriteString(fmt.Sprintf("Type: %s\n", memory.Type))
		}
		if memory.CollectionName != "" {
			builder.WriteString(fmt.Sprintf("Collection: %s\n", memory.CollectionName))
		}
		builder.WriteString("============================================================\n\n")
	}

	// 2. Description (if available and different from transcript)
	if memory.Description != "" && memory.Description != memory.Transcript {
		builder.WriteString(fmt.Sprintf("Summary: %s\n\n", memory.Description))
	}

	// 3. Primary content
	builder.WriteString("Content:\n")
	builder.WriteString(memory.Transcript)
	builder.WriteString("\n")

	// 4. Media references (based on config)
	mediaContext := config.MediaContext
	if mediaContext == "" {
		mediaContext = "detailed" // Rich strategy defaults to detailed
	}

	if mediaContext != "none" {
		mediaInfo := []string{}
		if memory.HasAudio() {
			if mediaContext == "detailed" && memory.GcsUri != "" {
				mediaInfo = append(mediaInfo, fmt.Sprintf("Audio: %s", memory.GcsUri))
			} else if mediaContext == "compact" {
				mediaInfo = append(mediaInfo, "Audio recording available")
			}
		}
		if memory.HasImage() {
			if mediaContext == "detailed" && memory.GcsUriImg != "" {
				mediaInfo = append(mediaInfo, fmt.Sprintf("Image: %s", memory.GcsUriImg))
			} else if mediaContext == "compact" {
				mediaInfo = append(mediaInfo, "Image attachment available")
			}
		}
		if len(mediaInfo) > 0 {
			builder.WriteString("\nAttachments:\n")
			builder.WriteString(strings.Join(mediaInfo, "\n"))
			builder.WriteString("\n")
		}
	}

	// 5. Location (if available)
	if memory.HasLocation() {
		builder.WriteString("\n")
		if config.EnrichLocation {
			// Perform location enrichment
			enricher := utils.NewLocationEnricher(true)
			if enriched, err := enricher.Enrich(*memory.LocationLat, *memory.LocationLon, utils.DefaultPrecision); err == nil {
				builder.WriteString(fmt.Sprintf("Location: %s (%.4f, %.4f)\n", enriched.PlaceName, *memory.LocationLat, *memory.LocationLon))
				if enriched.Timezone != "" {
					builder.WriteString(fmt.Sprintf("Timezone: %s\n", enriched.Timezone))
				}
			} else {
				builder.WriteString(fmt.Sprintf("Location: %.4f, %.4f\n", *memory.LocationLat, *memory.LocationLon))
			}
		} else {
			builder.WriteString(fmt.Sprintf("Location: %.4f, %.4f\n", *memory.LocationLat, *memory.LocationLon))
		}
	}

	// 6. Tags
	if len(memory.Tags) > 0 {
		builder.WriteString(fmt.Sprintf("\nTags: %s\n", strings.Join(memory.Tags, ", ")))
	}

	// Build rich metadata
	metadata := buildRichMetadata(memory, config)

	return builder.String(), metadata, nil
}

// buildCoreMetadata creates the core metadata fields (used by both strategies)
func buildCoreMetadata(memory *models.Memory, config TransformConfig) map[string]string {
	metadata := make(map[string]string)

	if !config.IncludeMetadata {
		return metadata
	}

	// === SOURCE TRACEABILITY ===
	// Use memory:// URI scheme as approved
	contextID := memory.ContextID
	if contextID == "" {
		contextID = config.ContextID // Fallback to config
	}
	metadata["memory_id"] = memory.ID
	metadata["context_id"] = contextID
	metadata["file_path"] = fmt.Sprintf("memory://%s/%s", contextID, memory.ID)

	// Source system URL (if configured)
	if config.SourceSystem != "" {
		metadata["source_system"] = config.SourceSystem
	}

	// === TEMPORAL METADATA ===
	metadata["created_at"] = memory.CreatedAt
	if parsedTime, err := memory.ParseCreatedAt(); err == nil {
		metadata["created_timestamp"] = fmt.Sprintf("%d", parsedTime.Unix())
		metadata["created_year"] = fmt.Sprintf("%d", parsedTime.Year())
		metadata["created_month"] = fmt.Sprintf("%02d", int(parsedTime.Month()))
		metadata["created_date"] = parsedTime.Format("2006-01-02")
	}

	if memory.UpdatedAt != nil && *memory.UpdatedAt != memory.CreatedAt {
		metadata["updated_at"] = *memory.UpdatedAt
	}

	// === CLASSIFICATION ===
	if memory.Type != "" {
		metadata["memory_type"] = memory.Type
	}
	if memory.CollectionName != "" {
		metadata["collection_name"] = memory.CollectionName
	}
	if memory.TranscriptStatus != "" {
		metadata["transcript_status"] = memory.TranscriptStatus
	}

	// === MEDIA FLAGS ===
	if memory.HasAudio() {
		metadata["has_audio"] = "true"
	} else {
		metadata["has_audio"] = "false"
	}
	if memory.HasImage() {
		metadata["has_image"] = "true"
	} else {
		metadata["has_image"] = "false"
	}

	// === LOCATION ===
	if memory.HasLocation() {
		metadata["location_lat"] = fmt.Sprintf("%.6f", *memory.LocationLat)
		metadata["location_lon"] = fmt.Sprintf("%.6f", *memory.LocationLon)

		// Add geohash for spatial indexing
		if geohash, err := utils.EncodeWithDefault(*memory.LocationLat, *memory.LocationLon); err == nil {
			metadata["location_geohash"] = geohash
		}
	}

	// === TAGS ===
	if len(memory.Tags) > 0 {
		metadata["tags"] = strings.Join(memory.Tags, ",")
		metadata["tag_count"] = fmt.Sprintf("%d", len(memory.Tags))
	}

	// === TRANSFORMATION INFO ===
	metadata["transformation_strategy"] = "standard"
	metadata["ingestion_timestamp"] = time.Now().UTC().Format(time.RFC3339)
	metadata["connector_id"] = config.ConnectorID

	return metadata
}

// buildRichMetadata creates rich metadata (superset of core metadata)
func buildRichMetadata(memory *models.Memory, config TransformConfig) map[string]string {
	// Start with core metadata
	metadata := buildCoreMetadata(memory, config)

	if !config.IncludeMetadata {
		return metadata
	}

	// Override transformation strategy
	metadata["transformation_strategy"] = "rich"

	// === MEDIA REFERENCES (detailed) ===
	if memory.HasAudio() && memory.GcsUri != "" {
		metadata["audio_uri"] = memory.GcsUri
		// Extract filename from GCS URI
		metadata["audio_filename"] = filepath.Base(memory.GcsUri)
	}

	if memory.HasImage() && memory.GcsUriImg != "" {
		metadata["image_uri"] = memory.GcsUriImg
		metadata["image_filename"] = filepath.Base(memory.GcsUriImg)
	}

	// === LOCATION ENRICHMENT ===
	if memory.HasLocation() && config.EnrichLocation {
		metadata["location_enriched"] = "true"

		// Perform location enrichment (reverse geocoding, timezone)
		enricher := utils.NewLocationEnricher(true) // Enable caching
		if enriched, err := enricher.Enrich(*memory.LocationLat, *memory.LocationLon, utils.DefaultPrecision); err == nil {
			metadata["location_place_name"] = enriched.PlaceName
			if enriched.City != "" {
				metadata["location_city"] = enriched.City
			}
			if enriched.State != "" {
				metadata["location_state"] = enriched.State
			}
			if enriched.Country != "" {
				metadata["location_country"] = enriched.Country
			}
			if enriched.CountryCode != "" {
				metadata["location_country_code"] = enriched.CountryCode
			}
			if enriched.Timezone != "" {
				metadata["location_timezone"] = enriched.Timezone
			}
		}
	}

	// === TEMPORAL METADATA (enhanced) ===
	if parsedTime, err := memory.ParseCreatedAt(); err == nil {
		metadata["created_day"] = fmt.Sprintf("%02d", parsedTime.Day())
		metadata["created_hour"] = fmt.Sprintf("%02d", parsedTime.Hour())
		metadata["created_weekday"] = parsedTime.Weekday().String()
	}

	return metadata
}
