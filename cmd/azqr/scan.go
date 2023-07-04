// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package azqr

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azqr/internal/scanners/afd"
	"github.com/Azure/azqr/internal/scanners/afw"
	"github.com/Azure/azqr/internal/scanners/agw"
	"github.com/Azure/azqr/internal/scanners/aks"
	"github.com/Azure/azqr/internal/scanners/apim"
	"github.com/Azure/azqr/internal/scanners/appcs"
	"github.com/Azure/azqr/internal/scanners/appi"
	"github.com/Azure/azqr/internal/scanners/cae"
	"github.com/Azure/azqr/internal/scanners/ci"
	"github.com/Azure/azqr/internal/scanners/cosmos"
	"github.com/Azure/azqr/internal/scanners/cr"
	"github.com/Azure/azqr/internal/scanners/evgd"
	"github.com/Azure/azqr/internal/scanners/evh"
	"github.com/Azure/azqr/internal/scanners/kv"
	"github.com/Azure/azqr/internal/scanners/mysql"
	"github.com/Azure/azqr/internal/scanners/plan"
	"github.com/Azure/azqr/internal/scanners/psql"
	"github.com/Azure/azqr/internal/scanners/redis"
	"github.com/Azure/azqr/internal/scanners/sb"
	"github.com/Azure/azqr/internal/scanners/sigr"
	"github.com/Azure/azqr/internal/scanners/sql"
	"github.com/Azure/azqr/internal/scanners/st"
	"github.com/Azure/azqr/internal/scanners/wps"
	"github.com/Azure/azqr/internal/scanners/lb"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/Azure/azqr/internal/renderers"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/subscription/armsubscription"
	"github.com/spf13/cobra"
)

func init() {
	scanCmd.PersistentFlags().StringP("subscription-id", "s", "", "Azure Subscription Id")
	scanCmd.PersistentFlags().StringP("resource-group", "g", "", "Azure Resource Group (Use with --subscription-id)")
	scanCmd.PersistentFlags().BoolP("defender", "d", true, "Scan Defender Status")
	scanCmd.PersistentFlags().BoolP("advisor", "a", true, "Scan Azure Advisor Recommendations")
	scanCmd.PersistentFlags().BoolP("costs", "c", false, "Scan Azure Costs")
	scanCmd.PersistentFlags().StringP("output-prefix", "o", "azqr_report", "Output file prefix")
	scanCmd.PersistentFlags().BoolP("mask", "m", true, "Mask the subscription id in the report")
	scanCmd.PersistentFlags().BoolP("debug", "", false, "Set log level to debug")

	rootCmd.AddCommand(scanCmd)
}

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan Azure Resources",
	Long:  "Scan Azure Resources",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		serviceScanners := []scanners.IAzureScanner{
			&aks.AKSScanner{},
			&apim.APIManagementScanner{},
			&agw.ApplicationGatewayScanner{},
			&cae.ContainerAppsScanner{},
			&ci.ContainerInstanceScanner{},
			&cosmos.CosmosDBScanner{},
			&cr.ContainerRegistryScanner{},
			&evh.EventHubScanner{},
			&evgd.EventGridScanner{},
			&kv.KeyVaultScanner{},
			&appcs.AppConfigurationScanner{},
			&plan.AppServiceScanner{},
			&redis.RedisScanner{},
			&sb.ServiceBusScanner{},
			&sigr.SignalRScanner{},
			&wps.WebPubSubScanner{},
			&st.StorageScanner{},
			&psql.PostgreScanner{},
			&psql.PostgreFlexibleScanner{},
			&sql.SQLScanner{},
			&afd.FrontDoorScanner{},
			&afw.FirewallScanner{},
			&mysql.MySQLScanner{},
			&mysql.MySQLFlexibleScanner{},
			&appi.AppInsightsScanner{},
			&lb.LoadBalancerScanner{},
		}

		scan(cmd, serviceScanners)
	},
}

