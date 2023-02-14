package scanners

import (
	"context"
	"reflect"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/webpubsub/armwebpubsub"
	"github.com/Azure/go-autorest/autorest/to"
)

func newWebPubSub(t *testing.T) *armwebpubsub.ResourceInfo {
	sku := "FreeF1"
	return &armwebpubsub.ResourceInfo{
		ID:       to.StringPtr("id"),
		Name:     to.StringPtr("wpsname"),
		Location: to.StringPtr("westeurope"),
		Type:     to.StringPtr("Microsoft.SignalRService/webPubSub"),
		SKU: &armwebpubsub.ResourceSKU{
			Name: &sku,
		},
		Properties: &armwebpubsub.Properties{
			PrivateEndpointConnections: []*armwebpubsub.PrivateEndpointConnection{},
		},
	}
}

func newWebPubSubWithAvailabilityZones(t *testing.T) *armwebpubsub.ResourceInfo {
	sku := "Premium"
	svc := newWebPubSub(t)
	svc.SKU.Name = &sku
	return svc
}

func newWebPubSubWithPrivateEndpoints(t *testing.T) *armwebpubsub.ResourceInfo {
	svc := newWebPubSub(t)
	svc.Properties.PrivateEndpointConnections = []*armwebpubsub.PrivateEndpointConnection{
		{
			ID: to.StringPtr("id"),
		},
	}
	return svc
}

func newWebPubSubResult(t *testing.T) AzureServiceResult {
	return AzureServiceResult{
		SubscriptionID:     "subscriptionId",
		ResourceGroup:      "resourceGroupName",
		ServiceName:        "wpsname",
		SKU:                "FreeF1",
		SLA:                "None",
		Type:               "Microsoft.SignalRService/webPubSub",
		Location:           "westeurope",
		CAFNaming:          true,
		AvailabilityZones:  false,
		PrivateEndpoints:   false,
		DiagnosticSettings: true,
	}
}

func newWebPubSubAvailabilityZonesResult(t *testing.T) AzureServiceResult {
	svc := newWebPubSubResult(t)
	svc.AvailabilityZones = true
	svc.SKU = "Premium"
	svc.SLA = "99.9%"
	return svc
}

func newWebPubSubPrivateEndpointResult(t *testing.T) AzureServiceResult {
	svc := newWebPubSubResult(t)
	svc.PrivateEndpoints = true
	return svc
}

func TestWebPubSubScanner_Scan(t *testing.T) {
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
		c       WebPubSubScanner
		args    args
		want    []IAzureServiceResult
		wantErr bool
	}{
		{
			name: "Test Scan",
			c: WebPubSubScanner{
				config: config,
				diagnosticsSettings: DiagnosticsSettings{
					config:                    config,
					diagnosticsSettingsClient: nil,
					hasDiagnosticsFunc: func(resourceId string) (bool, error) {
						return true, nil
					},
				},
				client: nil,
				listWebPubSubFunc: func(resourceGroupName string) ([]*armwebpubsub.ResourceInfo, error) {
					return []*armwebpubsub.ResourceInfo{
							newWebPubSub(t),
							newWebPubSubWithAvailabilityZones(t),
							newWebPubSubWithPrivateEndpoints(t),
						},
						nil
				},
			},
			args: args{
				resourceGroupName: "resourceGroupName",
			},
			want: []IAzureServiceResult{
				newWebPubSubResult(t),
				newWebPubSubAvailabilityZonesResult(t),
				newWebPubSubPrivateEndpointResult(t),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.c.Scan(tt.args.resourceGroupName)
			if (err != nil) != tt.wantErr {
				t.Errorf("WebPubSubScanner.Scan() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WebPubSubScanner.Scan() = %v, want %v", got, tt.want)
			}
		})
	}
}
