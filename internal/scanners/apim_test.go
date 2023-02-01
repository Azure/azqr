package scanners

import (
	"context"
	"reflect"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/apimanagement/armapimanagement"
	"github.com/Azure/go-autorest/autorest/to"
)

func newAPIM(t *testing.T) *armapimanagement.ServiceResource {
	sku := armapimanagement.SKUTypeDeveloper
	return &armapimanagement.ServiceResource{
		ID:       to.StringPtr("id"),
		Name:     to.StringPtr("apim-name"),
		Location: to.StringPtr("westeurope"),
		Type:     to.StringPtr("Microsoft.ApiManagement/service"),
		SKU: &armapimanagement.ServiceSKUProperties{
			Name: &sku,
		},
		Properties: &armapimanagement.ServiceProperties{
			PrivateEndpointConnections: []*armapimanagement.RemotePrivateEndpointConnectionWrapper{},
		},
		Zones: []*string{},
	}
}

func newAPIMWithAvailabilityZones(t *testing.T) *armapimanagement.ServiceResource {
	svc := newAPIM(t)
	svc.Zones = []*string{to.StringPtr("1"), to.StringPtr("2"), to.StringPtr("3")}
	return svc
}

func newAPIMWithPrivateEndpoints(t *testing.T) *armapimanagement.ServiceResource {
	svc := newAPIM(t)
	svc.Properties.PrivateEndpointConnections = []*armapimanagement.RemotePrivateEndpointConnectionWrapper{
		{
			ID: to.StringPtr("id"),
		},
	}
	return svc
}

func newAPIMResult(t *testing.T) AzureServiceResult {
	return AzureServiceResult{
		SubscriptionID:     "subscriptionId",
		ResourceGroup:      "resourceGroupName",
		ServiceName:        "apim-name",
		SKU:                "Developer",
		SLA:                "None",
		Type:               "Microsoft.ApiManagement/service",
		Location:           "westeurope",
		CAFNaming:          true,
		AvailabilityZones:  false,
		PrivateEndpoints:   false,
		DiagnosticSettings: true,
	}
}

func newAPIMAvailabilityZonesResult(t *testing.T) AzureServiceResult {
	svc := newAPIMResult(t)
	svc.AvailabilityZones = true
	return svc
}

func newAPIMPrivateEndpointResult(t *testing.T) AzureServiceResult {
	svc := newAPIMResult(t)
	svc.PrivateEndpoints = true
	return svc
}

func TestAPIMAnalyzer_Review(t *testing.T) {
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
		a       APIManagementScanner
		args    args
		want    []IAzureServiceResult
		wantErr bool
	}{
		{
			name: "Test Review",
			a: APIManagementScanner{
				config: config,
				diagnosticsSettings: DiagnosticsSettings{
					config:                    config,
					diagnosticsSettingsClient: nil,
					hasDiagnosticsFunc: func(resourceId string) (bool, error) {
						return true, nil
					},
				},
				serviceClient: nil,
				listServicesFunc: func(resourceGroupName string) ([]*armapimanagement.ServiceResource, error) {
					return []*armapimanagement.ServiceResource{
							newAPIM(t),
							newAPIMWithAvailabilityZones(t),
							newAPIMWithPrivateEndpoints(t),
						},
						nil
				},
			},
			args: args{
				resourceGroupName: "resourceGroupName",
			},
			want: []IAzureServiceResult{
				newAPIMResult(t),
				newAPIMAvailabilityZonesResult(t),
				newAPIMPrivateEndpointResult(t),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.a.Scan(tt.args.resourceGroupName)
			if (err != nil) != tt.wantErr {
				t.Errorf("APIManagementScanner.Scan() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("APIManagementScanner.Scan() = %v, want %v", got, tt.want)
			}
		})
	}
}
