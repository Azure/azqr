package agw

import (
	"reflect"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/cmendible/azqr/internal/scanners"
)

func TestApplicationGatewayScanner_Rules(t *testing.T) {
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
			name: "ApplicationGatewayScanner DiagnosticSettings",
			fields: fields{
				rule: "DiagnosticSettings",
				target: &armnetwork.ApplicationGateway{
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
			name: "ApplicationGatewayScanner SLA",
			fields: fields{
				rule: "SLA",
				target: &armnetwork.ApplicationGateway{
					ID: to.StringPtr("test"),
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
			name: "ApplicationGatewayScanner SKU",
			fields: fields{
				rule: "SKU",
				target: &armnetwork.ApplicationGateway{
					ID: to.StringPtr("test"),
					Properties: &armnetwork.ApplicationGatewayPropertiesFormat{
						SKU: &armnetwork.ApplicationGatewaySKU{
							Name: getSKUName(),
						},
					},
				},
				scanContext:         &scanners.ScanContext{},
				diagnosticsSettings: scanners.DiagnosticsSettings{},
			},
			want: want{
				broken: false,
				result: "Standard_v2",
			},
		},
		{
			name: "ApplicationGatewayScanner CAF",
			fields: fields{
				rule: "CAF",
				target: &armnetwork.ApplicationGateway{
					ID:   to.StringPtr("test"),
					Name: to.StringPtr("agw-test"),
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
			s := &ApplicationGatewayScanner{
				diagnosticsSettings: tt.fields.diagnosticsSettings,
			}
			rules := s.GetRules()
			b, w := rules[tt.fields.rule].Eval(tt.fields.target, tt.fields.scanContext)
			got := want{
				broken: b,
				result: w,
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ApplicationGatewayScanner Rule.Eval() = %v, want %v", got, tt.want)
			}
		})
	}
}

func getSKUName() *armnetwork.ApplicationGatewaySKUName {
	s := armnetwork.ApplicationGatewaySKUNameStandardV2
	return &s
}
