// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package pipeline

import (
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
		managementGroupScanner := scanners.ManagementGroupsScanner{}
		ctx.Subscriptions = managementGroupScanner.ListSubscriptions(
			ctx.Ctx,
			ctx.Cred,
			params.ManagementGroups,
			params.Filters,
			ctx.ClientOptions,
		)
	} else {
		subscriptionScanner := scanners.SubcriptionScanner{}
		ctx.Subscriptions = subscriptionScanner.ListSubscriptions(
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

	return nil
}
