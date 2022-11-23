namespace azqr;

public class ApplicationGatewayAnalyzer : IAzureServiceAnalyzer
{
    ApplicationGatewayData[] data;
    ArmClient client;
    string subscriptionId;
    string resourceGroup;

    public ApplicationGatewayAnalyzer(ArmClient client, string subscriptionId, string resourceGroup, ApplicationGatewayData[] data)
    {
        this.data = data;
        this.client = client;
        this.subscriptionId = subscriptionId;
        this.resourceGroup = resourceGroup;
    }

    public IEnumerable<AzureServiceResult> Review()
    {
        Console.WriteLine("Reviewing Application Gateway...");
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
                Type = item.ResourceType.ToString()!,
                AvaliabilityZones = item.AvailabilityZones.Count() > 0 ? "Yes" : "No",
                PrivateEndpoints = item.FrontendIPConfigurations.Count(c => !string.IsNullOrEmpty(c.PublicIPAddressId)) == 0,
                DiagnosticSettings = diagnosticsCount > 0,
                CAFNaming = item.Name.StartsWith("agw")
            };
        }
    }
}