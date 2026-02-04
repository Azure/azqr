// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package models

import (
	"testing"
)

func TestParseAndValidateStageParams(t *testing.T) {
	tests := []struct {
		name    string
		params  []string
		want    map[string]map[string]any
		wantErr bool
	}{
		{
			name:   "valid cost previousMonth true",
			params: []string{"cost.previousMonth=true"},
			want: map[string]map[string]any{
				"cost": {"previousMonth": true},
			},
		},
		{
			name:   "valid cost previousMonth false",
			params: []string{"cost.previousMonth=false"},
			want: map[string]map[string]any{
				"cost": {"previousMonth": false},
			},
		},
		{
			name:   "multiple params",
			params: []string{"cost.previousMonth=true", "cost.previousMonth=false"},
			want: map[string]map[string]any{
				"cost": {"previousMonth": false},
			},
		},
		{
			name:    "unknown stage",
			params:  []string{"unknown.key=value"},
			wantErr: true,
		},
		{
			name:    "unknown key for stage",
			params:  []string{"cost.unknownKey=value"},
			wantErr: true,
		},
		{
			name:    "type mismatch bool",
			params:  []string{"cost.previousMonth=notabool"},
			wantErr: true,
		},
		{
			name:    "missing equals",
			params:  []string{"cost.previousMonth"},
			wantErr: true,
		},
		{
			name:    "missing dot",
			params:  []string{"costpreviousMonth=true"},
			wantErr: true,
		},
		{
			name:   "empty params ignored",
			params: []string{"", "  ", "cost.previousMonth=true"},
			want: map[string]map[string]any{
				"cost": {"previousMonth": true},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseAndValidateStageParams(tt.params)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseAndValidateStageParams() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				return
			}

			if len(got) != len(tt.want) {
				t.Fatalf("got %d stages, want %d", len(got), len(tt.want))
			}

			for stage, wantOpts := range tt.want {
				gotOpts, ok := got[stage]
				if !ok {
					t.Fatalf("stage %q not found in result", stage)
				}

				for key, wantVal := range wantOpts {
					gotVal, ok := gotOpts[key]
					if !ok {
						t.Fatalf("key %q not found for stage %q", key, stage)
					}

					if gotVal != wantVal {
						t.Fatalf("stage %q key %q: got %v (%T), want %v (%T)", stage, key, gotVal, gotVal, wantVal, wantVal)
					}
				}
			}
		})
	}
}

func TestApplyStageParams(t *testing.T) {
	tests := []struct {
		name      string
		params    []string
		wantPrevM bool
		wantErr   bool
	}{
		{
			name:      "default (no params)",
			params:    []string{},
			wantPrevM: false,
		},
		{
			name:      "previousMonth true",
			params:    []string{"cost.previousMonth=true"},
			wantPrevM: true,
		},
		{
			name:      "previousMonth false",
			params:    []string{"cost.previousMonth=false"},
			wantPrevM: false,
		},
		{
			name:      "invalid param",
			params:    []string{"unknown.key=value"},
			wantErr:   true,
			wantPrevM: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configs := NewStageConfigs()
			err := configs.ApplyStageParams(tt.params)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ApplyStageParams() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				return
			}

			costOpts := configs.GetStageOptions(StageNameCost)
			previousMonth := false
			if costOpts != nil {
				if val, ok := costOpts["previousMonth"]; ok {
					if b, ok := val.(bool); ok {
						previousMonth = b
					}
				}
			}

			if previousMonth != tt.wantPrevM {
				t.Fatalf("previousMonth = %v, want %v", previousMonth, tt.wantPrevM)
			}
		})
	}
}
