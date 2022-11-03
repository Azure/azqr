// https://learn.microsoft.com/en-us/dotnet/azure/sdk/resource-management?tabs=dotnetcli
var credential = new DefaultAzureCredential();
var client = new ArmClient(credential);

var engine = LoadRulesEngine();

await Review(client, engine);

static async Task Review(ArmClient client, RulesEngine.RulesEngine engine)
{
    var results = new List<Results>();
    var subscription = await client.GetDefaultSubscriptionAsync();
    var subscriptionId = new ResourceIdentifier(subscription.Id!);
    var resourceGroupCollection = subscription.GetResourceGroups();
    foreach (var resourceGroupResource in resourceGroupCollection)
    {
        var rgId = new ResourceIdentifier(resourceGroupResource.Id!);

        var storageAccounts = resourceGroupResource.GetStorageAccounts().Select(x => x.Data).ToArray();
        results.AddRange(await ExecuteRules(client, engine, subscriptionId.Name, rgId.Name, "Storage", storageAccounts));

        var cosmosAccounts = resourceGroupResource.GetCosmosDBAccounts().Select(x => x.Data).ToArray();
        results.AddRange(await ExecuteRules(client, engine, subscriptionId.Name, rgId.Name, "CosmosDB", cosmosAccounts));

        var keyVaults = resourceGroupResource.GetKeyVaults().Select(x => x.Data).ToArray();
        results.AddRange(await ExecuteRules(client, engine, subscriptionId.Name, rgId.Name, "KeyVault", keyVaults));

        var plans = resourceGroupResource.GetAppServicePlans().Select(x => x.Data).ToArray();
        results.AddRange(await ExecuteRules(client, engine, subscriptionId.Name, rgId.Name, "AppServicePlan", plans));

        var redis = resourceGroupResource.GetAllRedis().Select(x => x.Data).ToArray();
        results.AddRange(await ExecuteRules(client, engine, subscriptionId.Name, rgId.Name, "Redis", redis));

        var apims = resourceGroupResource.GetApiManagementServices().Select(x => x.Data).ToArray();
        results.AddRange(await ExecuteRules(client, engine, subscriptionId.Name, rgId.Name, "ApiManagement", apims));

        var acrs = resourceGroupResource.GetContainerRegistries().Select(x => x.Data).ToArray();
        results.AddRange(await ExecuteRules(client, engine, subscriptionId.Name, rgId.Name, "ContainerRegistry", acrs));

        var aks = resourceGroupResource.GetContainerServiceManagedClusters().Select(x => x.Data).ToArray();
        results.AddRange(await ExecuteRules(client, engine, subscriptionId.Name, rgId.Name, "AKS", aks));

        var signalR = resourceGroupResource.GetSignalRs().Select(x => x.Data).ToArray();
        results.AddRange(await ExecuteRules(client, engine, subscriptionId.Name, rgId.Name, "SignalR", signalR));

        var serviceBusNamespaces = resourceGroupResource.GetServiceBusNamespaces().Select(x => x.Data).ToArray();
        results.AddRange(await ExecuteRules(client, engine, subscriptionId.Name, rgId.Name, "ServiceBus", serviceBusNamespaces));

        var eventGrids = resourceGroupResource.GetDomains().Select(x => x.Data).ToArray();
        results.AddRange(await ExecuteRules(client, engine, subscriptionId.Name, rgId.Name, "EventGrid", eventGrids));

        var applicationGateways = resourceGroupResource.GetApplicationGateways().Select(x => x.Data).ToArray();
        results.AddRange(await ExecuteNetworkRules(client, engine, subscriptionId.Name, rgId.Name, "ApplicationGateway", applicationGateways));
    }

    var reportTemplate = GetTemplate("Resources.Report.md");
    var resultsTable = WriteTable(results);

    var customer = "Contoso";

    var report = reportTemplate.Replace("{{date}}", $"{CultureInfo.CurrentCulture.DateTimeFormat.GetMonthName(DateTime.Now.Month)} {DateTime.Now.Year.ToString()}");
    report = report.Replace("{{customer}}", customer);
    report = report.Replace("{{results}}", resultsTable);
    report = report.Replace("{{recommendations}}", GetRecommendations(results));

    await File.WriteAllTextAsync("Report.md", report);
}

