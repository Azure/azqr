public class Results
{
    public string SubscriptionId { get; set; }
    public string ResourceGroup { get; set; }
    public string Type { get; set; }
    public string ServiceName { get; set; }
    public List<RuleResultTree> RulesResults { get; set; }
}