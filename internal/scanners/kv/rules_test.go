// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package kv

import (
	"reflect"
	"testing"

	"github.com/Azure/azqr/internal/azqr"
	"github.com/Azure/azqr/internal/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/keyvault/armkeyvault"
)

func TestKeyVaultScanner_Rules(t *testing.T) {
	type fields struct {
		rule        string
		target      interface{}
		scanContext *azqr.ScanContext
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
			name: "KeyVaultScanner DiagnosticSettings",
			fields: fields{
				rule: "kv-001",
				target: &armkeyvault.Vault{
					ID: to.Ptr("test"),
				},
				scanContext: &azqr.ScanContext{
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
			name: "KeyVaultScanner SLA",
			fields: fields{
				rule:        "kv-003",
				target:      &armkeyvault.Vault{},
				scanContext: &azqr.ScanContext{},
			},
			want: want{
				broken: false,
				result: "99.99%",
			},
		},
		{
			name: "KeyVaultScanner CAF",
			fields: fields{
				rule: "kv-006",
				target: &armkeyvault.Vault{
					Name: to.Ptr("kv-test"),
				},
				scanContext: &azqr.ScanContext{},
			},
			want: want{
				broken: false,
				result: "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &KeyVaultScanner{}
			rules := s.GetRecommendations()
			b, w := rules[tt.fields.rule].Eval(tt.fields.target, tt.fields.scanContext)
			got := want{
				broken: b,
				result: w,
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("KeyVaultScanner Rule.Eval() = %v, want %v", got, tt.want)
			}
		})
	}
}
