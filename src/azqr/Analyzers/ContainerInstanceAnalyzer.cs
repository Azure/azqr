namespace azqr;

public class ContainerInstanceAnalyzer : IAzureServiceAnalyzer
{
    ContainerGroupData[] data;
    ArmClient client;
    string subscriptionId;
    string resourceGroup;

    public ContainerInstanceAnalyzer(ArmClient client, string subscriptionId, string resourceGroup, ContainerGroupData[] data)
    {
        this.data = data;
        this.client = client;
        this.subscriptionId = subscriptionId;
        this.resourceGroup = resourceGroup;
    }

    public IEnumerable<AzureServiceResult> Review()
    {
        Console.WriteLine("Reviewing Container Instance...");
        foreach (var item in data)
        {
            var diagnostics = client.GetDiagnosticSettings(new ResourceIdentifier(item.Id!));
            var diagnosticsCount = diagnostics.Count();

            yield return new AzureServiceResult
            {
                SubscriptionId = subscriptionId,
                ResourceGroup = resourceGroup,
                ServiceName = item.Name,
                Sku = item.Sku.ToString()!,
                Sla = "99.9%",
                Type = item.ResourceType,
                AvaliabilityZones = item.Zones.Count() > 0 ? "Yes" : "No",
                PrivateEndpoints = item.SubnetIds.Count() > 0,
                DiagnosticSettings = diagnosticsCount > 0,
                CAFNaming = item.Name.StartsWith("ci")
            };
        }
    }
}