package st

import (
	"reflect"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/cmendible/azqr/internal/scanners"
)

func TestStorageScanner_Rules(t *testing.T) {
	type fields struct {
		rule                string
		target              interface{}
		scanContext         *scanners.ScanContext
		diagnosticsSettings scanners.DiagnosticsSettings
	}
	type want struct {
		broken bool
		result string
	}
	tests := []struct {
		name   string
		fields fields
		want   want
	}{
		{
			name: "StorageScanner DiagnosticSettings",
			fields: fields{
				rule: "DiagnosticSettings",
				target: &armstorage.Account{
					ID: to.StringPtr("test"),
				},
				scanContext: &scanners.ScanContext{},
				diagnosticsSettings: scanners.DiagnosticsSettings{
					HasDiagnosticsFunc: func(resourceId string) (bool, error) {
						return true, nil
					},
				},
			},
			want: want{
				broken: false,
				result: "",
			},
		},
		{
			name: "StorageScanner Availability Zones",
			fields: fields{
				rule: "AvailabilityZones",
				target: &armstorage.Account{
					SKU: &armstorage.SKU{
						Name: getPremiumZRSSKU(),
					},
				},
				scanContext:         &scanners.ScanContext{},
				diagnosticsSettings: scanners.DiagnosticsSettings{},
			},
			want: want{
				broken: false,
				result: "",
			},
		},
		{
			name: "StorageScanner SLA 99.9%",
			fields: fields{
				rule: "SLA",
				target: &armstorage.Account{
					SKU: &armstorage.SKU{
						Name: getPremiumZRSSKU(),
					},
					Properties: &armstorage.AccountProperties{
						AccessTier: getHotTier(),
					},
				},
				scanContext:         &scanners.ScanContext{},
				diagnosticsSettings: scanners.DiagnosticsSettings{},
			},
			want: want{
				broken: false,
				result: "99.9%",
			},
		},
		{
			name: "StorageScanner Private Endpoint",
			fields: fields{
				rule: "Private",
				target: &armstorage.Account{
					Properties: &armstorage.AccountProperties{
						PrivateEndpointConnections: []*armstorage.PrivateEndpointConnection{
							{
								ID: to.StringPtr("test"),
							},
						},
					},
				},
				scanContext:         &scanners.ScanContext{},
				diagnosticsSettings: scanners.DiagnosticsSettings{},
			},
			want: want{
				broken: false,
				result: "",
			},
		},
		{
			name: "StorageScanner SKU",
			fields: fields{
				rule: "SKU",
				target: &armstorage.Account{
					SKU: &armstorage.SKU{
						Name: getPremiumZRSSKU(),
					},
				},
				scanContext:         &scanners.ScanContext{},
				diagnosticsSettings: scanners.DiagnosticsSettings{},
			},
			want: want{
				broken: false,
				result: "Premium_ZRS",
			},
		},
		{
			name: "StorageScanner CAF",
			fields: fields{
				rule: "CAF",
				target: &armstorage.Account{
					Name: to.StringPtr("sttest"),
				},
				scanContext:         &scanners.ScanContext{},
				diagnosticsSettings: scanners.DiagnosticsSettings{},
			},
			want: want{
				broken: false,
				result: "",
			},
		},
		{
			name: "StorageScanner HTTPS only",
			fields: fields{
				rule: "st-007",
				target: &armstorage.Account{
					Properties: &armstorage.AccountProperties{
						EnableHTTPSTrafficOnly: to.BoolPtr(true),
					},
				},
				scanContext:         &scanners.ScanContext{},
				diagnosticsSettings: scanners.DiagnosticsSettings{},
			},
			want: want{
				broken: false,
				result: "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &StorageScanner{
				diagnosticsSettings: tt.fields.diagnosticsSettings,
			}
			rules := s.GetRules()
			b, w := rules[tt.fields.rule].Eval(tt.fields.target, tt.fields.scanContext)
			got := want{
				broken: b,
				result: w,
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("StorageScanner Rule.Eval() = %v, want %v", got, tt.want)
			}
		})
	}
}

func getPremiumZRSSKU() *armstorage.SKUName {
	s := armstorage.SKUNamePremiumZRS
	return &s
}

func getHotTier() *armstorage.AccessTier {
	s := armstorage.AccessTierHot
	return &s
}
