package scanners

import (
	"context"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/subscription/armsubscription"
	"github.com/rs/zerolog/log"
)

type SubcriptionScanner struct{}

func (sc SubcriptionScanner) ListSubscriptions(ctx context.Context, cred azcore.TokenCredential, subscriptionID string, filters *models.Filters, options *arm.ClientOptions) map[string]string {
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
		sid := *s.SubscriptionID
		// If subscriptionID is empty run the filter on all subscriptions.
		// If SubscriptionID is not empty exlude all subscriptions except the one specified.
		if subscriptionID == "" || sid == subscriptionID {
			if filters.Azqr.IsSubscriptionExcluded(sid) {
				log.Info().Msgf("Skipping subscriptions/...%s", sid[29:])
				continue
			}
			result[sid] = *s.DisplayName
		}
	}

	return result
}
