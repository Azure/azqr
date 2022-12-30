package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/cmendible/azqr/cmd/azqr/analyzers"
	"github.com/cmendible/azqr/cmd/azqr/report_templates"
	"github.com/fbiville/markdown-table-formatter/pkg/markdown"
	"golang.org/x/sync/semaphore"
)

const (
	DefaultConcurrency = 4
)

func main() {
	subscriptionPtr := flag.String("s", "", "Azure Subscription Id (Required)")
	resourceGroupPtr := flag.String("g", "", "Azure Resource Group")
	outputPtr := flag.String("o", "report.md", "Output file")
	customerPtr := flag.String("c", "<Replace with Customer Name>", "Customer Name")
	concurrency := flag.Int("p", DefaultConcurrency, fmt.Sprintf(`Parallel processes. Default to %d. A < 0 value will use the maxmimum concurrency. `, DefaultConcurrency))

	flag.Parse()

	subscriptionId := *subscriptionPtr
	resourceGroupName := *resourceGroupPtr
	outputFile := *outputPtr
	customer := *customerPtr

	if subscriptionId == "" {
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
		exists, err := checkExistenceResourceGroup(subscriptionId, resourceGroupName, ctx, cred)
		if err != nil {
			log.Fatal(err)
		}

		if !exists {
			log.Fatalf("Resource Group %s does not exist", resourceGroupName)
		}
		resourceGroups = append(resourceGroups, resourceGroupName)
	} else {
		rgs, err := listResourceGroup(subscriptionId, ctx, cred)
		if err != nil {
			log.Fatal(err)
		}
		for _, rg := range rgs {
			resourceGroups = append(resourceGroups, *rg.Name)
		}
	}

	svcanalyzers := []analyzers.AzureServiceAnalyzer{
		analyzers.NewAKSAnalyzer(subscriptionId, ctx, cred),
		analyzers.NewApiManagementAnalyzer(subscriptionId, ctx, cred),
		analyzers.NewApplicationGatewayAnalyzer(subscriptionId, ctx, cred),
		analyzers.NewContainerAppsAnalyzer(subscriptionId, ctx, cred),
		analyzers.NewContainerIntanceAnalyzer(subscriptionId, ctx, cred),
		analyzers.NewCosmosDBAnalyzer(subscriptionId, ctx, cred),
		analyzers.NewContainerRegistryAnalyzer(subscriptionId, ctx, cred),
		analyzers.NewEventHubAnalyzer(subscriptionId, ctx, cred),
		analyzers.NewEventGridAnalyzer(subscriptionId, ctx, cred),
		analyzers.NewKeyVaultAnalyzer(subscriptionId, ctx, cred),
		analyzers.NewAppServiceAnalyzer(subscriptionId, ctx, cred),
		analyzers.NewRedisAnalyzer(subscriptionId, ctx, cred),
		analyzers.NewServiceBusAnalyzer(subscriptionId, ctx, cred),
		analyzers.NewSignalRAnalyzer(subscriptionId, ctx, cred),
		analyzers.NewStorageAnalyzer(subscriptionId, ctx, cred),
		analyzers.NewPostgreAnalyzer(subscriptionId, ctx, cred),
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	var all []analyzers.AzureServiceResult
	rc := ReviewContext{
		Ctx:   ctx,
		ResCh: make(chan []analyzers.AzureServiceResult),
		ErrCh: make(chan error),
	}
	for _, r := range resourceGroups {
		log.Printf("Analyzing Resource Group %s", r)
		go reviewRunner(&rc, r, &svcanalyzers, *concurrency)
		res, err := waitForReviews(&rc, len(svcanalyzers))
		// As soon as any error happen, we cancel every still running analysis
		if err != nil {
			cancel()
			log.Fatal(err)
		}
		all = append(all, *res...)
	}
	resultsTable := renderTable(all)
	reportTemplate := report_templates.GetTemplates("Report.md")
	reportTemplate = strings.Replace(reportTemplate, "{{results}}", resultsTable, 1)
	reportTemplate = strings.Replace(reportTemplate, "{{date}}", time.Now().Format("2006-01-02"), 1)
	reportTemplate = strings.Replace(reportTemplate, "{{customer}}", customer, -1)

	recommendations := ""
	dict := map[string]bool{}
	for _, r := range all {
		parsedType := strings.Replace(r.Type, "/", ".", -1)
		if _, ok := dict[r.Type]; !ok {
			dict[r.Type] = true
			recommendations += "\n\n"
			recommendations += report_templates.GetTemplates(fmt.Sprintf("%s.md", parsedType))
		}
	}

	reportTemplate = strings.Replace(reportTemplate, "{{recommendations}}", recommendations, 1)

	err = os.WriteFile(outputFile, []byte(reportTemplate), 0644)
	if err != nil {
		log.Fatal(err)
	}
}

// ReviewContext A running resource group analysis support context
type ReviewContext struct {
	// Review context, will be passed to every created goroutines
	Ctx context.Context
	// Communication interface for each review results
	ResCh chan []analyzers.AzureServiceResult
	// Communication interface for errors
	ErrCh chan error
}

// Run a review on a peculiar resource group "r" with the appropriates analysers using "concurrency" goroutines
func reviewRunner(rc *ReviewContext, r string, svcAnalysers *[]analyzers.AzureServiceAnalyzer, concurrency int) {
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
		go func(a *analyzers.AzureServiceAnalyzer, r string) {
			defer sem.Release(1)
			// In case the analysis was cancelled, we don't need to execute the review
			if context.Canceled == rc.Ctx.Err() {
				return
			}
			res, err := (*a).Review(r)
			if err != nil {
				rc.ErrCh <- err
			}
			rc.ResCh <- res
		}(analyserPtr, r)
	}
}

// Wait for at least "nb" goroutines to hands their result and return them
func waitForReviews(rc *ReviewContext, nb int) (*[]analyzers.AzureServiceResult, error) {
	received := 0
	reviews := make([]analyzers.AzureServiceResult, nb)
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
			break
		}
	}
}

func checkExistenceResourceGroup(subscriptionId string, resourceGroupName string, ctx context.Context, cred azcore.TokenCredential) (bool, error) {
	resourceGroupClient, err := armresources.NewResourceGroupsClient(subscriptionId, cred, nil)
	if err != nil {
		return false, err
	}

	boolResp, err := resourceGroupClient.CheckExistence(ctx, resourceGroupName, nil)
	if err != nil {
		return false, err
	}
	return boolResp.Success, nil
}

func listResourceGroup(subscriptionId string, ctx context.Context, cred azcore.TokenCredential) ([]*armresources.ResourceGroup, error) {
	resourceGroupClient, err := armresources.NewResourceGroupsClient(subscriptionId, cred, nil)
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

func renderTable(results []analyzers.AzureServiceResult) string {
	rows := [][]string{}
	for _, r := range results {
		rows = append([][]string{
			{r.SubscriptionId, r.ResourceGroup, r.Location, r.Type, r.ServiceName, r.Sku, r.Sla, strconv.FormatBool(r.AvailabilityZones), strconv.FormatBool(r.PrivateEndpoints), strconv.FormatBool(r.DiagnosticSettings), strconv.FormatBool(r.CAFNaming)},
		}, rows...)
	}

	prettyPrintedTable, err := markdown.NewTableFormatterBuilder().
		WithPrettyPrint().
		Build("SubscriptionId", "ResourceGroup", "Location", "Type", "Name", "SKU", "SLA", "Zones", "P Endpoints", "Diag", "CAF").
		Format(rows)

	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("")
	fmt.Println(prettyPrintedTable)
	return prettyPrintedTable
}
