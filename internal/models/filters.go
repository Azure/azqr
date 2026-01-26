// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
package models

import (
	"fmt"
	"os"
	"strings"

	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

type (
	Filters struct {
		Azqr *AzqrFilter `yaml:"azqr" json:"azqr"`
	}

	AzqrFilter struct {
		Include          *IncludeFilter `yaml:"include" json:"include"`
		Exclude          *ExcludeFilter `yaml:"exclude" json:"exclude"`
		iSubscriptions   map[string]bool
		iResourceGroups  map[string]bool
		iResourceTypes   map[string]bool
		xSubscriptions   map[string]bool
		xResourceGroups  map[string]bool
		xServices        map[string]bool
		xRecommendations map[string]bool
		Scanners         []IAzureScanner
	}

	// ExcludeFilter - Struct for ExcludeFilter
	ExcludeFilter struct {
		Subscriptions   []string `yaml:"subscriptions,flow" json:"subscriptions"`
		ResourceGroups  []string `yaml:"resourceGroups,flow" json:"resourceGroups"`
		Services        []string `yaml:"services,flow" json:"services"`
		Recommendations []string `yaml:"recommendations,flow" json:"recommendations"`
	}

	// IncludeFilter - Struct for IncludeFilter
	IncludeFilter struct {
		Subscriptions  []string `yaml:"subscriptions,flow" json:"subscriptions"`
		ResourceGroups []string `yaml:"resourceGroups,flow" json:"resourceGroups"`
		ResourceTypes  []string `yaml:"resourceTypes,flow"`
	}
)

func (e *AzqrFilter) AddSubscription(subscriptionID string) {
	if e.iSubscriptions == nil {
		e.iSubscriptions = make(map[string]bool)
	}
	e.iSubscriptions[strings.ToLower(subscriptionID)] = true
	e.Include.Subscriptions = append(e.Include.Subscriptions, subscriptionID)
}

func (e *AzqrFilter) AddResourceGroup(resourceGroupID string) {
	if e.iResourceGroups == nil {
		e.iResourceGroups = make(map[string]bool)
	}
	e.iResourceGroups[strings.ToLower(resourceGroupID)] = true
	e.Include.ResourceGroups = append(e.Include.ResourceGroups, resourceGroupID)
}

func (e *AzqrFilter) IsSubscriptionExcluded(subscriptionID string) bool {
	_, ok := e.iSubscriptions[strings.ToLower(subscriptionID)]
	if ok {
		return false
	}

	// If not included, but there are included ressubscriptions, then exclude it
	if len(e.iSubscriptions) > 0 {
		return true
	}

	_, ok = e.xSubscriptions[strings.ToLower(subscriptionID)]
	return ok
}

func (e *AzqrFilter) IsServiceExcluded(resourceID string) bool {
	t := GetResourceTypeFromResourceID(resourceID)
	if _, included := e.iResourceTypes[strings.ToLower(t)]; included {
		sID := GetSubscriptionFromResourceID(resourceID)
		excluded := e.IsSubscriptionExcluded(sID)

		if !excluded {
			rgID := GetResourceGroupIDFromResourceID(resourceID)
			excluded = e.isResourceGroupExcluded(rgID)

			if !excluded {
				_, excluded = e.xServices[strings.ToLower(resourceID)]
			}
		}

		if excluded {
			log.Debug().Msgf("Service is excluded: %s", resourceID)
		}

		return excluded
	} else {
		log.Debug().Msgf("Service type is excluded: %s", t)
		return true
	}
}

func (e *AzqrFilter) IsRecommendationExcluded(recommendationID string) bool {
	_, ok := e.xRecommendations[strings.ToLower(recommendationID)]
	return ok
}

func (e *AzqrFilter) IsResourceTypeExcluded(resourceType string) bool {
	_, ok := e.iResourceTypes[strings.ToLower(resourceType)]
	return !ok
}

// validateResourceGroupID validates that a resource group ID matches the expected ARM format:
// /subscriptions/{subscription-id}/resourceGroups/{resource-group-name}
func validateResourceGroupID(resourceGroupID string) error {
	parts := strings.Split(resourceGroupID, "/")

	// ARM resource group ID should have exactly 5 parts: ["", "subscriptions", "sub-id", "resourceGroups", "rg-name"]
	if len(parts) != 5 {
		return fmt.Errorf("resource group ID '%s' has incorrect format. Expected format: /subscriptions/{subscription-id}/resourceGroups/{resource-group-name}", resourceGroupID)
	}

	// Validate the structure
	if parts[0] != "" || parts[1] != "subscriptions" || parts[3] != "resourceGroups" {
		return fmt.Errorf("resource group ID '%s' has incorrect format. Expected format: /subscriptions/{subscription-id}/resourceGroups/{resource-group-name}", resourceGroupID)
	}

	// Validate subscription ID and resource group name are not empty
	if parts[2] == "" {
		return fmt.Errorf("resource group ID '%s' has empty subscription ID. Expected format: /subscriptions/{subscription-id}/resourceGroups/{resource-group-name}", resourceGroupID)
	}

	if parts[4] == "" {
		return fmt.Errorf("resource group ID '%s' has empty resource group name. Expected format: /subscriptions/{subscription-id}/resourceGroups/{resource-group-name}", resourceGroupID)
	}

	return nil
}

func NewFilters() *Filters {
	filters := &Filters{
		Azqr: &AzqrFilter{
			Include: &IncludeFilter{
				Subscriptions:  []string{},
				ResourceGroups: []string{},
				ResourceTypes:  []string{},
			},
			Exclude: &ExcludeFilter{
				Subscriptions:   []string{},
				ResourceGroups:  []string{},
				Services:        []string{},
				Recommendations: []string{},
			},
			Scanners: []IAzureScanner{},
		},
	}
	return filters
}

func LoadFilters(filterFile string, scannerKeys []string) *Filters {
	filters := NewFilters()

	if filterFile != "" {
		data, err := os.ReadFile(filterFile)
		if err != nil {
			log.Fatal().Err(err).Msgf("failed reading data from file: %s", filterFile)
		}

		err = yaml.Unmarshal([]byte(data), &filters)
		if err != nil {
			log.Fatal().Err(err).Msgf("failed parsing yaml from file: %s", filterFile)
		}
	}

	filters.Azqr.iSubscriptions = make(map[string]bool)
	for _, id := range filters.Azqr.Include.Subscriptions {
		log.Debug().Msgf("Adding subscription to include: %s", id)
		filters.Azqr.iSubscriptions[strings.ToLower(id)] = true
	}

	// Validate resource group IDs in include list
	for _, id := range filters.Azqr.Include.ResourceGroups {
		if err := validateResourceGroupID(id); err != nil {
			log.Fatal().Err(err).Msgf("invalid resource group ID in include list")
		}
	}

	// Validate resource group IDs in exclude list
	for _, id := range filters.Azqr.Exclude.ResourceGroups {
		if err := validateResourceGroupID(id); err != nil {
			log.Fatal().Err(err).Msgf("invalid resource group ID in exclude list")
		}
	}

	filters.Azqr.iResourceGroups = make(map[string]bool)
	for _, id := range filters.Azqr.Include.ResourceGroups {
		log.Debug().Msgf("Adding resource group to include: %s", id)
		filters.Azqr.iResourceGroups[strings.ToLower(id)] = true
	}

	filters.Azqr.xResourceGroups = make(map[string]bool)
	for _, id := range filters.Azqr.Exclude.ResourceGroups {
		log.Debug().Msgf("Adding resource group to exclude: %s", id)
		filters.Azqr.xResourceGroups[strings.ToLower(id)] = true
	}

	filters.Azqr.xSubscriptions = make(map[string]bool)
	for _, id := range filters.Azqr.Exclude.Subscriptions {
		log.Debug().Msgf("Adding subscription to exclude: %s", id)
		filters.Azqr.xSubscriptions[strings.ToLower(id)] = true
	}

	filters.Azqr.xServices = make(map[string]bool)
	for _, id := range filters.Azqr.Exclude.Services {
		log.Debug().Msgf("Adding service to exclude: %s", id)
		filters.Azqr.xServices[strings.ToLower(id)] = true
	}

	filters.Azqr.xRecommendations = make(map[string]bool)
	for _, id := range filters.Azqr.Exclude.Recommendations {
		log.Debug().Msgf("Adding recommendation to exclude: %s", id)
		filters.Azqr.xRecommendations[strings.ToLower(id)] = true
	}

	s := []IAzureScanner{}

	switch {
	case len(scannerKeys) > 1 && len(filters.Azqr.Include.ResourceTypes) > 0:
		for _, key := range filters.Azqr.Include.ResourceTypes {
			if scannerList, exists := ScannerList[key]; exists {
				s = append(s, scannerList...)
			}
		}
		log.Debug().
			Strs("resource_types", filters.Azqr.Include.ResourceTypes).
			Int("scanners", len(s)).
			Msg("Loaded scanners by resource types")
	case len(scannerKeys) >= 1:
		for _, key := range scannerKeys {
			if scannerList, exists := ScannerList[key]; exists {
				s = append(s, scannerList...)
			}
		}
		log.Debug().
			Strs("scanner_keys", scannerKeys).
			Int("scanners", len(s)).
			Msg("Loaded scanners by keys")
	default:
		_, s = GetScanners()
		log.Debug().
			Int("scanners", len(s)).
			Msg("Loaded all scanners")
	}

	filters.Azqr.Scanners = s

	filters.Azqr.iResourceTypes = make(map[string]bool)
	for _, t := range s {
		for _, r := range t.ResourceTypes() {
			filters.Azqr.iResourceTypes[strings.ToLower(r)] = true
		}
	}

	return filters
}

func (e *AzqrFilter) isResourceGroupExcluded(resourceGroupID string) bool {
	// Check if the resource group is included
	_, ok := e.iResourceGroups[strings.ToLower(resourceGroupID)]
	if ok {
		return false
	}

	// If not included, but there are included resource groups, then exclude it
	if len(e.iResourceGroups) > 0 {
		return true
	}

	// Check if the resource group is excluded
	_, ok = e.xResourceGroups[strings.ToLower(resourceGroupID)]
	return ok
}
