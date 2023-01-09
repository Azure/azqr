package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/cmendible/azqr/internal/analyzers"
	"github.com/cmendible/azqr/internal/report/templates"
	"github.com/fbiville/markdown-table-formatter/pkg/markdown"
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
	outputPtr := flag.String("o", "report.md", "Output file")
	customerPtr := flag.String("c", "<Replace with Customer Name>", "Customer Name")
	concurrency := flag.Int("p", defaultConcurrency, fmt.Sprintf("Parallel processes. Default to %d. A < 0 value will use the maxmimum concurrency.", defaultConcurrency))
	ver := flag.Bool("v", false, "Print version and exit")

	flag.Parse()

	subscriptionID := *subscriptionPtr
	resourceGroupName := *resourceGroupPtr
	outputFile := *outputPtr
	customer := *customerPtr

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

	svcanalyzers := []analyzers.AzureServiceAnalyzer{
		analyzers.NewAKSAnalyzer(ctx, subscriptionID, cred),
		analyzers.NewAPIManagementAnalyzer(ctx, subscriptionID, cred),
		analyzers.NewApplicationGatewayAnalyzer(ctx, subscriptionID, cred),
		analyzers.NewContainerAppsAnalyzer(ctx, subscriptionID, cred),
		analyzers.NewContainerIntanceAnalyzer(ctx, subscriptionID, cred),
		analyzers.NewCosmosDBAnalyzer(ctx, subscriptionID, cred),
		analyzers.NewContainerRegistryAnalyzer(ctx, subscriptionID, cred),
		analyzers.NewEventHubAnalyzer(ctx, subscriptionID, cred),
		analyzers.NewEventGridAnalyzer(ctx, subscriptionID, cred),
		analyzers.NewKeyVaultAnalyzer(ctx, subscriptionID, cred),
		analyzers.NewAppServiceAnalyzer(ctx, subscriptionID, cred),
		analyzers.NewRedisAnalyzer(ctx, subscriptionID, cred),
		analyzers.NewServiceBusAnalyzer(ctx, subscriptionID, cred),
		analyzers.NewSignalRAnalyzer(ctx, subscriptionID, cred),
		analyzers.NewStorageAnalyzer(ctx, subscriptionID, cred),
		analyzers.NewPostgreAnalyzer(ctx, subscriptionID, cred),
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	var all []analyzers.IAzureServiceResult
	rc := ReviewContext{
		Ctx:   ctx,
		ResCh: make(chan []analyzers.IAzureServiceResult),
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

	var allFunctions []analyzers.IAzureServiceResult
	for _, r := range all {
		v, ok := r.(analyzers.AzureFunctionAppResult)
		if ok {
			allFunctions = append(allFunctions, v)
		}
	}

	reportTemplate := templates.GetTemplates("Report.md")
	reportTemplate = strings.Replace(reportTemplate, "{{results}}", resultsTable, 1)
	reportTemplate = strings.Replace(reportTemplate, "{{date}}", time.Now().Format("2006-01-02"), 1)
	reportTemplate = strings.Replace(reportTemplate, "{{customer}}", customer, -1)

	recommendations := ""
	dict := map[string]bool{}
	for _, r := range all {
		parsedType := strings.Replace(r.GetResourceType(), "/", ".", -1)
		if _, ok := dict[r.GetResourceType()]; !ok {
			dict[r.GetResourceType()] = true
			recommendations += "\n\n"
			recommendations += templates.GetTemplates(fmt.Sprintf("%s.md", parsedType))

			if r.GetResourceType() == "Microsoft.Web/serverfarms/sites" && len(allFunctions) > 0 {
				recommendations = strings.Replace(recommendations, "{{functions}}", renderDetailsTable(allFunctions), 1)
			}
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
	ResCh chan []analyzers.IAzureServiceResult
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
func waitForReviews(rc *ReviewContext, nb int) (*[]analyzers.IAzureServiceResult, error) {
	received := 0
	reviews := make([]analyzers.IAzureServiceResult, 0)
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

func renderTable(results []analyzers.IAzureServiceResult) string {
	heathers := results[0].GetProperties()

	rows := [][]string{}
	for _, r := range results {
		rows = append(mapToRow(heathers, r.ToMap()), rows...)
	}

	prettyPrintedTable, err := markdown.NewTableFormatterBuilder().
		WithPrettyPrint().
		Build(heathers...).
		Format(rows)

	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("")
	fmt.Println(prettyPrintedTable)
	return prettyPrintedTable
}

func renderDetailsTable(results []analyzers.IAzureServiceResult) string {
	heathers := results[0].GetDetailProperties()

	rows := [][]string{}
	for _, r := range results {
		rows = append(mapToRow(heathers, r.ToDetail()), rows...)
	}

	prettyPrintedTable, err := markdown.NewTableFormatterBuilder().
		WithPrettyPrint().
		Build(heathers...).
		Format(rows)

	if err != nil {
		log.Fatal(err)
	}

	return prettyPrintedTable
}

func mapToRow(heathers []string, m map[string]string) [][]string {
	v := make([]string, 0, len(m))

	for _, k := range heathers {
		v = append(v, m[k])
	}

	return [][]string{v}
}
