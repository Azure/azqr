// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package models

import "testing"

func boolPtr(b bool) *bool {
	return &b
}

func TestNewScanParamsWithDefaultsXlsxJsonDefaultOff(t *testing.T) {
	tests := []struct {
		name     string
		args     ScanArgs
		wantXlsx bool
		wantJSON bool
	}{
		{
			name:     "nil Xlsx and Json default to false",
			args:     ScanArgs{},
			wantXlsx: false,
			wantJSON: false,
		},
		{
			name:     "explicit true is honored",
			args:     ScanArgs{Xlsx: boolPtr(true), Json: boolPtr(true)},
			wantXlsx: true,
			wantJSON: true,
		},
		{
			name:     "explicit false stays false",
			args:     ScanArgs{Xlsx: boolPtr(false), Json: boolPtr(false)},
			wantXlsx: false,
			wantJSON: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewScanParamsWithDefaults(tt.args)
			if got.Xlsx != tt.wantXlsx {
				t.Errorf("NewScanParamsWithDefaults().Xlsx = %v, want %v", got.Xlsx, tt.wantXlsx)
			}
			if got.Json != tt.wantJSON {
				t.Errorf("NewScanParamsWithDefaults().Json = %v, want %v", got.Json, tt.wantJSON)
			}
		})
	}
}

func TestNewScanParamsForPluginsXlsxJsonDefaultOff(t *testing.T) {
	tests := []struct {
		name     string
		args     PluginScanArgs
		wantXlsx bool
		wantJSON bool
	}{
		{
			name:     "nil Xlsx and Json default to false",
			args:     PluginScanArgs{},
			wantXlsx: false,
			wantJSON: false,
		},
		{
			name:     "explicit true is honored",
			args:     PluginScanArgs{Xlsx: boolPtr(true), Json: boolPtr(true)},
			wantXlsx: true,
			wantJSON: true,
		},
		{
			name:     "explicit false stays false",
			args:     PluginScanArgs{Xlsx: boolPtr(false), Json: boolPtr(false)},
			wantXlsx: false,
			wantJSON: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewScanParamsForPlugins(tt.args)
			if got.Xlsx != tt.wantXlsx {
				t.Errorf("NewScanParamsForPlugins().Xlsx = %v, want %v", got.Xlsx, tt.wantXlsx)
			}
			if got.Json != tt.wantJSON {
				t.Errorf("NewScanParamsForPlugins().Json = %v, want %v", got.Json, tt.wantJSON)
			}
		})
	}
}
