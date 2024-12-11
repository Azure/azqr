// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package agw

import (
	"reflect"
	"testing"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v6"
)

func TestApplicationGatewayScanner_Rules(t *testing.T) {
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
			name: "ApplicationGatewayScanner DiagnosticSettings",
			fields: fields{
				rule: "agw-005",
				target: &armnetwork.ApplicationGateway{
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
			name: "ApplicationGatewayScanner SLA",
			fields: fields{
				rule:        "agw-103",
				target:      &armnetwork.ApplicationGateway{},
				scanContext: &models.ScanContext{},
			},
			want: want{
				broken: false,
				result: "99.95%",
			},
		},
		{
			name: "ApplicationGatewayScanner CAF",
			fields: fields{
				rule: "agw-105",
				target: &armnetwork.ApplicationGateway{
					Name: to.Ptr("agw-test"),
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
			s := &ApplicationGatewayScanner{}
			rules := s.GetRecommendations()
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
