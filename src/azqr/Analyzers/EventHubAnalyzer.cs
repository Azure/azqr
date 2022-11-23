namespace azqr;

public class EventHubAnalyzer : IAzureServiceAnalyzer
{
    EventHubsNamespaceData[] data;
    ArmClient client;
    string subscriptionId;
    string resourceGroup;

    public EventHubAnalyzer(ArmClient client, string subscriptionId, string resourceGroup, EventHubsNamespaceData[] data)
    {
        this.data = data;
        this.client = client;
        this.subscriptionId = subscriptionId;
        this.resourceGroup = resourceGroup;
    }

    public IEnumerable<AzureServiceResult> Review()
    {
        Console.WriteLine("Reviewing Event Hub...");
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
                Sla = item.Sku.Name.ToString().Contains("Basic") || item.Sku.Name.ToString().Contains("Standard") ? "99.95%" : "99.99%",
                Type = item.ResourceType,
                AvaliabilityZones = item.ZoneRedundant == true ? "Yes" : "No",
                PrivateEndpoints = item.PrivateEndpointConnections.Count() > 0,
                DiagnosticSettings = diagnosticsCount > 0,
                CAFNaming = item.Name.StartsWith("evh")
            };
        }
    }
}