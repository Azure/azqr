package analyzers

import (
	"context"
	"log"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice"
)

type AKSAnalyzer struct {
	diagnosticsSettings DiagnosticsSettings
	subscriptionId      string
	ctx                 context.Context
	cred                azcore.TokenCredential
}

func NewAKSAnalyzer(subscriptionId string, ctx context.Context, cred azcore.TokenCredential) *AKSAnalyzer {
	diagnosticsSettings, _ := NewDiagnosticsSettings(cred, ctx)
	analyzer := AKSAnalyzer{
		diagnosticsSettings: *diagnosticsSettings,
		subscriptionId:      subscriptionId,
		ctx:                 ctx,
		cred:                cred,
	}
	return &analyzer
}

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

		results = append(results, AzureServiceResult{
			SubscriptionId:     a.subscriptionId,
			ResourceGroup:      resourceGroupName,
			ServiceName:        *c.Name,
			Sku:                string(*c.SKU.Name),
			Sla:                "TODO",
			Type:               *c.Type,
			AvailabilityZones:  false,
			PrivateEndpoints:   false,
			DiagnosticSettings: hasDiagnostics,
			CAFNaming:          strings.HasPrefix(*c.Name, "aks"),
		})
	}
	return results, nil
}

func (a AKSAnalyzer) listClusters(resourceGroupName string) ([]*armcontainerservice.ManagedCluster, error) {

	clustersClient, err := armcontainerservice.NewManagedClustersClient(a.subscriptionId, a.cred, nil)
	if err != nil {
		return nil, err
	}

	pager := clustersClient.NewListByResourceGroupPager(resourceGroupName, nil)

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
