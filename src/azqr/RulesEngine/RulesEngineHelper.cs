namespace azqr;

public static class RulesEngineHelper
{
    public static RulesEngine.RulesEngine LoadRulesEngine()
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

    public static async ValueTask<List<Results>> ExecuteRules(
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

    public static async ValueTask<List<Results>> ExecuteNetworkRules(
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
}