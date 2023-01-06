package analyzers

import (
	"context"
	"log"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice"
)

// AKSAnalyzer - Analyzer for AKS Clusters
type AKSAnalyzer struct {
	diagnosticsSettings DiagnosticsSettings
	subscriptionID      string
	ctx                 context.Context
	cred                azcore.TokenCredential
	clustersClient      *armcontainerservice.ManagedClustersClient
	listClustersFunc    func(resourceGroupName string) ([]*armcontainerservice.ManagedCluster, error)
}

// NewAKSAnalyzer -Creates a new AKSAnalyzer
func NewAKSAnalyzer(ctx context.Context, subscriptionID string, cred azcore.TokenCredential) *AKSAnalyzer {
	diagnosticsSettings, _ := NewDiagnosticsSettings(ctx, cred)
	clustersClient, err := armcontainerservice.NewManagedClustersClient(subscriptionID, cred, nil)
	if err != nil {
		log.Fatal(err)
	}
	analyzer := AKSAnalyzer{
		diagnosticsSettings: *diagnosticsSettings,
		subscriptionID:      subscriptionID,
		ctx:                 ctx,
		cred:                cred,
		clustersClient:      clustersClient,
	}
	return &analyzer
}

// Review - Analyzes all AKS Clusters in a Resource Group
func (a AKSAnalyzer) Review(resourceGroupName string) ([]AzureServiceResult, error) {
	log.Printf("Analyzing AKS Clusters in Resource Group %s", resourceGroupName)

	clusters, err := a.listClusters(resourceGroupName)
	if err != nil {
		return nil, err
	}
	results := []AzureServiceResult{}
	for _, c := range clusters {
		hasDiagnostics, err := a.diagnosticsSettings.HasDiagnostics(*c.ID)
		if err != nil {
			return nil, err
		}

		zones := true
		for _, profile := range c.Properties.AgentPoolProfiles {
			if profile.AvailabilityZones == nil || (profile.AvailabilityZones != nil && len(profile.AvailabilityZones) <= 1) {
				zones = false
			}
		}

		sku := string(*c.SKU.Tier)
		sla := "None"
		if sku == "Paid" {
			sla = "99.9%"
			if zones {
				sla = "99.95%"
			}
		}

		privateEndpoints := false
		if c.Properties.APIServerAccessProfile != nil && *c.Properties.APIServerAccessProfile.EnablePrivateCluster {
			privateEndpoints = true
		}

		results = append(results, AzureServiceResult{
			AzureBaseServiceResult: AzureBaseServiceResult{
				SubscriptionID: a.subscriptionID,
				ResourceGroup:  resourceGroupName,
				ServiceName:    *c.Name,
				SKU:            sku,
				SLA:            sla,
				Type:           *c.Type,
				Location:       parseLocation(c.Location),
				CAFNaming:      strings.HasPrefix(*c.Name, "aks")},
			AvailabilityZones:  zones,
			PrivateEndpoints:   privateEndpoints,
			DiagnosticSettings: hasDiagnostics,
		})
	}
	return results, nil
}

func (a AKSAnalyzer) listClusters(resourceGroupName string) ([]*armcontainerservice.ManagedCluster, error) {
	if a.listClustersFunc == nil {
		pager := a.clustersClient.NewListByResourceGroupPager(resourceGroupName, nil)

		clusters := make([]*armcontainerservice.ManagedCluster, 0)
		for pager.More() {
			resp, err := pager.NextPage(a.ctx)
			if err != nil {
				return nil, err
			}
			clusters = append(clusters, resp.Value...)
		}
		return clusters, nil
	}

	return a.listClustersFunc(resourceGroupName)
}
