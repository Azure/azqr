// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package internal

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/Azure/azqr/internal/filters"
	"github.com/Azure/azqr/internal/graph"
	"github.com/Azure/azqr/internal/renderers"
	"github.com/Azure/azqr/internal/renderers/csv"
	"github.com/Azure/azqr/internal/renderers/excel"
	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azqr/internal/to"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v5"
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
	Xlsx                    bool
	Debug                   bool
	ServiceScanners         []scanners.IAzureScanner
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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	clientOptions := &arm.ClientOptions{
		ClientOptions: policy.ClientOptions{
			Retry: policy.RetryOptions{
				RetryDelay:    20 * time.Millisecond,
				MaxRetries:    3,
				MaxRetryDelay: 10 * time.Minute,
			},
		},
	}

	// list subscriptions
	subscriptions := listSubscriptions(ctx, cred, params.SubscriptionID, filters, clientOptions)

	// initialize scanners
	defenderScanner := scanners.DefenderScanner{}
	pipScanner := scanners.PublicIPScanner{}
	peScanner := scanners.PrivateEndpointScanner{}
	diagnosticsScanner := scanners.DiagnosticSettingsScanner{}
	advisorScanner := scanners.AdvisorScanner{}
	costScanner := scanners.CostScanner{}

	// build report data
	reportData := renderers.ReportData{
		OutputFileName: outputFile,
		Mask:           params.Mask,
		Recomendations: map[string]map[string]scanners.AprlRecommendation{},
		AzqrData:       []scanners.AzqrServiceResult{},
		AprlData:       []scanners.AprlResult{},
		DefenderData:   []scanners.DefenderResult{},
		AdvisorData:    []scanners.AdvisorResult{},
		CostData: &scanners.CostResult{
			Items: []*scanners.CostResultItem{},
		},
	}

	reportData.Recomendations, reportData.AprlData = aprlScan(ctx, cred, params, filters, subscriptions)

	if params.UseAzqrRecommendations {
		for _, s := range params.ServiceScanners {
			for i, r := range s.GetRecommendations() {
				if reportData.Recomendations[r.ResourceType] == nil {
					reportData.Recomendations[r.ResourceType] = map[string]scanners.AprlRecommendation{}
				}

				reportData.Recomendations[r.ResourceType][i] = r.ToAzureAprlRecommendation()
			}
		}
	}

	for sid, sn := range subscriptions {
		config := &scanners.ScannerConfig{
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
			peResults := scanPrivateEndpoints(config, &peScanner)

			// scan diagnostic settings
			diagResults := scanDiagnosticSettings(config, &diagnosticsScanner)

			// scan public IPs
			pips := scanPublicIPs(config, &pipScanner)

			scanContext := scanners.ScanContext{
				Exclusions:          filters.Azqr.Exclude,
				PrivateEndpoints:    peResults,
				DiagnosticsSettings: diagResults,
				PublicIPs:           pips,
			}

			for _, r := range resourceGroups {
				var wg sync.WaitGroup
				ch := make(chan []scanners.AzqrServiceResult, 5)
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

					go func(r string, s scanners.IAzureScanner) {
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
						if filters.Azqr.Exclude.IsServiceExcluded(r.ResourceID()) {
							continue
						}
						reportData.AzqrData = append(reportData.AzqrData, r)
					}
				}
			}
		}

		// scan defender
		reportData.DefenderData = append(reportData.DefenderData, scanDefender(params.Defender, config, &defenderScanner)...)

		// scan advisor
		reportData.AdvisorData = append(reportData.AdvisorData, scanAdvisor(params.Advisor, config, &advisorScanner)...)

		// scan costs
		costs := scanCosts(params.Cost, config, &costScanner)
		reportData.CostData.From = costs.From
		reportData.CostData.To = costs.To
		reportData.CostData.Items = append(reportData.CostData.Items, costs.Items...)
	}

	// render excel report
	if params.Xlsx {
		excel.CreateExcelReport(&reportData)
	}

	// render csv reports
	csv.CreateCsvReport(&reportData)

	log.Info().Msg("Scan completed.")
}

