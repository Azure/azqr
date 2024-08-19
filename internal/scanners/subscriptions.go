package scanners

import (
	"context"

	"github.com/Azure/azqr/internal/azqr"
	"github.com/Azure/azqr/internal/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/subscription/armsubscription"
	"github.com/rs/zerolog/log"
)

type SubcriptionScanner struct{}

func (sc SubcriptionScanner) ListSubscriptions(ctx context.Context, cred azcore.TokenCredential, subscriptionID string, filters *azqr.Filters, options *arm.ClientOptions) map[string]string {
	client, err := armsubscription.NewSubscriptionsClient(cred, options)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create subscriptions client")
	}

	resultPager := client.NewListPager(nil)

	subscriptions := make([]*armsubscription.Subscription, 0)
	for resultPager.More() {
		pageResp, err := resultPager.NextPage(ctx)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to list subscriptions")
		}

		for _, s := range pageResp.Value {
			if s.State != to.Ptr(armsubscription.SubscriptionStateDisabled) &&
				s.State != to.Ptr(armsubscription.SubscriptionStateDeleted) {
				subscriptions = append(subscriptions, s)
			}
		}
	}

	result := map[string]string{}
	for _, s := range subscriptions {
		// if subscriptionID is empty, return filtered subscriptions. Otherwise, return only the specified subscription
		sid := *s.SubscriptionID
		if subscriptionID == "" || subscriptionID == sid {
			if filters.Azqr.IsSubscriptionExcluded(sid) {
				log.Info().Msgf("Skipping subscriptions/...%s", sid[29:])
				continue
			}
			result[*s.SubscriptionID] = *s.DisplayName
		}
	}

	return result
}
