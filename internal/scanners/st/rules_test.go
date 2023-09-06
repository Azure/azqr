// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package st

import (
	"reflect"
	"testing"

	"github.com/Azure/azqr/internal/ref"
	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage"
)

func TestStorageScanner_Rules(t *testing.T) {
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
			name: "StorageScanner DiagnosticSettings",
			fields: fields{
				rule: "st-001",
				target: &armstorage.Account{
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
			name: "StorageScanner Availability Zones",
			fields: fields{
				rule: "st-002",
				target: &armstorage.Account{
					SKU: &armstorage.SKU{
						Name: getPremiumZRSSKU(),
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
			name: "StorageScanner SLA 99.9%",
			fields: fields{
				rule: "st-003",
				target: &armstorage.Account{
					SKU: &armstorage.SKU{
						Name: getPremiumZRSSKU(),
					},
					Properties: &armstorage.AccountProperties{
						AccessTier: getHotTier(),
					},
				},
				scanContext: &scanners.ScanContext{},
			},
			want: want{
				broken: false,
				result: "99.9%",
			},
		},
		{
			name: "StorageScanner Private Endpoint",
			fields: fields{
				rule: "st-004",
				target: &armstorage.Account{
					Properties: &armstorage.AccountProperties{
						PrivateEndpointConnections: []*armstorage.PrivateEndpointConnection{
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
			name: "StorageScanner SKU",
			fields: fields{
				rule: "st-005",
				target: &armstorage.Account{
					SKU: &armstorage.SKU{
						Name: getPremiumZRSSKU(),
					},
				},
				scanContext: &scanners.ScanContext{},
			},
			want: want{
				broken: false,
				result: "Premium_ZRS",
			},
		},
		{
			name: "StorageScanner CAF",
			fields: fields{
				rule: "st-006",
				target: &armstorage.Account{
					Name: ref.Of("sttest"),
				},
				scanContext: &scanners.ScanContext{},
			},
			want: want{
				broken: false,
				result: "",
			},
		},
		{
			name: "StorageScanner HTTPS only",
			fields: fields{
				rule: "st-007",
				target: &armstorage.Account{
					Properties: &armstorage.AccountProperties{
						EnableHTTPSTrafficOnly: ref.Of(true),
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
			name: "StorageScanner minimum TLS version",
			fields: fields{
				rule: "st-009",
				target: &armstorage.Account{
					Properties: &armstorage.AccountProperties{
						MinimumTLSVersion: getTLSVersion(),
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
			s := &StorageScanner{}
			rules := s.GetRules()
			b, w := rules[tt.fields.rule].Eval(tt.fields.target, tt.fields.scanContext)
			got := want{
				broken: b,
				result: w,
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("StorageScanner Rule.Eval() = %v, want %v", got, tt.want)
			}
		})
	}
}

func getPremiumZRSSKU() *armstorage.SKUName {
	s := armstorage.SKUNamePremiumZRS
	return &s
}

func getHotTier() *armstorage.AccessTier {
	s := armstorage.AccessTierHot
	return &s
}

func getTLSVersion() *armstorage.MinimumTLSVersion {
	s := armstorage.MinimumTLSVersionTLS12
	return &s
}
