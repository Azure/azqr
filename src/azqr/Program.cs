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
        results.AddRange(await ExecuteRules(engine, subscriptionId.Name, rgId.Name, "Storage", storageAccounts));

        var cosmosAccounts = resourceGroupResource.GetCosmosDBAccounts().Select(x => x.Data).ToArray();
        results.AddRange(await ExecuteRules(engine, subscriptionId.Name, rgId.Name, "CosmosDB", cosmosAccounts));

        var keyVaults = resourceGroupResource.GetKeyVaults().Select(x => x.Data).ToArray();
        results.AddRange(await ExecuteRules(engine, subscriptionId.Name, rgId.Name, "KeyVault", keyVaults));

        var appServices = resourceGroupResource.GetAppServicePlans().Select(x => x.Data).ToArray();
        results.AddRange(await ExecuteRules(engine, subscriptionId.Name, rgId.Name, "AppServicePlan", appServices));

        var redis = resourceGroupResource.GetAllRedis().Select(x => x.Data).ToArray();
        results.AddRange(await ExecuteRules(engine, subscriptionId.Name, rgId.Name, "Redis", redis));

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
            results.Add(new Results
            {
                SubscriptionId = subscriptionId,
                ResourceGroup = resourceGroup,
                Type = svc.ResourceType,
                ServiceName = svc.Name,
                RulesResults = await engine.ExecuteAllRulesAsync(workflowName, svc)
            });
        }
    }
    return results;
}

static string WriteTable(List<Results> results)
{
    var table = new ConsoleTable("SubscriptionId", "Resource Group", "Type", "Service Name", "Rule Name", "Result");
    foreach (var result in results)
    {
        foreach (var ruleResult in result.RulesResults)
        {
            table.AddRow(
                result.SubscriptionId,
                result.ResourceGroup,
                result.Type,
                result.ServiceName,
                ruleResult.Rule.RuleName,
                ruleResult.ActionResult.Output);
        }
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