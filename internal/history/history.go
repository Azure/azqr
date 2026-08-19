// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

// Package history stores and reads aggregate azqr scan snapshots.
package history

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/Azure/azqr/internal/findings"
	"github.com/Azure/azqr/internal/models"
	"github.com/rs/zerolog/log"
)

const (
	// SchemaVersion is the current JSON Lines record schema.
	SchemaVersion = 1
	lockTimeout   = 2 * time.Second
	staleLockAge  = 30 * time.Second
)

// RecommendationCount stores aggregate data for one impacted recommendation.
type RecommendationCount struct {
	ID           string `json:"id"`
	Impact       string `json:"impact"`
	Category     string `json:"category"`
	ResourceType string `json:"resourceType"`
	Count        int    `json:"count"`
}

// Record is one aggregate scan snapshot.
type Record struct {
	SchemaVersion        int                   `json:"schemaVersion"`
	Timestamp            time.Time             `json:"timestamp"`
	AzqrVersion          string                `json:"azqrVersion"`
	ScopeID              string                `json:"scopeId"`
	Resources            int                   `json:"resources"`
	ImpactedResources    int                   `json:"impactedResources"`
	RecommendationsFound int                   `json:"recommendationsFound"`
	ImpactedByImpact     map[string]int        `json:"impactedByImpact"`
	ImpactedByCategory   map[string]int        `json:"impactedByCategory"`
	Recommendations      []RecommendationCount `json:"recommendations"`
}

// DefaultPath returns the default per-user history path.
func DefaultPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolve user config directory: %w", err)
	}
	return filepath.Join(configDir, "azqr", "history.jsonl"), nil
}

// ResolvePath returns the override or default history path.
func ResolvePath(override string) (string, error) {
	if override != "" {
		return filepath.Clean(override), nil
	}
	return DefaultPath()
}

// ScopeID returns a stable opaque ID for recommendation-relevant scan scope.
func ScopeID(params *models.ScanParams) (string, error) {
	if params == nil {
		return "", errors.New("scan parameters are required")
	}
	descriptor := struct {
		ManagementGroups []string        `json:"managementGroups"`
		Subscriptions    []string        `json:"subscriptions"`
		ResourceGroups   []string        `json:"resourceGroups"`
		ScannerKeys      []string        `json:"scannerKeys"`
		Stages           []string        `json:"stages"`
		Filters          *models.Filters `json:"filters"`
	}{
		ManagementGroups: sortedCopy(params.ManagementGroups),
		Subscriptions:    sortedCopy(params.Subscriptions),
		ResourceGroups:   sortedCopy(params.ResourceGroups),
		ScannerKeys:      sortedCopy(params.ScannerKeys),
		Filters:          params.Filters,
	}
	if params.Stages != nil {
		descriptor.Stages = sortedCopy(params.Stages.GetEnabledStages())
	}

	data, err := json.Marshal(descriptor)
	if err != nil {
		return "", fmt.Errorf("marshal scan scope: %w", err)
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:8]), nil
}

// NewRecord creates a versioned snapshot from a canonical findings summary.
func NewRecord(summary *findings.Summary, scopeID, azqrVersion string, timestamp time.Time) Record {
	record := Record{
		SchemaVersion:      SchemaVersion,
		Timestamp:          timestamp.UTC(),
		AzqrVersion:        azqrVersion,
		ScopeID:            scopeID,
		ImpactedByImpact:   map[string]int{},
		ImpactedByCategory: map[string]int{},
		Recommendations:    []RecommendationCount{},
	}
	if summary == nil {
		return record
	}

	record.Resources = summary.Resources
	record.ImpactedResources = summary.ImpactedResources
	record.RecommendationsFound = summary.RecommendationsFound
	record.ImpactedByImpact = copyCounts(summary.ImpactedByImpact)
	record.ImpactedByCategory = copyCounts(summary.ImpactedByCategory)
	for _, recommendation := range summary.Recommendations {
		if recommendation.ImpactedResources == 0 {
			continue
		}
		record.Recommendations = append(record.Recommendations, RecommendationCount{
			ID:           recommendation.ID,
			Impact:       recommendation.Impact,
			Category:     recommendation.Category,
			ResourceType: recommendation.ResourceType,
			Count:        recommendation.ImpactedResources,
		})
	}
	return record
}

