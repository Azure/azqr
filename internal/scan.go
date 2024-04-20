// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package internal

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

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
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/subscription/armsubscription"

	"github.com/Azure/azqr/internal/scanners/adf"
	"github.com/Azure/azqr/internal/scanners/afd"
	"github.com/Azure/azqr/internal/scanners/afw"
	"github.com/Azure/azqr/internal/scanners/agw"
	"github.com/Azure/azqr/internal/scanners/aks"
	"github.com/Azure/azqr/internal/scanners/apim"
	"github.com/Azure/azqr/internal/scanners/appcs"
	"github.com/Azure/azqr/internal/scanners/appi"
	"github.com/Azure/azqr/internal/scanners/asp"
	"github.com/Azure/azqr/internal/scanners/ca"
	"github.com/Azure/azqr/internal/scanners/cae"
	"github.com/Azure/azqr/internal/scanners/ci"
	"github.com/Azure/azqr/internal/scanners/cog"
	"github.com/Azure/azqr/internal/scanners/cosmos"
	"github.com/Azure/azqr/internal/scanners/cr"
	"github.com/Azure/azqr/internal/scanners/dbw"
	"github.com/Azure/azqr/internal/scanners/dec"
	"github.com/Azure/azqr/internal/scanners/evgd"
	"github.com/Azure/azqr/internal/scanners/evh"
	"github.com/Azure/azqr/internal/scanners/kv"
	"github.com/Azure/azqr/internal/scanners/lb"
	"github.com/Azure/azqr/internal/scanners/logic"
	"github.com/Azure/azqr/internal/scanners/maria"
	"github.com/Azure/azqr/internal/scanners/mysql"
	"github.com/Azure/azqr/internal/scanners/psql"
	"github.com/Azure/azqr/internal/scanners/redis"
	"github.com/Azure/azqr/internal/scanners/sb"
	"github.com/Azure/azqr/internal/scanners/sigr"
	"github.com/Azure/azqr/internal/scanners/sql"
	"github.com/Azure/azqr/internal/scanners/st"
	"github.com/Azure/azqr/internal/scanners/syndp"
	"github.com/Azure/azqr/internal/scanners/synsp"
	"github.com/Azure/azqr/internal/scanners/synw"
	"github.com/Azure/azqr/internal/scanners/traf"
	"github.com/Azure/azqr/internal/scanners/vm"
	"github.com/Azure/azqr/internal/scanners/vnet"
	"github.com/Azure/azqr/internal/scanners/wps"
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
}

