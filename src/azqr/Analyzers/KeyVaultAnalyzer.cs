namespace azqr;

public class KeyVaultAnalyzer : IAzureServiceAnalyzer
{
    KeyVaultData[] data;
    ArmClient client;
    string subscriptionId;
    string resourceGroup;

    public KeyVaultAnalyzer(ArmClient client, string subscriptionId, string resourceGroup, KeyVaultData[] data)
    {
        this.data = data;
        this.client = client;
        this.subscriptionId = subscriptionId;
        this.resourceGroup = resourceGroup;
    }

    public IEnumerable<AzureServiceResult> Review()
    {
        Console.WriteLine("Reviewing Key Vault...");
        foreach (var item in data)
        {
            var diagnostics = client.GetDiagnosticSettings(new ResourceIdentifier(item.Id!));
            var diagnosticsCount = diagnostics.Count();

            yield return new AzureServiceResult
            {
                SubscriptionId = subscriptionId,
                ResourceGroup = resourceGroup,
                ServiceName = item.Name,
                Sku = item.Properties.Sku.Name.ToString(),
                Sla = "99.99%",
                Type = item.ResourceType,
                AvaliabilityZones = "Yes",
                PrivateEndpoints = item.Properties.PrivateEndpointConnections.Count() > 0,
                DiagnosticSettings = diagnosticsCount > 0,
                CAFNaming = item.Name.StartsWith("kv")
            };
        }
    }
}