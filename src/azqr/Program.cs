var rootCommand = new RootCommand("Azure Quick Review");
var subscriptionOption = new System.CommandLine.Option<string>(
    new string[] { "--subscriptionId", "-s" },
    "Id of the subscription to review.");
rootCommand.AddOption(subscriptionOption);

var resourceGroupOption = new System.CommandLine.Option<string>(
    new string[] { "--resource-group", "-g" },
    "Name of the resource group to review.");
rootCommand.AddOption(resourceGroupOption);

var customerOption = new System.CommandLine.Option<string>(
    new string[] { "--customer", "-c" },
    () => "<Replace with Customer Name>",
    "Name of the customer.");
rootCommand.AddOption(customerOption);

rootCommand.SetHandler<string, string, string>(async (subscriptionId, resourceGroup, customerName) =>
    {
        var credential = new DefaultAzureCredential();
        var client = new ArmClient(credential, subscriptionId);

        await Review(client, customerName, resourceGroup);
    },
    subscriptionOption,
    resourceGroupOption,
    customerOption);

return await rootCommand.InvokeAsync(args);

static async Task Review(ArmClient client, string customerName, string resourceGroup)
{
    // https://learn.microsoft.com/en-us/dotnet/azure/sdk/resource-management?tabs=dotnetcli
    var results = new List<AzureServiceResult>();
    var subscription = await client.GetDefaultSubscriptionAsync();
    var subscriptionId = new ResourceIdentifier(subscription.Id!);
    var resourceGroupCollection = subscription.GetResourceGroups();

    if (string.IsNullOrEmpty(resourceGroup))
    {
        await foreach (var rg in resourceGroupCollection.GetAllAsync())
        {
            var resourceGroupResult = ReviewResourceGroup(client, subscriptionId, rg);
            results.AddRange(resourceGroupResult);
        }
    }
    else
    {
        var rg = await resourceGroupCollection.GetAsync(resourceGroup);
        var resourceGroupResult = ReviewResourceGroup(client, subscriptionId, rg);
        results.AddRange(resourceGroupResult);
    }

    var reportTemplate = EmbeddedFilesHelper.GetTemplate("Resources.Report.md");
    var resultsTable = WriteTable(results);

    var report = reportTemplate
        .Replace("{{date}}", $"{CultureInfo.CurrentCulture.DateTimeFormat.GetMonthName(DateTime.Now.Month)} {DateTime.Now.Year.ToString()}")
        .Replace("{{customer}}", customerName)
        .Replace("{{results}}", resultsTable)
        .Replace("{{recommendations}}", EmbeddedFilesHelper.GetRecommendations(results));

    await File.WriteAllTextAsync("Report.md", report);

    Console.WriteLine("Review completed!");
}

static List<AzureServiceResult> ReviewResourceGroup(ArmClient client, ResourceIdentifier subscriptionId, ResourceGroupResource resourceGroupResource)
{
    var analyzers = new List<IAzureServiceAnalyzer>();
    var results = new List<AzureServiceResult>();

    var rgId = new ResourceIdentifier(resourceGroupResource.Id!);

    Console.WriteLine($"Reviewing Resource Group: {rgId}...");

    analyzers.Add(new StorageAccountAnalyzer(
        client,
        subscriptionId.Name,
        rgId.Name,
        resourceGroupResource.GetStorageAccounts().Select(x => x.Data).ToArray()));

    analyzers.Add(new CosmosDbAnalyzer(
        client,
        subscriptionId.Name,
        rgId.Name,
        resourceGroupResource.GetCosmosDBAccounts().Select(x => x.Data).ToArray()));

    analyzers.Add(new KeyVaultAnalyzer(
        client,
        subscriptionId.Name,
        rgId.Name,
        resourceGroupResource.GetKeyVaults().Select(x => x.Data).ToArray()));

    analyzers.Add(new RedisAnalyzer(
        client,
        subscriptionId.Name,
        rgId.Name,
        resourceGroupResource.GetAllRedis().Select(x => x.Data).ToArray()));

    analyzers.Add(new ApimAnalyzer(
        client,
        subscriptionId.Name,
        rgId.Name,
        resourceGroupResource.GetApiManagementServices().Select(x => x.Data).ToArray()));

    analyzers.Add(new ContainerRegistryAnalyzer(
        client,
        subscriptionId.Name,
        rgId.Name,
        resourceGroupResource.GetContainerRegistries().Select(x => x.Data).ToArray()));

    analyzers.Add(new AksAnalyzer(
        client,
        subscriptionId.Name,
        rgId.Name,
        resourceGroupResource.GetContainerServiceManagedClusters().Select(x => x.Data).ToArray()));

    analyzers.Add(new ContainerAppEnvironmentAnalyzer(
        client,
        subscriptionId.Name,
        rgId.Name,
        resourceGroupResource.GetManagedEnvironments().Select(x => x.Data).ToArray()));

    analyzers.Add(new ContainerInstanceAnalyzer(
        client,
        subscriptionId.Name,
        rgId.Name,
        resourceGroupResource.GetContainerGroups().Select(x => x.Data).ToArray()));

    analyzers.Add(new SignalRAnalyzer(
        client,
        subscriptionId.Name,
        rgId.Name,
        resourceGroupResource.GetSignalRs().Select(x => x.Data).ToArray()));

    analyzers.Add(new ServiceBusAnalyzer(
        client,
        subscriptionId.Name,
        rgId.Name,
        resourceGroupResource.GetServiceBusNamespaces().Select(x => x.Data).ToArray()));

    analyzers.Add(new EventHubAnalyzer(
        client,
        subscriptionId.Name,
        rgId.Name,
        resourceGroupResource.GetEventHubsNamespaces().Select(x => x.Data).ToArray()));

    analyzers.Add(new EventGridAnalyzer(
        client,
        subscriptionId.Name,
        rgId.Name,
        resourceGroupResource.GetDomains().Select(x => x.Data).ToArray()));

    analyzers.Add(new ApplicationGatewayAnalyzer(
        client,
        subscriptionId.Name,
        rgId.Name,
        resourceGroupResource.GetApplicationGateways().Select(x => x.Data).ToArray()));

    analyzers.Add(new AppServicePlanAnalyzer(
        client,
        subscriptionId.Name,
        rgId.Name,
        resourceGroupResource.GetAppServicePlans().Select(x => x.Data).ToArray()));

    foreach (var analyzer in analyzers)
    {
        results.AddRange(analyzer.Review());
    }

    return results;
}

static string WriteTable(List<AzureServiceResult> results)
{
    var table = new ConsoleTable(
        ColumnNames.SubscriptionId,
        ColumnNames.ResourceGroup,
        ColumnNames.Type,
        ColumnNames.ServiceName,
        ColumnNames.SKU,
        ColumnNames.AvaliabilityZones,
        ColumnNames.SLA,
        ColumnNames.PrivateEndpoints,
        ColumnNames.DiagnosticSettings,
        ColumnNames.CAFNaming);

    foreach (var result in results)
    {
        table.AddRow(
            result.SubscriptionId,
            result.ResourceGroup,
            result.Type,
            result.ServiceName,
            result.Sku,
            result.AvaliabilityZones,
            result.Sla,
            result.PrivateEndpoints,
            result.DiagnosticSettings,
            result.CAFNaming);
    }

    return table.ToMarkDownString();
}
