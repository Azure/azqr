// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package skus

import "testing"

func TestLookup_KnownSKU(t *testing.T) {
	got, ok := Lookup("Standard_D4s_v5")
	if !ok {
		t.Fatal("Lookup(Standard_D4s_v5) not found")
	}
	if got.VCPUs != 4 {
		t.Errorf("Lookup(Standard_D4s_v5).VCPUs = %d, want 4", got.VCPUs)
	}
}

func TestLookup_UnknownSKU(t *testing.T) {
	_, ok := Lookup("does-not-exist")
	if ok {
		t.Error("Lookup(does-not-exist) should return false")
	}
}

func TestLookup_EmptyString(t *testing.T) {
	_, ok := Lookup("")
	if ok {
		t.Error("Lookup(\"\") should return false")
	}
}
