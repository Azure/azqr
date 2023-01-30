package scanners

import (
	"context"
	"reflect"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage"
	"github.com/Azure/go-autorest/autorest/to"
)

func newStorage(t *testing.T) *armstorage.Account {
	sku := armstorage.SKUNamePremiumLRS
	tier := armstorage.AccessTierHot
	return &armstorage.Account{
		ID:       to.StringPtr("id"),
		Name:     to.StringPtr("stname"),
		Location: to.StringPtr("westeurope"),
		Type:     to.StringPtr("Microsoft.Storage/storageAccounts"),
		SKU: &armstorage.SKU{
			Name: &sku,
		},
		Properties: &armstorage.AccountProperties{
			AccessTier:                 &tier,
			PrivateEndpointConnections: []*armstorage.PrivateEndpointConnection{},
		},
	}
}

func newStorageWithAvailabilityZones(t *testing.T) *armstorage.Account {
	sku := armstorage.SKUNamePremiumZRS
	svc := newStorage(t)
	svc.SKU.Name = &sku
	return svc
}

func newStorageWithPrivateEndpoints(t *testing.T) *armstorage.Account {
	svc := newStorage(t)
	svc.Properties.PrivateEndpointConnections = []*armstorage.PrivateEndpointConnection{
		{
			ID: to.StringPtr("id"),
		},
	}
	return svc
}

func newStorageResult(t *testing.T) AzureServiceResult {
	return AzureServiceResult{
		SubscriptionID:     "subscriptionId",
		ResourceGroup:      "resourceGroupName",
		ServiceName:        "stname",
		SKU:                "Premium_LRS",
		SLA:                "99.99%",
		Type:               "Microsoft.Storage/storageAccounts",
		Location:           "westeurope",
		CAFNaming:          true,
		AvailabilityZones:  false,
		PrivateEndpoints:   false,
		DiagnosticSettings: true,
	}
}

func newStorageAvailabilityZonesResult(t *testing.T) AzureServiceResult {
	svc := newStorageResult(t)
	svc.AvailabilityZones = true
	svc.SKU = "Premium_ZRS"
	return svc
}

func newStoragePrivateEndpointResult(t *testing.T) AzureServiceResult {
	svc := newStorageResult(t)
	svc.PrivateEndpoints = true
	return svc
}

func TestStorageScanner_Review(t *testing.T) {
	type args struct {
		resourceGroupName string
	}
	tests := []struct {
		name    string
		c       StorageScanner
		args    args
		want    []IAzureServiceResult
		wantErr bool
	}{
		{
			name: "Test Review",
			c: StorageScanner{
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
				storageClient:  nil,
				listStorageFunc: func(resourceGroupName string) ([]*armstorage.Account, error) {
					return []*armstorage.Account{
							newStorage(t),
							newStorageWithAvailabilityZones(t),
							newStorageWithPrivateEndpoints(t),
						},
						nil
				},
			},
			args: args{
				resourceGroupName: "resourceGroupName",
			},
			want: []IAzureServiceResult{
				newStorageResult(t),
				newStorageAvailabilityZonesResult(t),
				newStoragePrivateEndpointResult(t),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.c.Review(tt.args.resourceGroupName)
			if (err != nil) != tt.wantErr {
				t.Errorf("StorageScanner.Review() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("StorageScanner.Review() = %v, want %v", got, tt.want)
			}
		})
	}
}
