package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/cmendible/azqr/internal/renderers"
	"github.com/cmendible/azqr/internal/scanners"
	"github.com/cmendible/azqr/internal/scanners/afd"
	"github.com/cmendible/azqr/internal/scanners/agw"
	"github.com/cmendible/azqr/internal/scanners/aks"
	"github.com/cmendible/azqr/internal/scanners/apim"
	"github.com/cmendible/azqr/internal/scanners/appcs"
	"github.com/cmendible/azqr/internal/scanners/cae"
	"github.com/cmendible/azqr/internal/scanners/ci"
	"github.com/cmendible/azqr/internal/scanners/cosmos"
	"github.com/cmendible/azqr/internal/scanners/cr"
	"github.com/cmendible/azqr/internal/scanners/evgd"
	"github.com/cmendible/azqr/internal/scanners/evh"
	"github.com/cmendible/azqr/internal/scanners/kv"
	"github.com/cmendible/azqr/internal/scanners/plan"
	"github.com/cmendible/azqr/internal/scanners/psql"
	"github.com/cmendible/azqr/internal/scanners/redis"
	"github.com/cmendible/azqr/internal/scanners/sb"
	"github.com/cmendible/azqr/internal/scanners/sigr"
	"github.com/cmendible/azqr/internal/scanners/sql"
	"github.com/cmendible/azqr/internal/scanners/st"
	"github.com/cmendible/azqr/internal/scanners/wps"
	"golang.org/x/sync/semaphore"
)

const (
	defaultConcurrency = 4
)

var (
	version = "dev"
)

