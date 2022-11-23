namespace azqr;

public class CosmosDbAnalyzer : IAzureServiceAnalyzer
{
    CosmosDBAccountData[] data;
    ArmClient client;
    string subscriptionId;
    string resourceGroup;

    public CosmosDbAnalyzer(ArmClient client, string subscriptionId, string resourceGroup, CosmosDBAccountData[] data)
    {
        this.data = data;
        this.client = client;
        this.subscriptionId = subscriptionId;
        this.resourceGroup = resourceGroup;
    }

    public IEnumerable<AzureServiceResult> Review()
    {
        Console.WriteLine("Reviewing Cosmos DB...");
        foreach (var item in data)
        {
            var diagnostics = client.GetDiagnosticSettings(new ResourceIdentifier(item.Id!));
            var diagnosticsCount = diagnostics.Count();

            yield return new AzureServiceResult
            {
                SubscriptionId = subscriptionId,
                ResourceGroup = resourceGroup,
                ServiceName = item.Name,
                Sku = item.CapacityTotalThroughputLimit.ToString()!,
                Sla = "TODO",
                Type = item.ResourceType,
                AvaliabilityZones = item.Locations.Count(c => c.IsZoneRedundant == true) > 0 ? "Yes" : "No",
                PrivateEndpoints = item.PrivateEndpointConnections.Count() > 0,
                DiagnosticSettings = diagnosticsCount > 0,
                CAFNaming = item.Name.StartsWith("cosmos")
            };
        }
    }
}