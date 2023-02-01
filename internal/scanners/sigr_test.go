package scanners

import (
	"context"
	"reflect"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/signalr/armsignalr"
	"github.com/Azure/go-autorest/autorest/to"
)

func newSignalR(t *testing.T) *armsignalr.ResourceInfo {
	return &armsignalr.ResourceInfo{
		ID:       to.StringPtr("id"),
		Name:     to.StringPtr("sigr-name"),
		Location: to.StringPtr("westeurope"),
		Type:     to.StringPtr("Microsoft.SignalRService/SignalR"),
		SKU: &armsignalr.ResourceSKU{
			Name: to.StringPtr("StandardS1"),
		},
		Properties: &armsignalr.Properties{
			PrivateEndpointConnections: []*armsignalr.PrivateEndpointConnection{},
		},
	}
}

func newSignalRWithAvailabilityZones(t *testing.T) *armsignalr.ResourceInfo {
	svc := newSignalR(t)
	svc.SKU = &armsignalr.ResourceSKU{
		Name: to.StringPtr("Premium"),
	}
	return svc
}

func newSignalRWithPrivateEndpoints(t *testing.T) *armsignalr.ResourceInfo {
	svc := newSignalR(t)
	svc.Properties.PrivateEndpointConnections = []*armsignalr.PrivateEndpointConnection{
		{
			ID: to.StringPtr("id"),
		},
	}
	return svc
}

func newSignalRResult(t *testing.T) AzureServiceResult {
	return AzureServiceResult{
		SubscriptionID:     "subscriptionId",
		ResourceGroup:      "resourceGroupName",
		ServiceName:        "sigr-name",
		SKU:                "StandardS1",
		SLA:                "99.9%",
		Type:               "Microsoft.SignalRService/SignalR",
		Location:           "westeurope",
		CAFNaming:          true,
		AvailabilityZones:  false,
		PrivateEndpoints:   false,
		DiagnosticSettings: true,
	}
}

func newSignalRAvailabilityZonesResult(t *testing.T) AzureServiceResult {
	svc := newSignalRResult(t)
	svc.AvailabilityZones = true
	svc.SKU = "Premium"
	return svc
}

func newSignalRPrivateEndpointResult(t *testing.T) AzureServiceResult {
	svc := newSignalRResult(t)
	svc.PrivateEndpoints = true
	return svc
}

func TestSignalRScanner_Review(t *testing.T) {
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
		c       SignalRScanner
		args    args
		want    []IAzureServiceResult
		wantErr bool
	}{
		{
			name: "Test Review",
			c: SignalRScanner{
				config: config,
				diagnosticsSettings: DiagnosticsSettings{
					config:                    config,
					diagnosticsSettingsClient: nil,
					hasDiagnosticsFunc: func(resourceId string) (bool, error) {
						return true, nil
					},
				},
				signalrClient: nil,
				listSignalRFunc: func(resourceGroupName string) ([]*armsignalr.ResourceInfo, error) {
					return []*armsignalr.ResourceInfo{
							newSignalR(t),
							newSignalRWithAvailabilityZones(t),
							newSignalRWithPrivateEndpoints(t),
						},
						nil
				},
			},
			args: args{
				resourceGroupName: "resourceGroupName",
			},
			want: []IAzureServiceResult{
				newSignalRResult(t),
				newSignalRAvailabilityZonesResult(t),
				newSignalRPrivateEndpointResult(t),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.c.Scan(tt.args.resourceGroupName)
			if (err != nil) != tt.wantErr {
				t.Errorf("SignalRScanner.Scan() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SignalRScanner.Scan() = %v, want %v", got, tt.want)
			}
		})
	}
}
