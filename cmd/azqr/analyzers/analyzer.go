package analyzers

type AzureServiceAnalyzer interface {
	Review(resourceGroupName string) ([]AzureServiceResult, error)
}

type AzureServiceResult struct {
	SubscriptionId     string
	ResourceGroup      string
	ServiceName        string
	Sku                string
	Sla                string
	Type               string
	AvailabilityZones  bool
	PrivateEndpoints   bool
	DiagnosticSettings bool
	CAFNaming          bool
}
