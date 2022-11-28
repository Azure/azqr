package main

import (
	"context"
	"flag"
	"log"
	"os"
	"strconv"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/cmendibl3/azqr/cmd/azqr/analyzers"
	"github.com/olekukonko/tablewriter"
)

func main() {
	subscriptionPtr := flag.String("s", "", "Azure Subscription Id (Required)")
	resourceGroupPtr := flag.String("g", "", "Azure Resource Group")

	flag.Parse()

	subscriptionId := *subscriptionPtr
	resourceGroupName := *resourceGroupPtr

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
		analyzers.NewRedisAnalyzer(subscriptionId, ctx, cred),
		analyzers.NewServiceBusAnalyzer(subscriptionId, ctx, cred),
		analyzers.NewSignalRAnalyzer(subscriptionId, ctx, cred),
		analyzers.NewStorageAnalyzer(subscriptionId, ctx, cred),
	}

	all := make([]analyzers.AzureServiceResult, 0)
	for _, r := range resourceGroups {
		log.Printf("Analyzing Resource Group %s", r)
		for _, a := range svcanalyzers {
			results, err := a.Review(r)
			if err != nil {
				log.Fatal(err)
			}
			all = append(all, results...)
		}
	}

	renderTable(all)
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

func renderTable(results []analyzers.AzureServiceResult) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"SubscriptionId", "ResourceGroup", "ServiceName", "Sku", "Sla", "Type", "AvailabilityZones", "PrivateEndpoints", "DiagnosticSettings", "CAFNaming"})
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetHeaderLine(false)
	table.SetBorder(false)
	table.SetTablePadding("\t")
	table.SetNoWhiteSpace(true)

	for _, r := range results {
		table.Append([]string{r.SubscriptionId, r.ResourceGroup, r.ServiceName, r.Sku, r.Sla, r.Type, strconv.FormatBool(r.AvailabilityZones), strconv.FormatBool(r.PrivateEndpoints), strconv.FormatBool(r.DiagnosticSettings), strconv.FormatBool(r.CAFNaming)})
	}
	table.Render()
}
