// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package scanners

import (
	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/throttling"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v6"
	"github.com/rs/zerolog/log"
)

// PrivateEndpointScanner - Scanner for Private Endpoints
type PrivateEndpointScanner struct {
	config                 *models.ScannerConfig
	client                 *armnetwork.PrivateEndpointsClient
	hasPrivateEndpointFunc func() (map[string]bool, error)
}

// Init - Initializes the PrivateEndpointScanner
func (s *PrivateEndpointScanner) Init(config *models.ScannerConfig) error {
	s.config = config
	var err error
	s.client, err = armnetwork.NewPrivateEndpointsClient(s.config.SubscriptionID, s.config.Cred, config.ClientOptions)
	if err != nil {
		return err
	}
	return nil
}

// ListResourcesWithPrivateEndpoints - Lists all resources with private endpoints
func (s *PrivateEndpointScanner) ListResourcesWithPrivateEndpoints() (map[string]bool, error) {
	models.LogSubscriptionScan(s.config.SubscriptionID, "Private Endpoints")

	res := map[string]bool{}
	if s.hasPrivateEndpointFunc == nil {
		opt := armnetwork.PrivateEndpointsClientListBySubscriptionOptions{}

		pager := s.client.NewListBySubscriptionPager(&opt)

		for pager.More() {
			// Wait for a token from the burstLimiter channel before making the request
			<-throttling.ARMLimiter
			resp, err := pager.NextPage(s.config.Ctx)
			if err != nil {
				return nil, err
			}

			for _, v := range resp.Value {
				for _, c := range v.Properties.PrivateLinkServiceConnections {
					if len(*c.Properties.PrivateLinkServiceID) > 0 {
						res[*c.Properties.PrivateLinkServiceID] = true
					}
				}
			}
		}

		return res, nil
	}

	return s.hasPrivateEndpointFunc()
}

func (s *PrivateEndpointScanner) Scan(config *models.ScannerConfig) map[string]bool {
	err := s.Init(config)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize Private Endpoint Scanner")
	}
	peResults, err := s.ListResourcesWithPrivateEndpoints()
	if err != nil {
		if models.ShouldSkipError(err) {
			peResults = map[string]bool{}
		} else {
			log.Fatal().Err(err).Msg("Failed to list resources with Private Endpoints")
		}
	}
	return peResults
}
