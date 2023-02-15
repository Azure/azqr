package scanners

import (
	"context"
	"reflect"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerregistry/armcontainerregistry"
	"github.com/Azure/go-autorest/autorest/to"
)

func newContainerRegistry(t *testing.T) *armcontainerregistry.Registry {
	sku := armcontainerregistry.SKUNameBasic
	zoneRedundancy := armcontainerregistry.ZoneRedundancyDisabled
	return &armcontainerregistry.Registry{
		ID:       to.StringPtr("id"),
		Name:     to.StringPtr("cr-name"),
		Location: to.StringPtr("westeurope"),
		Type:     to.StringPtr("Microsoft.ContainerRegistry/registries"),
		SKU: &armcontainerregistry.SKU{
			Name: &sku,
		},
		Properties: &armcontainerregistry.RegistryProperties{
			ZoneRedundancy:             &zoneRedundancy,
			PrivateEndpointConnections: []*armcontainerregistry.PrivateEndpointConnection{},
		},
	}
}

func newContainerRegistryWithAvailabilityZones(t *testing.T) *armcontainerregistry.Registry {
	zoneRedundancy := armcontainerregistry.ZoneRedundancyEnabled
	svc := newContainerRegistry(t)
	svc.Properties.ZoneRedundancy = &zoneRedundancy
	return svc
}

func newContainerRegistryWithPrivateEndpoints(t *testing.T) *armcontainerregistry.Registry {
	svc := newContainerRegistry(t)
	svc.Properties.PrivateEndpointConnections = []*armcontainerregistry.PrivateEndpointConnection{
		{
			ID: to.StringPtr("id"),
		},
	}
	return svc
}

func newContainerRegistryResult(t *testing.T) AzureServiceResult {
	return AzureServiceResult{
		SubscriptionID:     "subscriptionId",
		ResourceGroup:      "resourceGroupName",
		ServiceName:        "cr-name",
		SKU:                "Basic",
		SLA:                "99.95%",
		Type:               "Microsoft.ContainerRegistry/registries",
		Location:           "westeurope",
		CAFNaming:          true,
		AvailabilityZones:  false,
		PrivateEndpoints:   false,
		DiagnosticSettings: true,
	}
}

func newContainerRegistryAvailabilityZonesResult(t *testing.T) AzureServiceResult {
	svc := newContainerRegistryResult(t)
	svc.AvailabilityZones = true
	return svc
}

func newContainerRegistryPrivateEndpointResult(t *testing.T) AzureServiceResult {
	svc := newContainerRegistryResult(t)
	svc.PrivateEndpoints = true
	return svc
}

func TestContainerRegistryScanner_Scan(t *testing.T) {
	type args struct {
		resourceGroupName string
	}
	config := &ScannerConfig{
		SubscriptionID: "subscriptionId",
		Cred:           nil,
		Ctx:            context.TODO(),
	}
	tests := []struct {
		name    string
		c       ContainerRegistryScanner
		args    args
		want    []IAzureServiceResult
		wantErr bool
	}{
		{
			name: "Test Scan",
			c: ContainerRegistryScanner{
				config: config,
				diagnosticsSettings: DiagnosticsSettings{
					config:                    config,
					diagnosticsSettingsClient: nil,
					hasDiagnosticsFunc: func(resourceId string) (bool, error) {
						return true, nil
					},
				},
				registriesClient: nil,
				listRegistriesFunc: func(resourceGroupName string) ([]*armcontainerregistry.Registry, error) {
					return []*armcontainerregistry.Registry{
							newContainerRegistry(t),
							newContainerRegistryWithAvailabilityZones(t),
							newContainerRegistryWithPrivateEndpoints(t),
						},
						nil
				},
			},
			args: args{
				resourceGroupName: "resourceGroupName",
			},
			want: []IAzureServiceResult{
				newContainerRegistryResult(t),
				newContainerRegistryAvailabilityZonesResult(t),
				newContainerRegistryPrivateEndpointResult(t),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.c.Scan(tt.args.resourceGroupName, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("ContainerRegistryScanner.Scan() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ContainerRegistryScanner.Scan() = %v, want %v", got, tt.want)
			}
		})
	}
}
