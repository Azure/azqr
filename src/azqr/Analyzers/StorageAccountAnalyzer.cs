namespace azqr;

public class StorageAccountAnalyzer : IAzureServiceAnalyzer
{
    StorageAccountData[] data;
    ArmClient client;
    string subscriptionId;
    string resourceGroup;

    public StorageAccountAnalyzer(ArmClient client, string subscriptionId, string resourceGroup, StorageAccountData[] data)
    {
        this.data = data;
        this.client = client;
        this.subscriptionId = subscriptionId;
        this.resourceGroup = resourceGroup;
    }

    public IEnumerable<AzureServiceResult> Review()
    {
        Console.WriteLine("Reviewing Storage Accounts...");
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
                Sla = item.AccessTier.ToString() == "Hot" && item.Sku.Name.ToString().Contains("RAGRS") ? "99.99%" : "99.9%",
                Type = item.ResourceType,
                AvaliabilityZones = item.Sku.Name.ToString().Contains("ZRS") ? "Yes" : "No",
                PrivateEndpoints = item.PrivateEndpointConnections.Count() > 0,
                DiagnosticSettings = diagnosticsCount > 0,
                CAFNaming = item.Name.StartsWith("st")
            };
        }
    }
}