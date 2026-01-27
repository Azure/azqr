// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package pipeline

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/renderers"
	"github.com/Azure/azqr/internal/scanners"
	"github.com/rs/zerolog/log"
)

// AzqrScanStage executes AZQR service scanners for each subscription.
type AzqrScanStage struct {
	*BaseStage
}

func NewAzqrScanStage() *AzqrScanStage {
	return &AzqrScanStage{
		BaseStage: NewBaseStage("AZQR Service Scan", false),
	}
}

func (s *AzqrScanStage) CanSkip(ctx *ScanContext) bool {
	return !ctx.Params.UseAzqrRecommendations
}

func (s *AzqrScanStage) Execute(ctx *ScanContext) error {
	log.Debug().
		Int("subscriptions", len(ctx.Subscriptions)).
		Msg("Starting AZQR service scan")

	// Get service scanners from filters
	serviceScanners := ctx.Params.Filters.Azqr.Scanners

	// Build subscription resource type map for efficient filtering
	subscriptionResourceTypeMap := s.buildSubscriptionResourceTypeMap(
		ctx.ReportData.ResourceTypeCount,
		ctx.Subscriptions,
	)

	// For each service scanner, get the recommendations list
	// This populates the Recommendations map with all possible recommendations
	log.Debug().
		Int("service_scanners", len(serviceScanners)).
		Int("existing_recommendation_types", len(ctx.ReportData.Recommendations)).
		Msg("Starting AZQR recommendation collection")

	azqrRecommendationCount := 0
	excludedCount := 0
	nonRecommendationCount := 0

	for _, scanner := range serviceScanners {
		recs := scanner.GetRecommendations()
		log.Debug().
			Str("scanner_type", fmt.Sprintf("%T", scanner)).
			Int("recommendations", len(recs)).
			Msg("Processing scanner recommendations")

		for i, r := range recs {
			if ctx.Params.Filters.Azqr.IsRecommendationExcluded(r.RecommendationID) {
				excludedCount++
				continue
			}

			if r.RecommendationType != models.TypeRecommendation {
				nonRecommendationCount++
				continue
			}

			if ctx.ReportData.Recommendations[strings.ToLower(r.ResourceType)] == nil {
				ctx.ReportData.Recommendations[strings.ToLower(r.ResourceType)] = map[string]models.AprlRecommendation{}
			}

			ctx.ReportData.Recommendations[strings.ToLower(r.ResourceType)][i] = r.ToAzureAprlRecommendation()
			azqrRecommendationCount++
		}
	}

	log.Debug().
		Int("azqr_recommendations_added", azqrRecommendationCount).
		Int("excluded", excludedCount).
		Int("non_recommendation_type", nonRecommendationCount).
		Int("total_recommendation_types", len(ctx.ReportData.Recommendations)).
		Msg("AZQR recommendation collection completed")

	// Initialize scanners used for all subscriptions
	peScanner := scanners.PrivateEndpointScanner{}
	pipScanner := scanners.PublicIPScanner{}
	costScanner := scanners.CostScanner{}
	diagnosticsScanner := scanners.DiagnosticSettingsScanner{}

	// Initialize diagnostic settings scanner
	err := diagnosticsScanner.Init(ctx.Ctx, ctx.Cred, ctx.ClientOptions)
	if err != nil {
		log.Error().Err(err).Msg("Failed to initialize diagnostic settings scanner")
		return err
	}
	diagResults := diagnosticsScanner.Scan(ctx.ReportData.Resources)

	// Scan each subscription with AZQR scanners
	for sid, sn := range ctx.Subscriptions {
		config := &models.ScannerConfig{
			Ctx:              ctx.Ctx,
			SubscriptionID:   sid,
			SubscriptionName: sn,
			Cred:             ctx.Cred,
			ClientOptions:    ctx.ClientOptions,
		}

		// Filter service scanners for this subscription only
		subscriptionFilteredScanners := s.buildSubscriptionFilteredScanners(
			serviceScanners,
			subscriptionResourceTypeMap[sid],
			sid,
			ctx.Params.Mask,
		)

		// Skip scanning if no scanners are needed for this subscription
		if len(subscriptionFilteredScanners) == 0 {
			log.Info().Msgf("No scanners needed for subscription %s, skipping AZQR scan",
				renderers.MaskSubscriptionID(sid, ctx.Params.Mask))
			continue
		}

		// Scan private endpoints
		peResults := peScanner.Scan(config)

		// Scan public IPs
		pips := pipScanner.Scan(config)

		// Initialize scan context
		scanContext := models.ScanContext{
			Filters:             ctx.Params.Filters,
			PrivateEndpoints:    peResults,
			DiagnosticsSettings: diagResults,
			PublicIPs:           pips,
		}

		// Worker pool to limit concurrent scanner goroutines
		const numScannerWorkers = 10
		jobs := make(chan models.IAzureScanner, len(subscriptionFilteredScanners))
		results := make(chan []*models.AzqrServiceResult, len(subscriptionFilteredScanners))

		// Start worker pool
		var workerWg sync.WaitGroup
		for w := 0; w < numScannerWorkers; w++ {
			workerWg.Add(1)
			go func() {
				defer workerWg.Done()
				for scanner := range jobs {
					// Initialize scanner with this subscription's config
					err := scanner.Init(config)
					if err != nil {
						log.Error().Err(err).Msg("Failed to initialize scanner")
						continue
					}

					res, err := s.retry(3, 10*time.Millisecond, scanner, &scanContext)
					if err != nil {
						log.Error().Err(err).Msg("Scanner failed after retries")
						continue
					}
					results <- res
				}
			}()
		}

		// Send scanner jobs to workers
		go func() {
			for _, s := range subscriptionFilteredScanners {
				jobs <- s
			}
			close(jobs)
		}()

		// Wait for workers to finish and close results channel
		go func() {
			workerWg.Wait()
			close(results)
		}()

		// Collect results from all scanners
		azqrResultsForSubscription := 0
		for res := range results {
			for _, r := range res {
				// Check if the resource is excluded
				if ctx.Params.Filters.Azqr.IsServiceExcluded(r.ResourceID()) {
					continue
				}
				ctx.ReportData.Azqr = append(ctx.ReportData.Azqr, r)
				azqrResultsForSubscription++
			}
		}

		log.Debug().
			Str("subscription", renderers.MaskSubscriptionID(sid, ctx.Params.Mask)).
			Int("results", azqrResultsForSubscription).
			Msg("AZQR scan completed for subscription")

		// Scan costs
		costs := costScanner.Scan(ctx.Params.Cost, config)
		ctx.ReportData.Cost.From = costs.From
		ctx.ReportData.Cost.To = costs.To
		ctx.ReportData.Cost.Items = append(ctx.ReportData.Cost.Items, costs.Items...)
	}

	log.Debug().
		Int("subscriptions", len(ctx.Subscriptions)).
		Int("results", len(ctx.ReportData.Azqr)).
		Msg("AZQR service scan completed")

	return nil
}

