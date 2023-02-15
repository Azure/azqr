package scanners

import (
	"context"
	"reflect"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/cdn/armcdn"
	"github.com/Azure/go-autorest/autorest/to"
)

func newFrontDoor(t *testing.T) *armcdn.Profile {
	sku := armcdn.SKUNameStandardMicrosoft
	return &armcdn.Profile{
		ID:       to.StringPtr("id"),
		Name:     to.StringPtr("afd-name"),
		Location: to.StringPtr("global"),
		Type:     to.StringPtr("Microsoft.Cdn/profiles"),
		SKU: &armcdn.SKU{
			Name: &sku,
		},
	}
}

func newFrontDoorResult(t *testing.T) AzureServiceResult {
	return AzureServiceResult{
		SubscriptionID:     "subscriptionId",
		ResourceGroup:      "resourceGroupName",
		ServiceName:        "afd-name",
		SKU:                "Standard_Microsoft",
		SLA:                "99.99%",
		Type:               "Microsoft.Cdn/profiles",
		Location:           "global",
		CAFNaming:          true,
		AvailabilityZones:  true,
		PrivateEndpoints:   false,
		DiagnosticSettings: true,
	}
}

func TestFrontDoor_Scan(t *testing.T) {
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
		a       FrontDoorScanner
		args    args
		want    []IAzureServiceResult
		wantErr bool
	}{
		{
			name: "Test Scan",

			a: FrontDoorScanner{
				config: config,
				diagnosticsSettings: DiagnosticsSettings{
					config:                    config,
					diagnosticsSettingsClient: nil,
					hasDiagnosticsFunc: func(resourceId string) (bool, error) {
						return true, nil
					},
				},
				listFunc: func(resourceGroupName string) ([]*armcdn.Profile, error) {
					return []*armcdn.Profile{
							newFrontDoor(t),
						},
						nil
				},
			},
			args: args{
				resourceGroupName: "resourceGroupName",
			},
			want: []IAzureServiceResult{
				newFrontDoorResult(t),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.a.Scan(tt.args.resourceGroupName, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("FrontDoorScanner.Scan() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FrontDoorScanner.Scan() = %v, want %v", got, tt.want)
			}
		})
	}
}
