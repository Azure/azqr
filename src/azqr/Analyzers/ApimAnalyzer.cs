namespace azqr;

public class ApimAnalyzer : IAzureServiceAnalyzer
{
    ApiManagementServiceData[] data;
    ArmClient client;
    string subscriptionId;
    string resourceGroup;

    public ApimAnalyzer(ArmClient client, string subscriptionId, string resourceGroup, ApiManagementServiceData[] data)
    {
        this.data = data;
        this.client = client;
        this.subscriptionId = subscriptionId;
        this.resourceGroup = resourceGroup;
    }

    public IEnumerable<AzureServiceResult> Review()
    {
        Console.WriteLine("Reviewing APIM...");
        foreach (var item in data)
        {
            var diagnostics = client.GetDiagnosticSettings(new ResourceIdentifier(item.Id!));
            var diagnosticsCount = diagnostics.Count();

            yield return new AzureServiceResult
            {
                SubscriptionId = subscriptionId,
                ResourceGroup = resourceGroup,
                ServiceName = item.Name,
                Sku = item.Sku.Name.ToString(),
                Sla = "TODO",
                Type = item.ResourceType,
                AvaliabilityZones = item.Zones.Count() > 0 ? "Yes" : "No",
                PrivateEndpoints = item.PrivateEndpointConnections.Count() > 0,
                DiagnosticSettings = diagnosticsCount > 0,
                CAFNaming = item.Name.StartsWith("apim")
            };
        }
    }
}