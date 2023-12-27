// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package scanners

import (
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
)

// PublicIPScanner - Scanner for Public IPs
type PublicIPScanner struct {
	config *ScannerConfig
	client *armnetwork.PublicIPAddressesClient
}

// Init - Initializes the PublicIPScanner
func (s *PublicIPScanner) Init(config *ScannerConfig) error {
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
	LogSubscriptionScan(s.config.SubscriptionID, "Public IPs")

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
