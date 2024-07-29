// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package st

import (
	"reflect"
	"testing"

	"github.com/Azure/azqr/internal/azqr"
	"github.com/Azure/azqr/internal/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage"
)

func TestStorageScanner_Rules(t *testing.T) {
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
			name: "StorageScanner DiagnosticSettings",
			fields: fields{
				rule: "st-001",
				target: &armstorage.Account{
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
			name: "StorageScanner SLA 99.9%",
			fields: fields{
				rule: "st-003",
				target: &armstorage.Account{
					SKU: &armstorage.SKU{
						Name: to.Ptr(armstorage.SKUNamePremiumZRS),
					},
					Properties: &armstorage.AccountProperties{
						AccessTier: to.Ptr(armstorage.AccessTierHot),
					},
				},
				scanContext: &azqr.ScanContext{},
			},
			want: want{
				broken: false,
				result: "99.9%",
			},
		},
		{
			name: "StorageScanner SKU",
			fields: fields{
				rule: "st-005",
				target: &armstorage.Account{
					SKU: &armstorage.SKU{
						Name: to.Ptr(armstorage.SKUNamePremiumZRS),
					},
				},
				scanContext: &azqr.ScanContext{},
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
					Name: to.Ptr("sttest"),
				},
				scanContext: &azqr.ScanContext{},
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
						EnableHTTPSTrafficOnly: to.Ptr(true),
					},
				},
				scanContext: &azqr.ScanContext{},
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
						MinimumTLSVersion: to.Ptr(armstorage.MinimumTLSVersionTLS12),
					},
				},
				scanContext: &azqr.ScanContext{},
			},
			want: want{
				broken: false,
				result: "",
			},
		},
		{
			name: "StorageScanner inmutable storage versioning disabled",
			fields: fields{
				rule: "st-010",
				target: &armstorage.Account{
					Properties: &armstorage.AccountProperties{
						ImmutableStorageWithVersioning: &armstorage.ImmutableStorageAccount{
							Enabled: to.Ptr(false),
						},
					},
				},
				scanContext: &azqr.ScanContext{},
			},
			want: want{
				broken: true,
				result: "",
			},
		},
		{
			name: "StorageScanner inmutable storage versioning enabled",
			fields: fields{
				rule: "st-010",
				target: &armstorage.Account{
					Properties: &armstorage.AccountProperties{
						ImmutableStorageWithVersioning: &armstorage.ImmutableStorageAccount{
							Enabled: to.Ptr(true),
						},
					},
				},
				scanContext: &azqr.ScanContext{},
			},
			want: want{
				broken: false,
				result: "",
			},
		},
		{
			name: "StorageScanner inmutable storage versioning enabled",
			fields: fields{
				rule: "st-011",
				target: &armstorage.Account{
					Properties: &armstorage.AccountProperties{},
				},
				scanContext: &azqr.ScanContext{
					BlobServiceProperties: &armstorage.BlobServicesClientGetServicePropertiesResponse{
						BlobServiceProperties: armstorage.BlobServiceProperties{
							BlobServiceProperties: &armstorage.BlobServicePropertiesProperties{
								ContainerDeleteRetentionPolicy: &armstorage.DeleteRetentionPolicy{
									Enabled: to.Ptr(true),
								},
							},
						},
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
			s := &StorageScanner{}
			rules := s.GetRecommendations()
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
