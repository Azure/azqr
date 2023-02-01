package scanners

import (
	"context"
	"reflect"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/redis/armredis"
	"github.com/Azure/go-autorest/autorest/to"
)

func newRedis(t *testing.T) *armredis.ResourceInfo {
	sku := armredis.SKUNameBasic
	return &armredis.ResourceInfo{
		ID:       to.StringPtr("id"),
		Name:     to.StringPtr("redis-name"),
		Location: to.StringPtr("westeurope"),
		Type:     to.StringPtr("Microsoft.Cache/Redis"),
		Zones:    []*string{},
		Properties: &armredis.Properties{
			SKU: &armredis.SKU{
				Name: &sku,
			},
			PrivateEndpointConnections: []*armredis.PrivateEndpointConnection{},
		},
	}
}

func newRedisWithAvailabilityZones(t *testing.T) *armredis.ResourceInfo {
	svc := newRedis(t)
	svc.Zones = []*string{to.StringPtr("1"), to.StringPtr("2"), to.StringPtr("3")}
	return svc
}

func newRedisWithPrivateEndpoints(t *testing.T) *armredis.ResourceInfo {
	svc := newRedis(t)
	svc.Properties.PrivateEndpointConnections = []*armredis.PrivateEndpointConnection{
		{
			ID: to.StringPtr("id"),
		},
	}
	return svc
}

func newRedisResult(t *testing.T) AzureServiceResult {
	return AzureServiceResult{
		SubscriptionID:     "subscriptionId",
		ResourceGroup:      "resourceGroupName",
		ServiceName:        "redis-name",
		SKU:                "Basic",
		SLA:                "99.9%",
		Type:               "Microsoft.Cache/Redis",
		Location:           "westeurope",
		CAFNaming:          true,
		AvailabilityZones:  false,
		PrivateEndpoints:   false,
		DiagnosticSettings: true,
	}
}

func newRedisAvailabilityZonesResult(t *testing.T) AzureServiceResult {
	svc := newRedisResult(t)
	svc.AvailabilityZones = true
	return svc
}

func newRedisPrivateEndpointResult(t *testing.T) AzureServiceResult {
	svc := newRedisResult(t)
	svc.PrivateEndpoints = true
	return svc
}

func TestRedisScanner_Review(t *testing.T) {
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
		c       RedisScanner
		args    args
		want    []IAzureServiceResult
		wantErr bool
	}{
		{
			name: "Test Review",
			c: RedisScanner{
				config: config,
				diagnosticsSettings: DiagnosticsSettings{
					config:                    config,
					diagnosticsSettingsClient: nil,
					hasDiagnosticsFunc: func(resourceId string) (bool, error) {
						return true, nil
					},
				},
				redisClient:    nil,
				listRedisFunc: func(resourceGroupName string) ([]*armredis.ResourceInfo, error) {
					return []*armredis.ResourceInfo{
							newRedis(t),
							newRedisWithAvailabilityZones(t),
							newRedisWithPrivateEndpoints(t),
						},
						nil
				},
			},
			args: args{
				resourceGroupName: "resourceGroupName",
			},
			want: []IAzureServiceResult{
				newRedisResult(t),
				newRedisAvailabilityZonesResult(t),
				newRedisPrivateEndpointResult(t),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.c.Review(tt.args.resourceGroupName)
			if (err != nil) != tt.wantErr {
				t.Errorf("RedisScanner.Review() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RedisScanner.Review() = %v, want %v", got, tt.want)
			}
		})
	}
}
