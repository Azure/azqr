// https://learn.microsoft.com/en-us/dotnet/azure/sdk/resource-management?tabs=dotnetcli
var credential = new DefaultAzureCredential();
var client = new ArmClient(credential);

var engine = LoadRulesEngine();

await Review(client, engine);

static async Task Review(ArmClient client, RulesEngine.RulesEngine engine)
{
    var results = new List<Results>();
    var subscription = await client.GetDefaultSubscriptionAsync();
    var subscriptionId = new ResourceIdentifier(subscription.Id);
    var resourceGroupCollection = subscription.GetResourceGroups();
    foreach (var resourceGroupResource in resourceGroupCollection)
    {
        var rgId = new ResourceIdentifier(resourceGroupResource.Id);

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
    WriteTable(results);
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
        var workflows = JsonConvert.DeserializeObject<List<Workflow>>(fileData).ToArray();
        allWorkflows.AddRange(workflows);
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

static void WriteTable(List<Results> results)
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

    table.Write(Format.MarkDown);
}