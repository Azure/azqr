package scanners

import (
	"context"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/managementgroups/armmanagementgroups"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/subscription/armsubscription"
	"github.com/rs/zerolog/log"
)

type ManagementGroupsScanner struct{}

func (sc ManagementGroupsScanner) ListSubscriptions(ctx context.Context, cred azcore.TokenCredential, groups []string, filters *models.Filters, options *arm.ClientOptions) map[string]string {
	client, err := armmanagementgroups.NewClientFactory(cred, options)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create management groups client")
	}
	result := map[string]string{}

	for _, group := range groups {
		resultPager := client.NewManagementGroupSubscriptionsClient().NewGetSubscriptionsUnderManagementGroupPager(group, nil)

		subscriptions := make([]*armmanagementgroups.SubscriptionUnderManagementGroup, 0)
		for resultPager.More() {
			pageResp, err := resultPager.NextPage(ctx)
			if err != nil {
				log.Fatal().Err(err).Msg("Failed to list management group subscriptions")
			}

			for _, s := range pageResp.Value {
				if s.Properties.State != to.Ptr(string(armsubscription.SubscriptionStateDisabled)) &&
					s.Properties.State != to.Ptr(string(armsubscription.SubscriptionStateDeleted)) {
					subscriptions = append(subscriptions, s)
				}
			}
		}

		for _, s := range subscriptions {
			sid := *s.Name
			if filters.Azqr.IsSubscriptionExcluded(sid) {
				log.Info().Msgf("Skipping subscriptions/...%s", sid[29:])
				continue
			}
			result[sid] = *s.Properties.DisplayName
		}
	}

	return result
}
