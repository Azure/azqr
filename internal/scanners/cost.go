// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
package scanners

import (
	"fmt"
	"time"

	"github.com/Azure/azqr/internal/az"
	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/costmanagement/armcostmanagement"
	"github.com/rs/zerolog/log"
)

// CostScanner - Cost scanner
type CostScanner struct {
	config     *models.ScannerConfig
	httpClient *az.HttpClient
}

func (s *CostScanner) init(config *models.ScannerConfig) error {
	s.config = config
	// Cost Management's query operation has no generated pager, so a raw HttpClient is used
	// instead of a typed armcostmanagement.QueryClient - see az.QueryCostManagementPages.
	// Transport/ThrottlingPolicy are carried over from config.ClientOptions (when supplied,
	// e.g. by tests) without mutating the caller's arm.ClientOptions.
	httpOpts := az.DefaultHttpClientOptions(60 * time.Second)
	if config.ClientOptions != nil {
		if config.ClientOptions.Transport != nil {
			httpOpts.Transport = config.ClientOptions.Transport
		}
		if len(config.ClientOptions.PerRetryPolicies) > 0 {
			httpOpts.ThrottlingPolicy = config.ClientOptions.PerRetryPolicies[0]
		}
	}
	s.httpClient = az.NewHttpClient(config.Cred, httpOpts)
	return nil
}

// QueryCosts - Query Costs.
func (s *CostScanner) QueryCosts() ([]*models.CostResult, error) {
	models.LogSubscriptionScan(s.config.SubscriptionID, "Costs")
	timeframeType := armcostmanagement.TimeframeTypeCustom
	etype := armcostmanagement.ExportTypeActualCost
	fromTime, toTime := costTimeRange(time.Now().UTC())
	sum := armcostmanagement.FunctionTypeSum
	dimension := armcostmanagement.QueryColumnTypeDimension
	qd := armcostmanagement.QueryDefinition{
		Type:      &etype,
		Timeframe: &timeframeType,
		TimePeriod: &armcostmanagement.QueryTimePeriod{
			From: &fromTime,
			To:   &toTime,
		},
		Dataset: &armcostmanagement.QueryDataset{
			// Granularity: &daily,
			Aggregation: map[string]*armcostmanagement.QueryAggregation{
				"TotalCost": {
					Name:     to.Ptr("Cost"),
					Function: &sum,
				},
			},
			Grouping: []*armcostmanagement.QueryGrouping{
				{
					Name: to.Ptr("ServiceName"),
					Type: &dimension,
				},
			},
		},
	}

	scope := fmt.Sprintf("/subscriptions/%s", s.config.SubscriptionID)
	result := []*models.CostResult{}
	err := az.QueryCostManagementPages(s.config.Ctx, s.httpClient, scope, qd, func(properties *armcostmanagement.QueryProperties) error {
		for _, v := range properties.Rows {
			result = append(result, &models.CostResult{
				From:             fromTime,
				To:               toTime,
				SubscriptionID:   s.config.SubscriptionID,
				SubscriptionName: s.config.SubscriptionName,
				ServiceName:      fmt.Sprintf("%v", v[1]),
				Value:            fmt.Sprintf("%v", v[0]),
				Currency:         fmt.Sprintf("%v", v[2]),
			})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *CostScanner) Scan(config *models.ScannerConfig) []*models.CostResult {
	costResult := []*models.CostResult{}
	err := s.init(config)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize Cost Scanner")
	}
	costs, err := s.QueryCosts()
	if err != nil && !models.ShouldSkipError(err) {
		log.Fatal().Err(err).Msg("Failed to query costs")
	}
	costResult = append(costResult, costs...)
	return costResult
}

func costTimeRange(now time.Time) (time.Time, time.Time) {
	start := time.Date(now.Year(), now.Month()-1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC).Add(-time.Nanosecond)
	return start, end
}
