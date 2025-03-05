// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
package scanners

import (
	"os"
	"strings"

	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

type (
	Filters struct {
		Azqr *AzqrFilter `yaml:"azqr"`
	}

	AzqrFilter struct {
		Include          *IncludeFilter `yaml:"include"`
		Exclude          *ExcludeFilter `yaml:"exclude"`
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
		Subscriptions   []string `yaml:"subscriptions,flow"`
		ResourceGroups  []string `yaml:"resourceGroups,flow"`
		Services        []string `yaml:"services,flow"`
		Recommendations []string `yaml:"recommendations,flow"`
	}

	// IncludeFilter - Struct for IncludeFilter
	IncludeFilter struct {
		Subscriptions  []string `yaml:"subscriptions,flow"`
		ResourceGroups []string `yaml:"resourceGroups,flow"`
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

func LoadFilters(filterFile string, scannerKeys []string) *Filters {
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

	filters.Azqr.xResourceGroups = make(map[string]bool)
	for _, id := range filters.Azqr.Exclude.ResourceGroups {
		filters.Azqr.xResourceGroups[strings.ToLower(id)] = true
	}

	filters.Azqr.xSubscriptions = make(map[string]bool)
	for _, id := range filters.Azqr.Exclude.Subscriptions {
		filters.Azqr.xSubscriptions[strings.ToLower(id)] = true
	}

	filters.Azqr.xServices = make(map[string]bool)
	for _, id := range filters.Azqr.Exclude.Services {
		filters.Azqr.xServices[strings.ToLower(id)] = true
	}

	filters.Azqr.xRecommendations = make(map[string]bool)
	for _, id := range filters.Azqr.Exclude.Recommendations {
		filters.Azqr.xRecommendations[strings.ToLower(id)] = true
	}

	s := []IAzureScanner{}

	if len(scannerKeys) > 1 && len(filters.Azqr.Include.ResourceTypes) > 0 {
		for _, key := range filters.Azqr.Include.ResourceTypes {
			if scannerList, exists := ScannerList[key]; exists {
				s = append(s, scannerList...)
			}
		}
	} else if len(scannerKeys) == 1 {
		s = append(s, ScannerList[scannerKeys[0]]...)
	} else {
		_, s = GetScanners()
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
