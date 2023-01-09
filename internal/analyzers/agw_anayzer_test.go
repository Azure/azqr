package analyzers

import (
	"context"
	"reflect"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"github.com/Azure/go-autorest/autorest/to"
)

func newApplicationGateway(t *testing.T) *armnetwork.ApplicationGateway {
	sku := armnetwork.ApplicationGatewaySKUNameStandardV2
	return &armnetwork.ApplicationGateway{
		ID:       to.StringPtr("id"),
		Name:     to.StringPtr("agw-name"),
		Location: to.StringPtr("westeurope"),
		Type:     to.StringPtr("Microsoft.Network/applicationGateways"),
		Properties: &armnetwork.ApplicationGatewayPropertiesFormat{
			SKU: &armnetwork.ApplicationGatewaySKU{
				Name: &sku,
			},
		},
	}
}

func newApplicationGatewayWithAvailabilityZones(t *testing.T) *armnetwork.ApplicationGateway {
	svc := newApplicationGateway(t)
	svc.Zones = []*string{to.StringPtr("1"), to.StringPtr("2"), to.StringPtr("3")}
	return svc
}

func newApplicationGatewayWithPrivateEndpoints(t *testing.T) *armnetwork.ApplicationGateway {
	svc := newApplicationGateway(t)
	svc.Properties.PrivateEndpointConnections = []*armnetwork.ApplicationGatewayPrivateEndpointConnection{
		{
			ID: to.StringPtr("id"),
		},
	}
	return svc
}

func newApplicationGatewayResult(t *testing.T) AzureServiceResult {
	return AzureServiceResult{
		SubscriptionID:     "subscriptionId",
		ResourceGroup:      "resourceGroupName",
		ServiceName:        "agw-name",
		SKU:                "Standard_v2",
		SLA:                "99.95%",
		Type:               "Microsoft.Network/applicationGateways",
		Location:           "westeurope",
		CAFNaming:          true,
		AvailabilityZones:  false,
		PrivateEndpoints:   false,
		DiagnosticSettings: true,
	}
}

func newApplicationGatewayAvailabilityZonesResult(t *testing.T) AzureServiceResult {
	svc := newApplicationGatewayResult(t)
	svc.AvailabilityZones = true
	return svc
}

func newApplicationGatewayPrivateEndpointResult(t *testing.T) AzureServiceResult {
	svc := newApplicationGatewayResult(t)
	svc.PrivateEndpoints = true
	return svc
}

func TestApplicationGatewayAnalyzer_Review(t *testing.T) {
	type args struct {
		resourceGroupName string
	}
	tests := []struct {
		name    string
		a       ApplicationGatewayAnalyzer
		args    args
		want    []IAzureServiceResult
		wantErr bool
	}{
		{
			name: "Test Review",
			a: ApplicationGatewayAnalyzer{
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
				gatewaysClient: nil,
				listGatewaysFunc: func(resourceGroupName string) ([]*armnetwork.ApplicationGateway, error) {
					return []*armnetwork.ApplicationGateway{
							newApplicationGateway(t),
							newApplicationGatewayWithAvailabilityZones(t),
							newApplicationGatewayWithPrivateEndpoints(t),
						},
						nil
				},
			},
			args: args{
				resourceGroupName: "resourceGroupName",
			},
			want: []IAzureServiceResult{
				newApplicationGatewayResult(t),
				newApplicationGatewayAvailabilityZonesResult(t),
				newApplicationGatewayPrivateEndpointResult(t),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.a.Review(tt.args.resourceGroupName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ApplicationGatewayAnalyzer.Review() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ApplicationGatewayAnalyzer.Review() = %v, want %v", got, tt.want)
			}
		})
	}
}