func aprlScan(ctx context.Context, cred azcore.TokenCredential, params *ScanParams, filters *filters.Filters, subscriptions map[string]string) (map[string]map[string]scanners.AprlRecommendation, []scanners.AprlResult) {
	recommendations := map[string]map[string]scanners.AprlRecommendation{}
	results := []scanners.AprlResult{}
	rules := []scanners.AprlRecommendation{}
	graph := graph.NewGraphQuery(cred)

	// get APRL recommendations
	aprl := GetAprlRecommendations()

	for _, s := range params.ServiceScanners {
		for _, t := range s.ResourceTypes() {
			scanners.LogResourceTypeScan(t)
			gr := scanners.GetGraphRules(t, aprl)
			for _, r := range gr {
				rules = append(rules, r)
			}

			for i, r := range gr {
				if recommendations[t] == nil {
					recommendations[t] = map[string]scanners.AprlRecommendation{}
				}
				recommendations[t][i] = r
			}
		}
	}

	batches := int(math.Ceil(float64(len(rules)) / 12))

	var wg sync.WaitGroup
	ch := make(chan []scanners.AprlResult, 12)
	wg.Add(batches)

	go func() {
		wg.Wait()
		close(ch)
	}()

	batchSzie := 12
	batchNumber := 0
	for i := 0; i < len(rules); i += batchSzie {
		j := i + batchSzie
		if j > len(rules) {
			j = len(rules)
		}

		go func(r []scanners.AprlRecommendation, b int) {
			defer wg.Done()
			if b > 0 {
				// Staggering queries to avoid throttling. Max 15 queries each 5 seconds.
				// https://learn.microsoft.com/en-us/azure/governance/resource-graph/concepts/guidance-for-throttled-requests#staggering-queries
				s := time.Duration(b * 7)
				time.Sleep(s * time.Second)
			}
			res, err := graphScan(ctx, graph, r, subscriptions)
			if err != nil {
				log.Fatal().Err(err).Msg("Failed to scan")
			}
			ch <- res
		}(rules[i:j], batchNumber)

		batchNumber++
	}

	for i := 0; i < batches; i++ {
		res := <-ch
		for _, r := range res {
			if filters.Azqr.Exclude.IsServiceExcluded(r.ResourceID) {
				continue
			}
			results = append(results, r)
		}
	}

	return recommendations, results
}

