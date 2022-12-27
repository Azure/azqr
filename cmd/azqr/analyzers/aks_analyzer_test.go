package analyzers

import (
	"context"
	"reflect"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice"
	"github.com/Azure/go-autorest/autorest/to"
)

func newAKS(t *testing.T) *armcontainerservice.ManagedCluster {
	sku := armcontainerservice.ManagedClusterSKUNameBasic
	tier := armcontainerservice.ManagedClusterSKUTierFree
	return &armcontainerservice.ManagedCluster{
		ID:       to.StringPtr("id"),
		Name:     to.StringPtr("aks-name"),
		Location: to.StringPtr("westeurope"),
		Type:     to.StringPtr("Microsoft.ContainerService/managedClusters"),
		SKU: &armcontainerservice.ManagedClusterSKU{
			Name: &sku,
			Tier: &tier,
		},
		Properties: &armcontainerservice.ManagedClusterProperties{
			APIServerAccessProfile: &armcontainerservice.ManagedClusterAPIServerAccessProfile{
				EnablePrivateCluster: to.BoolPtr(false),
			},
			AgentPoolProfiles: []*armcontainerservice.ManagedClusterAgentPoolProfile{
				{
					AvailabilityZones: []*string{},
				},
			},
		},
	}
}

func newAKSWithAvailabilityZones(t *testing.T) *armcontainerservice.ManagedCluster {
	svc := newAKS(t)
	svc.Properties.AgentPoolProfiles = []*armcontainerservice.ManagedClusterAgentPoolProfile{
		{
			AvailabilityZones: []*string{to.StringPtr("1"), to.StringPtr("2"), to.StringPtr("3")},
		},
	}
	return svc
}

func newAKSWithPrivateEndpoints(t *testing.T) *armcontainerservice.ManagedCluster {
	svc := newAKS(t)
	svc.Properties.APIServerAccessProfile.EnablePrivateCluster = to.BoolPtr(true)
	return svc
}

func newAKSResult(t *testing.T) AzureServiceResult {
	return AzureServiceResult{
		AzureBaseServiceResult: AzureBaseServiceResult{
			SubscriptionId: "subscriptionId",
			ResourceGroup:  "resourceGroupName",
			ServiceName:    "aks-name",
			Sku:            "Free",
			Sla:            "None",
			Type:           "Microsoft.ContainerService/managedClusters",
			Location:       "westeurope",
			CAFNaming:      true,
		},
		AvailabilityZones:  false,
		PrivateEndpoints:   false,
		DiagnosticSettings: true,
	}
}

func newAKSAvailabilityZonesResult(t *testing.T) AzureServiceResult {
	svc := newAKSResult(t)
	svc.AvailabilityZones = true
	return svc
}

func newAKSPrivateEndpointResult(t *testing.T) AzureServiceResult {
	svc := newAKSResult(t)
	svc.PrivateEndpoints = true
	return svc
}

func TestAKSAnalyzer_Review(t *testing.T) {
	type fields struct {
		diagnosticsSettings DiagnosticsSettings
		subscriptionId      string
		ctx                 context.Context
		cred                azcore.TokenCredential
		clustersClient      *armcontainerservice.ManagedClustersClient
		listClustersFunc    func(resourceGroupName string) ([]*armcontainerservice.ManagedCluster, error)
	}
	type args struct {
		resourceGroupName string
	}
	f := fields{
		diagnosticsSettings: DiagnosticsSettings{
			diagnosticsSettingsClient: nil,
			ctx:                       context.TODO(),
			hasDiagnosticsFunc: func(resourceId string) (bool, error) {
				return true, nil
			},
		},
		subscriptionId: "subscriptionId",
		ctx:            context.TODO(),
		cred:           nil,
		clustersClient: nil,
		listClustersFunc: func(resourceGroupName string) ([]*armcontainerservice.ManagedCluster, error) {
			return []*armcontainerservice.ManagedCluster{
					newAKS(t),
					newAKSWithAvailabilityZones(t),
					newAKSWithPrivateEndpoints(t),
				},
				nil
		},
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []AzureServiceResult
		wantErr bool
	}{
		{
			name:   "Test Review",
			fields: f,
			args: args{
				resourceGroupName: "resourceGroupName",
			},
			want: []AzureServiceResult{
				newAKSResult(t),
				newAKSAvailabilityZonesResult(t),
				newAKSPrivateEndpointResult(t),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := AKSAnalyzer{
				diagnosticsSettings: tt.fields.diagnosticsSettings,
				subscriptionId:      tt.fields.subscriptionId,
				ctx:                 tt.fields.ctx,
				cred:                tt.fields.cred,
				clustersClient:      tt.fields.clustersClient,
				listClustersFunc:    tt.fields.listClustersFunc,
			}
			got, err := a.Review(tt.args.resourceGroupName)
			if (err != nil) != tt.wantErr {
				t.Errorf("AKSAnalyzer.Review() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AKSAnalyzer.Review() = %v, want %v", got, tt.want)
			}
		})
	}
}
