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

        var engine = RulesEngineHelper.LoadRulesEngine();

        await Review(client, engine, customerName, resourceGroup);
    },
    subscriptionOption,
    resourceGroupOption,
    customerOption);

return await rootCommand.InvokeAsync(args);

static async Task Review(ArmClient client, RulesEngine.RulesEngine engine, string customerName, string resourceGroup)
{
    // https://learn.microsoft.com/en-us/dotnet/azure/sdk/resource-management?tabs=dotnetcli
    var results = new List<Results>();
    var subscription = await client.GetDefaultSubscriptionAsync();
    var subscriptionId = new ResourceIdentifier(subscription.Id!);
    var resourceGroupCollection = subscription.GetResourceGroups();

    if (string.IsNullOrEmpty(resourceGroup))
    {
        await foreach (var rg in resourceGroupCollection.GetAllAsync())
        {
            var resourceGroupResult = await ReviewResourceGroup(client, engine, subscriptionId, rg);
            results.AddRange(resourceGroupResult);
        }
    }
    else
    {
        var rg = await resourceGroupCollection.GetAsync(resourceGroup);
        var resourceGroupResult = await ReviewResourceGroup(client, engine, subscriptionId, rg);
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

static async Task<List<Results>> ReviewResourceGroup(ArmClient client, RulesEngine.RulesEngine engine, ResourceIdentifier subscriptionId, ResourceGroupResource resourceGroupResource)
{
    var results = new List<Results>();

    var rgId = new ResourceIdentifier(resourceGroupResource.Id!);

    Console.WriteLine($"Reviewing Subscription Id: {subscriptionId} and Resource Group: {rgId.Name}...");

    var storageAccounts = resourceGroupResource.GetStorageAccounts().Select(x => x.Data).ToArray();
    results.AddRange(await RulesEngineHelper.ExecuteRules(client, engine, subscriptionId.Name, rgId.Name, "Storage", storageAccounts));

    var cosmosAccounts = resourceGroupResource.GetCosmosDBAccounts().Select(x => x.Data).ToArray();
    results.AddRange(await RulesEngineHelper.ExecuteRules(client, engine, subscriptionId.Name, rgId.Name, "CosmosDB", cosmosAccounts));

    var keyVaults = resourceGroupResource.GetKeyVaults().Select(x => x.Data).ToArray();
    results.AddRange(await RulesEngineHelper.ExecuteRules(client, engine, subscriptionId.Name, rgId.Name, "KeyVault", keyVaults));

    var plans = resourceGroupResource.GetAppServicePlans().Select(x => x.Data).ToArray();
    results.AddRange(await RulesEngineHelper.ExecuteRules(client, engine, subscriptionId.Name, rgId.Name, "AppServicePlan", plans));

    var redis = resourceGroupResource.GetAllRedis().Select(x => x.Data).ToArray();
    results.AddRange(await RulesEngineHelper.ExecuteRules(client, engine, subscriptionId.Name, rgId.Name, "Redis", redis));

    var apims = resourceGroupResource.GetApiManagementServices().Select(x => x.Data).ToArray();
    results.AddRange(await RulesEngineHelper.ExecuteRules(client, engine, subscriptionId.Name, rgId.Name, "ApiManagement", apims));

    var acrs = resourceGroupResource.GetContainerRegistries().Select(x => x.Data).ToArray();
    results.AddRange(await RulesEngineHelper.ExecuteRules(client, engine, subscriptionId.Name, rgId.Name, "ContainerRegistry", acrs));

    var aks = resourceGroupResource.GetContainerServiceManagedClusters().Select(x => x.Data).ToArray();
    results.AddRange(await RulesEngineHelper.ExecuteRules(client, engine, subscriptionId.Name, rgId.Name, "AKS", aks));

    var ci = resourceGroupResource.GetContainerGroups().Select(x => x.Data).ToArray();
    results.AddRange(await RulesEngineHelper.ExecuteRules(client, engine, subscriptionId.Name, rgId.Name, "ContainerInstance", ci));

    var signalR = resourceGroupResource.GetSignalRs().Select(x => x.Data).ToArray();
    results.AddRange(await RulesEngineHelper.ExecuteRules(client, engine, subscriptionId.Name, rgId.Name, "SignalR", signalR));

    var serviceBusNamespaces = resourceGroupResource.GetServiceBusNamespaces().Select(x => x.Data).ToArray();
    results.AddRange(await RulesEngineHelper.ExecuteRules(client, engine, subscriptionId.Name, rgId.Name, "ServiceBus", serviceBusNamespaces));

    var evenHubs = resourceGroupResource.GetEventHubsNamespaces().Select(x => x.Data).ToArray();
    results.AddRange(await RulesEngineHelper.ExecuteRules(client, engine, subscriptionId.Name, rgId.Name, "EventHubs", evenHubs));

    var eventGrids = resourceGroupResource.GetDomains().Select(x => x.Data).ToArray();
    results.AddRange(await RulesEngineHelper.ExecuteRules(client, engine, subscriptionId.Name, rgId.Name, "EventGrid", eventGrids));

    var applicationGateways = resourceGroupResource.GetApplicationGateways().Select(x => x.Data).ToArray();
    results.AddRange(await RulesEngineHelper.ExecuteNetworkRules(client, engine, subscriptionId.Name, rgId.Name, "ApplicationGateway", applicationGateways));

    return results;
}

static string WriteTable(List<Results> results)
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
            result.RulesResults.FirstOrDefault(x => x.Rule.RuleName == ColumnNames.SKU)?.ActionResult.Output,
            result.RulesResults.FirstOrDefault(x => x.Rule.RuleName == ColumnNames.AvaliabilityZones)?.ActionResult.Output,
            result.RulesResults.FirstOrDefault(x => x.Rule.RuleName == ColumnNames.SLA)?.ActionResult.Output,
            result.RulesResults.FirstOrDefault(x => x.Rule.RuleName == ColumnNames.PrivateEndpoints)?.ActionResult.Output,
            result.RulesResults.FirstOrDefault(x => x.Rule.RuleName == ColumnNames.DiagnosticSettings)?.ActionResult.Output,
            result.RulesResults.FirstOrDefault(x => x.Rule.RuleName == ColumnNames.CAFNaming)?.ActionResult.Output);
    }

    return table.ToMarkDownString();
}