func main() {
	subscriptionPtr := flag.String("s", "", "Azure Subscription Id (Required)")
	resourceGroupPtr := flag.String("g", "", "Azure Resource Group")
	outputPtr := flag.String("o", "azqr_report", "Output file prefix")
	detail := flag.Bool("d", false, "Enable more details in the report")
	maskPtr := flag.Bool("m", false, "Mask the subscription id in the report")
	concurrency := flag.Int("p", defaultConcurrency, fmt.Sprintf("Parallel processes. Default to %d. A < 0 value will use the maxmimum concurrency.", defaultConcurrency))
	ver := flag.Bool("v", false, "Print version and exit")

	flag.Parse()

	subscriptionID := *subscriptionPtr
	resourceGroupName := *resourceGroupPtr
	outputFilePrefix := *outputPtr

	current_time := time.Now()
	outputFileStamp := fmt.Sprintf("%d_%02d_%02d_T%02d%02d%02d",
		current_time.Year(), current_time.Month(), current_time.Day(),
		current_time.Hour(), current_time.Minute(), current_time.Second())

	outputFile := fmt.Sprintf("%s_%s", outputFilePrefix, outputFileStamp)

	if *ver {
		fmt.Printf("azqr version: %s", version)
		os.Exit(0)
	}

	if subscriptionID == "" {
		flag.Usage()
		os.Exit(1)
	}

	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()

	resourceGroups := []string{}
	if resourceGroupName != "" {
		exists, err := checkExistenceResourceGroup(ctx, subscriptionID, resourceGroupName, cred)
		if err != nil {
			log.Fatal(err)
		}

		if !exists {
			log.Fatalf("Resource Group %s does not exist", resourceGroupName)
		}
		resourceGroups = append(resourceGroups, resourceGroupName)
	} else {
		rgs, err := listResourceGroup(ctx, subscriptionID, cred)
		if err != nil {
			log.Fatal(err)
		}
		for _, rg := range rgs {
			resourceGroups = append(resourceGroups, *rg.Name)
		}
	}

	svcScanners := []scanners.IAzureScanner{
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
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	config := &scanners.ScannerConfig{
		Ctx:                ctx,
		SubscriptionID:     subscriptionID,
		Cred:               cred,
		EnableDetailedScan: *detail,
	}

	peScanner := scanners.PrivateEndpointScanner{}
	err = peScanner.Init(config)
	if err != nil {
		log.Fatal(err)
	}
	peResults, err := peScanner.ListResourcesWithPrivateEndpoints()
	if err != nil {
		log.Fatal(err)
	}

	scanContext := scanners.ScanContext{
		PrivateEndpoints: peResults,
	}

	for _, a := range svcScanners {
		err := a.Init(config)
		if err != nil {
			log.Fatal(err)
		}
	}

	var all []scanners.AzureServiceResult
	rc := ReviewContext{
		Ctx:   ctx,
		ResCh: make(chan []scanners.AzureServiceResult),
		ErrCh: make(chan error),
	}
	for _, r := range resourceGroups {
		log.Printf("Scanning Resource Group %s", r)
		go scanRunner(&rc, r, &scanContext, &svcScanners, *concurrency)
		res, err := waitForReviews(&rc, len(svcScanners))
		// As soon as any error happen, we cancel every still running analysis
		if err != nil {
			cancel()
			log.Fatal(err)
		}
		all = append(all, *res...)
	}

	defenderScanner := scanners.DefenderScanner{}
	err = defenderScanner.Init(config)
	if err != nil {
		log.Fatal(err)
	}

	defenderResults, err := defenderScanner.ListConfiguration()
	if err != nil {
		log.Fatal(err)
	}

	reportData := renderers.ReportData{
		OutputFileName:     outputFile,
		EnableDetailedScan: config.EnableDetailedScan,
		Mask:               *maskPtr,
		MainData:           all,
		DefenderData:       defenderResults,
	}

	renderers.CreateExcelReport(reportData)

	log.Println("Scan completed.")
}

// ReviewContext A running resource group analysis support context
type ReviewContext struct {
	// Review context, will be passed to every created goroutines
	Ctx context.Context
	// Communication interface for each review results
	ResCh chan []scanners.AzureServiceResult
	// Communication interface for errors
	ErrCh chan error
}

// Run a scan on a particular resource group "r" with the appropriates scanners using "concurrency" goroutines
func scanRunner(rc *ReviewContext, r string, scanContext *scanners.ScanContext, svcAnalysers *[]scanners.IAzureScanner, concurrency int) {
	if concurrency <= 0 {
		concurrency = len(*svcAnalysers)
	}
	sem := semaphore.NewWeighted(int64(concurrency))
	for i := range *svcAnalysers {
		if err := sem.Acquire(rc.Ctx, 1); err != nil {
			rc.ErrCh <- err
			return
		}
		// When starting a goroutine from a loop, we cannot directly use
		// the iteration variable, as only the last element of the loop will
		// be processed
		analyserPtr := &(*svcAnalysers)[i]
		go func(a *scanners.IAzureScanner, r string) {
			defer sem.Release(1)
			// In case the analysis was cancelled, we don't need to execute the review
			if context.Canceled == rc.Ctx.Err() {
				return
			}
			res, err := (*a).Scan(r, scanContext)
			if err != nil {
				rc.ErrCh <- err
			}
			rc.ResCh <- res
		}(analyserPtr, r)
	}
}

// Wait for at least "nb" goroutines to hands their result and return them
func waitForReviews(rc *ReviewContext, nb int) (*[]scanners.AzureServiceResult, error) {
	received := 0
	reviews := make([]scanners.AzureServiceResult, 0)
	for {
		select {
		// In case a timeout is set
		case <-rc.Ctx.Done():
			return nil, rc.Ctx.Err()
		case err := <-rc.ErrCh:
			return nil, err
		case res := <-rc.ResCh:
			received++
			reviews = append(reviews, res...)
			if received >= nb {
				return &reviews, nil
			}
		}
	}
}

func checkExistenceResourceGroup(ctx context.Context, subscriptionID string, resourceGroupName string, cred azcore.TokenCredential) (bool, error) {
	resourceGroupClient, err := armresources.NewResourceGroupsClient(subscriptionID, cred, nil)
	if err != nil {
		return false, err
	}

	boolResp, err := resourceGroupClient.CheckExistence(ctx, resourceGroupName, nil)
	if err != nil {
		return false, err
	}
	return boolResp.Success, nil
}

func listResourceGroup(ctx context.Context, subscriptionID string, cred azcore.TokenCredential) ([]*armresources.ResourceGroup, error) {
	resourceGroupClient, err := armresources.NewResourceGroupsClient(subscriptionID, cred, nil)
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