func retry(attempts int, sleep time.Duration, a scanners.IAzureScanner, r string, scanContext *scanners.ScanContext) ([]scanners.AzqrServiceResult, error) {
	var err error
	for i := 0; ; i++ {
		res, err := a.Scan(r, scanContext)
		if err == nil {
			return res, nil
		}

		if shouldSkipError(err) {
			return []scanners.AzqrServiceResult{}, nil
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

func shouldSkipError(err error) bool {
	var respErr *azcore.ResponseError
	if errors.As(err, &respErr) {
		switch respErr.ErrorCode {
		case "MissingRegistrationForResourceProvider", "MissingSubscriptionRegistration", "DisallowedOperation":
			log.Warn().Msgf("Subscription failed with code: %s. Skipping Scan...", respErr.ErrorCode)
			return true
		}
	}
	return false
}

func graphScan(ctx context.Context, graphClient *graph.GraphQuery, rules []scanners.AprlRecommendation, subscriptions map[string]string) ([]scanners.AprlResult, error) {
	results := []scanners.AprlResult{}
	subs := make([]*string, 0, len(subscriptions))
	for s := range subscriptions {
		subs = append(subs, &s)
	}

	batchSize := 300
	for i := 0; i < len(subs); i += batchSize {
		j := i + batchSize
		if j > len(subs) {
			j = len(subs)
		}

		for _, rule := range rules {
			if rule.GraphQuery != "" {
				result := graphClient.Query(ctx, rule.GraphQuery, subs[i:j])
				if result.Data != nil {
					for _, row := range result.Data {
						m := row.(map[string]interface{})

						tags := ""
						// if m["tags"] != nil {
						// 	tags = m["tags"].(string)
						// }

						param1 := ""
						if m["param1"] != nil {
							param1 = m["param1"].(string)
						}

						param2 := ""
						if m["param2"] != nil {
							param2 = m["param2"].(string)
						}

						param3 := ""
						if m["param3"] != nil {
							param3 = m["param3"].(string)
						}

						param4 := ""
						if m["param4"] != nil {
							param4 = m["param4"].(string)
						}

						param5 := ""
						if m["param5"] != nil {
							param5 = m["param5"].(string)
						}

						log.Debug().Msg(rule.GraphQuery)

						subscription := scanners.GetSubsctiptionFromResourceID(m["id"].(string))
						subscriptionName := subscriptions[subscription]

						results = append(results, scanners.AprlResult{
							RecommendationID:    rule.RecommendationID,
							Category:            scanners.RecommendationCategory(rule.Category),
							Recommendation:      rule.Recommendation,
							ResourceType:        rule.ResourceType,
							LongDescription:     rule.LongDescription,
							PotentialBenefits:   rule.PotentialBenefits,
							Impact:              scanners.RecommendationImpact(rule.Impact),
							Name:                m["name"].(string),
							ResourceID:          m["id"].(string),
							SubscriptionID:      subscription,
							SubscriptionName:    subscriptionName,
							ResourceGroup:       scanners.GetResourceGroupFromResourceID(m["id"].(string)),
							Tags:                tags,
							Param1:              param1,
							Param2:              param2,
							Param3:              param3,
							Param4:              param4,
							Param5:              param5,
							Learn:               rule.LearnMoreLink[0].Url,
							AutomationAvailable: rule.AutomationAvailable,
							Source:              "APRL",
						})
					}
				}
			}
		}
	}

	return results, nil
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

func scanCosts(scan bool, config *scanners.ScannerConfig, costScanner *scanners.CostScanner) *scanners.CostResult {
	costResult := &scanners.CostResult{
		Items: []*scanners.CostResultItem{},
	}
	if scan {
		err := costScanner.Init(config)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to initialize Cost Scanner")
		}
		costs, err := costScanner.QueryCosts()
		if err != nil && !shouldSkipError(err) {
			log.Fatal().Err(err).Msg("Failed to query costs")
		}
		costResult.From = costs.From
		costResult.To = costs.To
		costResult.Items = append(costResult.Items, costs.Items...)
	}
	return costResult
}

func scanDefender(scan bool, config *scanners.ScannerConfig, defenderScanner *scanners.DefenderScanner) []scanners.DefenderResult {
	defenderResults := []scanners.DefenderResult{}
	if scan {
		err := defenderScanner.Init(config)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to initialize Defender Scanner")
		}

		res, err := defenderScanner.ListConfiguration()
		if err != nil {
			if shouldSkipError(err) {
				res = []scanners.DefenderResult{}
			} else {
				log.Fatal().Err(err).Msg("Failed to list Defender configuration")
			}
		}
		defenderResults = append(defenderResults, res...)
	}
	return defenderResults
}

func scanAdvisor(scan bool, config *scanners.ScannerConfig, advisorScanner *scanners.AdvisorScanner) []scanners.AdvisorResult {
	advisorResults := []scanners.AdvisorResult{}
	if scan {
		err := advisorScanner.Init(config)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to initialize Advisor Scanner")
		}

		rec, err := advisorScanner.ListRecommendations()
		if err != nil {
			if shouldSkipError(err) {
				rec = []scanners.AdvisorResult{}
			} else {
				log.Fatal().Err(err).Msg("Failed to list Advisor recommendations")
			}
		}
		advisorResults = append(advisorResults, rec...)
	}
	return advisorResults
}

func scanPrivateEndpoints(config *scanners.ScannerConfig, peScanner *scanners.PrivateEndpointScanner) map[string]bool {
	err := peScanner.Init(config)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize Private Endpoint Scanner")
	}
	peResults, err := peScanner.ListResourcesWithPrivateEndpoints()
	if err != nil {
		if shouldSkipError(err) {
			peResults = map[string]bool{}
		} else {
			log.Fatal().Err(err).Msg("Failed to list resources with Private Endpoints")
		}
	}
	return peResults
}

func scanPublicIPs(config *scanners.ScannerConfig, pipScanner *scanners.PublicIPScanner) map[string]*armnetwork.PublicIPAddress {
	err := pipScanner.Init(config)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize Diagnostic Settings Scanner")
	}
	pips, err := pipScanner.ListPublicIPs()
	if err != nil {
		if shouldSkipError(err) {
			pips = map[string]*armnetwork.PublicIPAddress{}
		} else {
			log.Fatal().Err(err).Msg("Failed to list Public IPs")
		}
	}
	return pips
}

func scanDiagnosticSettings(config *scanners.ScannerConfig, diagnosticsScanner *scanners.DiagnosticSettingsScanner) map[string]bool {
	err := diagnosticsScanner.Init(config)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize Diagnostic Settings Scanner")
	}
	diagResults, err := diagnosticsScanner.ListResourcesWithDiagnosticSettings()
	if err != nil {
		if shouldSkipError(err) {
			diagResults = map[string]bool{}
		} else {
			log.Fatal().Err(err).Msg("Failed to list resources with Diagnostic Settings")
		}
	}
	return diagResults
}
