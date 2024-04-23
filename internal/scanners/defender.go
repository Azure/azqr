// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package scanners

import (
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/security/armsecurity"
)

// DefenderResult - Defender result
type DefenderResult struct {
	SubscriptionID, Name, Tier string
	Deprecated                 bool
}

// DefenderScanner - Defender scanner
type DefenderScanner struct {
	config       *ScannerConfig
	client       *armsecurity.PricingsClient
	defenderFunc func() ([]DefenderResult, error)
}

// Init - Initializes the Defender Scanner
func (s *DefenderScanner) Init(config *ScannerConfig) error {
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
	LogSubscriptionScan(s.config.SubscriptionID, "Defender Status")
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
				SubscriptionID: s.config.SubscriptionID,
				Name:           *v.Name,
				Tier:           string(*v.Properties.PricingTier),
				Deprecated:     deprecated,
			}

			results = append(results, result)
		}
		return results, nil
	}

	return s.defenderFunc()
}