func Scan(params *ScanParams) {
	subscriptionID := params.SubscriptionID
	resourceGroupName := params.ResourceGroup
	outputFileName := params.OutputName
	defender := params.Defender
	advisor := params.Advisor
	cost := params.Cost
	createXlsx := params.Xlsx
	mask := params.Mask
	debug := params.Debug
	forceAzureCliCredential := params.ForceAzureCliCredential

	// Default level for this example is info, unless debug flag is present
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		log.Debug().Msg("Debug logging enabled")
	}

	if subscriptionID == "" && resourceGroupName != "" {
		log.Fatal().Msg("Resource Group name can only be used with a Subscription Id")
	}

	outputFile := outputFileName
	if outputFile == "" {
		current_time := time.Now()
		outputFileStamp := fmt.Sprintf("%d_%02d_%02d_T%02d%02d%02d",
			current_time.Year(), current_time.Month(), current_time.Day(),
			current_time.Hour(), current_time.Minute(), current_time.Second())

		outputFile = fmt.Sprintf("%s_%s", "azqr_report", outputFileStamp)
	}

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

	ctx := context.Background()

	clientOptions := &arm.ClientOptions{
		ClientOptions: policy.ClientOptions{
			Retry: policy.RetryOptions{
				RetryDelay:    20 * time.Millisecond,
				MaxRetries:    3,
				MaxRetryDelay: 10 * time.Minute,
			},
		},
	}

	subscriptions := []string{}
	if subscriptionID != "" {
		subscriptions = append(subscriptions, subscriptionID)
	} else {
		subs, err := listSubscriptions(ctx, cred, clientOptions)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to list subscriptions")
		}
		for _, s := range subs {
			subscriptions = append(subscriptions, *s.SubscriptionID)
		}
	}

	var ruleResults []scanners.AzureServiceResult
	var defenderResults []scanners.DefenderResult
	var advisorResults []scanners.AdvisorResult
	costResult := &scanners.CostResult{
		Items: []*scanners.CostResultItem{},
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	defenderScanner := scanners.DefenderScanner{}
	peScanner := scanners.PrivateEndpointScanner{}
	pipScanner := scanners.PublicIPScanner{}
	diagnosticsScanner := scanners.DiagnosticSettingsScanner{}
	advisorScanner := scanners.AdvisorScanner{}
	costScanner := scanners.CostScanner{}

	for _, s := range subscriptions {
		resourceGroups := []string{}
		if resourceGroupName != "" {
			exists, err := checkExistenceResourceGroup(ctx, s, resourceGroupName, cred, clientOptions)
			if err != nil {
				log.Fatal().Err(err).Msg("Failed to check existence of Resource Group")
			}

			if !exists {
				log.Fatal().Msgf("Resource Group %s does not exist", resourceGroupName)
			}
			resourceGroups = append(resourceGroups, resourceGroupName)
		} else {
			rgs, err := listResourceGroup(ctx, s, cred, clientOptions)
			if err != nil {
				log.Fatal().Err(err).Msg("Failed to list Resource Groups")
			}
			for _, rg := range rgs {
				resourceGroups = append(resourceGroups, *rg.Name)
			}
		}

		config := &scanners.ScannerConfig{
			Ctx:            ctx,
			SubscriptionID: s,
			Cred:           cred,
			ClientOptions:  clientOptions,
		}

		err = peScanner.Init(config)
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

		err = diagnosticsScanner.Init(config)
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

		err = pipScanner.Init(config)
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

		scanContext := scanners.ScanContext{
			PrivateEndpoints:    peResults,
			DiagnosticsSettings: diagResults,
			PublicIPs:           pips,
		}

		for _, a := range params.ServiceScanners {
			err := a.Init(config)
			if err != nil {
				log.Fatal().Err(err).Msg("Failed to initialize scanner")
			}
		}

		for _, r := range resourceGroups {
			var wg sync.WaitGroup
			ch := make(chan []scanners.AzureServiceResult, 5)
			wg.Add(len(params.ServiceScanners))

			go func() {
				wg.Wait()
				close(ch)
			}()

			for _, s := range params.ServiceScanners {
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
				ruleResults = append(ruleResults, res...)
			}
		}

		if defender {
			err = defenderScanner.Init(config)
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

		if advisor {
			err = advisorScanner.Init(config)
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

		if cost {
			err = costScanner.Init(config)
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
	}

	reportData := renderers.ReportData{
		OutputFileName: outputFile,
		Mask:           mask,
		MainData:       ruleResults,
		DefenderData:   defenderResults,
		AdvisorData:    advisorResults,
		CostData:       costResult,
	}

	if createXlsx {
		excel.CreateExcelReport(&reportData)
	}

	csv.CreateCsvReport(&reportData)

	log.Info().Msg("Scan completed.")
}

func retry(attempts int, sleep time.Duration, a scanners.IAzureScanner, r string, scanContext *scanners.ScanContext) ([]scanners.AzureServiceResult, error) {
	var err error
	for i := 0; ; i++ {
		res, err := a.Scan(r, scanContext)
		if err == nil {
			return res, nil
		}

		if shouldSkipError(err) {
			return []scanners.AzureServiceResult{}, nil
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

func listSubscriptions(ctx context.Context, cred azcore.TokenCredential, options *arm.ClientOptions) ([]*armsubscription.Subscription, error) {
	client, err := armsubscription.NewSubscriptionsClient(cred, options)
	if err != nil {
		return nil, err
	}

	resultPager := client.NewListPager(nil)

	subscriptions := make([]*armsubscription.Subscription, 0)
	for resultPager.More() {
		pageResp, err := resultPager.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, s := range pageResp.Value {
			if s.State != to.Ptr(armsubscription.SubscriptionStateDisabled) &&
				s.State != to.Ptr(armsubscription.SubscriptionStateDeleted) {
				subscriptions = append(subscriptions, s)
			}
		}
	}
	return subscriptions, nil
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

func GetScanners() []scanners.IAzureScanner {
	return []scanners.IAzureScanner{
		&dbw.DatabricksScanner{},
		&adf.DataFactoryScanner{},
		&afd.FrontDoorScanner{},
		&afw.FirewallScanner{},
		&agw.ApplicationGatewayScanner{},
		&aks.AKSScanner{},
		&apim.APIManagementScanner{},
		&appcs.AppConfigurationScanner{},
		&appi.AppInsightsScanner{},
		&cae.ContainerAppsEnvironmentScanner{},
		&ca.ContainerAppsScanner{},
		&ci.ContainerInstanceScanner{},
		&cog.CognitiveScanner{},
		&cosmos.CosmosDBScanner{},
		&cr.ContainerRegistryScanner{},
		&dec.DataExplorerScanner{},
		&evgd.EventGridScanner{},
		&evh.EventHubScanner{},
		&kv.KeyVaultScanner{},
		&lb.LoadBalancerScanner{},
		&logic.LogicAppScanner{},
		&maria.MariaScanner{},
		&mysql.MySQLFlexibleScanner{},
		&mysql.MySQLScanner{},
		&asp.AppServiceScanner{},
		&psql.PostgreFlexibleScanner{},
		&psql.PostgreScanner{},
		&redis.RedisScanner{},
		&sb.ServiceBusScanner{},
		&sigr.SignalRScanner{},
		&sql.SQLScanner{},
		&synw.SynapseWorkspaceScanner{},
		&synsp.SynapseSparkPoolScanner{},
		&syndp.SynapseSqlPoolScanner{},
		&traf.TrafficManagerScanner{},
		&st.StorageScanner{},
		&vm.VirtualMachineScanner{},
		&vnet.VirtualNetworkScanner{},
		&wps.WebPubSubScanner{},
	}
}
