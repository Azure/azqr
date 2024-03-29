// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package cog

import (
	"reflect"
	"testing"

	"github.com/Azure/azqr/internal/ref"
	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/cognitiveservices/armcognitiveservices"
)

func TestCognitiveScanner_Rules(t *testing.T) {
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
			name: "CognitiveScanner DiagnosticSettings",
			fields: fields{
				rule: "cog-001",
				target: &armcognitiveservices.Account{
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
			name: "CognitiveScanner SLA 99.99%",
			fields: fields{
				rule:        "cog-003",
				target:      &armcognitiveservices.Account{},
				scanContext: &scanners.ScanContext{},
			},
			want: want{
				broken: false,
				result: "99.9%",
			},
		},
		{
			name: "CognitiveScanner Private Endpoint",
			fields: fields{
				rule: "cog-004",
				target: &armcognitiveservices.Account{
					Properties: &armcognitiveservices.AccountProperties{
						PrivateEndpointConnections: []*armcognitiveservices.PrivateEndpointConnection{
							{
								ID: ref.Of("test"),
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
			name: "CognitiveScanner SKU",
			fields: fields{
				rule: "cog-005",
				target: &armcognitiveservices.Account{
					SKU: &armcognitiveservices.SKU{
						Name: ref.Of("test"),
					},
				},
				scanContext: &scanners.ScanContext{},
			},
			want: want{
				broken: false,
				result: "test",
			},
		},
		{
			name: "CognitiveScanner CAF",
			fields: fields{
				rule: "cog-006",
				target: &armcognitiveservices.Account{
					Name: ref.Of("cog-test"),
				},
				scanContext: &scanners.ScanContext{},
			},
			want: want{
				broken: false,
				result: "",
			},
		},
		{
			name: "CognitiveScanner DisableLocalAuth nil",
			fields: fields{
				rule: "cog-008",
				target: &armcognitiveservices.Account{
					Properties: &armcognitiveservices.AccountProperties{},
				},
				scanContext: &scanners.ScanContext{},
			},
			want: want{
				broken: true,
				result: "",
			},
		},
		{
			name: "CognitiveScanner DisableLocalAuth true",
			fields: fields{
				rule: "cog-008",
				target: &armcognitiveservices.Account{
					Properties: &armcognitiveservices.AccountProperties{
						DisableLocalAuth: ref.Of(true),
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
			s := &CognitiveScanner{}
			rules := s.GetRules()
			b, w := rules[tt.fields.rule].Eval(tt.fields.target, tt.fields.scanContext)
			got := want{
				broken: b,
				result: w,
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CognitiveScanner Rule.Eval() = %v, want %v", got, tt.want)
			}
		})
	}
}