// Append safely appends one record to a history file.
func Append(path string, record Record) error {
	if record.SchemaVersion != SchemaVersion {
		return fmt.Errorf("unsupported history schema version %d", record.SchemaVersion)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return fmt.Errorf("create history directory: %w", err)
	}

	release, err := acquireLock(path + ".lock")
	if err != nil {
		return err
	}
	defer release()

	data, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("marshal history record: %w", err)
	}
	data = append(data, '\n')

	file, err := os.OpenFile(filepath.Clean(path), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("open history file: %w", err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			log.Warn().Err(closeErr).Msg("Failed to close history file")
		}
	}()
	if err := file.Chmod(0600); err != nil {
		return fmt.Errorf("secure history file: %w", err)
	}
	if _, err := file.Write(data); err != nil {
		return fmt.Errorf("append history record: %w", err)
	}
	if err := file.Sync(); err != nil {
		return fmt.Errorf("sync history file: %w", err)
	}
	return nil
}

// Read loads and validates all complete history records.
func Read(path string) ([]Record, error) {
	data, err := os.ReadFile(filepath.Clean(path))
	if errors.Is(err, os.ErrNotExist) {
		return []Record{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read history file: %w", err)
	}

	lines := bytes.Split(data, []byte{'\n'})
	records := make([]Record, 0, len(lines))
	for index, line := range lines {
		if len(bytes.TrimSpace(line)) == 0 {
			continue
		}
		var record Record
		if err := json.Unmarshal(line, &record); err != nil {
			if index == len(lines)-1 && !bytes.HasSuffix(data, []byte{'\n'}) {
				log.Warn().Msg("Ignoring truncated final history record")
				break
			}
			return nil, fmt.Errorf("parse history record %d: %w", index+1, err)
		}
		if record.SchemaVersion != SchemaVersion {
			return nil, fmt.Errorf("history record %d uses unsupported schema version %d", index+1, record.SchemaVersion)
		}
		records = append(records, record)
	}
	return records, nil
}

// Select returns the requested scope and most recent records.
func Select(records []Record, scopeID string, last int) ([]Record, string, error) {
	if last < 1 {
		return nil, "", errors.New("last must be greater than zero")
	}

	if scopeID == "" {
		if len(records) == 0 {
			return []Record{}, "", nil
		}
		scopeID = records[len(records)-1].ScopeID
	}

	selected := make([]Record, 0, len(records))
	for _, record := range records {
		if record.ScopeID == scopeID {
			selected = append(selected, record)
		}
	}
	sort.SliceStable(selected, func(i, j int) bool {
		return selected[i].Timestamp.Before(selected[j].Timestamp)
	})
	if len(selected) > last {
		selected = selected[len(selected)-last:]
	}
	return selected, scopeID, nil
}

// Latest returns the most recent record for a scope.
func Latest(records []Record, scopeID string) *Record {
	var latest *Record
	for index := range records {
		record := &records[index]
		if record.ScopeID != scopeID {
			continue
		}
		if latest == nil || record.Timestamp.After(latest.Timestamp) {
			latest = record
		}
	}
	return latest
}

func acquireLock(path string) (func(), error) {
	deadline := time.Now().Add(lockTimeout)
	for {
		file, err := os.OpenFile(filepath.Clean(path), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if err == nil {
			_ = file.Close()
			return func() {
				if removeErr := os.Remove(path); removeErr != nil && !errors.Is(removeErr, os.ErrNotExist) {
					log.Warn().Err(removeErr).Msg("Failed to remove history lock")
				}
			}, nil
		}
		if !errors.Is(err, os.ErrExist) {
			return nil, fmt.Errorf("create history lock: %w", err)
		}
		if stat, statErr := os.Stat(path); statErr == nil && time.Since(stat.ModTime()) > staleLockAge {
			if removeErr := os.Remove(path); removeErr == nil {
				continue
			}
		}
		if time.Now().After(deadline) {
			return nil, fmt.Errorf("timed out waiting for history lock %s", filepath.Base(path))
		}
		time.Sleep(25 * time.Millisecond)
	}
}

func sortedCopy(values []string) []string {
	result := append([]string(nil), values...)
	for i := range result {
		result[i] = strings.ToLower(strings.TrimSpace(result[i]))
	}
	sort.Strings(result)
	return result
}

func copyCounts(source map[string]int) map[string]int {
	result := make(map[string]int, len(source))
	for key, value := range source {
		result[key] = value
	}
	return result
}
