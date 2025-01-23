// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package synw

import (
	"reflect"
	"testing"

	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azqr/internal/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/synapse/armsynapse"
)

func TestSynapseWorkspaceScanner_WorkspaceRules(t *testing.T) {
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
			name: "SynapseWorkspaceScanner DiagnosticSettings",
			fields: fields{
				rule: "synw-001",
				target: &armsynapse.Workspace{
					ID: to.Ptr("test"),
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
			name: "SynapseWorkspaceScanner Private Endpoint",
			fields: fields{
				rule: "synw-002",
				target: &armsynapse.Workspace{
					Properties: &armsynapse.WorkspaceProperties{
						PrivateEndpointConnections: []*armsynapse.PrivateEndpointConnection{
							{
								ID: to.Ptr("test"),
							},
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
			name: "SynapseWorkspaceScanner SLA",
			fields: fields{
				rule:        "synw-003",
				target:      &armsynapse.Workspace{},
				scanContext: &scanners.ScanContext{},
			},
			want: want{
				broken: false,
				result: "99.9%",
			},
		},
		{
			name: "SynapseWorkspaceScanner CAF",
			fields: fields{
				rule: "synw-004",
				target: &armsynapse.Workspace{
					Name: to.Ptr("synw-test"),
				},
				scanContext: &scanners.ScanContext{},
			},
			want: want{
				broken: false,
				result: "",
			},
		},
		{
			name: "SynapseWorkspaceScanner ManagedVirtualNetwork default",
			fields: fields{
				rule: "synw-006",
				target: &armsynapse.Workspace{
					Name: to.Ptr("synw-test"),
					Properties: &armsynapse.WorkspaceProperties{
						ManagedVirtualNetwork: to.Ptr("default"),
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
			name: "SynapseWorkspaceScanner ManagedVirtualNetwork null",
			fields: fields{
				rule: "synw-006",
				target: &armsynapse.Workspace{
					Name: to.Ptr("synw-test"),
					Properties: &armsynapse.WorkspaceProperties{
						ManagedVirtualNetwork: nil,
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
			name: "SynapseWorkspaceScanner Security Profile",
			fields: fields{
				rule: "synw-007",
				target: &armsynapse.Workspace{
					Name: to.Ptr("synw-test"),
					Properties: &armsynapse.WorkspaceProperties{
						PublicNetworkAccess: &armsynapse.PossibleWorkspacePublicNetworkAccessValues()[0],
					},
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
			s := &SynapseWorkspaceScanner{}
			rules := s.getWorkspaceRules()
			b, w := rules[tt.fields.rule].Eval(tt.fields.target, tt.fields.scanContext)
			got := want{
				broken: b,
				result: w,
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SynapseWorkspaceScanner Rule.Eval() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSynapseWorkspaceScanner_SparkPoolRules(t *testing.T) {
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
			name: "SynapseSparkPoolScanner SLA",
			fields: fields{
				rule:        "synsp-002",
				target:      &armsynapse.Workspace{},
				scanContext: &scanners.ScanContext{},
			},
			want: want{
				broken: false,
				result: "99.9%",
			},
		},
		{
			name: "SynapseSparkPoolScanner CAF",
			fields: fields{
				rule: "synsp-001",
				target: &armsynapse.BigDataPoolResourceInfo{
					Name: to.Ptr("synsp-test"),
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
			s := &SynapseWorkspaceScanner{}
			rules := s.getSparkPoolRules()
			b, w := rules[tt.fields.rule].Eval(tt.fields.target, tt.fields.scanContext)
			got := want{
				broken: b,
				result: w,
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SynapseSparkPoolScanner Rule.Eval() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSynapseWorkspaceScanner_SqlPoolRules(t *testing.T) {
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
			name: "SynapseSqlPoolScanner SLA",
			fields: fields{
				rule:        "syndp-002",
				target:      &armsynapse.Workspace{},
				scanContext: &scanners.ScanContext{},
			},
			want: want{
				broken: false,
				result: "99.9%",
			},
		},
		{
			name: "SynapseSqlPoolScanner CAF",
			fields: fields{
				rule: "syndp-001",
				target: &armsynapse.SQLPool{
					Name: to.Ptr("syndp-test"),
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
			s := &SynapseWorkspaceScanner{}
			rules := s.getSqlPoolRules()
			b, w := rules[tt.fields.rule].Eval(tt.fields.target, tt.fields.scanContext)
			got := want{
				broken: b,
				result: w,
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SynapseSqlPoolScanner Rule.Eval() = %v, want %v", got, tt.want)
			}
		})
	}
}
