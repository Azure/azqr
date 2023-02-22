package appcs

// import (
// 	"context"
// 	"reflect"
// 	"testing"

// 	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appconfiguration/armappconfiguration"
// 	"github.com/Azure/go-autorest/autorest/to"
// )

// func newAppConfiguration(t *testing.T) *armappconfiguration.ConfigurationStore {
// 	sku := "Free"
// 	return &armappconfiguration.ConfigurationStore{
// 		ID:       to.StringPtr("id"),
// 		Name:     to.StringPtr("appcs-name"),
// 		Location: to.StringPtr("westeurope"),
// 		Type:     to.StringPtr("Microsoft.AppConfiguration/configurationStores"),
// 		SKU: &armappconfiguration.SKU{
// 			Name: &sku,
// 		},
// 		Properties: &armappconfiguration.ConfigurationStoreProperties{
// 			PrivateEndpointConnections: []*armappconfiguration.PrivateEndpointConnectionReference{},
// 		},
// 	}
// }

// func newAppConfigurationWithSLA(t *testing.T) *armappconfiguration.ConfigurationStore {
// 	sku := "Standard"
// 	svc := newAppConfiguration(t)
// 	svc.SKU.Name = &sku
// 	return svc
// }

// func newAppConfigurationWithPrivateEndpoints(t *testing.T) *armappconfiguration.ConfigurationStore {
// 	svc := newAppConfiguration(t)
// 	svc.Properties.PrivateEndpointConnections = []*armappconfiguration.PrivateEndpointConnectionReference{
// 		{
// 			ID: to.StringPtr("id"),
// 		},
// 	}
// 	return svc
// }

// func newAppConfigurationResult(t *testing.T) AzureServiceResult {
// 	return AzureServiceResult{
// 		SubscriptionID:     "subscriptionId",
// 		ResourceGroup:      "resourceGroupName",
// 		ServiceName:        "appcs-name",
// 		SKU:                "Free",
// 		SLA:                "None",
// 		Type:               "Microsoft.AppConfiguration/configurationStores",
// 		Location:           "westeurope",
// 		CAFNaming:          true,
// 		AvailabilityZones:  false,
// 		PrivateEndpoints:   false,
// 		DiagnosticSettings: true,
// 	}
// }

// func newAppConfigurationWithSLAResult(t *testing.T) AzureServiceResult {
// 	svc := newAppConfigurationResult(t)
// 	svc.SKU = "Standard"
// 	svc.SLA = "99.9%"
// 	return svc
// }

// func newAppConfigurationPrivateEndpointResult(t *testing.T) AzureServiceResult {
// 	svc := newAppConfigurationResult(t)
// 	svc.PrivateEndpoints = true
// 	return svc
// }

// func TestAppConfigurationScanner_Scan(t *testing.T) {
// 	config := &ScannerConfig{
// 		SubscriptionID: "subscriptionId",
// 		Cred:           nil,
// 		Ctx:            context.TODO(),
// 	}
// 	type args struct {
// 		resourceGroupName string
// 	}
// 	tests := []struct {
// 		name    string
// 		a       AppConfigurationScanner
// 		args    args
// 		want    []AzureServiceResult
// 		wantErr bool
// 	}{
// 		{
// 			name: "Test Scan",
// 			a: AppConfigurationScanner{
// 				config: config,
// 				diagnosticsSettings: DiagnosticsSettings{
// 					config:                    config,
// 					diagnosticsSettingsClient: nil,
// 					hasDiagnosticsFunc: func(resourceId string) (bool, error) {
// 						return true, nil
// 					},
// 				},
// 				client: nil,
// 				listFunc: func(resourceGroupName string) ([]*armappconfiguration.ConfigurationStore, error) {
// 					return []*armappconfiguration.ConfigurationStore{
// 							newAppConfiguration(t),
// 							newAppConfigurationWithSLA(t),
// 							newAppConfigurationWithPrivateEndpoints(t),
// 						},
// 						nil
// 				},
// 			},
// 			args: args{
// 				resourceGroupName: "resourceGroupName",
// 			},
// 			want: []AzureServiceResult{
// 				newAppConfigurationResult(t),
// 				newAppConfigurationWithSLAResult(t),
// 				newAppConfigurationPrivateEndpointResult(t),
// 			},
// 			wantErr: false,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			got, err := tt.a.Scan(tt.args.resourceGroupName, nil)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("AppConfigurationScanner.Scan() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("AppConfigurationScanner.Scan() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }
