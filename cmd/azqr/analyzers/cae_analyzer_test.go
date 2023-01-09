package analyzers

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
		AzureBaseServiceResult: AzureBaseServiceResult{
			SubscriptionID: "subscriptionId",
			ResourceGroup:  "resourceGroupName",
			ServiceName:    "cae-name",
			SKU:            "None",
			SLA:            "99.95%",
			Type:           "Microsoft.App/managedEnvironments",
			Location:       "westeurope",
			CAFNaming:      true,
		},
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

func TestContainerAppsAnalyzer_Review(t *testing.T) {
	type args struct {
		resourceGroupName string
	}
	tests := []struct {
		name    string
		c       ContainerAppsAnalyzer
		args    args
		want    []IAzureServiceResult
		wantErr bool
	}{
		{
			name: "Test Review",
			c: ContainerAppsAnalyzer{
				diagnosticsSettings: DiagnosticsSettings{
					diagnosticsSettingsClient: nil,
					ctx:                       context.TODO(),
					hasDiagnosticsFunc: func(resourceId string) (bool, error) {
						return true, nil
					},
				},
				subscriptionID: "subscriptionId",
				ctx:            context.TODO(),
				cred:           nil,
				appsClient:     nil,
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
			got, err := tt.c.Review(tt.args.resourceGroupName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ContainerAppsAnalyzer.Review() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ContainerAppsAnalyzer.Review() = %v, want %v", got, tt.want)
			}
		})
	}
}
