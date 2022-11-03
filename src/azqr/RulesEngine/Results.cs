namespace azqr;

public class Results
{
    public string SubscriptionId { get; set; } = string.Empty;
    public string ResourceGroup { get; set; } = string.Empty;
    public string Type { get; set; } = string.Empty;
    public string ServiceName { get; set; } = string.Empty;
    public List<RuleResultTree> RulesResults { get; set; } = new List<RuleResultTree>();
}