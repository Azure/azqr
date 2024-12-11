// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package cae

import (
	"reflect"
	"testing"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appcontainers/armappcontainers/v2"
)

func TestContainerAppsEnvironmentScanner_Rules(t *testing.T) {
	type fields struct {
		rule        string
		target      interface{}
		scanContext *models.ScanContext
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
			name: "ContainerAppsEnvironmentScanner DiagnosticSettings",
			fields: fields{
				rule: "cae-001",
				target: &armappcontainers.ManagedEnvironment{
					ID: to.Ptr("test"),
				},
				scanContext: &models.ScanContext{
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
			name: "ContainerAppsEnvironmentScanner SLA",
			fields: fields{
				rule:        "cae-003",
				target:      &armappcontainers.ManagedEnvironment{},
				scanContext: &models.ScanContext{},
			},
			want: want{
				broken: false,
				result: "99.95%",
			},
		},
		{
			name: "ContainerAppsEnvironmentScanner VnetConfiguration not present",
			fields: fields{
				rule: "cae-004",
				target: &armappcontainers.ManagedEnvironment{
					Properties: &armappcontainers.ManagedEnvironmentProperties{
						VnetConfiguration: nil,
					},
				},
				scanContext: &models.ScanContext{},
			},
			want: want{
				broken: true,
				result: "",
			},
		},
		{
			name: "ContainerAppsEnvironmentScanner VnetConfiguration Internal",
			fields: fields{
				rule: "cae-004",
				target: &armappcontainers.ManagedEnvironment{
					Properties: &armappcontainers.ManagedEnvironmentProperties{
						VnetConfiguration: &armappcontainers.VnetConfiguration{
							Internal: to.Ptr(true),
						},
					},
				},
				scanContext: &models.ScanContext{},
			},
			want: want{
				broken: false,
				result: "",
			},
		},
		{
			name: "ContainerAppsEnvironmentScanner CAF",
			fields: fields{
				rule: "cae-006",
				target: &armappcontainers.ManagedEnvironment{
					Name: to.Ptr("cae-test"),
				},
				scanContext: &models.ScanContext{},
			},
			want: want{
				broken: false,
				result: "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ContainerAppsEnvironmentScanner{}
			rules := s.GetRecommendations()
			b, w := rules[tt.fields.rule].Eval(tt.fields.target, tt.fields.scanContext)
			got := want{
				broken: b,
				result: w,
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ContainerAppsEnvironmentScanner Rule.Eval() = %v, want %v", got, tt.want)
			}
		})
	}
}
