namespace azqr;

public class AppServicePlanAnalyzer : IAzureServiceAnalyzer
{
    AppServicePlanData[] data;
    ArmClient client;
    string subscriptionId;
    string resourceGroup;

    public AppServicePlanAnalyzer(ArmClient client, string subscriptionId, string resourceGroup, AppServicePlanData[] data)
    {
        this.data = data;
        this.client = client;
        this.subscriptionId = subscriptionId;
        this.resourceGroup = resourceGroup;
    }

    public IEnumerable<AzureServiceResult> Review()
    {
        Console.WriteLine("Reviewing App Service Plan...");
        foreach (var item in data)
        {
            var diagnostics = client.GetDiagnosticSettings(new ResourceIdentifier(item.Id!));
            var diagnosticsCount = diagnostics.Count();

            yield return new AzureServiceResult
            {
                SubscriptionId = subscriptionId,
                ResourceGroup = resourceGroup,
                ServiceName = item.Name,
                Sku = item.Sku.Name.ToString()!,
                Sla = "TODO",
                Type = item.ResourceType,
                AvaliabilityZones = item.IsZoneRedundant == true ? "Yes" : "No",
                PrivateEndpoints = false,
                DiagnosticSettings = diagnosticsCount > 0,
                CAFNaming = item.Name.StartsWith("plan")
            };
        }
    }
}