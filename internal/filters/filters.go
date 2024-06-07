// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
package filters

import (
	"os"
	"strings"
	"gopkg.in/yaml.v3"
	"github.com/rs/zerolog/log"
)

type (
	Filters struct {
		Azqr *AzqrFilter `yaml:"azqr"`
	}

	AzqrFilter struct {
		Exclude *Exclude `yaml:"exclude"`
	}

	// Exclude - Struct for Exclude
	Exclude struct {
		Subscriptions   []string `yaml:"subscriptions,flow"`
		ResourceGroups  []string `yaml:"resourceGroups,flow"`
		Services        []string `yaml:"services,flow"`
		Recommendations []string `yaml:"recommendations,flow"`
		subscriptions   map[string]bool
		resourceGroups  map[string]bool
		services        map[string]bool
		recommendations map[string]bool
	}
)

func (e *Exclude) IsSubscriptionExcluded(subscriptionID string) bool {
	if e.subscriptions == nil {
		e.subscriptions = make(map[string]bool)
		for _, id := range e.Subscriptions {
			e.subscriptions[strings.ToLower(id)] = true
		}
	}

	_, ok := e.subscriptions[strings.ToLower(subscriptionID)]

	return ok
}

func (e *Exclude) IsResourceGroupExcluded(resourceGroupID string) bool {
	if e.resourceGroups == nil {
		e.resourceGroups = make(map[string]bool)
		for _, id := range e.ResourceGroups {
			e.resourceGroups[strings.ToLower(id)] = true
		}
	}

	_, ok := e.resourceGroups[strings.ToLower(resourceGroupID)]

	return ok
}

func (e *Exclude) IsServiceExcluded(serviceID string) bool {
	if e.services == nil {
		e.services = make(map[string]bool)
		for _, id := range e.Services {
			e.services[strings.ToLower(id)] = true
		}
	}

	_, ok := e.services[strings.ToLower(serviceID)]

	return ok
}

func (e *Exclude) IsRecommendationExcluded(recommendationID string) bool {
	if e.recommendations == nil {
		e.recommendations = make(map[string]bool)
		for _, id := range e.Recommendations {
			e.recommendations[strings.ToLower(id)] = true
		}
	}

	_, ok := e.recommendations[strings.ToLower(recommendationID)]

	return ok
}

func LoadFilters(exclusionsFile string) (*Filters) {
	exclusions := &Filters{
		Azqr: &AzqrFilter{
			Exclude: &Exclude{
				Subscriptions:  []string{},
				ResourceGroups: []string{},
				Services:       []string{},
			},
		},
	}

	if exclusionsFile != "" {
		data, err := os.ReadFile(exclusionsFile)
		if err != nil {
			log.Fatal().Err(err).Msgf("failed reading data from file: %s", exclusionsFile)
		}

		err = yaml.Unmarshal([]byte(data), &exclusions)
		if err != nil {
			log.Fatal().Err(err).Msgf("failed parsing yaml from file: %s", exclusionsFile)
		}
	}

	return exclusions
}
