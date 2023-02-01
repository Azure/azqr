package scanners

import (
	"context"
	"reflect"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/sql/armsql"
	"github.com/Azure/go-autorest/autorest/to"
)

func newSQLServer(t *testing.T) *armsql.Server {
	return &armsql.Server{
		ID:       to.StringPtr("id"),
		Name:     to.StringPtr("sql-name"),
		Location: to.StringPtr("westeurope"),
		Type:     to.StringPtr("Microsoft.Sql/servers"),
		Properties: &armsql.ServerProperties{
			PrivateEndpointConnections: []*armsql.ServerPrivateEndpointConnection{},
		},
	}
}

func newSQLServerWithPrivateEndpoints(t *testing.T) *armsql.Server {
	svc := newSQLServer(t)
	svc.Properties.PrivateEndpointConnections = []*armsql.ServerPrivateEndpointConnection{
		{
			ID: to.StringPtr("id"),
		},
	}
	return svc
}

func newSQLServerResult(t *testing.T) AzureServiceResult {
	return AzureServiceResult{
		SubscriptionID:     "subscriptionId",
		ResourceGroup:      "resourceGroupName",
		ServiceName:        "sql-name",
		SKU:                "N/A",
		SLA:                "N/A",
		Type:               "Microsoft.Sql/servers",
		Location:           "westeurope",
		CAFNaming:          true,
		AvailabilityZones:  false,
		PrivateEndpoints:   false,
		DiagnosticSettings: true,
	}
}

func newSQLServerPrivateEndpointResult(t *testing.T) AzureServiceResult {
	svc := newSQLServerResult(t)
	svc.PrivateEndpoints = true
	return svc
}

func TestSQLScanner_Review(t *testing.T) {
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
		c       SQLScanner
		args    args
		want    []IAzureServiceResult
		wantErr bool
	}{
		{
			name: "Test Review",
			c: SQLScanner{
				config: config,
				diagnosticsSettings: DiagnosticsSettings{
					config:                    config,
					diagnosticsSettingsClient: nil,
					hasDiagnosticsFunc: func(resourceId string) (bool, error) {
						return true, nil
					},
				},
				sqlClient:          nil,
				sqlDatabasedClient: nil,
				listServersFunc: func(resourceGroupName string) ([]*armsql.Server, error) {
					return []*armsql.Server{
							newSQLServer(t),
							newSQLServerWithPrivateEndpoints(t),
						},
						nil
				},
				listDatabasesFunc: func(resourceGroupName, serverName string) ([]*armsql.Database, error) {
					return []*armsql.Database{}, nil
				},
			},
			args: args{
				resourceGroupName: "resourceGroupName",
			},
			want: []IAzureServiceResult{
				newSQLServerResult(t),
				newSQLServerPrivateEndpointResult(t),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.c.Review(tt.args.resourceGroupName)
			if (err != nil) != tt.wantErr {
				t.Errorf("SQLScanner.Review() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SQLScanner.Review() = %v, want %v", got, tt.want)
			}
		})
	}
}
