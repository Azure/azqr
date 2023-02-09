package scanners

import (
	"log"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/cdn/armcdn"
)

// FrontDoorScanner - Scanner for Front Door
type FrontDoorScanner struct {
	config              *ScannerConfig
	diagnosticsSettings DiagnosticsSettings
	client              *armcdn.ProfilesClient
	listFunc            func(resourceGroupName string) ([]*armcdn.Profile, error)
}

// Init - Initializes the FrontDoor Scanner
func (a *FrontDoorScanner) Init(config *ScannerConfig) error {
	a.config = config
	var err error
	a.client, err = armcdn.NewProfilesClient(config.SubscriptionID, a.config.Cred, nil)
	if err != nil {
		return err
	}
	a.diagnosticsSettings = DiagnosticsSettings{}
	err = a.diagnosticsSettings.Init(config)
	if err != nil {
		return err
	}
	return nil
}

// Scan - Scans all Front Doors in a Resource Group
func (a *FrontDoorScanner) Scan(resourceGroupName string) ([]IAzureServiceResult, error) {
	log.Printf("Analyzing Front Doors in Resource Group %s", resourceGroupName)

	gateways, err := a.list(resourceGroupName)
	if err != nil {
		return nil, err
	}
	results := []IAzureServiceResult{}
	for _, g := range gateways {
		hasDiagnostics, err := a.diagnosticsSettings.HasDiagnostics(*g.ID)
		if err != nil {
			return nil, err
		}

		results = append(results, AzureServiceResult{
			SubscriptionID:     a.config.SubscriptionID,
			ResourceGroup:      resourceGroupName,
			ServiceName:        *g.Name,
			SKU:                string(*g.SKU.Name),
			SLA:                "99.99%",
			Type:               *g.Type,
			Location:           *g.Location,
			CAFNaming:          strings.HasPrefix(*g.Name, "afd"),
			AvailabilityZones:  true,
			PrivateEndpoints:   false,
			DiagnosticSettings: hasDiagnostics,
		})
	}
	return results, nil
}

func (a *FrontDoorScanner) list(resourceGroupName string) ([]*armcdn.Profile, error) {
	if a.listFunc == nil {
		pager := a.client.NewListByResourceGroupPager(resourceGroupName, nil)

		services := make([]*armcdn.Profile, 0)
		for pager.More() {
			resp, err := pager.NextPage(a.config.Ctx)
			if err != nil {
				return nil, err
			}
			services = append(services, resp.Value...)
		}
		return services, nil
	}

	return a.listFunc(resourceGroupName)
}
