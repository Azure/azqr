// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
package scanners

import (
	"fmt"
	"time"

	"github.com/Azure/azqr/internal/ref"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/costmanagement/armcostmanagement"
)

// CostResult - Cost result
type CostResult struct {
	From, To time.Time
	Items    []*CostResultItem
}

// CostResultItem - Cost result ite,
type CostResultItem struct {
	SubscriptionID, ServiceName, Value, Currency string
}

// CostScanner - Cost scanner
type CostScanner struct {
	config *ScannerConfig
	client *armcostmanagement.QueryClient
}

// GetProperties - Returns the properties of the CostResult
func (d CostResult) GetProperties() []string {
	return []string{
		"SubscriptionID",
		"ServiceName",
		"Value",
		"Currency",
	}
}

// ToMap - Returns the properties of the CostResult as a map
func (r CostResultItem) ToMap(mask bool) map[string]string {
	return map[string]string{
		"SubscriptionID": MaskSubscriptionID(r.SubscriptionID, mask),
		"ServiceName":    r.ServiceName,
		"Value":          r.Value,
		"Currency":       r.Currency,
	}
}

// Init - Initializes the Cost Scanner
func (s *CostScanner) Init(config *ScannerConfig) error {
	s.config = config
	var err error
	s.client, err = armcostmanagement.NewQueryClient(config.Cred, config.ClientOptions)
	if err != nil {
		return err
	}
	return nil
}

// QueryCosts - Query Costs.
func (s *CostScanner) QueryCosts() (*CostResult, error) {
	LogSubscriptionScan(s.config.SubscriptionID, "Costs")
	timeframeType := armcostmanagement.TimeframeTypeCustom
	etype := armcostmanagement.ExportTypeActualCost
	toTime := time.Now().UTC()
	fromTime := time.Date(toTime.Year(), toTime.Month(), 1, 0, 0, 0, 0, time.UTC)
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
					Name:     ref.Of("Cost"),
					Function: &sum,
				},
			},
			Grouping: []*armcostmanagement.QueryGrouping{
				{
					Name: ref.Of("ServiceName"),
					Type: &dimension,
				},
			},
		},
	}

	resp, err := s.client.Usage(s.config.Ctx, fmt.Sprintf("/subscriptions/%s", s.config.SubscriptionID), qd, nil)
	if err != nil {
		return nil, err
	}

	result := CostResult{
		From:  fromTime,
		To:    toTime,
		Items: []*CostResultItem{},
	}

	for _, v := range resp.Properties.Rows {
		result.Items = append(result.Items, &CostResultItem{
			SubscriptionID: s.config.SubscriptionID,
			ServiceName:    fmt.Sprintf("%v", v[1]),
			Value:          fmt.Sprintf("%v", v[0]),
			Currency:       fmt.Sprintf("%v", v[2]),
		})
	}
	return &result, nil
}
