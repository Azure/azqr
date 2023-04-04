package cr

import (
	"reflect"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerregistry/armcontainerregistry"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/cmendible/azqr/internal/scanners"
)

func TestContainerRegistryScanner_Rules(t *testing.T) {
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
			name: "ContainerRegistryScanner DiagnosticSettings",
			fields: fields{
				rule: "DiagnosticSettings",
				target: &armcontainerregistry.Registry{
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
			name: "ContainerRegistryScanner Availability Zones",
			fields: fields{
				rule: "AvailabilityZones",
				target: &armcontainerregistry.Registry{
					Properties: &armcontainerregistry.RegistryProperties{
						ZoneRedundancy: getZoneRedundancy(),
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
			name: "ContainerRegistryScanner SLA",
			fields: fields{
				rule:                "SLA",
				target:              &armcontainerregistry.Registry{},
				scanContext:         &scanners.ScanContext{},
				diagnosticsSettings: scanners.DiagnosticsSettings{},
			},
			want: want{
				broken: false,
				result: "99.95%",
			},
		},
		{
			name: "ContainerRegistryScanner Private Endpoint",
			fields: fields{
				rule: "Private",
				target: &armcontainerregistry.Registry{
					Properties: &armcontainerregistry.RegistryProperties{
						PrivateEndpointConnections: []*armcontainerregistry.PrivateEndpointConnection{
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
			name: "ContainerRegistryScanner SKU",
			fields: fields{
				rule: "SKU",
				target: &armcontainerregistry.Registry{
					SKU: &armcontainerregistry.SKU{
						Name: getSKUName(),
					},
				},
				scanContext:         &scanners.ScanContext{},
				diagnosticsSettings: scanners.DiagnosticsSettings{},
			},
			want: want{
				broken: false,
				result: "Standard",
			},
		},
		{
			name: "ContainerRegistryScanner CAF",
			fields: fields{
				rule: "CAF",
				target: &armcontainerregistry.Registry{
					Name: to.StringPtr("cr-test"),
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
			name: "ContainerRegistryScanner AnonymousPullEnabled not present",
			fields: fields{
				rule: "cr-007",
				target: &armcontainerregistry.Registry{
					Properties: &armcontainerregistry.RegistryProperties{},
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
			name: "ContainerRegistryScanner AnonymousPull Disabled",
			fields: fields{
				rule: "cr-007",
				target: &armcontainerregistry.Registry{
					Properties: &armcontainerregistry.RegistryProperties{
						AnonymousPullEnabled: to.BoolPtr(false),
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
			name: "ContainerRegistryScanner AdminUserEnabled not present",
			fields: fields{
				rule: "cr-008",
				target: &armcontainerregistry.Registry{
					Properties: &armcontainerregistry.RegistryProperties{},
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
			name: "ContainerRegistryScanner AdminUser Disabled",
			fields: fields{
				rule: "cr-008",
				target: &armcontainerregistry.Registry{
					Properties: &armcontainerregistry.RegistryProperties{
						AdminUserEnabled: to.BoolPtr(false),
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
			name: "ContainerRegistryScanner Policies not present",
			fields: fields{
				rule: "cr-010",
				target: &armcontainerregistry.Registry{
					Properties: &armcontainerregistry.RegistryProperties{},
				},
				scanContext:         &scanners.ScanContext{},
				diagnosticsSettings: scanners.DiagnosticsSettings{},
			},
			want: want{
				broken: true,
				result: "",
			},
		},
		{
			name: "ContainerRegistryScanner Retention Policies disabled",
			fields: fields{
				rule: "cr-010",
				target: &armcontainerregistry.Registry{
					Properties: &armcontainerregistry.RegistryProperties{
						Policies: &armcontainerregistry.Policies{
							RetentionPolicy: &armcontainerregistry.RetentionPolicy{
								Status: getPolicyStatusDisabled(),
							},
						},
					},
				},
				scanContext:         &scanners.ScanContext{},
				diagnosticsSettings: scanners.DiagnosticsSettings{},
			},
			want: want{
				broken: true,
				result: "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ContainerRegistryScanner{
				diagnosticsSettings: tt.fields.diagnosticsSettings,
			}
			rules := s.GetRules()
			b, w := rules[tt.fields.rule].Eval(tt.fields.target, tt.fields.scanContext)
			got := want{
				broken: b,
				result: w,
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ContainerRegistryScanner Rule.Eval() = %v, want %v", got, tt.want)
			}
		})
	}
}

func getZoneRedundancy() *armcontainerregistry.ZoneRedundancy {
	s := armcontainerregistry.ZoneRedundancyEnabled
	return &s
}

func getSKUName() *armcontainerregistry.SKUName {
	s := armcontainerregistry.SKUNameStandard
	return &s
}

func getPolicyStatusDisabled() *armcontainerregistry.PolicyStatus {
	s := armcontainerregistry.PolicyStatusDisabled
	return &s
}