// retry retries the Azure scanner Scan, a number of times with an increasing delay between retries
func (s *AzqrScanStage) retry(attempts int, sleep time.Duration, a models.IAzureScanner, scanContext *models.ScanContext) ([]*models.AzqrServiceResult, error) {
	var err error
	for i := 0; ; i++ {
		res, err := a.Scan(scanContext)
		if err == nil {
			return res, nil
		}

		if models.ShouldSkipError(err) {
			return []*models.AzqrServiceResult{}, nil
		}

		errAsString := err.Error()

		if i >= (attempts - 1) {
			log.Info().Msgf("Retry limit reached. Error: %s", errAsString)
			break
		}

		log.Debug().Msgf("Retrying after error: %s", errAsString)

		time.Sleep(sleep)
		sleep *= 2
	}
	return nil, err
}

// buildSubscriptionResourceTypeMap creates a map of subscription ID -> resource type -> count
func (s *AzqrScanStage) buildSubscriptionResourceTypeMap(resourceTypeCounts []models.ResourceTypeCount, subscriptions map[string]string) map[string]map[string]float64 {
	subscriptionResourceTypeMap := make(map[string]map[string]float64)
	for _, rtc := range resourceTypeCounts {
		// Find subscription ID by name
		for sid, sname := range subscriptions {
			if sname == rtc.Subscription {
				if subscriptionResourceTypeMap[sid] == nil {
					subscriptionResourceTypeMap[sid] = make(map[string]float64)
				}
				subscriptionResourceTypeMap[sid][strings.ToLower(rtc.ResourceType)] = rtc.Count
				break
			}
		}
	}
	return subscriptionResourceTypeMap
}

// buildSubscriptionFilteredScanners filters service scanners based on resource types
func (s *AzqrScanStage) buildSubscriptionFilteredScanners(serviceScanners []models.IAzureScanner, subscriptionResourceTypes map[string]float64, subscriptionID string, mask bool) []models.IAzureScanner {
	if subscriptionResourceTypes == nil {
		subscriptionResourceTypes = make(map[string]float64)
	}

	var filteredScanners []models.IAzureScanner
	for _, scanner := range serviceScanners {
		add := true
		for _, resourceType := range scanner.ResourceTypes() {
			resourceType = strings.ToLower(resourceType)

			// Check if the resource type is in this subscription's resource types
			if count, exists := subscriptionResourceTypes[resourceType]; !exists || count <= 0 {
				log.Debug().Msgf("Skipping scanner for resource type %s in subscription %s as it has no resources",
					resourceType, renderers.MaskSubscriptionID(subscriptionID, mask))
				continue
			} else if add {
				filteredScanners = append(filteredScanners, scanner)
				add = false
				log.Info().Msgf("Scanner for resource type %s will be used in subscription %s",
					resourceType, renderers.MaskSubscriptionID(subscriptionID, mask))
			}
		}
	}
	return filteredScanners
}
