package st

// import (
// 	"context"
// 	"reflect"
// 	"testing"

// 	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage"
// 	"github.com/Azure/go-autorest/autorest/to"
// )

// func newStorage(t *testing.T) *armstorage.Account {
// 	sku := armstorage.SKUNamePremiumLRS
// 	tier := armstorage.AccessTierHot
// 	return &armstorage.Account{
// 		ID:       to.StringPtr("id"),
// 		Name:     to.StringPtr("stname"),
// 		Location: to.StringPtr("westeurope"),
// 		Type:     to.StringPtr("Microsoft.Storage/storageAccounts"),
// 		SKU: &armstorage.SKU{
// 			Name: &sku,
// 		},
// 		Properties: &armstorage.AccountProperties{
// 			AccessTier:                 &tier,
// 			PrivateEndpointConnections: []*armstorage.PrivateEndpointConnection{},
// 		},
// 	}
// }

// func newStorageStandard_RAGRS_Hot(t *testing.T) *armstorage.Account {
// 	sku := armstorage.SKUNameStandardRAGRS
// 	svc := newStorage(t)
// 	svc.SKU.Name = &sku
// 	return svc
// }

// func newStorageStandard_RAGRS_Cool(t *testing.T) *armstorage.Account {
// 	sku := armstorage.SKUNameStandardRAGRS
// 	svc := newStorage(t)
// 	svc.SKU.Name = &sku
// 	tier := armstorage.AccessTierCool
// 	svc.Properties.AccessTier = &tier

// 	return svc
// }

// func newStorageStandard_LRS_Hot(t *testing.T) *armstorage.Account {
// 	sku := armstorage.SKUNameStandardLRS
// 	svc := newStorage(t)
// 	svc.SKU.Name = &sku
// 	return svc
// }

// func newStorageStandard_LRS_Cool(t *testing.T) *armstorage.Account {
// 	sku := armstorage.SKUNameStandardLRS
// 	svc := newStorage(t)
// 	svc.SKU.Name = &sku
// 	tier := armstorage.AccessTierCool
// 	svc.Properties.AccessTier = &tier

// 	return svc
// }

// func newStorageWithAvailabilityZones(t *testing.T) *armstorage.Account {
// 	sku := armstorage.SKUNamePremiumZRS
// 	svc := newStorage(t)
// 	svc.SKU.Name = &sku
// 	return svc
// }

// func newStorageWithPrivateEndpoints(t *testing.T) *armstorage.Account {
// 	svc := newStorage(t)
// 	svc.Properties.PrivateEndpointConnections = []*armstorage.PrivateEndpointConnection{
// 		{
// 			ID: to.StringPtr("id"),
// 		},
// 	}
// 	return svc
// }

// func newStorageResult(t *testing.T) AzureServiceResult {
// 	return AzureServiceResult{
// 		SubscriptionID:     "subscriptionId",
// 		ResourceGroup:      "resourceGroupName",
// 		ServiceName:        "stname",
// 		SKU:                "Premium_LRS",
// 		SLA:                "99.9%",
// 		Type:               "Microsoft.Storage/storageAccounts",
// 		Location:           "westeurope",
// 		CAFNaming:          true,
// 		AvailabilityZones:  false,
// 		PrivateEndpoints:   false,
// 		DiagnosticSettings: true,
// 	}
// }

// func newStorageAvailabilityZonesResult(t *testing.T) AzureServiceResult {
// 	svc := newStorageResult(t)
// 	svc.AvailabilityZones = true
// 	svc.SKU = "Premium_ZRS"
// 	return svc
// }

// func newStoragePrivateEndpointResult(t *testing.T) AzureServiceResult {
// 	svc := newStorageResult(t)
// 	svc.PrivateEndpoints = true
// 	return svc
// }

// func newStorageStandard_RAGRS_Hot_Result(t *testing.T) AzureServiceResult {
// 	svc := newStorageResult(t)
// 	svc.SKU = "Standard_RAGRS"
// 	svc.SLA = "99.99%"
// 	return svc
// }

// func newStorageStandard_RAGRS_Cool_Result(t *testing.T) AzureServiceResult {
// 	svc := newStorageResult(t)
// 	svc.SKU = "Standard_RAGRS"
// 	svc.SLA = "99.9%"
// 	return svc
// }

// func newStorageStandard_LRS_Hot_Result(t *testing.T) AzureServiceResult {
// 	svc := newStorageResult(t)
// 	svc.SKU = "Standard_LRS"
// 	svc.SLA = "99.9%"
// 	return svc
// }

// func newStorageStandard_LRS_Cool_Result(t *testing.T) AzureServiceResult {
// 	svc := newStorageResult(t)
// 	svc.SKU = "Standard_LRS"
// 	svc.SLA = "99%"
// 	return svc
// }

// func TestStorageScanner_Scan(t *testing.T) {
// 	type args struct {
// 		resourceGroupName string
// 	}
// 	config := &ScannerConfig{
// 		SubscriptionID: "subscriptionId",
// 		Cred:           nil,
// 		Ctx:            context.TODO(),
// 	}
// 	tests := []struct {
// 		name    string
// 		c       StorageScanner
// 		args    args
// 		want    []AzureServiceResult
// 		wantErr bool
// 	}{
// 		{
// 			name: "Test Scan",
// 			c: StorageScanner{
// 				config: config,
// 				diagnosticsSettings: DiagnosticsSettings{
// 					config:                    config,
// 					diagnosticsSettingsClient: nil,
// 					hasDiagnosticsFunc: func(resourceId string) (bool, error) {
// 						return true, nil
// 					},
// 				},
// 				storageClient: nil,
// 				listStorageFunc: func(resourceGroupName string) ([]*armstorage.Account, error) {
// 					return []*armstorage.Account{
// 							newStorage(t),
// 							newStorageWithAvailabilityZones(t),
// 							newStorageWithPrivateEndpoints(t),
// 							newStorageStandard_RAGRS_Hot(t),
// 							newStorageStandard_RAGRS_Cool(t),
// 							newStorageStandard_LRS_Hot(t),
// 							newStorageStandard_LRS_Cool(t),
// 						},
// 						nil
// 				},
// 			},
// 			args: args{
// 				resourceGroupName: "resourceGroupName",
// 			},
// 			want: []AzureServiceResult{
// 				newStorageResult(t),
// 				newStorageAvailabilityZonesResult(t),
// 				newStoragePrivateEndpointResult(t),
// 				newStorageStandard_RAGRS_Hot_Result(t),
// 				newStorageStandard_RAGRS_Cool_Result(t),
// 				newStorageStandard_LRS_Hot_Result(t),
// 				newStorageStandard_LRS_Cool_Result(t),
// 			},
// 			wantErr: false,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			got, err := tt.c.Scan(tt.args.resourceGroupName, nil)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("StorageScanner.Scan() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("StorageScanner.Scan() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }
