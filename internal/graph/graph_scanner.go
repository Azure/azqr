// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package graph

import (
	"context"
	"embed"
	"io/fs"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

//go:embed aprl/azure-resources/**/**/*.yaml
//go:embed aprl/azure-resources/**/**/kql/*.kql
//go:embed aprl/azure-specialized-workloads/**/*.yaml
//go:embed aprl/azure-specialized-workloads/**/kql/*.kql
//go:embed azure-orphan-resources/**/*.yaml
//go:embed azure-orphan-resources/**/kql/*.kql
var embededFiles embed.FS

type (
	AprlScanner struct {
		scanType        []ScanType
		serviceScanners []models.IAzureScanner
		filters         *models.Filters
		subscriptions   map[string]string
	}

	ScanType string
)

const (
	AprlScanType   ScanType = "aprl/azure-resources"
	OrphanScanType ScanType = "azure-orphan-resources"
)

// create a new APRL scanner
func NewAprlScanner(serviceScanners []models.IAzureScanner, filters *models.Filters, subscriptions map[string]string) AprlScanner {
	return AprlScanner{
		scanType: []ScanType{
			AprlScanType,
			OrphanScanType,
		},
		serviceScanners: serviceScanners,
		filters:         filters,
		subscriptions:   subscriptions,
	}
}

// GetAprlRecommendations returns a map with all APRL recommendations
func (a AprlScanner) GetAprlRecommendations() map[string]map[string]models.AprlRecommendation {
	recommendations := map[string]map[string]models.AprlRecommendation{}
	for _, t := range a.scanType {
		source := "APRL"
		if t == OrphanScanType {
			source = "AOR"
		}
		rs := a.getAprlRecommendations(string(t))
		for t, r := range rs {
			for _, r := range r {
				if recommendations[t] == nil {
					recommendations[t] = map[string]models.AprlRecommendation{}
				}
				r.Source = source
				recommendations[t][r.RecommendationID] = r
			}
		}
	}
	return recommendations
}

func (a AprlScanner) getAprlRecommendations(path string) map[string]map[string]models.AprlRecommendation {
	r := map[string]map[string]models.AprlRecommendation{}

	fsys, err := fs.Sub(embededFiles, path)
	if err != nil {
		return nil
	}

	q := map[string]string{}
	err = fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(path, ".kql") {
			content, err := fs.ReadFile(fsys, path)
			if err != nil {
				return err
			}

			fileName := strings.TrimSuffix(d.Name(), ".kql")
			q[fileName] = string(content)
		}
		return nil
	})
	if err != nil {
		return nil
	}

	err = fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(path, ".yaml") {
			content, err := fs.ReadFile(fsys, path)
			if err != nil {
				return err
			}

			var recommendations []models.AprlRecommendation
			err = yaml.Unmarshal(content, &recommendations)
			if err != nil {
				return err
			}

			for _, recommendation := range recommendations {
				t := strings.ToLower(recommendation.ResourceType)
				if _, ok := r[t]; !ok {
					r[t] = map[string]models.AprlRecommendation{}
				}

				if i, ok := q[recommendation.RecommendationID]; ok {
					recommendation.GraphQuery = i
				}

				r[t][recommendation.RecommendationID] = recommendation
			}

		}
		return nil
	})
	if err != nil {
		return nil
	}

	return r
}

// AprlScan scans Azure resources using Azure Proactive Resiliency Library v2 (APRL)
func (a AprlScanner) Scan(ctx context.Context, cred azcore.TokenCredential) (map[string]map[string]models.AprlRecommendation, []models.AprlResult) {
	recommendations := map[string]map[string]models.AprlRecommendation{}
	results := []models.AprlResult{}
	rules := []models.AprlRecommendation{}
	graph := NewGraphQuery(cred)

	// get APRL recommendations
	aprl := a.GetAprlRecommendations()

	for _, s := range a.serviceScanners {
		for _, t := range s.ResourceTypes() {
			models.LogResourceTypeScan(t)
			gr := a.getGraphRules(t, aprl)
			for _, r := range gr {
				rules = append(rules, r)
			}

			for i, r := range gr {
				if recommendations[strings.ToLower(t)] == nil {
					recommendations[strings.ToLower(t)] = map[string]models.AprlRecommendation{}
				}
				recommendations[strings.ToLower(t)][i] = r
			}
		}
	}

	batches := int(math.Ceil(float64(len(rules)) / 12))

	jobs := make(chan []models.AprlRecommendation, batches)
	ch := make(chan []models.AprlResult, batches)
	var wg sync.WaitGroup

	// Start workers
	numWorkers := 12 // Define the number of workers in the pool
	for w := 0; w < numWorkers; w++ {
		go a.worker(ctx, graph, a.subscriptions, jobs, ch, &wg)
	}
	wg.Add(batches)

	batchSize := 12
	for i := 0; i < len(rules); i += batchSize {
		j := i + batchSize
		if j > len(rules) {
			j = len(rules)
		}

		jobs <- rules[i:j]

		// Staggering queries to avoid throttling. Max 15 queries each 5 seconds.
		// https://learn.microsoft.com/en-us/azure/governance/resource-graph/concepts/guidance-for-throttled-requests#staggering-queries
		time.Sleep(5 * time.Second)
	}

	// Wait for all workers to finish
	close(jobs)
	wg.Wait()

	for i := 0; i < batches; i++ {
		res := <-ch
		for _, r := range res {
			if a.filters.Azqr.IsServiceExcluded(r.ResourceID) {
				continue
			}
			results = append(results, r)
		}
	}

	return recommendations, results
}

