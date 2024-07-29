// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package internal

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Azure/azqr/internal/azqr"
	"github.com/Azure/azqr/internal/filters"
	"github.com/Azure/azqr/internal/renderers"
	"github.com/Azure/azqr/internal/renderers/csv"
	"github.com/Azure/azqr/internal/renderers/excel"
	"github.com/Azure/azqr/internal/renderers/json"
	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azqr/internal/to"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/subscription/armsubscription"
)

type ScanParams struct {
	SubscriptionID          string
	ResourceGroup           string
	OutputName              string
	Defender                bool
	Advisor                 bool
	Cost                    bool
	Mask                    bool
	Csv                     bool
	Debug                   bool
	ServiceScanners         []azqr.IAzureScanner
	ForceAzureCliCredential bool
	FilterFile              string
	UseAzqrRecommendations  bool
	UseAprlRecommendations  bool
}

func Scan(params *ScanParams) {
	// Default level for this example is info, unless debug flag is present
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if params.Debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		log.Debug().Msg("Debug logging enabled")
	}

	// validate input
	if params.SubscriptionID == "" && params.ResourceGroup != "" {
		log.Fatal().Msg("Resource Group name can only be used with a Subscription Id")
	}

	// generate output file name
	outputFile := generateOutputFileName(params.OutputName)

	// load filters
	filters := filters.LoadFilters(params.FilterFile)

	// create Azure credentials
	cred := newAzureCredential(params.ForceAzureCliCredential)

	// create a cancelable context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// create ARM client options
	clientOptions := &arm.ClientOptions{
		ClientOptions: policy.ClientOptions{
			Retry: policy.RetryOptions{
				RetryDelay:    20 * time.Millisecond,
				MaxRetries:    3,
				MaxRetryDelay: 10 * time.Minute,
			},
		},
	}

	// list subscriptions. Key is subscription ID, value is subscription name
	subscriptions := listSubscriptions(ctx, cred, params.SubscriptionID, filters, clientOptions)

	// initialize scanners
	defenderScanner := scanners.DefenderScanner{}
	pipScanner := scanners.PublicIPScanner{}
	peScanner := scanners.PrivateEndpointScanner{}
	diagnosticsScanner := scanners.DiagnosticSettingsScanner{}
	advisorScanner := scanners.AdvisorScanner{}
	costScanner := scanners.CostScanner{}

	// initialize report data
	reportData := renderers.ReportData{
		OutputFileName: outputFile,
		Mask:           params.Mask,
		Recomendations: map[string]map[string]azqr.AprlRecommendation{},
		AzqrData:       []azqr.AzqrServiceResult{},
		AprlData:       []azqr.AprlResult{},
		DefenderData:   []scanners.DefenderResult{},
		AdvisorData:    []scanners.AdvisorResult{},
		CostData: &scanners.CostResult{
			Items: []*scanners.CostResultItem{},
		},
		ResourceTypeCount: []azqr.ResourceTypeCount{},
	}

	// get the APRL scan results
	reportData.Recomendations, reportData.AprlData = AprlScan(ctx, cred, params, filters, subscriptions)

	// For each service scanner, get the recommendations list
	if params.UseAzqrRecommendations {
		for _, s := range params.ServiceScanners {
			for i, r := range s.GetRecommendations() {
				if reportData.Recomendations[strings.ToLower(r.ResourceType)] == nil {
					reportData.Recomendations[strings.ToLower(r.ResourceType)] = map[string]azqr.AprlRecommendation{}
				}

				reportData.Recomendations[strings.ToLower(r.ResourceType)][i] = r.ToAzureAprlRecommendation()
			}
		}
	}

	// scan each subscription with AZQR scanners
	for sid, sn := range subscriptions {
		config := &azqr.ScannerConfig{
			Ctx:              ctx,
			SubscriptionID:   sid,
			SubscriptionName: sn,
			Cred:             cred,
			ClientOptions:    clientOptions,
		}

		if params.UseAzqrRecommendations {
			// list resource groups
			resourceGroups := listResourceGroups(ctx, cred, params.ResourceGroup, sid, filters, clientOptions)

			// scan private endpoints
			peResults := peScanner.Scan(config)

			// scan diagnostic settings
			diagResults := diagnosticsScanner.Scan(config)

			// scan public IPs
			pips := pipScanner.Scan(config)

			// initialize scan context
			scanContext := azqr.ScanContext{
				Exclusions:          filters.Azqr.Exclude,
				PrivateEndpoints:    peResults,
				DiagnosticsSettings: diagResults,
				PublicIPs:           pips,
			}

			// scan each resource group
			for _, r := range resourceGroups {
				var wg sync.WaitGroup
				ch := make(chan []azqr.AzqrServiceResult, 5)
				wg.Add(len(params.ServiceScanners))

				go func() {
					wg.Wait()
					close(ch)
				}()

				for _, s := range params.ServiceScanners {
					err := s.Init(config)
					if err != nil {
						log.Fatal().Err(err).Msg("Failed to initialize scanner")
					}

					go func(r string, s azqr.IAzureScanner) {
						defer wg.Done()

						res, err := retry(3, 10*time.Millisecond, s, r, &scanContext)
						if err != nil {
							cancel()
							log.Fatal().Err(err).Msg("Failed to scan")
						}
						ch <- res
					}(r, s)
				}

				for i := 0; i < len(params.ServiceScanners); i++ {
					res := <-ch
					for _, r := range res {
						// check if the resource is excluded
						if filters.Azqr.Exclude.IsServiceExcluded(r.ResourceID()) {
							continue
						}
						reportData.AzqrData = append(reportData.AzqrData, r)
					}
				}
			}
		}

		// scan defender
		reportData.DefenderData = append(reportData.DefenderData, defenderScanner.Scan(params.Defender, config)...)

		// scan advisor
		reportData.AdvisorData = append(reportData.AdvisorData, advisorScanner.Scan(params.Advisor, config)...)

		// scan costs
		costs := costScanner.Scan(params.Cost, config)
		reportData.CostData.From = costs.From
		reportData.CostData.To = costs.To
		reportData.CostData.Items = append(reportData.CostData.Items, costs.Items...)
	}

	reportData.ResourceTypeCount = getCountPerResourceType(ctx, cred, subscriptions, reportData.Recomendations)

	// render csv reports
	if params.Csv {
		csv.CreateCsvReport(&reportData)
	}

	// render excel report
	excel.CreateExcelReport(&reportData)

	// render json report
	json.CreateJsonReport(&reportData)

	log.Info().Msg("Scan completed.")
}

