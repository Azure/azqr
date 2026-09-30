// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package aigov

import (
	"bytes"
	"context"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/Azure/azqr/internal/scanners/plugins/aigov/internal/testutil"
	"github.com/Azure/azqr/internal/scanners/plugins/aigov/metrics"
	"github.com/Azure/azqr/internal/scanners/plugins/aigov/modelinfo"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func TestIsStandardTier(t *testing.T) {
	for tier, want := range map[string]bool{"standard": true, metrics.StatusUnknown: true, "priority": false, "flex": false} {
		if isStandardTier(tier) != want {
			t.Errorf("isStandardTier(%q) != %v", tier, want)
		}
	}
}

func TestPriceBookLookup_FailedFetchIsLoggedAndCached(t *testing.T) {
	var calls int32
	client := testutil.NewPricingClient(t, func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddInt32(&calls, 1)
		http.Error(w, "boom", http.StatusBadRequest)
	})

	var buf bytes.Buffer
	prev := log.Logger
	log.Logger = zerolog.New(&buf).Level(zerolog.DebugLevel)
	t.Cleanup(func() { log.Logger = prev })

	pb := newPriceBook(client)
	if v := pb.lookup(context.Background(), "eastus", "gpt-4o", "", modelinfo.ModeGlobal, "short"); v.ok {
		t.Fatal("expected no price when the retail fetch fails")
	}
	if !strings.Contains(buf.String(), "Retail prices unavailable") || !strings.Contains(buf.String(), `"region":"eastus"`) {
		t.Errorf("expected debug log naming the region, got %q", buf.String())
	}

	afterFirst := atomic.LoadInt32(&calls)
	if afterFirst == 0 {
		t.Fatal("expected at least one retail request")
	}
	if v := pb.lookup(context.Background(), "eastus", "gpt-4.1", "", modelinfo.ModeGlobal, "short"); v.ok {
		t.Fatal("expected no price for a region whose fetch failed")
	}
	if got := atomic.LoadInt32(&calls); got != afterFirst {
		t.Errorf("failed region was refetched: %d requests, want %d", got, afterFirst)
	}
}
