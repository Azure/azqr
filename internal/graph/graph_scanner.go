// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package graph

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"math"
	"strings"
	"sync"

	"github.com/Azure/azqr/internal/models"
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
//go:embed azqr/azure-resources/**/*.yaml
//go:embed azqr/azure-resources/**/kql/*.kql
//go:embed azqr/azure-resources/**/**/*.yaml
//go:embed azqr/azure-resources/**/**/kql/*.kql
var embededFiles embed.FS

type (
	GraphScanner struct {
		scanType        []ScanType
		serviceScanners []models.IAzureScanner
		filters         *models.Filters
		subscriptions   map[string]string
		externalQueries map[string]map[string]models.GraphRecommendation // External YAML plugin queries by resource type
	}

	ScanType string
)

const (
	AprlScanType   ScanType = "aprl/azure-resources"
	OrphanScanType ScanType = "azure-orphan-resources"
	AzqrScanType   ScanType = "azqr/azure-resources"
	bucketCapacity          = 10 // matches graphLimiter burst in internal/throttling/policy.go
)

var (
	embeddedRecsOnce  sync.Once
	embeddedRecsCache map[string]map[string]models.GraphRecommendation
)

// GetRecommendations returns all embedded Graph recommendations grouped by resource type.
// Results are computed once and cached for the lifetime of the process — the embedded
// filesystem is immutable so repeated FS walks and YAML parses are pure waste.
func (a *GraphScanner) GetRecommendations() map[string]map[string]models.GraphRecommendation {
	embeddedRecsOnce.Do(func() {
		result := map[string]map[string]models.GraphRecommendation{}
		for _, scanType := range a.scanType {
			var source string
			switch scanType {
			case OrphanScanType:
				source = "AOR"
			case AzqrScanType:
				source = "AZQR"
			default:
				source = "APRL"
			}

			typeRecs := a.getRecommendations(string(scanType))
			for resourceType, recs := range typeRecs {
				for _, rec := range recs {
					if result[resourceType] == nil {
						result[resourceType] = map[string]models.GraphRecommendation{}
					}
					rec.Source = source
					result[resourceType][rec.RecommendationID] = rec
				}
			}
		}
		embeddedRecsCache = result
	})
	return embeddedRecsCache
}

// NewScanner creates a new Graph scanner.
func NewScanner(serviceScanners []models.IAzureScanner, filters *models.Filters, subscriptions map[string]string) GraphScanner {
	return GraphScanner{
		scanType: []ScanType{
			AprlScanType,
			OrphanScanType,
			AzqrScanType,
		},
		serviceScanners: serviceScanners,
		filters:         filters,
		subscriptions:   subscriptions,
		externalQueries: make(map[string]map[string]models.GraphRecommendation),
	}
}

// RegisterExternalQuery adds an external YAML plugin query to the scanner
func (a *GraphScanner) RegisterExternalQuery(resourceType string, recommendation models.GraphRecommendation) {
	resourceType = strings.ToLower(resourceType)
	if a.externalQueries[resourceType] == nil {
		a.externalQueries[resourceType] = make(map[string]models.GraphRecommendation)
	}
	a.externalQueries[resourceType][recommendation.RecommendationID] = recommendation
}