func (a *AprlScanner) worker(ctx context.Context, graph *GraphQuery, subscriptions map[string]string, jobs <-chan []models.AprlRecommendation, results chan<- []models.AprlResult, wg *sync.WaitGroup) {
	for r := range jobs {
		res, err := a.graphScan(ctx, graph, r, subscriptions)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to scan")
		}
		results <- res
		wg.Done()
	}
}

func (a AprlScanner) graphScan(ctx context.Context, graphClient *GraphQuery, rules []models.AprlRecommendation, subscriptions map[string]string) ([]models.AprlResult, error) {
	results := []models.AprlResult{}
	subs := make([]*string, 0, len(subscriptions))
	for s := range subscriptions {
		subs = append(subs, &s)
	}

	sentQueries := 0
	for _, rule := range rules {
		if rule.GraphQuery != "" {
			result := graphClient.Query(ctx, rule.GraphQuery, subs)
			if result.Data != nil {
				for _, row := range result.Data {
					m := row.(map[string]interface{})

					log.Debug().Msg(rule.GraphQuery)

					// Check if "id" is present in the map
					if _, ok := m["id"]; !ok {
						log.Warn().Msgf("Skipping result: 'id' field is missing in the response for recommendation: %s", rule.RecommendationID)
						break
					}

					subscription := models.GetSubscriptionFromResourceID(m["id"].(string))
					subscriptionName, ok := subscriptions[subscription]
					if !ok {
						subscriptionName = ""
					}

					results = append(results, models.AprlResult{
						RecommendationID:    rule.RecommendationID,
						Category:            models.RecommendationCategory(rule.Category),
						Recommendation:      rule.Recommendation,
						ResourceType:        rule.ResourceType,
						LongDescription:     rule.LongDescription,
						PotentialBenefits:   rule.PotentialBenefits,
						Impact:              models.RecommendationImpact(rule.Impact),
						Name:                to.String(m["name"]),
						ResourceID:          to.String(m["id"]),
						SubscriptionID:      subscription,
						SubscriptionName:    subscriptionName,
						ResourceGroup:       models.GetResourceGroupFromResourceID(m["id"].(string)),
						Tags:                to.String(m["tags"]),
						Param1:              to.String(m["param1"]),
						Param2:              to.String(m["param2"]),
						Param3:              to.String(m["param3"]),
						Param4:              to.String(m["param4"]),
						Param5:              to.String(m["param5"]),
						Learn:               rule.LearnMoreLink[0].Url,
						AutomationAvailable: rule.AutomationAvailable,
						Source:              rule.Source,
					})
				}
			}
			sentQueries++
			if sentQueries == 2 {
				// Staggering queries to avoid throttling. Max 10 queries each 5 seconds.
				// https://learn.microsoft.com/en-us/azure/governance/resource-graph/concepts/guidance-for-throttled-requests#staggering-queries
				time.Sleep(1 * time.Second)
			}
		}
	}

	return results, nil
}

func (a AprlScanner) getGraphRules(service string, aprl map[string]map[string]models.AprlRecommendation) map[string]models.AprlRecommendation {
	r := map[string]models.AprlRecommendation{}
	if i, ok := aprl[strings.ToLower(service)]; ok {
		for _, recommendation := range i {
			if a.filters.Azqr.IsRecommendationExcluded(recommendation.RecommendationID) ||
				strings.Contains(recommendation.GraphQuery, "cannot-be-validated-with-arg") ||
				strings.Contains(recommendation.GraphQuery, "under-development") ||
				strings.Contains(recommendation.GraphQuery, "under development") {
				continue
			}

			r[recommendation.RecommendationID] = recommendation
		}
	}
	return r
}
