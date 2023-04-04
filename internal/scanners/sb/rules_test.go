package sb

import (
	"reflect"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/servicebus/armservicebus"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/cmendible/azqr/internal/scanners"
)

func TestServiceBusScanner_Rules(t *testing.T) {
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
			name: "ServiceBusScanner DiagnosticSettings",
			fields: fields{
				rule: "DiagnosticSettings",
				target: &armservicebus.SBNamespace{
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
			name: "ServiceBusScanner Availability Zones",
			fields: fields{
				rule: "AvailabilityZones",
				target: &armservicebus.SBNamespace{
					SKU: &armservicebus.SBSKU{
						Name: getSKUNamePremium(),
					},
					Properties: &armservicebus.SBNamespaceProperties{
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
			name: "ServiceBusScanner SLA 99.9%",
			fields: fields{
				rule: "SLA",
				target: &armservicebus.SBNamespace{
					SKU: &armservicebus.SBSKU{
						Name: getSKUNameStandard(),
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
			name: "ServiceBusScanner SLA 99.95%",
			fields: fields{
				rule: "SLA",
				target: &armservicebus.SBNamespace{
					SKU: &armservicebus.SBSKU{
						Name: getSKUNamePremium(),
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
			name: "ServiceBusScanner Private Endpoint",
			fields: fields{
				rule: "Private",
				target: &armservicebus.SBNamespace{
					Properties: &armservicebus.SBNamespaceProperties{
						PrivateEndpointConnections: []*armservicebus.PrivateEndpointConnection{
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
			name: "ServiceBusScanner SKU",
			fields: fields{
				rule: "SKU",
				target: &armservicebus.SBNamespace{
					SKU: &armservicebus.SBSKU{
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
			name: "ServiceBusScanner CAF",
			fields: fields{
				rule: "CAF",
				target: &armservicebus.SBNamespace{
					Name: to.StringPtr("sb-test"),
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
			name: "ServiceBusScanner Disable Local Auth",
			fields: fields{
				rule: "sb-008",
				target: &armservicebus.SBNamespace{
					Properties: &armservicebus.SBNamespaceProperties{
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
			s := &ServiceBusScanner{
				diagnosticsSettings: tt.fields.diagnosticsSettings,
			}
			rules := s.GetRules()
			b, w := rules[tt.fields.rule].Eval(tt.fields.target, tt.fields.scanContext)
			got := want{
				broken: b,
				result: w,
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ServiceBusScanner Rule.Eval() = %v, want %v", got, tt.want)
			}
		})
	}
}

func getSKUNameStandard() *armservicebus.SKUName {
	s := armservicebus.SKUNameStandard
	return &s
}


func getSKUNamePremium() *armservicebus.SKUName {
	s := armservicebus.SKUNamePremium
	return &s
}
