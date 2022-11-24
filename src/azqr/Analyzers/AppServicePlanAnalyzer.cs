namespace azqr;

public class AppServicePlanAnalyzer : IAzureServiceAnalyzer
{
    AppServicePlanData[] data;
    ArmClient client;
    string subscriptionId;
    string resourceGroup;
    WebSiteCollection webSites;

    public AppServicePlanAnalyzer(
        ArmClient client, 
        string subscriptionId, 
        string resourceGroup, 
        AppServicePlanData[] data, 
        WebSiteCollection webSites)
    {
        this.data = data;
        this.client = client;
        this.subscriptionId = subscriptionId;
        this.resourceGroup = resourceGroup;
        this.webSites = webSites;
    }

    public IEnumerable<AzureServiceResult> Review()
    {
        Console.WriteLine("Reviewing App Service Plan...");
        foreach (var item in data)
        {
            var diagnostics = client.GetDiagnosticSettings(new ResourceIdentifier(item.Id!));
            var sla = "TODO";

            yield return new AzureServiceResult
            {
                SubscriptionId = subscriptionId,
                ResourceGroup = resourceGroup,
                ServiceName = item.Name,
                Sku = item.Sku.Name.ToString()!,
                Sla = sla,
                Type = item.ResourceType,
                AvaliabilityZones = item.IsZoneRedundant == true ? "Yes" : "No",
                PrivateEndpoints = false,
                DiagnosticSettings = diagnostics.Any(),
                CAFNaming = item.Name.StartsWith("plan")
            };
            
            foreach (var webSite in webSites.Where(w => w.Data.AppServicePlanId == item.Id))
            {   
                diagnostics = client.GetDiagnosticSettings(new ResourceIdentifier(item.Id!));

                var webSiteData = webSite.Data;
                var privateEndpoints = webSite.GetSitePrivateEndpointConnections().Any();

                yield return new AzureServiceResult
                {
                    SubscriptionId = subscriptionId,
                    ResourceGroup = resourceGroup,
                    ServiceName = webSiteData.Name,
                    Sku = item.Sku.Name.ToString()!,
                    Sla = sla,
                    Type = webSiteData.ResourceType,
                    AvaliabilityZones = item.IsZoneRedundant == true ? "Yes" : "No",
                    PrivateEndpoints = privateEndpoints,
                    DiagnosticSettings = diagnostics.Any(),
                    CAFNaming = webSiteData.Name.StartsWith("app") || webSiteData.Name.StartsWith("func")
                };
            }
        }
    }
}