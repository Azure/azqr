// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package quota

import (
	"strings"
	"testing"
)

func TestSQLQuotaFiltering(t *testing.T) {
	shouldKeep := func(item usageItem) bool {
		for suffix := range sqlSkipList {
			if strings.HasSuffix(item.Name.Value, suffix) {
				return false
			}
		}
		return item.Limit > 0
	}

	tests := []struct {
		name string
		item usageItem
		want bool
	}{
		{
			name: "keeps servers quota",
			item: usageItem{Name: usageName{Value: "Servers"}, Limit: 10},
			want: true,
		},
		{
			name: "keeps elastic pools quota",
			item: usageItem{Name: usageName{Value: "ElasticPools"}, Limit: 20},
			want: true,
		},
		{
			name: "keeps DTUs quota",
			item: usageItem{Name: usageName{Value: "DTUs"}, Limit: 100},
			want: true,
		},
		{
			name: "keeps vCores quota",
			item: usageItem{Name: usageName{Value: "vCores"}, Limit: 64},
			want: true,
		},
		{
			name: "skips per-server sub-limit",
			item: usageItem{Name: usageName{Value: "vCoresPerServer"}, Limit: 64},
			want: false,
		},
		{
			name: "skips per-database sub-limit",
			item: usageItem{Name: usageName{Value: "DTUsPerDatabase"}, Limit: 1600},
			want: false,
		},
		{
			name: "skips zero-limit entries",
			item: usageItem{Name: usageName{Value: "Servers"}, Limit: 0},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shouldKeep(tt.item)
			if got != tt.want {
				t.Errorf("shouldKeep(%q) = %t, want %t", tt.item.Name.Value, got, tt.want)
			}
		})
	}
}
