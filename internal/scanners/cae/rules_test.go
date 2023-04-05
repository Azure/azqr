// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package cae

import (
	"reflect"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appcontainers/armappcontainers"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/cmendible/azqr/internal/scanners"
)

func TestContainerAppsScanner_Rules(t *testing.T) {
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
			name: "ContainerAppsScanner DiagnosticSettings",
			fields: fields{
				rule: "DiagnosticSettings",
				target: &armappcontainers.ManagedEnvironment{
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
			name: "ContainerAppsScanner Availability Zones",
			fields: fields{
				rule: "AvailabilityZones",
				target: &armappcontainers.ManagedEnvironment{
					Properties: &armappcontainers.ManagedEnvironmentProperties{
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
			name: "ContainerAppsScanner SLA",
			fields: fields{
				rule:                "SLA",
				target:              &armappcontainers.ManagedEnvironment{},
				scanContext:         &scanners.ScanContext{},
				diagnosticsSettings: scanners.DiagnosticsSettings{},
			},
			want: want{
				broken: false,
				result: "99.95%",
			},
		},
		{
			name: "ContainerAppsScanner VnetConfiguration not present",
			fields: fields{
				rule: "Private",
				target: &armappcontainers.ManagedEnvironment{
					Properties: &armappcontainers.ManagedEnvironmentProperties{
						VnetConfiguration: nil,
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
		{
			name: "ContainerAppsScanner VnetConfiguration Internal",
			fields: fields{
				rule: "Private",
				target: &armappcontainers.ManagedEnvironment{
					Properties: &armappcontainers.ManagedEnvironmentProperties{
						VnetConfiguration: &armappcontainers.VnetConfiguration{
							Internal: to.BoolPtr(true),
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
			name: "ContainerAppsScanner CAF",
			fields: fields{
				rule: "CAF",
				target: &armappcontainers.ManagedEnvironment{
					Name: to.StringPtr("cae-test"),
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
			s := &ContainerAppsScanner{
				diagnosticsSettings: tt.fields.diagnosticsSettings,
			}
			rules := s.GetRules()
			b, w := rules[tt.fields.rule].Eval(tt.fields.target, tt.fields.scanContext)
			got := want{
				broken: b,
				result: w,
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ContainerAppsScanner Rule.Eval() = %v, want %v", got, tt.want)
			}
		})
	}
}
