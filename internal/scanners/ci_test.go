package scanners

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

func newContainerInstanceWithoutIPAdress(t *testing.T) *armcontainerinstance.ContainerGroup {
	svc := newContainerInstance(t)
	svc.Properties.IPAddress = nil
	return svc
}

func newContainerInstanceResult(t *testing.T) AzureServiceResult {
	return AzureServiceResult{
		SubscriptionID:     "subscriptionId",
		ResourceGroup:      "resourceGroupName",
		ServiceName:        "ci-name",
		SKU:                "Standard",
		SLA:                "99.9%",
		Type:               "Microsoft.ContainerInstance/containerGroups",
		Location:           "westeurope",
		CAFNaming:          true,
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

func newContainerInstanceWithoutIPAddressResult(t *testing.T) AzureServiceResult {
	svc := newContainerInstanceResult(t)
	svc.PrivateEndpoints = false
	return svc
}


func TestContainerInstanceScanner_Scan(t *testing.T) {
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
		c       ContainerInstanceScanner
		args    args
		want    []IAzureServiceResult
		wantErr bool
	}{
		{
			name: "Test Scan",
			c: ContainerInstanceScanner{
				config: config,
				diagnosticsSettings: DiagnosticsSettings{
					config:                    config,
					diagnosticsSettingsClient: nil,
					hasDiagnosticsFunc: func(resourceId string) (bool, error) {
						return true, nil
					},
				},
				instancesClient: nil,
				listInstancesFunc: func(resourceGroupName string) ([]*armcontainerinstance.ContainerGroup, error) {
					return []*armcontainerinstance.ContainerGroup{
							newContainerInstance(t),
							newContainerInstanceWithAvailabilityZones(t),
							newContainerInstanceWithPrivateEndpoints(t),
							newContainerInstanceWithoutIPAdress(t),
						},
						nil
				},
			},
			args: args{
				resourceGroupName: "resourceGroupName",
			},
			want: []IAzureServiceResult{
				newContainerInstanceResult(t),
				newContainerInstanceAvailabilityZonesResult(t),
				newContainerInstancePrivateEndpointResult(t),
				newContainerInstanceWithoutIPAddressResult(t),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.c.Scan(tt.args.resourceGroupName, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("ContainerInstanceScanner.Scan() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ContainerInstanceScanner.Scan() = %v, want %v", got, tt.want)
			}
		})
	}
}
