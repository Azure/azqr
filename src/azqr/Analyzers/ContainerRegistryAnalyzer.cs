namespace azqr;

public class ContainerRegistryAnalyzer : IAzureServiceAnalyzer
{
    ContainerRegistryData[] data;
    ArmClient client;
    string subscriptionId;
    string resourceGroup;

    public ContainerRegistryAnalyzer(ArmClient client, string subscriptionId, string resourceGroup, ContainerRegistryData[] data)
    {
        this.data = data;
        this.client = client;
        this.subscriptionId = subscriptionId;
        this.resourceGroup = resourceGroup;
    }

    public IEnumerable<AzureServiceResult> Review()
    {
        Console.WriteLine("Reviewing Container Registry...");
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
                Sla = "99.95%",
                Type = item.ResourceType,
                AvaliabilityZones = item.ZoneRedundancy == ContainerRegistryZoneRedundancy.Enabled ? "Yes" : "No",
                PrivateEndpoints = item.PrivateEndpointConnections.Count() > 0,
                DiagnosticSettings = diagnosticsCount > 0,
                CAFNaming = item.Name.StartsWith("cr")
            };
        }
    }
}