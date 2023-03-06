package afd

import (
	"reflect"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/cdn/armcdn"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/cmendible/azqr/internal/scanners"
)

func TestFrontDoorScanner_Rules(t *testing.T) {
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
			name: "FrontDoorScanner DiagnosticSettings",
			fields: fields{
				rule: "DiagnosticSettings",
				target: &armcdn.Profile{
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
			name: "FrontDoorScanner SLA",
			fields: fields{
				rule: "SLA",
				target: &armcdn.Profile{
					ID: to.StringPtr("test"),
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
			name: "FrontDoorScanner SKU",
			fields: fields{
				rule: "SKU",
				target: &armcdn.Profile{
					ID: to.StringPtr("test"),
					SKU: &armcdn.SKU{
						Name: getSKU(),
					},
				},
				scanContext:         &scanners.ScanContext{},
				diagnosticsSettings: scanners.DiagnosticsSettings{},
			},
			want: want{
				broken: false,
				result: "Standard_Microsoft",
			},
		},
		{
			name: "FrontDoorScanner CAF",
			fields: fields{
				rule: "CAF",
				target: &armcdn.Profile{
					ID:   to.StringPtr("test"),
					Name: to.StringPtr("afd-test"),
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
			s := &FrontDoorScanner{
				diagnosticsSettings: tt.fields.diagnosticsSettings,
			}
			rules := s.GetRules()
			b, w := rules[tt.fields.rule].Eval(tt.fields.target, tt.fields.scanContext)
			got := want{
				broken: b,
				result: w,
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FrontDoorScanner Rule.Eval() = %v, want %v", got, tt.want)
			}
		})
	}
}

func getSKU() *armcdn.SKUName {
	s := armcdn.SKUNameStandardMicrosoft
	return &s
}
