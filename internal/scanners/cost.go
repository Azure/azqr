// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
package scanners

import (
	"fmt"
	"time"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/throttling"
	"github.com/Azure/azqr/internal/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/costmanagement/armcostmanagement"
	"github.com/rs/zerolog/log"
)

// CostScanner - Cost scanner
type CostScanner struct {
	config *models.ScannerConfig
	client *armcostmanagement.QueryClient
}

// Init - Initializes the Cost Scanner
func (s *CostScanner) Init(config *models.ScannerConfig) error {
	s.config = config
	var err error
	s.client, err = armcostmanagement.NewQueryClient(config.Cred, config.ClientOptions)
	if err != nil {
		return err
	}
	return nil
}

// QueryCosts - Query Costs.
func (s *CostScanner) QueryCosts() (*models.CostResult, error) {
	models.LogSubscriptionScan(s.config.SubscriptionID, "Costs")
	timeframeType := armcostmanagement.TimeframeTypeCustom
	etype := armcostmanagement.ExportTypeActualCost
	toTime := time.Now().UTC()
	fromTime := time.Date(toTime.Year(), toTime.Month()-3, 1, 0, 0, 0, 0, time.UTC)
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

	// Wait for a token from the burstLimiter channel before making the request
	_ = throttling.WaitARM(s.config.Ctx); // nolint:errcheck
	resp, err := s.client.Usage(s.config.Ctx, fmt.Sprintf("/subscriptions/%s", s.config.SubscriptionID), qd, nil)
	if err != nil {
		return nil, err
	}

	result := models.CostResult{
		From:  fromTime,
		To:    toTime,
		Items: []*models.CostResultItem{},
	}

	for _, v := range resp.Properties.Rows {
		result.Items = append(result.Items, &models.CostResultItem{
			SubscriptionID:   s.config.SubscriptionID,
			SubscriptionName: s.config.SubscriptionName,
			ServiceName:      fmt.Sprintf("%v", v[1]),
			Value:            fmt.Sprintf("%v", v[0]),
			Currency:         fmt.Sprintf("%v", v[2]),
		})
	}
	return &result, nil
}

func (s *CostScanner) Scan(scan bool, config *models.ScannerConfig) *models.CostResult {
	costResult := &models.CostResult{
		Items: []*models.CostResultItem{},
	}
	if scan {
		err := s.Init(config)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to initialize Cost Scanner")
		}
		costs, err := s.QueryCosts()
		if err != nil && !models.ShouldSkipError(err) {
			log.Fatal().Err(err).Msg("Failed to query costs")
		}
		costResult.From = costs.From
		costResult.To = costs.To
		costResult.Items = append(costResult.Items, costs.Items...)
	}
	return costResult
}
