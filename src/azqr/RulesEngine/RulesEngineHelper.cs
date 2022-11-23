namespace azqr;

public static class RulesEngineHelper
{
    public static RulesEngine.RulesEngine LoadRulesEngine()
    {
        var allWorkflows = new List<Workflow>();
        var files = EmbeddedFilesHelper.GetTemplates("Rules");

        foreach (var kv in files)
        {
            var workflows = JsonConvert.DeserializeObject<List<Workflow>>(kv.Value)!.ToArray();
            allWorkflows.AddRange(workflows);
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