func scan(cmd *cobra.Command, serviceScanners []scanners.IAzureScanner) {
	subscriptionID, _ := cmd.Flags().GetString("subscription-id")
	resourceGroupName, _ := cmd.Flags().GetString("resource-group")
	outputFilePrefix, _ := cmd.Flags().GetString("output-prefix")
	defender, _ := cmd.Flags().GetBool("defender")
	advisor, _ := cmd.Flags().GetBool("advisor")
	cost, _ := cmd.Flags().GetBool("costs")
	mask, _ := cmd.Flags().GetBool("mask")
	debug, _ := cmd.Flags().GetBool("debug")

	// Default level for this example is info, unless debug flag is present
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		log.Debug().Msg("Debug logging enabled")
	}

	if subscriptionID == "" && resourceGroupName != "" {
		log.Fatal().Msg("Resource Group name can only be used with a Subscription Id")
	}

	current_time := time.Now()
	outputFileStamp := fmt.Sprintf("%d_%02d_%02d_T%02d%02d%02d",
		current_time.Year(), current_time.Month(), current_time.Day(),
		current_time.Hour(), current_time.Minute(), current_time.Second())

	outputFile := fmt.Sprintf("%s_%s", outputFilePrefix, outputFileStamp)

	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		log.Fatal().Err(err)
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
			log.Fatal().Err(err)
		}
		for _, s := range subs {
			subscriptions = append(subscriptions, *s.SubscriptionID)
		}
	}

	var ruleResults []scanners.AzureServiceResult
	var defenderResults []scanners.DefenderResult
	var advisorResults []scanners.AdvisorResult
	var costResult *scanners.CostResult

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	defenderScanner := scanners.DefenderScanner{}
	peScanner := scanners.PrivateEndpointScanner{}
	diagnosticsScanner := scanners.DiagnosticSettingsScanner{}
	advisorScanner := scanners.AdvisorScanner{}
	costScanner := scanners.CostScanner{}

	for _, s := range subscriptions {
		resourceGroups := []string{}
		if resourceGroupName != "" {
			exists, err := checkExistenceResourceGroup(ctx, s, resourceGroupName, cred, clientOptions)
			if err != nil {
				log.Fatal().Err(err)
			}

			if !exists {
				log.Fatal().Msgf("Resource Group %s does not exist", resourceGroupName)
			}
			resourceGroups = append(resourceGroups, resourceGroupName)
		} else {
			rgs, err := listResourceGroup(ctx, s, cred, clientOptions)
			if err != nil {
				log.Fatal().Err(err)
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
			log.Fatal().Err(err)
		}
		peResults, err := peScanner.ListResourcesWithPrivateEndpoints()
		if err != nil {
			log.Fatal().Err(err)
		}

		err = diagnosticsScanner.Init(config)
		if err != nil {
			log.Fatal().Err(err)
		}
		diagResults, err := diagnosticsScanner.ListResourcesWithDiagnosticSettings()
		if err != nil {
			log.Fatal().Err(err)
		}

		scanContext := scanners.ScanContext{
			PrivateEndpoints:    peResults,
			DiagnosticsSettings: diagResults,
		}

		for _, a := range serviceScanners {
			err := a.Init(config)
			if err != nil {
				log.Fatal().Err(err)
			}
		}

		for _, r := range resourceGroups {
			log.Info().Msgf("Scanning Resource Group %s", r)
			var wg sync.WaitGroup
			ch := make(chan []scanners.AzureServiceResult, 5)
			wg.Add(len(serviceScanners))

			go func() {
				wg.Wait()
				close(ch)
			}()

			for _, s := range serviceScanners {
				go func(r string, s scanners.IAzureScanner) {
					defer wg.Done()

					res, err := retry(3, 10*time.Millisecond, s, r, &scanContext)
					if err != nil {
						cancel()
						log.Fatal().Err(err)
					}
					ch <- res
				}(r, s)
			}

			for i := 0; i < len(serviceScanners); i++ {
				res := <-ch
				ruleResults = append(ruleResults, res...)
			}
		}

		if defender {
			err = defenderScanner.Init(config)
			if err != nil {
				log.Fatal().Err(err)
			}

			res, err := defenderScanner.ListConfiguration()
			if err != nil {
				log.Fatal().Err(err)
			}
			defenderResults = append(defenderResults, res...)
		}

		if advisor {
			err = advisorScanner.Init(config)
			if err != nil {
				log.Fatal().Err(err)
			}

			rec, err := advisorScanner.ListRecommendations()
			if err != nil {
				log.Fatal().Err(err)
			}
			advisorResults = append(advisorResults, rec...)
		}

		if cost {
			err = costScanner.Init(config)
			if err != nil {
				log.Fatal().Err(err)
			}
			costs, err := costScanner.QueryCosts()
			if err != nil {
				log.Fatal().Err(err)
			}
			if costResult == nil {
				costResult = costs
			} else {
				costResult.Items = append(costResult.Items, costs.Items...)
			}
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

	renderers.CreateExcelReport(reportData)

	xslx := fmt.Sprintf("%s.xlsx", reportData.OutputFileName)
	renderers.CreatePBIReport(xslx)

	log.Info().Msg("Scan completed.")
}

func retry(attempts int, sleep time.Duration, a scanners.IAzureScanner, r string, scanContext *scanners.ScanContext) ([]scanners.AzureServiceResult, error) {
	var err error
	for i := 0; ; i++ {
		res, err := a.Scan(r, scanContext)
		if err == nil {
			return res, nil
		}

		errAsString := err.Error()

		if strings.Contains(errAsString, "ERROR CODE: Subscription Not Registered") {
			log.Info().Msg("Subscription Not Registered. Skipping Scan...")
			return []scanners.AzureServiceResult{}, nil
		}

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
		subscriptions = append(subscriptions, pageResp.Value...)
	}
	return subscriptions, nil
}
