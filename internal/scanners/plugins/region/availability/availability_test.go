// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package availability

import (
	"context"
	"errors"
	"testing"

	"github.com/Azure/azqr/internal/scanners/plugins/region/types"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
)

func TestFetchMergedProviders_FailedSubscriptionDoesNotFlagTypes(t *testing.T) {
	orig := providerFetcher
	defer func() { providerFetcher = orig }()
	providerFetcher = func(_ context.Context, _ azcore.TokenCredential, subID string) (*types.ResourceTypeLocationData, error) {
		if subID == "bad" {
			return nil, errors.New("boom")
		}
		return &types.ResourceTypeLocationData{Data: map[string]map[string]map[string]struct{}{
			"microsoft.compute": {"disks": {"swedencentral": {}}},
		}}, nil
	}

	merged := FetchMergedProviders(context.Background(), nil, []string{"bad", "good"})
	if merged == nil || !merged.IsAvailable("microsoft.compute/disks", "swedencentral") {
		t.Fatal("disks should be available from the successful subscription")
	}

	if got := FetchMergedProviders(context.Background(), nil, []string{"bad"}); got != nil {
		t.Error("all fetches failed: expected nil provider data")
	}
}
