// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package hub

import (
	"reflect"
	"testing"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/machinelearning/armmachinelearning"
)

func TestAIFoundryHubScanner_Rules(t *testing.T) {
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
			name: "AIFoundryHubScanner private enpoint enabled",
			fields: fields{
				rule: "hub-005",
				target: &armmachinelearning.Workspace{
					Properties: &armmachinelearning.WorkspaceProperties{
						PrivateEndpointConnections: []*armmachinelearning.PrivateEndpointConnection{
							{},
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
			name: "AIFoundryHubScanner no private enpoint",
			fields: fields{
				rule: "hub-005",
				target: &armmachinelearning.Workspace{
					Properties: &armmachinelearning.WorkspaceProperties{},
				},
				scanContext: &models.ScanContext{},
			},
			want: want{
				broken: true,
				result: "",
			},
		},
		{
			name: "AIFoundryHubScanner SLA",
			fields: fields{
				rule:        "hub-002",
				target:      &armmachinelearning.Workspace{},
				scanContext: &models.ScanContext{},
			},
			want: want{
				broken: false,
				result: "99.9%",
			},
		},
		{
			name: "AIFoundryHubScanner CAF (hub)",
			fields: fields{
				rule: "hub-001",
				target: &armmachinelearning.Workspace{
					Name: to.Ptr("hub-test"),
				},
				scanContext: &models.ScanContext{},
			},
			want: want{
				broken: false,
				result: "",
			},
		},
		{
			name: "AIFoundryHubScanner CAF (proj)",
			fields: fields{
				rule: "hub-001",
				target: &armmachinelearning.Workspace{
					Name: to.Ptr("proj-test"),
				},
				scanContext: &models.ScanContext{},
			},
			want: want{
				broken: false,
				result: "",
			},
		},
		{
			name: "AIFoundryHubScanner CAF (mlw)",
			fields: fields{
				rule: "hub-001",
				target: &armmachinelearning.Workspace{
					Name: to.Ptr("mlw-test"),
				},
				scanContext: &models.ScanContext{},
			},
			want: want{
				broken: false,
				result: "",
			},
		},
		{
			name: "AIFoundryHubScanner Public network enabled",
			fields: fields{
				rule: "hub-004",
				target: &armmachinelearning.Workspace{
					Properties: &armmachinelearning.WorkspaceProperties{
						PublicNetworkAccess: to.Ptr(armmachinelearning.PublicNetworkAccessEnabled),
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
			name: "AIFoundryHubScanner Public network disabled",
			fields: fields{
				rule: "hub-004",
				target: &armmachinelearning.Workspace{
					Properties: &armmachinelearning.WorkspaceProperties{
						PublicNetworkAccess: to.Ptr(armmachinelearning.PublicNetworkAccessDisabled),
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
			name: "AIFoundryHubScanner Public network nil",
			fields: fields{
				rule: "hub-004",
				target: &armmachinelearning.Workspace{
					Properties: &armmachinelearning.WorkspaceProperties{},
				},
				scanContext: &models.ScanContext{},
			},
			want: want{
				broken: true,
				result: "",
			},
		},
		{
			name: "AIFoundryHubScanner DiagnosticSettings",
			fields: fields{
				rule: "hub-006",
				target: &armmachinelearning.Workspace{
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rules := getRecommendations()
			b, w := rules[tt.fields.rule].Eval(tt.fields.target, tt.fields.scanContext)
			got := want{
				broken: b,
				result: w,
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AIFoundryHubScanner Rule.Eval() = %v, want %v", got, tt.want)
			}
		})
	}
}
