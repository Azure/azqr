// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package cog

import (
	"reflect"
	"testing"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/cognitiveservices/armcognitiveservices"
)

func TestCognitiveScanner_Rules(t *testing.T) {
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
			name: "CognitiveScanner DiagnosticSettings",
			fields: fields{
				rule: "cog-001",
				target: &armcognitiveservices.Account{
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
			name: "CognitiveScanner SLA 99.99%",
			fields: fields{
				rule:        "cog-003",
				target:      &armcognitiveservices.Account{},
				scanContext: &models.ScanContext{},
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
			name: "CognitiveScanner OpenAi CAF",
			fields: fields{
				rule: "cog-006",
				target: &armcognitiveservices.Account{
					Name: to.Ptr("oai-test"),
					Kind: to.Ptr("OpenAi"),
				},
				scanContext: &models.ScanContext{},
			},
			want: want{
				broken: false,
				result: "",
			},
		},
		{
			name: "CognitiveScanner ContentModerator CAF",
			fields: fields{
				rule: "cog-006",
				target: &armcognitiveservices.Account{
					Name: to.Ptr("cm-test"),
					Kind: to.Ptr("ContentModerator"),
				},
				scanContext: &models.ScanContext{},
			},
			want: want{
				broken: false,
				result: "",
			},
		},
		{
			name: "CognitiveScanner ContentSafety CAF",
			fields: fields{
				rule: "cog-006",
				target: &armcognitiveservices.Account{
					Name: to.Ptr("cs-test"),
					Kind: to.Ptr("ContentSafety"),
				},
				scanContext: &models.ScanContext{},
			},
			want: want{
				broken: false,
				result: "",
			},
		},
		{
			name: "CognitiveScanner CustomVision.Prediction CAF",
			fields: fields{
				rule: "cog-006",
				target: &armcognitiveservices.Account{
					Name: to.Ptr("cstv-test"),
					Kind: to.Ptr("CustomVision.Prediction"),
				},
				scanContext: &models.ScanContext{},
			},
			want: want{
				broken: false,
				result: "",
			},
		},
		{
			name: "CognitiveScanner CustomVision.Training CAF",
			fields: fields{
				rule: "cog-006",
				target: &armcognitiveservices.Account{
					Name: to.Ptr("cstvt-test"),
					Kind: to.Ptr("CustomVision.Training"),
				},
				scanContext: &models.ScanContext{},
			},
			want: want{
				broken: false,
				result: "",
			},
		},
		{
			name: "CognitiveScanner FormRecognizer CAF",
			fields: fields{
				rule: "cog-006",
				target: &armcognitiveservices.Account{
					Name: to.Ptr("di-test"),
					Kind: to.Ptr("FormRecognizer"),
				},
				scanContext: &models.ScanContext{},
			},
			want: want{
				broken: false,
				result: "",
			},
		},
		{
			name: "CognitiveScanner Face CAF",
			fields: fields{
				rule: "cog-006",
				target: &armcognitiveservices.Account{
					Name: to.Ptr("face-test"),
					Kind: to.Ptr("Face"),
				},
				scanContext: &models.ScanContext{},
			},
			want: want{
				broken: false,
				result: "",
			},
		},
		{
			name: "CognitiveScanner HealthInsights CAF",
			fields: fields{
				rule: "cog-006",
				target: &armcognitiveservices.Account{
					Name: to.Ptr("hi-test"),
					Kind: to.Ptr("HealthInsights"),
				},
				scanContext: &models.ScanContext{},
			},
			want: want{
				broken: false,
				result: "",
			},
		},
		{
			name: "CognitiveScanner ImmersiveReader CAF",
			fields: fields{
				rule: "cog-006",
				target: &armcognitiveservices.Account{
					Name: to.Ptr("ir-test"),
					Kind: to.Ptr("ImmersiveReader"),
				},
				scanContext: &models.ScanContext{},
			},
			want: want{
				broken: false,
				result: "",
			},
		},
		{
			name: "CognitiveScanner TextAnalytics CAF",
			fields: fields{
				rule: "cog-006",
				target: &armcognitiveservices.Account{
					Name: to.Ptr("lang-test"),
					Kind: to.Ptr("TextAnalytics"),
				},
				scanContext: &models.ScanContext{},
			},
			want: want{
				broken: false,
				result: "",
			},
		},
		{
			name: "CognitiveScanner SpeechServices CAF",
			fields: fields{
				rule: "cog-006",
				target: &armcognitiveservices.Account{
					Name: to.Ptr("spch-test"),
					Kind: to.Ptr("SpeechServices"),
				},
				scanContext: &models.ScanContext{},
			},
			want: want{
				broken: false,
				result: "",
			},
		},
		{
			name: "CognitiveScanner TextTranslation CAF",
			fields: fields{
				rule: "cog-006",
				target: &armcognitiveservices.Account{
					Name: to.Ptr("trsl-test"),
					Kind: to.Ptr("TextTranslation"),
				},
				scanContext: &models.ScanContext{},
			},
			want: want{
				broken: false,
				result: "",
			},
		},
		{
			name: "CognitiveScanner ComputerVision CAF",
			fields: fields{
				rule: "cog-006",
				target: &armcognitiveservices.Account{
					Name: to.Ptr("cv-test"),
					Kind: to.Ptr("ComputerVision"),
				},
				scanContext: &models.ScanContext{},
			},
			want: want{
				broken: false,
				result: "",
			},
		},
		{
			name: "CognitiveScanner CAF",
			fields: fields{
				rule: "cog-006",
				target: &armcognitiveservices.Account{
					Name: to.Ptr("cog-test"),
					Kind: to.Ptr("cog"),
				},
				scanContext: &models.ScanContext{},
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
				scanContext: &models.ScanContext{},
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
						DisableLocalAuth: to.Ptr(true),
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
			s := &CognitiveScanner{}
			rules := s.GetRecommendations()
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