static RulesEngine.RulesEngine LoadRulesEngine()
{
    var allWorkflows = new List<Workflow>();
    var files = Directory.GetFiles(Directory.GetCurrentDirectory(), "*.azqr.json", SearchOption.AllDirectories);
    if (files == null || files.Length == 0)
        throw new Exception("Rules not found.");

    foreach (var file in files)
    {
        var fileData = File.ReadAllText(file);
        if (fileData != null)
        {
            var workflows = JsonConvert.DeserializeObject<List<Workflow>>(fileData)!.ToArray();
            allWorkflows.AddRange(workflows);
        }
    }

    return new RulesEngine.RulesEngine(allWorkflows.ToArray(), null);
}

static async ValueTask<List<Results>> ExecuteRules(
    ArmClient client,
    RulesEngine.RulesEngine engine,
    string subscriptionId,
    string resourceGroup,
    string workflowName,
    ResourceData[] services)
{
    var results = new List<Results>();
    if (engine.ContainsWorkflow(workflowName))
    {
        foreach (var svc in services)
        {
            var diagnostics = client.GetDiagnosticSettings(new ResourceIdentifier(svc.Id!));
            var diagnosticsCount = diagnostics.Count();

            results.Add(new Results
            {
                SubscriptionId = subscriptionId,
                ResourceGroup = resourceGroup,
                Type = svc.ResourceType,
                ServiceName = svc.Name,
                RulesResults = await engine.ExecuteAllRulesAsync(workflowName, svc, diagnosticsCount)
            });
        }
    }
    return results;
}

static async ValueTask<List<Results>> ExecuteNetworkRules(
    ArmClient client,
    RulesEngine.RulesEngine engine,
    string subscriptionId,
    string resourceGroup,
    string workflowName,
    NetworkTrackedResourceData[] services)
{
    var results = new List<Results>();
    if (engine.ContainsWorkflow(workflowName))
    {
        foreach (var svc in services)
        {
            var diagnostics = client.GetDiagnosticSettings(new ResourceIdentifier(svc.Id!));
            var diagnosticsCount = diagnostics.Count();

            results.Add(new Results
            {
                SubscriptionId = subscriptionId,
                ResourceGroup = resourceGroup,
                Type = svc.ResourceType!,
                ServiceName = svc.Name,
                RulesResults = await engine.ExecuteAllRulesAsync(workflowName, svc, diagnosticsCount)
            });
        }
    }
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
            result.RulesResults.FirstOrDefault(x => x.Rule.RuleName ==  ColumnNames.SKU)?.ActionResult.Output,
            result.RulesResults.FirstOrDefault(x => x.Rule.RuleName ==  ColumnNames.AvaliabilityZones)?.ActionResult.Output,
            result.RulesResults.FirstOrDefault(x => x.Rule.RuleName ==  ColumnNames.SLA)?.ActionResult.Output,
            result.RulesResults.FirstOrDefault(x => x.Rule.RuleName ==  ColumnNames.PrivateEndpoints)?.ActionResult.Output,
            result.RulesResults.FirstOrDefault(x => x.Rule.RuleName ==  ColumnNames.DiagnosticSettings)?.ActionResult.Output,
            result.RulesResults.FirstOrDefault(x => x.Rule.RuleName ==  ColumnNames.CAFNaming)?.ActionResult.Output);
    }

    return table.ToMarkDownString();
}

static string GetTemplate(string templateName)
{
    var embeddedProvider = new EmbeddedFileProvider(Assembly.GetExecutingAssembly());
    var fileInfo = embeddedProvider.GetFileInfo(templateName);
    if (fileInfo == null || !fileInfo.Exists)
        return string.Empty;

    using (var stream = fileInfo.CreateReadStream())
    {
        using (var reader = new StreamReader(stream))
        {
            return reader.ReadToEnd();
        }
    }
}

static string GetRecommendations(List<Results> results)
{
    var recommendations = string.Empty;
    var types = results.Select(x => x.Type).Distinct();
    foreach (var type in types)
    {
        var recommendationsTemplate = GetTemplate($"Resources.{type.Replace("/", ".")}.md");
        recommendations += recommendationsTemplate + Environment.NewLine;
    }
    return recommendations;
}