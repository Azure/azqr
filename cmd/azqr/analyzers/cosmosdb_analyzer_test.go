package analyzers

import (
	"context"
	"reflect"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/cosmos/armcosmos"
	"github.com/Azure/go-autorest/autorest/to"
)

func newCosmosDB(t *testing.T) *armcosmos.DatabaseAccountGetResults {
	return &armcosmos.DatabaseAccountGetResults{
		ID:       to.StringPtr("id"),
		Name:     to.StringPtr("cosmosdb-name"),
		Location: to.StringPtr("westeurope"),
		Type:     to.StringPtr("Microsoft.DocumentDB/databaseAccounts"),
		Properties: &armcosmos.DatabaseAccountGetProperties{
			DatabaseAccountOfferType: to.StringPtr("Standard"),
			Locations: []*armcosmos.Location{
				{
					LocationName:    to.StringPtr("westeurope"),
					IsZoneRedundant: to.BoolPtr(false),
				},
			},
			PrivateEndpointConnections: []*armcosmos.PrivateEndpointConnection{},
		},
	}
}

func newCosmosDBWithAvailabilityZones(t *testing.T) *armcosmos.DatabaseAccountGetResults {
	svc := newCosmosDB(t)
	svc.Properties.Locations = []*armcosmos.Location{
		{
			LocationName:    to.StringPtr("westeurope"),
			IsZoneRedundant: to.BoolPtr(true),
		},
	}
	return svc
}

func newCosmosDBWithPrivateEndpoints(t *testing.T) *armcosmos.DatabaseAccountGetResults {
	svc := newCosmosDB(t)
	svc.Properties.PrivateEndpointConnections = []*armcosmos.PrivateEndpointConnection{
		{
			ID: to.StringPtr("id"),
		},
	}
	return svc
}

func newCosmosDBResult(t *testing.T) AzureServiceResult {
	return AzureServiceResult{
		AzureBaseServiceResult: AzureBaseServiceResult{
			SubscriptionID: "subscriptionId",
			ResourceGroup:  "resourceGroupName",
			ServiceName:    "cosmosdb-name",
			SKU:            "Standard",
			SLA:            "99.99%",
			Type:           "Microsoft.DocumentDB/databaseAccounts",
			Location:       "westeurope",
			CAFNaming:      true,
		},
		AvailabilityZones:  false,
		PrivateEndpoints:   false,
		DiagnosticSettings: true,
	}
}

func newCosmosDBAvailabilityZonesResult(t *testing.T) AzureServiceResult {
	svc := newCosmosDBResult(t)
	svc.AvailabilityZones = true
	svc.SLA = "99.995%"
	return svc
}

func newCosmosDBPrivateEndpointResult(t *testing.T) AzureServiceResult {
	svc := newCosmosDBResult(t)
	svc.PrivateEndpoints = true
	return svc
}

func TestCosmosDBAnalyzer_Review(t *testing.T) {
	type args struct {
		resourceGroupName string
	}
	tests := []struct {
		name    string
		c       CosmosDBAnalyzer
		args    args
		want    []AzureServiceResult
		wantErr bool
	}{
		{
			name: "Test Review",
			c: CosmosDBAnalyzer{
				diagnosticsSettings: DiagnosticsSettings{
					diagnosticsSettingsClient: nil,
					ctx:                       context.TODO(),
					hasDiagnosticsFunc: func(resourceId string) (bool, error) {
						return true, nil
					},
				},
				subscriptionID:  "subscriptionId",
				ctx:             context.TODO(),
				cred:            nil,
				databasesClient: nil,
				listDatabasesFunc: func(resourceGroupName string) ([]*armcosmos.DatabaseAccountGetResults, error) {
					return []*armcosmos.DatabaseAccountGetResults{
							newCosmosDB(t),
							newCosmosDBWithAvailabilityZones(t),
							newCosmosDBWithPrivateEndpoints(t),
						},
						nil
				},
			},
			args: args{
				resourceGroupName: "resourceGroupName",
			},
			want: []AzureServiceResult{
				newCosmosDBResult(t),
				newCosmosDBAvailabilityZonesResult(t),
				newCosmosDBPrivateEndpointResult(t),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.c.Review(tt.args.resourceGroupName)
			if (err != nil) != tt.wantErr {
				t.Errorf("CosmosDBAnalyzer.Review() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CosmosDBAnalyzer.Review() = %v, want %v", got, tt.want)
			}
		})
	}
}
