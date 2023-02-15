package scanners

import (
	"context"
	"reflect"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appcontainers/armappcontainers"
	"github.com/Azure/go-autorest/autorest/to"
)

func newContainerApps(t *testing.T) *armappcontainers.ManagedEnvironment {
	return &armappcontainers.ManagedEnvironment{
		ID:       to.StringPtr("id"),
		Name:     to.StringPtr("cae-name"),
		Location: to.StringPtr("westeurope"),
		Type:     to.StringPtr("Microsoft.App/managedEnvironments"),
		Properties: &armappcontainers.ManagedEnvironmentProperties{
			ZoneRedundant: to.BoolPtr(false),
			VnetConfiguration: &armappcontainers.VnetConfiguration{
				Internal: to.BoolPtr(false),
			},
		},
	}
}

func newContainerAppsWithAvailabilityZones(t *testing.T) *armappcontainers.ManagedEnvironment {
	svc := newContainerApps(t)
	svc.Properties.ZoneRedundant = to.BoolPtr(true)
	return svc
}

func newContainerAppsWithPrivateEndpoints(t *testing.T) *armappcontainers.ManagedEnvironment {
	svc := newContainerApps(t)
	svc.Properties.VnetConfiguration.Internal = to.BoolPtr(true)
	return svc
}

func newContainerAppsResult(t *testing.T) AzureServiceResult {
	return AzureServiceResult{
		SubscriptionID:     "subscriptionId",
		ResourceGroup:      "resourceGroupName",
		ServiceName:        "cae-name",
		SKU:                "None",
		SLA:                "99.95%",
		Type:               "Microsoft.App/managedEnvironments",
		Location:           "westeurope",
		CAFNaming:          true,
		AvailabilityZones:  false,
		PrivateEndpoints:   false,
		DiagnosticSettings: true,
	}
}

func newContainerAppsAvailabilityZonesResult(t *testing.T) AzureServiceResult {
	svc := newContainerAppsResult(t)
	svc.AvailabilityZones = true
	return svc
}

func newContainerAppsPrivateEndpointResult(t *testing.T) AzureServiceResult {
	svc := newContainerAppsResult(t)
	svc.PrivateEndpoints = true
	return svc
}

func TestContainerAppsScanner_Scan(t *testing.T) {
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
		c       ContainerAppsScanner
		args    args
		want    []IAzureServiceResult
		wantErr bool
	}{
		{
			name: "Test Scan",
			c: ContainerAppsScanner{
				config: config,
				diagnosticsSettings: DiagnosticsSettings{
					config:                    config,
					diagnosticsSettingsClient: nil,
					hasDiagnosticsFunc: func(resourceId string) (bool, error) {
						return true, nil
					},
				},
				appsClient: nil,
				listAppsFunc: func(resourceGroupName string) ([]*armappcontainers.ManagedEnvironment, error) {
					return []*armappcontainers.ManagedEnvironment{
							newContainerApps(t),
							newContainerAppsWithAvailabilityZones(t),
							newContainerAppsWithPrivateEndpoints(t),
						},
						nil
				},
			},
			args: args{
				resourceGroupName: "resourceGroupName",
			},
			want: []IAzureServiceResult{
				newContainerAppsResult(t),
				newContainerAppsAvailabilityZonesResult(t),
				newContainerAppsPrivateEndpointResult(t),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.c.Scan(tt.args.resourceGroupName, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("ContainerAppsScanner.Scan() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ContainerAppsScanner.Scan() = %v, want %v", got, tt.want)
			}
		})
	}
}