// retry retries the Azure scanner Scan, a number of times with an increasing delay between retries
func retry(attempts int, sleep time.Duration, a azqr.IAzureScanner, r string, scanContext *azqr.ScanContext) ([]azqr.AzqrServiceResult, error) {
	var err error
	for i := 0; ; i++ {
		res, err := a.Scan(r, scanContext)
		if err == nil {
			return res, nil
		}

		if azqr.ShouldSkipError(err) {
			return []azqr.AzqrServiceResult{}, nil
		}

		errAsString := err.Error()

		if i >= (attempts - 1) {
			log.Info().Msgf("Retry limit reached. Error: %s", errAsString)
			break
		}

		log.Debug().Msgf("Retrying after error: %s", errAsString)

		time.Sleep(sleep)
		sleep *= 2
	}
	return nil, err
}

func checkExistenceResourceGroup(ctx context.Context, subscriptionID string, resourceGroupName string, cred azcore.TokenCredential, options *arm.ClientOptions) (bool, error) {
	resourceGroupClient, err := armresources.NewResourceGroupsClient(subscriptionID, cred, options)
	if err != nil {
		return false, err
	}

	boolResp, err := resourceGroupClient.CheckExistence(ctx, resourceGroupName, nil)
	if err != nil {
		return false, err
	}
	return boolResp.Success, nil
}

