// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package apim

import (
	"reflect"
	"testing"
	"time"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/apimanagement/armapimanagement"
)

func TestAPIManagementScanner_Rules(t *testing.T) {
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
			name: "APIManagementScanner DiagnosticSettings",
			fields: fields{
				rule: "apim-001",
				target: &armapimanagement.ServiceResource{
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
			name: "APIManagementScanner SLA Developer SKU",
			fields: fields{
				rule: "apim-003",
				target: &armapimanagement.ServiceResource{
					SKU: &armapimanagement.ServiceSKUProperties{
						Name: to.Ptr(armapimanagement.SKUTypeDeveloper),
					},
					Zones: []*string{},
					Properties: &armapimanagement.ServiceProperties{
						AdditionalLocations: []*armapimanagement.AdditionalLocation{},
					},
				},
				scanContext: &models.ScanContext{},
			},
			want: want{
				broken: true,
				result: "None",
			},
		},
		{
			name: "APIManagementScanner SLA Free Premimum SKU and Availability Zones",
			fields: fields{
				rule: "apim-003",
				target: &armapimanagement.ServiceResource{
					SKU: &armapimanagement.ServiceSKUProperties{
						Name: to.Ptr(armapimanagement.SKUTypePremium),
					},
					Zones: []*string{to.Ptr("1"), to.Ptr("2"), to.Ptr("3")},
					Properties: &armapimanagement.ServiceProperties{
						AdditionalLocations: []*armapimanagement.AdditionalLocation{},
					},
				},
				scanContext: &models.ScanContext{},
			},
			want: want{
				broken: false,
				result: "99.99%",
			},
		},
		{
			name: "APIManagementScanner SLA Consumption SKU",
			fields: fields{
				rule: "apim-003",
				target: &armapimanagement.ServiceResource{
					SKU: &armapimanagement.ServiceSKUProperties{
						Name: to.Ptr(armapimanagement.SKUTypeConsumption),
					},
					Zones: []*string{},
					Properties: &armapimanagement.ServiceProperties{
						AdditionalLocations: []*armapimanagement.AdditionalLocation{},
					},
				},
				scanContext: &models.ScanContext{},
			},
			want: want{
				broken: false,
				result: "99.95%",
			},
		},
		{
			name: "APIManagementScanner Private Endpoint",
			fields: fields{
				rule: "apim-004",
				target: &armapimanagement.ServiceResource{
					Properties: &armapimanagement.ServiceProperties{
						PrivateEndpointConnections: []*armapimanagement.RemotePrivateEndpointConnectionWrapper{
							{
								ID: to.Ptr("test"),
							},
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
			name: "APIManagementScanner CAF",
			fields: fields{
				rule: "apim-006",
				target: &armapimanagement.ServiceResource{
					Name: to.Ptr("apim-test"),
				},
				scanContext: &models.ScanContext{},
			},
			want: want{
				broken: false,
				result: "",
			},
		},
		{
			name: "APIManagementScanner ManagedIdentity None",
			fields: fields{
				rule: "apim-008",
				target: &armapimanagement.ServiceResource{
					Identity: &armapimanagement.ServiceIdentity{
						Type: to.Ptr(armapimanagement.ApimIdentityTypeNone),
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
			name: "APIManagementScanner TLS 1.2",
			fields: fields{
				rule: "apim-009",
				target: &armapimanagement.ServiceResource{
					Properties: &armapimanagement.ServiceProperties{
						CustomProperties: map[string]*string{
							"Microsoft.WindowsAzure.ApiManagement.Gateway.Security.Protocols.Tls10":         to.Ptr("false"),
							"Microsoft.WindowsAzure.ApiManagement.Gateway.Security.Protocols.Tls11":         to.Ptr("false"),
							"Microsoft.WindowsAzure.ApiManagement.Gateway.Security.Protocols.Ssl30":         to.Ptr("false"),
							"Microsoft.WindowsAzure.ApiManagement.Gateway.Security.Backend.Protocols.Tls10": to.Ptr("false"),
							"Microsoft.WindowsAzure.ApiManagement.Gateway.Security.Backend.Protocols.Tls11": to.Ptr("false"),
							"Microsoft.WindowsAzure.ApiManagement.Gateway.Security.Backend.Protocols.Ssl30": to.Ptr("false"),
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
			name: "APIManagementScanner CustomProperties nil",
			fields: fields{
				rule: "apim-009",
				target: &armapimanagement.ServiceResource{
					Properties: &armapimanagement.ServiceProperties{
						CustomProperties: nil,
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
			name: "APIManagementScanner Cyphers",
			fields: fields{
				rule: "apim-010",
				target: &armapimanagement.ServiceResource{
					Properties: &armapimanagement.ServiceProperties{
						CustomProperties: map[string]*string{
							"Microsoft.WindowsAzure.ApiManagement.Gateway.Security.Ciphers.TripleDes168":                       to.Ptr("false"),
							"Microsoft.WindowsAzure.ApiManagement.Gateway.Security.Ciphers.TLS_RSA_WITH_AES_128_CBC_SHA":       to.Ptr("false"),
							"Microsoft.WindowsAzure.ApiManagement.Gateway.Security.Ciphers.TLS_RSA_WITH_AES_256_CBC_SHA":       to.Ptr("false"),
							"Microsoft.WindowsAzure.ApiManagement.Gateway.Security.Ciphers.TLS_RSA_WITH_AES_128_CBC_SHA256":    to.Ptr("false"),
							"Microsoft.WindowsAzure.ApiManagement.Gateway.Security.Ciphers.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA": to.Ptr("false"),
							"Microsoft.WindowsAzure.ApiManagement.Gateway.Security.Ciphers.TLS_RSA_WITH_AES_256_CBC_SHA256":    to.Ptr("false"),
							"Microsoft.WindowsAzure.ApiManagement.Gateway.Security.Ciphers.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA": to.Ptr("false"),
							"Microsoft.WindowsAzure.ApiManagement.Gateway.Security.Ciphers.TLS_RSA_WITH_AES_128_GCM_SHA256":    to.Ptr("false"),
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
			name: "APIManagementScanner CustomProperties nil",
			fields: fields{
				rule: "apim-010",
				target: &armapimanagement.ServiceResource{
					Properties: &armapimanagement.ServiceProperties{
						CustomProperties: nil,
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
			name: "APIManagementScanner Certificate expiring",
			fields: fields{
				rule: "apim-011",
				target: &armapimanagement.ServiceResource{
					Properties: &armapimanagement.ServiceProperties{
						HostnameConfigurations: []*armapimanagement.HostnameConfiguration{
							{
								Certificate: &armapimanagement.CertificateInformation{
									Expiry: to.Ptr(time.Now()),
								},
							},
						},
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
			name: "APIManagementScanner Certificate not expiring",
			fields: fields{
				rule: "apim-011",
				target: &armapimanagement.ServiceResource{
					Properties: &armapimanagement.ServiceProperties{
						HostnameConfigurations: []*armapimanagement.HostnameConfiguration{
							{
								Certificate: &armapimanagement.CertificateInformation{
									Expiry: to.Ptr(time.Now().Add(time.Hour * 24 * 45)),
								},
							},
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
				t.Errorf("APIManagementScanner Rule.Eval() = %v, want %v", got, tt.want)
			}
		})
	}
}
