namespace azqr;

public interface IAzureServiceAnalyzer
{
    public IEnumerable<AzureServiceResult> Review();
}

public struct AzureServiceResult
{
    public string SubscriptionId { get; set; }
    public string ResourceGroup { get; set; }
    public string ServiceName { get; set; }
    public string Sku { get; set; }
    public string Sla { get; set; }
    public ResourceType Type { get; set; }

    public string AvaliabilityZones { get; set; }
    public bool PrivateEndpoints { get; set; }
    public bool DiagnosticSettings { get; set; }
    public bool CAFNaming { get; set; }
}