func listResourceGroup(ctx context.Context, subscriptionID string, cred azcore.TokenCredential, options *arm.ClientOptions) ([]*armresources.ResourceGroup, error) {
	resourceGroupClient, err := armresources.NewResourceGroupsClient(subscriptionID, cred, options)
	if err != nil {
		return nil, err
	}

	resultPager := resourceGroupClient.NewListPager(nil)

	resourceGroups := make([]*armresources.ResourceGroup, 0)
	for resultPager.More() {
		pageResp, err := resultPager.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		resourceGroups = append(resourceGroups, pageResp.ResourceGroupListResult.Value...)
	}
	return resourceGroups, nil
}

func listResourceGroups(ctx context.Context, cred azcore.TokenCredential, resourceGroup string, subscriptionID string, exclusions *filters.Filters, options *arm.ClientOptions) []string {
	resourceGroups := []string{}
	if resourceGroup != "" {
		exists, err := checkExistenceResourceGroup(ctx, subscriptionID, resourceGroup, cred, options)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to check existence of Resource Group")
		}

		if !exists {
			log.Fatal().Msgf("Resource Group %s does not exist", resourceGroup)
		}

		if exclusions.Azqr.Exclude.IsResourceGroupExcluded(fmt.Sprintf("/subscriptions/%s/resourceGroups/%s", subscriptionID, resourceGroup)) {
			log.Info().Msgf("Skipping subscriptions/...%s/resourceGroups/%s", subscriptionID[29:], resourceGroup)
			return resourceGroups
		}

		resourceGroups = append(resourceGroups, resourceGroup)
	} else {
		rgs, err := listResourceGroup(ctx, subscriptionID, cred, options)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to list Resource Groups")
		}
		for _, rg := range rgs {
			if exclusions.Azqr.Exclude.IsResourceGroupExcluded(fmt.Sprintf("/subscriptions/%s/resourceGroups/%s", subscriptionID, *rg.Name)) {
				log.Info().Msgf("Skipping subscriptions/...%s/resourceGroups/%s", subscriptionID[29:], *rg.Name)
				continue
			}
			resourceGroups = append(resourceGroups, *rg.Name)
		}
	}
	return resourceGroups
}

func listSubscriptions(ctx context.Context, cred azcore.TokenCredential, subscriptionID string, filters *filters.Filters, options *arm.ClientOptions) map[string]string {
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
		// if subscriptionID is empty, return all subscriptions. Otherwise, return only the specified subscription
		sid := *s.SubscriptionID
		if subscriptionID == "" || subscriptionID == sid {
			if filters.Azqr.Exclude.IsSubscriptionExcluded(sid) {
				log.Info().Msgf("Skipping subscriptions/...%s", sid[29:])
				continue
			}
			result[*s.SubscriptionID] = *s.DisplayName
		}
	}

	return result
}

func newAzureCredential(forceAzureCliCredential bool) azcore.TokenCredential {
	var cred azcore.TokenCredential
	var err error
	if !forceAzureCliCredential {
		cred, err = azidentity.NewDefaultAzureCredential(nil)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to get Azure credentials")
		}
	} else {
		cred, err = azidentity.NewAzureCLICredential(nil)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to get Azure CLI credentials")
		}
	}
	return cred
}

func generateOutputFileName(outputName string) string {
	outputFile := outputName
	if outputFile == "" {
		current_time := time.Now()
		outputFileStamp := fmt.Sprintf("%d_%02d_%02d_T%02d%02d%02d",
			current_time.Year(), current_time.Month(), current_time.Day(),
			current_time.Hour(), current_time.Minute(), current_time.Second())

		outputFile = fmt.Sprintf("%s_%s", "azqr_report", outputFileStamp)
	}
	return outputFile
}
