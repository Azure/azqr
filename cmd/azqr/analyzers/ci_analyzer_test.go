package analyzers

import (
	"context"
	"reflect"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerinstance/armcontainerinstance"
	"github.com/Azure/go-autorest/autorest/to"
)

func newContainerInstance(t *testing.T) *armcontainerinstance.ContainerGroup {
	ipType := armcontainerinstance.ContainerGroupIPAddressTypePublic
	sku := armcontainerinstance.ContainerGroupSKUStandard
	return &armcontainerinstance.ContainerGroup{
		ID:       to.StringPtr("id"),
		Name:     to.StringPtr("ci-name"),
		Location: to.StringPtr("westeurope"),
		Type:     to.StringPtr("Microsoft.ContainerInstance/containerGroups"),
		Zones:    []*string{},
		Properties: &armcontainerinstance.ContainerGroupProperties{
			SKU: &sku,
			IPAddress: &armcontainerinstance.IPAddress{
				Type: &ipType,
			},
		},
	}
}

func newContainerInstanceWithAvailabilityZones(t *testing.T) *armcontainerinstance.ContainerGroup {
	svc := newContainerInstance(t)
	svc.Zones = []*string{to.StringPtr("1"), to.StringPtr("2"), to.StringPtr("3")}
	return svc
}

func newContainerInstanceWithPrivateEndpoints(t *testing.T) *armcontainerinstance.ContainerGroup {
	ipType := armcontainerinstance.ContainerGroupIPAddressTypePrivate
	svc := newContainerInstance(t)
	svc.Properties.IPAddress.Type = &ipType
	return svc
}

func newContainerInstanceResult(t *testing.T) AzureServiceResult {
	return AzureServiceResult{
		AzureBaseServiceResult: AzureBaseServiceResult{
			SubscriptionId: "subscriptionId",
			ResourceGroup:  "resourceGroupName",
			ServiceName:    "ci-name",
			Sku:            "Standard",
			Sla:            "99.9%",
			Type:           "Microsoft.ContainerInstance/containerGroups",
			Location:       "westeurope",
			CAFNaming:      true,
		},
		AvailabilityZones:  false,
		PrivateEndpoints:   false,
		DiagnosticSettings: true,
	}
}

func newContainerInstanceAvailabilityZonesResult(t *testing.T) AzureServiceResult {
	svc := newContainerInstanceResult(t)
	svc.AvailabilityZones = true
	return svc
}

func newContainerInstancePrivateEndpointResult(t *testing.T) AzureServiceResult {
	svc := newContainerInstanceResult(t)
	svc.PrivateEndpoints = true
	return svc
}

func TestContainerInstanceAnalyzer_Review(t *testing.T) {
	type args struct {
		resourceGroupName string
	}
	tests := []struct {
		name    string
		c       ContainerInstanceAnalyzer
		args    args
		want    []AzureServiceResult
		wantErr bool
	}{
		{
			name: "Test Review",
			c: ContainerInstanceAnalyzer{
				diagnosticsSettings: DiagnosticsSettings{
					diagnosticsSettingsClient: nil,
					ctx:                       context.TODO(),
					hasDiagnosticsFunc: func(resourceId string) (bool, error) {
						return true, nil
					},
				},
				subscriptionId:  "subscriptionId",
				ctx:             context.TODO(),
				cred:            nil,
				instancesClient: nil,
				listInstancesFunc: func(resourceGroupName string) ([]*armcontainerinstance.ContainerGroup, error) {
					return []*armcontainerinstance.ContainerGroup{
							newContainerInstance(t),
							newContainerInstanceWithAvailabilityZones(t),
							newContainerInstanceWithPrivateEndpoints(t),
						},
						nil
				},
			},
			args: args{
				resourceGroupName: "resourceGroupName",
			},
			want: []AzureServiceResult{
				newContainerInstanceResult(t),
				newContainerInstanceAvailabilityZonesResult(t),
				newContainerInstancePrivateEndpointResult(t),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.c.Review(tt.args.resourceGroupName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ContainerInstanceAnalyzer.Review() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ContainerInstanceAnalyzer.Review() = %v, want %v", got, tt.want)
			}
		})
	}
}
