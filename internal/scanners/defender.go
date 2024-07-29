// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package scanners

import (
	"fmt"
	"strings"

	"github.com/Azure/azqr/internal/azqr"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/security/armsecurity"
	"github.com/rs/zerolog/log"
)

// DefenderResult - Defender result
type DefenderResult struct {
	SubscriptionID, SubscriptionName, Name, Tier string
	Deprecated                                   bool
}

// DefenderScanner - Defender scanner
type DefenderScanner struct {
	config       *azqr.ScannerConfig
	client       *armsecurity.PricingsClient
	defenderFunc func() ([]DefenderResult, error)
}

// Init - Initializes the Defender Scanner
func (s *DefenderScanner) Init(config *azqr.ScannerConfig) error {
	s.config = config
	var err error
	s.client, err = armsecurity.NewPricingsClient(config.Cred, config.ClientOptions)
	if err != nil {
		return err
	}
	return nil
}

// ListConfiguration - Lists Microsoft Defender for Cloud pricing configurations in the subscription.
func (s *DefenderScanner) ListConfiguration() ([]DefenderResult, error) {
	azqr.LogSubscriptionScan(s.config.SubscriptionID, "Defender Status")
	if s.defenderFunc == nil {
		resp, err := s.client.List(s.config.Ctx, fmt.Sprintf("subscriptions/%s", s.config.SubscriptionID), nil)
		if err != nil {
			if strings.Contains(err.Error(), "ERROR CODE: Subscription Not Registered") {
				log.Info().Msg("Subscription Not Registered for Defender. Skipping Defender Scan...")
				return []DefenderResult{}, nil
			}

			return nil, err
		}

		results := make([]DefenderResult, 0, len(resp.Value))
		for _, v := range resp.Value {
			deprecated := false
			if v.Properties.Deprecated != nil {
				deprecated = *v.Properties.Deprecated
			}
			result := DefenderResult{
				SubscriptionID:   s.config.SubscriptionID,
				SubscriptionName: s.config.SubscriptionName,
				Name:             *v.Name,
				Tier:             string(*v.Properties.PricingTier),
				Deprecated:       deprecated,
			}

			results = append(results, result)
		}
		return results, nil
	}

	return s.defenderFunc()
}

func (s *DefenderScanner) Scan(scan bool, config *azqr.ScannerConfig) []DefenderResult {
	defenderResults := []DefenderResult{}
	if scan {
		err := s.Init(config)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to initialize Defender Scanner")
		}

		res, err := s.ListConfiguration()
		if err != nil {
			if azqr.ShouldSkipError(err) {
				res = []DefenderResult{}
			} else {
				log.Fatal().Err(err).Msg("Failed to list Defender configuration")
			}
		}
		defenderResults = append(defenderResults, res...)
	}
	return defenderResults
}
