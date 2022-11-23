namespace azqr;

public class ContainerAppEnvironmentAnalyzer : IAzureServiceAnalyzer
{
    ManagedEnvironmentData[] data;
    ArmClient client;
    string subscriptionId;
    string resourceGroup;

    public ContainerAppEnvironmentAnalyzer(ArmClient client, string subscriptionId, string resourceGroup, ManagedEnvironmentData[] data)
    {
        this.data = data;
        this.client = client;
        this.subscriptionId = subscriptionId;
        this.resourceGroup = resourceGroup;
    }

    public IEnumerable<AzureServiceResult> Review()
    {
        Console.WriteLine("Reviewing Container App Environment...");
        foreach (var item in data)
        {
            var diagnostics = client.GetDiagnosticSettings(new ResourceIdentifier(item.Id!));
            var diagnosticsCount = diagnostics.Count();

            yield return new AzureServiceResult
            {
                SubscriptionId = subscriptionId,
                ResourceGroup = resourceGroup,
                ServiceName = item.Name,
                Sku = "None",
                Sla = "99.95%",
                Type = item.ResourceType,
                AvaliabilityZones = item.ZoneRedundant == true ? "Yes" : "No",
                PrivateEndpoints = (bool)item.VnetConfiguration.Internal!,
                DiagnosticSettings = diagnosticsCount > 0,
                CAFNaming = item.Name.StartsWith("cae")
            };
        }
    }
}