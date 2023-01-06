package analyzers

import (
	"context"
	"reflect"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/eventhub/armeventhub"
	"github.com/Azure/go-autorest/autorest/to"
)

func newEventHub(t *testing.T) *armeventhub.EHNamespace {
	sku := armeventhub.SKUNameBasic
	return &armeventhub.EHNamespace{
		ID:       to.StringPtr("id"),
		Name:     to.StringPtr("evh-name"),
		Location: to.StringPtr("westeurope"),
		Type:     to.StringPtr("Microsoft.EventHub/Namespaces"),
		SKU: &armeventhub.SKU{
			Name: &sku,
		},
		Properties: &armeventhub.EHNamespaceProperties{
			ZoneRedundant: to.BoolPtr(false),
			PrivateEndpointConnections: []*armeventhub.PrivateEndpointConnection{},
		},
	}
}

func newEventHubWithAvailabilityZones(t *testing.T) *armeventhub.EHNamespace {
	svc := newEventHub(t)
	svc.Properties.ZoneRedundant = to.BoolPtr(true)
	return svc
}

func newEventHubWithPrivateEndpoints(t *testing.T) *armeventhub.EHNamespace {
	svc := newEventHub(t)
	svc.Properties.PrivateEndpointConnections = []*armeventhub.PrivateEndpointConnection{
		{
			ID: to.StringPtr("id"),
		},
	}
	return svc
}

func newEventHubResult(t *testing.T) AzureServiceResult {
	return AzureServiceResult{
		AzureBaseServiceResult: AzureBaseServiceResult{
			SubscriptionID: "subscriptionId",
			ResourceGroup:  "resourceGroupName",
			ServiceName:    "evh-name",
			SKU:            "Basic",
			SLA:            "99.95%",
			Type:           "Microsoft.EventHub/Namespaces",
			Location:       "westeurope",
			CAFNaming:      true,
		},
		AvailabilityZones:  false,
		PrivateEndpoints:   false,
		DiagnosticSettings: true,
	}
}

func newEventHubAvailabilityZonesResult(t *testing.T) AzureServiceResult {
	svc := newEventHubResult(t)
	svc.AvailabilityZones = true
	return svc
}

func newEventHubPrivateEndpointResult(t *testing.T) AzureServiceResult {
	svc := newEventHubResult(t)
	svc.PrivateEndpoints = true
	return svc
}

func TestEventHubAnalyzer_Review(t *testing.T) {
	type args struct {
		resourceGroupName string
	}
	tests := []struct {
		name    string
		c       EventHubAnalyzer
		args    args
		want    []AzureServiceResult
		wantErr bool
	}{
		{
			name: "Test Review",
			c: EventHubAnalyzer{
				diagnosticsSettings: DiagnosticsSettings{
					diagnosticsSettingsClient: nil,
					ctx:                       context.TODO(),
					hasDiagnosticsFunc: func(resourceId string) (bool, error) {
						return true, nil
					},
				},
				subscriptionID:   "subscriptionId",
				ctx:              context.TODO(),
				cred:             nil,
				client: nil,
				listEventHubsFunc: func(resourceGroupName string) ([]*armeventhub.EHNamespace, error) {
					return []*armeventhub.EHNamespace{
							newEventHub(t),
							newEventHubWithAvailabilityZones(t),
							newEventHubWithPrivateEndpoints(t),
						},
						nil
				},
			},
			args: args{
				resourceGroupName: "resourceGroupName",
			},
			want: []AzureServiceResult{
				newEventHubResult(t),
				newEventHubAvailabilityZonesResult(t),
				newEventHubPrivateEndpointResult(t),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.c.Review(tt.args.resourceGroupName)
			if (err != nil) != tt.wantErr {
				t.Errorf("EventHubAnalyzer.Review() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("EventHubAnalyzer.Review() = %v, want %v", got, tt.want)
			}
		})
	}
}
