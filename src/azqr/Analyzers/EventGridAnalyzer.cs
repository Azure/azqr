namespace azqr;

public class EventGridAnalyzer : IAzureServiceAnalyzer
{
    DomainData[] data;
    ArmClient client;
    string subscriptionId;
    string resourceGroup;

    public EventGridAnalyzer(ArmClient client, string subscriptionId, string resourceGroup, DomainData[] data)
    {
        this.data = data;
        this.client = client;
        this.subscriptionId = subscriptionId;
        this.resourceGroup = resourceGroup;
    }

    public IEnumerable<AzureServiceResult> Review()
    {
        Console.WriteLine("Reviewing Event Grid...");
        foreach (var item in data)
        {
            var diagnostics = client.GetDiagnosticSettings(new ResourceIdentifier(item.Id!));
            var diagnosticsCount = diagnostics.Count();

            yield return new AzureServiceResult
            {
                SubscriptionId = subscriptionId,
                ResourceGroup = resourceGroup,
                ServiceName = item.Name,
                Sku = "Nne",
                Sla = "TODO",
                Type = item.ResourceType,
                AvaliabilityZones = "Yes",
                PrivateEndpoints = item.PrivateEndpointConnections.Count()  > 0,
                DiagnosticSettings = diagnosticsCount > 0,
                CAFNaming = item.Name.StartsWith("evgd")
            };
        }
    }
}