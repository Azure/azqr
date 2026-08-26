// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package pipeline

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"strings"

	"github.com/Azure/azqr/internal/scanners"
	"github.com/rs/zerolog/log"
)

// SubscriptionDiscoveryStage discovers subscriptions to scan.
type SubscriptionDiscoveryStage struct {
	*BaseStage
}

func NewSubscriptionDiscoveryStage() *SubscriptionDiscoveryStage {
	return &SubscriptionDiscoveryStage{
		BaseStage: NewBaseStage("Subscription Discovery", true),
	}
}

func (s *SubscriptionDiscoveryStage) Execute(ctx *ScanContext) error {
	params := ctx.Params

	if len(params.ManagementGroups) > 0 {
		scanner := scanners.ManagementGroupDiscovery{}
		ctx.Subscriptions = scanner.ListSubscriptions(
			ctx.Ctx,
			ctx.Cred,
			params.ManagementGroups,
			params.Filters,
			ctx.ClientOptions,
		)
	} else {
		scanner := scanners.SubcriptionDiscovery{}
		ctx.Subscriptions = scanner.ListSubscriptions(
			ctx.Ctx,
			ctx.Cred,
			params.Subscriptions,
			params.Filters,
			ctx.ClientOptions,
		)
	}

	log.Info().
		Int("subscriptions", len(ctx.Subscriptions)).
		Msg("Discovered subscriptions")

	ids := make([]string, 0, len(ctx.Subscriptions))
	for id := range ctx.Subscriptions {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	sum := sha256.Sum256([]byte(strings.Join(ids, ",")))
	ctx.ReportData.ScopeID = hex.EncodeToString(sum[:])

	return nil
}
