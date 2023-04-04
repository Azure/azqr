package evh

import (
	"reflect"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/eventhub/armeventhub"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/cmendible/azqr/internal/scanners"
)

func TestEventHubScanner_Rules(t *testing.T) {
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
			name: "EventHubScanner DiagnosticSettings",
			fields: fields{
				rule: "DiagnosticSettings",
				target: &armeventhub.EHNamespace{
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
			name: "EventHubScanner Availability Zones",
			fields: fields{
				rule: "AvailabilityZones",
				target: &armeventhub.EHNamespace{
					Properties: &armeventhub.EHNamespaceProperties{
						ZoneRedundant: to.BoolPtr(true),
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
			name: "EventHubScanner SLA 99.95%",
			fields: fields{
				rule: "SLA",
				target: &armeventhub.EHNamespace{
					SKU: &armeventhub.SKU{
						Name: getSKUNameStandard(),
					},
				},
				scanContext:         &scanners.ScanContext{},
				diagnosticsSettings: scanners.DiagnosticsSettings{},
			},
			want: want{
				broken: false,
				result: "99.95%",
			},
		},
		{
			name: "EventHubScanner SLA 99.99%",
			fields: fields{
				rule: "SLA",
				target: &armeventhub.EHNamespace{
					SKU: &armeventhub.SKU{
						Name: getSKUNamePremium(),
					},
				},
				scanContext:         &scanners.ScanContext{},
				diagnosticsSettings: scanners.DiagnosticsSettings{},
			},
			want: want{
				broken: false,
				result: "99.99%",
			},
		},
		{
			name: "EventHubScanner Private Endpoint",
			fields: fields{
				rule: "Private",
				target: &armeventhub.EHNamespace{
					Properties: &armeventhub.EHNamespaceProperties{
						PrivateEndpointConnections: []*armeventhub.PrivateEndpointConnection{
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
			name: "EventHubScanner SKU",
			fields: fields{
				rule: "SKU",
				target: &armeventhub.EHNamespace{
					SKU: &armeventhub.SKU{
						Name: getSKUNamePremium(),
					},
				},
				scanContext:         &scanners.ScanContext{},
				diagnosticsSettings: scanners.DiagnosticsSettings{},
			},
			want: want{
				broken: false,
				result: "Premium",
			},
		},
		{
			name: "EventHubScanner CAF",
			fields: fields{
				rule: "CAF",
				target: &armeventhub.EHNamespace{
					Name: to.StringPtr("evh-test"),
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
			name: "EventHubScanner Disable Local Auth",
			fields: fields{
				rule: "evh-008",
				target: &armeventhub.EHNamespace{
					Properties: &armeventhub.EHNamespaceProperties{
						DisableLocalAuth: to.BoolPtr(true),
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
			s := &EventHubScanner{
				diagnosticsSettings: tt.fields.diagnosticsSettings,
			}
			rules := s.GetRules()
			b, w := rules[tt.fields.rule].Eval(tt.fields.target, tt.fields.scanContext)
			got := want{
				broken: b,
				result: w,
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("EventHubScanner Rule.Eval() = %v, want %v", got, tt.want)
			}
		})
	}
}

func getSKUNameStandard() *armeventhub.SKUName {
	s := armeventhub.SKUNameStandard
	return &s
}

func getSKUNamePremium() *armeventhub.SKUName {
	s := armeventhub.SKUNamePremium
	return &s
}
