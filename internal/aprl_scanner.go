// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package internal

import (
	"context"
	"embed"
	"io/fs"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/Azure/azqr/internal/graph"
	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azqr/internal/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

//go:embed aprl/azure-resources/**/**/*.yaml
//go:embed aprl/azure-resources/**/**/kql/*.kql
//go:embed aprl/azure-specialized-workloads/**/*.yaml
//go:embed aprl/azure-specialized-workloads/**/kql/*.kql
var embededFiles embed.FS

type (
	AprlScanner struct{}
)

// GetAprlRecommendations returns a map with all APRL recommendations
func (sc AprlScanner) GetAprlRecommendations() map[string]map[string]scanners.AprlRecommendation {
	r := map[string]map[string]scanners.AprlRecommendation{}

	fsys, err := fs.Sub(embededFiles, "aprl/azure-resources")
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

			var recommendations []scanners.AprlRecommendation
			err = yaml.Unmarshal(content, &recommendations)
			if err != nil {
				return err
			}

			for _, recommendation := range recommendations {
				t := strings.ToLower(recommendation.ResourceType)
				if _, ok := r[t]; !ok {
					r[t] = map[string]scanners.AprlRecommendation{}
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
func (sc AprlScanner) Scan(ctx context.Context, cred azcore.TokenCredential, serviceScanners []scanners.IAzureScanner, filters *scanners.Filters, subscriptions map[string]string) (map[string]map[string]scanners.AprlRecommendation, []scanners.AprlResult) {
	recommendations := map[string]map[string]scanners.AprlRecommendation{}
	results := []scanners.AprlResult{}
	rules := []scanners.AprlRecommendation{}
	graph := graph.NewGraphQuery(cred)

	// get APRL recommendations
	aprl := sc.GetAprlRecommendations()

	for _, s := range serviceScanners {
		for _, t := range s.ResourceTypes() {
			scanners.LogResourceTypeScan(t)
			gr := sc.getGraphRules(t, filters, aprl)
			for _, r := range gr {
				rules = append(rules, r)
			}

			for i, r := range gr {
				if recommendations[strings.ToLower(t)] == nil {
					recommendations[strings.ToLower(t)] = map[string]scanners.AprlRecommendation{}
				}
				recommendations[strings.ToLower(t)][i] = r
			}
		}
	}

	batches := int(math.Ceil(float64(len(rules)) / 12))

	jobs := make(chan []scanners.AprlRecommendation, batches)
	ch := make(chan []scanners.AprlResult, batches)
	var wg sync.WaitGroup

	// Start workers
	numWorkers := 12 // Define the number of workers in the pool
	for w := 0; w < numWorkers; w++ {
		go sc.worker(ctx, graph, subscriptions, jobs, ch, &wg)
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
			if filters.Azqr.IsServiceExcluded(r.ResourceID) {
				continue
			}
			results = append(results, r)
		}
	}

	return recommendations, results
}

func (sc *AprlScanner) worker(ctx context.Context, graph *graph.GraphQuery, subscriptions map[string]string, jobs <-chan []scanners.AprlRecommendation, results chan<- []scanners.AprlResult, wg *sync.WaitGroup) {
	for r := range jobs {
		res, err := sc.graphScan(ctx, graph, r, subscriptions)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to scan")
		}
		results <- res
		wg.Done()
	}
}

func (sc AprlScanner) graphScan(ctx context.Context, graphClient *graph.GraphQuery, rules []scanners.AprlRecommendation, subscriptions map[string]string) ([]scanners.AprlResult, error) {
	results := []scanners.AprlResult{}
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

					subscription := scanners.GetSubscriptionFromResourceID(m["id"].(string))
					subscriptionName, ok := subscriptions[subscription]
					if !ok {
						subscriptionName = ""
					}

					results = append(results, scanners.AprlResult{
						RecommendationID:    rule.RecommendationID,
						Category:            scanners.RecommendationCategory(rule.Category),
						Recommendation:      rule.Recommendation,
						ResourceType:        rule.ResourceType,
						LongDescription:     rule.LongDescription,
						PotentialBenefits:   rule.PotentialBenefits,
						Impact:              scanners.RecommendationImpact(rule.Impact),
						Name:                to.String(m["name"]),
						ResourceID:          to.String(m["id"]),
						SubscriptionID:      subscription,
						SubscriptionName:    subscriptionName,
						ResourceGroup:       scanners.GetResourceGroupFromResourceID(m["id"].(string)),
						Tags:                to.String(m["tags"]),
						Param1:              to.String(m["param1"]),
						Param2:              to.String(m["param2"]),
						Param3:              to.String(m["param3"]),
						Param4:              to.String(m["param4"]),
						Param5:              to.String(m["param5"]),
						Learn:               rule.LearnMoreLink[0].Url,
						AutomationAvailable: rule.AutomationAvailable,
						Source:              "APRL",
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

func (sc AprlScanner) getGraphRules(service string, filters *scanners.Filters, aprl map[string]map[string]scanners.AprlRecommendation) map[string]scanners.AprlRecommendation {
	r := map[string]scanners.AprlRecommendation{}
	if i, ok := aprl[strings.ToLower(service)]; ok {
		for _, recommendation := range i {
			if filters.Azqr.IsRecommendationExcluded(recommendation.RecommendationID) ||
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
