// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package quota

import "testing"

func TestKeepAppServiceUsage(t *testing.T) {
	tests := []struct {
		name string
		item usageItem
		want bool
	}{
		{
			name: "keeps dedicated worker quota",
			item: usageItem{
				Limit: 10,
				Name:  usageName{Value: "DedicatedWorkers"},
			},
			want: true,
		},
		{
			name: "keeps sku capacity quota",
			item: usageItem{
				Limit: 25,
				Name:  usageName{Value: "sku-p1v3"},
			},
			want: true,
		},
		{
			name: "skips custom domains per app",
			item: usageItem{
				Limit: 500,
				Name:  usageName{Value: "CustomDomains"},
			},
			want: false,
		},
		{
			name: "skips ssl connections per app",
			item: usageItem{
				Limit: 1000,
				Name:  usageName{Value: "SslConnections"},
			},
			want: false,
		},
		{
			name: "skips subscription certificate inventory",
			item: usageItem{
				Limit: 100,
				Name:  usageName{Value: "Certificates"},
			},
			want: false,
		},
		{
			name: "keeps unlimited entry (limit=0 filtered by fetchUsages, not keepFn)",
			item: usageItem{
				Limit: 0,
				Name:  usageName{Value: "Sites"},
			},
			want: true, // keepFn passes it; fetchUsages will drop it with a debug log
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := keepAppServiceUsage(tt.item); got != tt.want {
				t.Errorf("keepAppServiceUsage() = %v, want %v", got, tt.want)
			}
		})
	}
}
