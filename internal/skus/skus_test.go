// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package skus

import "testing"

func TestLookup_KnownSKU(t *testing.T) {
	got := Lookup("Standard_D4s_v5")
	if got != 4 {
		t.Errorf("Lookup(Standard_D4s_v5) = %d, want 4", got)
	}
}

func TestLookup_UnknownSKU(t *testing.T) {
	got := Lookup("does-not-exist")
	if got != 0 {
		t.Errorf("Lookup(does-not-exist) = %d, want 0", got)
	}
}

func TestLookup_EmptyString(t *testing.T) {
	got := Lookup("")
	if got != 0 {
		t.Errorf("Lookup(\"\") = %d, want 0", got)
	}
}