func (a *GraphScanner) getRecommendations(path string) map[string]map[string]models.GraphRecommendation {
	r := map[string]map[string]models.GraphRecommendation{}

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

			var recommendations []models.GraphRecommendation
			err = yaml.Unmarshal(content, &recommendations)
			if err != nil {
				return err
			}

			for _, recommendation := range recommendations {
				t := strings.ToLower(recommendation.ResourceType)
				if _, ok := r[t]; !ok {
					r[t] = map[string]models.GraphRecommendation{}
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

func (a *GraphScanner) ListRecommendations() (map[string]map[string]*models.GraphRecommendation, []*models.GraphRecommendation) {
	recommendations := map[string]map[string]*models.GraphRecommendation{}
	rules := []*models.GraphRecommendation{}

	rec := a.GetRecommendations()

	for _, s := range a.serviceScanners {
		for _, t := range s.ResourceTypes() {
			gr := a.getGraphRules(t, rec)
			lowerT := strings.ToLower(t)
			for id, r := range gr {
				rule := r
				rules = append(rules, &rule)

				if recommendations[lowerT] == nil {
					recommendations[lowerT] = map[string]*models.GraphRecommendation{}
				}
				recommendations[lowerT][id] = &rule
			}
		}
	}
	return recommendations, rules
}

// Scan scans Azure resources using Graph queries
func (a *GraphScanner) Scan(ctx context.Context, cred azcore.TokenCredential) []*models.GraphResult {
	results := []*models.GraphResult{}
	graph := NewGraphQuery(cred)

	_, rules := a.ListRecommendations()

	batchSize := bucketCapacity
	batches := int(math.Ceil(float64(len(rules)) / float64(batchSize)))

	log.Debug().Msgf("Using %d rules to scan in %d batches", len(rules), batches)

	// Buffer the jobs and results channels to the number of rules to avoid deadlocks.
	jobs := make(chan *models.GraphRecommendation, len(rules))
	ch := make(chan []*models.GraphResult, len(rules))

	var wg sync.WaitGroup

	// Worker count matches graphLimiter burst capacity so no goroutine ever blocks
	// waiting for a token while another worker is idle.
	numWorkers := bucketCapacity
	for w := 0; w < numWorkers; w++ {
		go a.worker(ctx, graph, a.subscriptions, jobs, ch, &wg)
	}

	wg.Add(len(rules))

	for i := 0; i < len(rules); i += batchSize {
		j := i + batchSize
		if j > len(rules) {
			j = len(rules)
		}

		for _, r := range rules[i:j] {
			jobs <- r
		}
	}

	// Wait for all workers to finish
	close(jobs)
	wg.Wait()

	// Receive results from workers
	for i := 0; i < len(rules); i++ {
		res := <-ch
		for _, r := range res {
			if a.filters.Azqr.IsServiceExcluded(r.ResourceID) {
				continue
			}
			results = append(results, r)
		}
	}

	return results
}

func (a *GraphScanner) worker(ctx context.Context, graph *GraphQueryClient, subscriptions map[string]string, jobs <-chan *models.GraphRecommendation, results chan<- []*models.GraphResult, wg *sync.WaitGroup) {
	// worker processes batches of Graph recommendations from the jobs channel
	for r := range jobs {
		models.LogGraphRecommendationScan(r.ResourceType, r.RecommendationID)
		res, err := a.graphScan(ctx, graph, r, subscriptions)
		if err != nil {
			if shouldSkipUnsupportedGraphLogicalTableError(err) {
				log.Warn().
					Err(err).
					Str("recommendationId", r.RecommendationID).
					Str("resourceType", r.ResourceType).
					Msg("Skipping recommendation due to unsupported resource graph logical table")
				results <- []*models.GraphResult{}
				wg.Done()
				continue
			}
			log.Fatal().Err(err).Msg("Failed to scan")
		}
		results <- res
		wg.Done()
	}
}

func (a *GraphScanner) graphScan(ctx context.Context, graphClient *GraphQueryClient, rule *models.GraphRecommendation, subscriptions map[string]string) ([]*models.GraphResult, error) {
	results := []*models.GraphResult{}
	if rule.GraphQuery != "" {
		log.Debug().Msg(rule.GraphQuery)
		result, err := graphClient.Query(ctx, rule.GraphQuery, subscriptions)
		if err != nil {
			return nil, fmt.Errorf("recommendation %s query failed: %w", rule.RecommendationID, err)
		}

		if result.Data != nil {
			// graphScanRow matches the fields returned by APRL/azqr KQL queries.
			// param1-5 and tags use RawMessage because KQL may project them as objects or primitives.
			type graphScanRow struct {
				ID     string          `json:"id"`
				Name   string          `json:"name"`
				Tags   json.RawMessage `json:"tags"`
				Param1 json.RawMessage `json:"param1"`
				Param2 json.RawMessage `json:"param2"`
				Param3 json.RawMessage `json:"param3"`
				Param4 json.RawMessage `json:"param4"`
				Param5 json.RawMessage `json:"param5"`
			}

			for _, r := range UnmarshalRows[graphScanRow](result.Data, rule.RecommendationID) {
				if r.ID == "" {
					log.Warn().Msgf("Skipping result: 'id' field is missing in the response for recommendation: %s", rule.RecommendationID)
					break
				}

				subscription := models.GetSubscriptionFromResourceID(r.ID)
				subscriptionName, ok := subscriptions[subscription]
				if !ok {
					subscriptionName = ""
				}

				resourceType := models.GetResourceTypeFromResourceID(r.ID)
				if resourceType == "" {
					resourceType = rule.ResourceType
				}

				results = append(results, &models.GraphResult{
					RecommendationID:    rule.RecommendationID,
					Category:            models.RecommendationCategory(rule.Category),
					Recommendation:      rule.Recommendation,
					ResourceType:        resourceType,
					LongDescription:     rule.LongDescription,
					PotentialBenefits:   rule.PotentialBenefits,
					Impact:              models.RecommendationImpact(rule.Impact),
					Name:                r.Name,
					ResourceID:          r.ID,
					SubscriptionID:      subscription,
					SubscriptionName:    subscriptionName,
					ResourceGroup:       models.GetResourceGroupFromResourceID(r.ID),
					Tags:                rawMessageToString(r.Tags),
					Param1:              rawMessageToString(r.Param1),
					Param2:              rawMessageToString(r.Param2),
					Param3:              rawMessageToString(r.Param3),
					Param4:              rawMessageToString(r.Param4),
					Param5:              rawMessageToString(r.Param5),
					Learn:               rule.LearnMoreLink[0].Url,
					AutomationAvailable: rule.AutomationAvailable,
					Source:              rule.Source,
				})
			}
		}
	}

	return results, nil
}

// rawMessageToString converts a RawMessage field to a plain string.
// Quoted JSON strings are unquoted; objects/arrays keep their JSON form; null → "".
func rawMessageToString(b json.RawMessage) string {
	if len(b) == 0 || bytes.Equal(b, []byte("null")) {
		return ""
	}
	// If it's a JSON string, unwrap the quotes.
	if len(b) >= 2 && b[0] == '"' && b[len(b)-1] == '"' {
		var s string
		if json.Unmarshal(b, &s) == nil {
			return s
		}
	}
	// Otherwise (object, array, number, bool) return raw JSON text.
	return string(b)
}

func (a *GraphScanner) getGraphRules(service string, rec map[string]map[string]models.GraphRecommendation) map[string]models.GraphRecommendation {
	r := map[string]models.GraphRecommendation{}

	// Add embedded recommendations
	if i, ok := rec[strings.ToLower(service)]; ok {
		for _, recommendation := range i {
			if a.filters.Azqr.IsRecommendationExcluded(recommendation.RecommendationID) ||
				strings.Contains(recommendation.GraphQuery, "cannot-be-validated-with-arg") ||
				strings.Contains(recommendation.GraphQuery, "under-development") ||
				strings.Contains(recommendation.GraphQuery, "under development") ||
				strings.EqualFold(recommendation.MetadataState, "disabled") {
				continue
			}

			r[recommendation.RecommendationID] = recommendation
		}
	}

	// Add external YAML plugin queries
	if external, ok := a.externalQueries[strings.ToLower(service)]; ok {
		for id, recommendation := range external {
			if a.filters.Azqr.IsRecommendationExcluded(recommendation.RecommendationID) {
				continue
			}
			r[id] = recommendation
		}
	}

	return r
}

func shouldSkipUnsupportedGraphLogicalTableError(err error) bool {
	if err == nil {
		return false
	}

	var respErr *azcore.ResponseError
	if !errors.As(err, &respErr) {
		return isUnsupportedLogicalTableErrorMessage(err.Error())
	}

	if strings.EqualFold(respErr.ErrorCode, "DisallowedLogicalTableName") {
		return true
	}

	errMsg := respErr.Error()
	if isUnsupportedLogicalTableErrorMessage(errMsg) {
		return true
	}

	if respErr.RawResponse == nil || respErr.RawResponse.Body == nil {
		return false
	}

	body, readErr := io.ReadAll(respErr.RawResponse.Body)
	if readErr != nil {
		return false
	}

	// Restore body so other consumers can still inspect it.
	respErr.RawResponse.Body = io.NopCloser(bytes.NewReader(body))

	var payload struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
			Details []struct {
				Code    string `json:"code"`
				Message string `json:"message"`
			} `json:"details"`
		} `json:"error"`
	}

	if unmarshalErr := json.Unmarshal(body, &payload); unmarshalErr != nil {
		return false
	}

	if strings.EqualFold(payload.Error.Code, "DisallowedLogicalTableName") {
		return true
	}

	for _, detail := range payload.Error.Details {
		if strings.EqualFold(detail.Code, "DisallowedLogicalTableName") {
			return true
		}

		if isUnsupportedLogicalTableErrorMessage(detail.Message) {
			return true
		}
	}

	return false
}

func isUnsupportedLogicalTableErrorMessage(message string) bool {
	msg := strings.ToLower(message)

	if strings.Contains(msg, "disallowedlogicaltablename") {
		return true
	}

	return strings.Contains(msg, "invalid, unsupported or disallowed") &&
		strings.Contains(msg, "logical table")
}
