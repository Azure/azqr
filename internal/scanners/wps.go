package scanners

import (
	"log"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/webpubsub/armwebpubsub"
)

// WebPubSubScanner - Scanner for WebPubSub
type WebPubSubScanner struct {
	config              *ScannerConfig
	diagnosticsSettings DiagnosticsSettings
	client              *armwebpubsub.Client
	listWebPubSubFunc   func(resourceGroupName string) ([]*armwebpubsub.ResourceInfo, error)
}

// Init - Initializes the WebPubSubScanner
func (c *WebPubSubScanner) Init(config *ScannerConfig) error {
	c.config = config
	var err error
	c.client, err = armwebpubsub.NewClient(config.SubscriptionID, config.Cred, nil)
	if err != nil {
		return err
	}
	c.diagnosticsSettings = DiagnosticsSettings{}
	err = c.diagnosticsSettings.Init(config)
	if err != nil {
		return err
	}
	return nil
}

// Scan - Scans all WebPubSub in a Resource Group
func (c *WebPubSubScanner) Scan(resourceGroupName string) ([]IAzureServiceResult, error) {
	log.Printf("Analyzing WebPubSub in Resource Group %s", resourceGroupName)

	WebPubSub, err := c.listWebPubSub(resourceGroupName)
	if err != nil {
		return nil, err
	}
	results := []IAzureServiceResult{}
	for _, WebPubSub := range WebPubSub {
		hasDiagnostics, err := c.diagnosticsSettings.HasDiagnostics(*WebPubSub.ID)
		if err != nil {
			return nil, err
		}

		sku := string(*WebPubSub.SKU.Name)
		sla := "99.9%"
		if strings.Contains(sku, "Free") {
			sla = "None"
		}
		zones := false
		if strings.Contains(sku, "Premium") {
			zones = true
		}

		results = append(results, AzureServiceResult{
			SubscriptionID:     c.config.SubscriptionID,
			ResourceGroup:      resourceGroupName,
			ServiceName:        *WebPubSub.Name,
			SKU:                sku,
			SLA:                sla,
			Type:               *WebPubSub.Type,
			Location:           *WebPubSub.Location,
			CAFNaming:          strings.HasPrefix(*WebPubSub.Name, "wps"),
			AvailabilityZones:  zones,
			PrivateEndpoints:   len(WebPubSub.Properties.PrivateEndpointConnections) > 0,
			DiagnosticSettings: hasDiagnostics,
		})
	}
	return results, nil
}

func (c *WebPubSubScanner) listWebPubSub(resourceGroupName string) ([]*armwebpubsub.ResourceInfo, error) {
	if c.listWebPubSubFunc == nil {
		pager := c.client.NewListByResourceGroupPager(resourceGroupName, nil)

		WebPubSubs := make([]*armwebpubsub.ResourceInfo, 0)
		for pager.More() {
			resp, err := pager.NextPage(c.config.Ctx)
			if err != nil {
				return nil, err
			}
			WebPubSubs = append(WebPubSubs, resp.Value...)
		}
		return WebPubSubs, nil
	}

	return c.listWebPubSubFunc(resourceGroupName)
}
