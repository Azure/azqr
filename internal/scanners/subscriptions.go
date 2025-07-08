package scanners

import (
	"context"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/throttling"
	"github.com/Azure/azqr/internal/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/subscription/armsubscription"
	"github.com/rs/zerolog/log"
)

type SubcriptionScanner struct{}

func (sc SubcriptionScanner) ListSubscriptions(ctx context.Context, cred azcore.TokenCredential, subscriptions []string, filters *models.Filters, options *arm.ClientOptions) map[string]string {
	client, err := armsubscription.NewSubscriptionsClient(cred, options)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create subscriptions client")
	}

	resultPager := client.NewListPager(nil)

	subs := make([]*armsubscription.Subscription, 0)
	for resultPager.More() {
		// Wait for a token from the burstLimiter channel before making the request
		<-throttling.ARMLimiter
		pageResp, err := resultPager.NextPage(ctx)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to list subscriptions")
		}

		for _, s := range pageResp.Value {
			if s.State != to.Ptr(armsubscription.SubscriptionStateDisabled) &&
				s.State != to.Ptr(armsubscription.SubscriptionStateDeleted) {
				subs = append(subs, s)
			}
		}
	}

	result := map[string]string{}
	for _, s := range subs {
		sid := *s.SubscriptionID
		// If subscriptionns is empty run the filter on all subscriptions.
		// If Subscriptions is not empty exlude all subscriptions except the ones specified.
		if len(subscriptions) == 0 || containsString(subscriptions, sid) {
			if filters.Azqr.IsSubscriptionExcluded(sid) {
				log.Info().Msgf("Skipping subscriptions/...%s", sid[29:])
				continue
			}
			result[sid] = *s.DisplayName
		}
	}

	return result
}

// Chek if string is in slice
func containsString(slice []string, str string) bool {
	for _, item := range slice {
		if item == str {
			return true
		}
	}
	return false
}
