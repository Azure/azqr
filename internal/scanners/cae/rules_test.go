// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package cae

import (
	"reflect"
	"testing"

	"github.com/Azure/azqr/internal/ref"
	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appcontainers/armappcontainers"
)

func TestContainerAppsScanner_Rules(t *testing.T) {
	type fields struct {
		rule        string
		target      interface{}
		scanContext *scanners.ScanContext
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
				rule: "cae-001",
				target: &armappcontainers.ManagedEnvironment{
					ID: ref.Of("test"),
				},
				scanContext: &scanners.ScanContext{
					DiagnosticsSettings: map[string]bool{
						"test": true,
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
				rule: "cae-002",
				target: &armappcontainers.ManagedEnvironment{
					Properties: &armappcontainers.ManagedEnvironmentProperties{
						ZoneRedundant: ref.Of(true),
					},
				},
				scanContext: &scanners.ScanContext{},
			},
			want: want{
				broken: false,
				result: "",
			},
		},
		{
			name: "ContainerAppsScanner SLA",
			fields: fields{
				rule:        "cae-003",
				target:      &armappcontainers.ManagedEnvironment{},
				scanContext: &scanners.ScanContext{},
			},
			want: want{
				broken: false,
				result: "99.95%",
			},
		},
		{
			name: "ContainerAppsScanner VnetConfiguration not present",
			fields: fields{
				rule: "cae-004",
				target: &armappcontainers.ManagedEnvironment{
					Properties: &armappcontainers.ManagedEnvironmentProperties{
						VnetConfiguration: nil,
					},
				},
				scanContext: &scanners.ScanContext{},
			},
			want: want{
				broken: true,
				result: "",
			},
		},
		{
			name: "ContainerAppsScanner VnetConfiguration Internal",
			fields: fields{
				rule: "cae-004",
				target: &armappcontainers.ManagedEnvironment{
					Properties: &armappcontainers.ManagedEnvironmentProperties{
						VnetConfiguration: &armappcontainers.VnetConfiguration{
							Internal: ref.Of(true),
						},
					},
				},
				scanContext: &scanners.ScanContext{},
			},
			want: want{
				broken: false,
				result: "",
			},
		},
		{
			name: "ContainerAppsScanner CAF",
			fields: fields{
				rule: "cae-006",
				target: &armappcontainers.ManagedEnvironment{
					Name: ref.Of("cae-test"),
				},
				scanContext: &scanners.ScanContext{},
			},
			want: want{
				broken: false,
				result: "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ContainerAppsScanner{}
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
