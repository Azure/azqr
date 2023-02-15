package scanners

import (
	"context"
	"reflect"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/servicebus/armservicebus"
	"github.com/Azure/go-autorest/autorest/to"
)

func newServiceBus(t *testing.T) *armservicebus.SBNamespace {
	sku := armservicebus.SKUNameBasic
	return &armservicebus.SBNamespace{
		ID:       to.StringPtr("id"),
		Name:     to.StringPtr("sb-name"),
		Location: to.StringPtr("westeurope"),
		Type:     to.StringPtr("Microsoft.ServiceBus/namespaces"),
		SKU: &armservicebus.SBSKU{
			Name: &sku,
		},
		Properties: &armservicebus.SBNamespaceProperties{
			PrivateEndpointConnections: []*armservicebus.PrivateEndpointConnection{},
		},
	}
}

func newServiceBusWithPrivateEndpoints(t *testing.T) *armservicebus.SBNamespace {
	svc := newServiceBus(t)
	svc.Properties.PrivateEndpointConnections = []*armservicebus.PrivateEndpointConnection{
		{
			ID: to.StringPtr("id"),
		},
	}
	return svc
}

func newServiceBusResult(t *testing.T) AzureServiceResult {
	return AzureServiceResult{
		SubscriptionID:     "subscriptionId",
		ResourceGroup:      "resourceGroupName",
		ServiceName:        "sb-name",
		SKU:                "Basic",
		SLA:                "99.9%",
		Type:               "Microsoft.ServiceBus/namespaces",
		Location:           "westeurope",
		CAFNaming:          true,
		AvailabilityZones:  true,
		PrivateEndpoints:   false,
		DiagnosticSettings: true,
	}
}

func newServiceBusPrivateEndpointResult(t *testing.T) AzureServiceResult {
	svc := newServiceBusResult(t)
	svc.PrivateEndpoints = true
	return svc
}

func TestServiceBusScanner_Scan(t *testing.T) {
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
		c       ServiceBusScanner
		args    args
		want    []IAzureServiceResult
		wantErr bool
	}{
		{
			name: "Test Scan",
			c: ServiceBusScanner{
				config: config,
				diagnosticsSettings: DiagnosticsSettings{
					config:                    config,
					diagnosticsSettingsClient: nil,
					hasDiagnosticsFunc: func(resourceId string) (bool, error) {
						return true, nil
					},
				},
				servicebusClient: nil,
				listServiceBusFunc: func(resourceGroupName string) ([]*armservicebus.SBNamespace, error) {
					return []*armservicebus.SBNamespace{
							newServiceBus(t),
							newServiceBusWithPrivateEndpoints(t),
						},
						nil
				},
			},
			args: args{
				resourceGroupName: "resourceGroupName",
			},
			want: []IAzureServiceResult{
				newServiceBusResult(t),
				newServiceBusPrivateEndpointResult(t),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.c.Scan(tt.args.resourceGroupName, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("ServiceBusScanner.Scan() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ServiceBusScanner.Scan() = %v, want %v", got, tt.want)
			}
		})
	}
}
