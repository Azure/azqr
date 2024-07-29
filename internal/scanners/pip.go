// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package scanners

import (
	"github.com/Azure/azqr/internal/azqr"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v5"
	"github.com/rs/zerolog/log"
)

// PublicIPScanner - Scanner for Public IPs
type PublicIPScanner struct {
	config *azqr.ScannerConfig
	client *armnetwork.PublicIPAddressesClient
}

// Init - Initializes the PublicIPScanner
func (s *PublicIPScanner) Init(config *azqr.ScannerConfig) error {
	s.config = config
	var err error
	s.client, err = armnetwork.NewPublicIPAddressesClient(s.config.SubscriptionID, s.config.Cred, config.ClientOptions)
	if err != nil {
		return err
	}
	return nil
}

// ListPublicIPs - Lists all Public IPs
func (s *PublicIPScanner) ListPublicIPs() (map[string]*armnetwork.PublicIPAddress, error) {
	azqr.LogSubscriptionScan(s.config.SubscriptionID, "Public IPs")

	res := map[string]*armnetwork.PublicIPAddress{}
	opt := armnetwork.PublicIPAddressesClientListAllOptions{}

	pager := s.client.NewListAllPager(&opt)

	for pager.More() {
		resp, err := pager.NextPage(s.config.Ctx)
		if err != nil {
			return nil, err
		}

		for _, v := range resp.Value {
			res[*v.ID] = v
		}
	}

	return res, nil
}

func (s *PublicIPScanner) Scan(config *azqr.ScannerConfig) map[string]*armnetwork.PublicIPAddress {
	err := s.Init(config)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize Diagnostic Settings Scanner")
	}
	pips, err := s.ListPublicIPs()
	if err != nil {
		if azqr.ShouldSkipError(err) {
			pips = map[string]*armnetwork.PublicIPAddress{}
		} else {
			log.Fatal().Err(err).Msg("Failed to list Public IPs")
		}
	}
	return pips
}
