// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package scanners

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Azure/azqr/internal/az"
	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/monitor/armmonitor"
	"github.com/rs/zerolog/log"
)

// DiagnosticSettingRecommendation holds the metadata for diagnostic settings recommendations
// extracted from each service scanner's rules.go file
type DiagnosticSettingRecommendation struct {
	RecommendationID string // The unique ID for this recommendation (e.g., "vwa-001")
	Recommendation   string // The recommendation text to display
	LearnMoreUrl     string // URL to documentation for more information
}

// typesWithDiagnosticSettingsSupport maps Azure resource types (lowercase) to their diagnostic settings recommendations.
// Resources with nil values support diagnostic settings but don't have dedicated scanner rules (legacy/system resources).
var typesWithDiagnosticSettingsSupport = map[string]*DiagnosticSettingRecommendation{
	"microsoft.datafactory/factories": {
		RecommendationID: "adf-001",
		Recommendation:   "Azure Data Factory should have diagnostic settings enabled",
		LearnMoreUrl:     "https://learn.microsoft.com/en-us/azure/data-factory/monitor-configure-diagnostics",
	},
	"microsoft.cdn/profiles": {
		RecommendationID: "afd-001",
		Recommendation:   "Azure FrontDoor should have diagnostic settings enabled",
		LearnMoreUrl:     "https://learn.microsoft.com/en-us/azure/frontdoor/standard-premium/how-to-logs",
	},
	"microsoft.network/azurefirewalls": {
		RecommendationID: "afw-001",
		Recommendation:   "Azure Firewall should have diagnostic settings enabled",
		LearnMoreUrl:     "https://docs.microsoft.com/en-us/azure/firewall/logs-and-metrics",
	},
	"microsoft.network/applicationgateways": {
		RecommendationID: "agw-005",
		Recommendation:   "Application Gateway: Monitor and Log the configurations and traffic",
		LearnMoreUrl:     "https://learn.microsoft.com/en-us/azure/application-gateway/application-gateway-diagnostics#diagnostic-logging",
	},
	"microsoft.cognitiveservices/accounts": {
		RecommendationID: "aif-001",
		Recommendation:   "Service should have diagnostic settings enabled",
		LearnMoreUrl:     "https://learn.microsoft.com/en-us/azure/event-hubs/monitor-event-hubs#collection-and-routing",
	},
	"microsoft.containerservice/managedclusters": {
		RecommendationID: "aks-001",
		Recommendation:   "AKS Cluster should have diagnostic settings enabled",
		LearnMoreUrl:     "https://learn.microsoft.com/en-us/azure/aks/monitor-aks#collect-resource-logs",
	},
	"microsoft.apimanagement/service": {
		RecommendationID: "apim-001",
		Recommendation:   "APIM should have diagnostic settings enabled",
		LearnMoreUrl:     "https://learn.microsoft.com/en-us/azure/api-management/api-management-howto-use-azure-monitor#resource-logs",
	},
	"microsoft.appconfiguration/configurationstores": {
		RecommendationID: "appcs-001",
		Recommendation:   "AppConfiguration should have diagnostic settings enabled",
		LearnMoreUrl:     "https://learn.microsoft.com/en-us/azure/azure-app-configuration/monitor-app-configuration?tabs=portal",
	},
	"microsoft.analysisservices/servers": {
		RecommendationID: "as-001",
		Recommendation:   "Azure Analysis Service should have diagnostic settings enabled",
		LearnMoreUrl:     "https://learn.microsoft.com/en-us/azure/analysis-services/analysis-services-logging",
	},
	"microsoft.web/serverfarms": {
		RecommendationID: "asp-001",
		Recommendation:   "Plan should have diagnostic settings enabled",
		LearnMoreUrl:     "",
	},
	"microsoft.web/sites": {
		RecommendationID: "app-001",
		Recommendation:   "App should have diagnostic settings enabled",
		LearnMoreUrl:     "https://learn.microsoft.com/en-us/azure/app-service/troubleshoot-diagnostic-logs#send-logs-to-azure-monitor",
	},
	"microsoft.app/managedenvironments": {
		RecommendationID: "cae-001",
		Recommendation:   "Container Apps Environment should have diagnostic settings enabled",
		LearnMoreUrl:     "https://learn.microsoft.com/en-us/azure/container-apps/log-options#diagnostic-settings",
	},
	"microsoft.documentdb/databaseaccounts": {
		RecommendationID: "cosmos-001",
		Recommendation:   "CosmosDB should have diagnostic settings enabled",
		LearnMoreUrl:     "https://learn.microsoft.com/en-us/azure/cosmos-db/monitor-resource-logs",
	},
	"microsoft.containerregistry/registries": {
		RecommendationID: "cr-001",
		Recommendation:   "ContainerRegistry should have diagnostic settings enabled",
		LearnMoreUrl:     "https://learn.microsoft.com/en-us/azure/container-registry/monitor-service",
	},
	"microsoft.databricks/workspaces": {
		RecommendationID: "dbw-001",
		Recommendation:   "Azure Databricks should have diagnostic settings enabled",
		LearnMoreUrl:     "https://learn.microsoft.com/en-us/azure/databricks/administration-guide/account-settings/audit-log-delivery",
	},
	"microsoft.kusto/clusters": {
		RecommendationID: "dec-001",
		Recommendation:   "Azure Data Explorer should have diagnostic settings enabled",
		LearnMoreUrl:     "https://learn.microsoft.com/en-us/azure/data-explorer/using-diagnostic-logs",
	},
	"microsoft.eventgrid/domains": {
		RecommendationID: "evgd-001",
		Recommendation:   "Event Grid Domain should have diagnostic settings enabled",
		LearnMoreUrl:     "https://learn.microsoft.com/en-us/azure/event-grid/diagnostic-logs",
	},
	"microsoft.eventhub/namespaces": {
		RecommendationID: "evh-001",
		Recommendation:   "Event Hub Namespace should have diagnostic settings enabled",
		LearnMoreUrl:     "https://learn.microsoft.com/en-us/azure/event-hubs/monitor-event-hubs#collection-and-routing",
	},
	"microsoft.machinelearningservices/workspaces": {
		RecommendationID: "hub-006",
		Recommendation:   "Service should have diagnostic settings enabled",
		LearnMoreUrl:     "https://learn.microsoft.com/en-us/azure/event-hubs/monitor-event-hubs#collection-and-routing",
	},
	"microsoft.keyvault/vaults": {
		RecommendationID: "kv-001",
		Recommendation:   "Key Vault should have diagnostic settings enabled",
		LearnMoreUrl:     "https://learn.microsoft.com/en-us/azure/key-vault/general/monitor-key-vault",
	},
	"microsoft.network/loadbalancers": {
		RecommendationID: "lb-001",
		Recommendation:   "Load Balancer should have diagnostic settings enabled",
		LearnMoreUrl:     "https://learn.microsoft.com/en-us/azure/load-balancer/monitor-load-balancer#creating-a-diagnostic-setting",
	},
	"microsoft.logic/workflows": {
		RecommendationID: "logic-001",
		Recommendation:   "Logic App should have diagnostic settings enabled",
		LearnMoreUrl:     "https://learn.microsoft.com/en-us/azure/logic-apps/monitor-workflows-collect-diagnostic-data",
	},
	"microsoft.dbformariadb/servers": {
		RecommendationID: "maria-001",
		Recommendation:   "MariaDB should have diagnostic settings enabled",
		LearnMoreUrl:     "",
	},
	"microsoft.dbformysql/servers": {
		RecommendationID: "mysql-001",
		Recommendation:   "Azure Database for MySQL - Single Server should have diagnostic settings enabled",
		LearnMoreUrl:     "https://learn.microsoft.com/en-us/azure/mysql/single-server/concepts-monitoring#server-logs",
	},
	"microsoft.dbformysql/flexibleservers": {
		RecommendationID: "mysqlf-001",
		Recommendation:   "Azure Database for MySQL - Flexible Server should have diagnostic settings enabled",
		LearnMoreUrl:     "https://learn.microsoft.com/en-us/azure/mysql/flexible-server/tutorial-query-performance-insights#set-up-diagnostics",
	},
	"microsoft.network/natgateways": {
		RecommendationID: "ng-001",
		Recommendation:   "NAT Gateway should have diagnostic settings enabled",
		LearnMoreUrl:     "https://learn.microsoft.com/en-us/azure/nat-gateway/nat-metrics",
	},
	"microsoft.network/networksecuritygroups": {
		RecommendationID: "nsg-001",
		Recommendation:   "NSG should have diagnostic settings enabled",
		LearnMoreUrl:     "https://learn.microsoft.com/en-us/azure/virtual-network/virtual-network-nsg-manage-log",
	},
	"microsoft.dbforpostgresql/servers": {
		RecommendationID: "psql-001",
		Recommendation:   "PostgreSQL should have diagnostic settings enabled",
		LearnMoreUrl:     "https://learn.microsoft.com/en-us/azure/postgresql/single-server/concepts-server-logs#resource-logs",
	},
	"microsoft.dbforpostgresql/flexibleservers": {
		RecommendationID: "psqlf-001",
		Recommendation:   "PostgreSQL should have diagnostic settings enabled",
		LearnMoreUrl:     "https://learn.microsoft.com/en-us/azure/postgresql/flexible-server/howto-configure-and-access-logs",
	},
	"microsoft.cache/redis": {
		RecommendationID: "redis-001",
		Recommendation:   "Redis should have diagnostic settings enabled",
		LearnMoreUrl:     "https://learn.microsoft.com/en-us/azure/azure-cache-for-redis/cache-monitor-diagnostic-settings",
	},
	"microsoft.servicebus/namespaces": {
		RecommendationID: "sb-001",
		Recommendation:   "Service Bus should have diagnostic settings enabled",
		LearnMoreUrl:     "https://learn.microsoft.com/en-us/azure/service-bus-messaging/monitor-service-bus#collection-and-routing",
	},
	"microsoft.signalrservice/signalr": {
		RecommendationID: "sigr-001",
		Recommendation:   "SignalR should have diagnostic settings enabled",
		LearnMoreUrl:     "https://learn.microsoft.com/en-us/azure/azure-signalr/signalr-howto-diagnostic-logs",
	},
	"microsoft.sql/servers/databases": {
		RecommendationID: "sqldb-001",
		Recommendation:   "SQL Database should have diagnostic settings enabled",
		LearnMoreUrl:     "",
	},
	"microsoft.search/searchservices": {
		RecommendationID: "srch-006",
		Recommendation:   "Azure AI Search should have diagnostic settings enabled",
		LearnMoreUrl:     "https://learn.microsoft.com/en-us/azure/search/search-monitor-enable-logging",
	},
	"microsoft.storage/storageaccounts": {
		RecommendationID: "st-001",
		Recommendation:   "Storage should have diagnostic settings enabled",
		LearnMoreUrl:     "https://learn.microsoft.com/en-us/azure/storage/blobs/monitor-blob-storage",
	},
	"microsoft.synapse/workspaces": {
		RecommendationID: "synw-001",
		Recommendation:   "Azure Synapse Workspace should have diagnostic settings enabled",
		LearnMoreUrl:     "https://learn.microsoft.com/en-us/azure/data-factory/monitor-configure-diagnostics",
	},
	"microsoft.network/trafficmanagerprofiles": {
		RecommendationID: "traf-001",
		Recommendation:   "Traffic Manager should have diagnostic settings enabled",
		LearnMoreUrl:     "https://learn.microsoft.com/en-us/azure/traffic-manager/traffic-manager-diagnostic-logs",
	},
	"microsoft.network/virtualnetworkgateways": {
		RecommendationID: "vgw-001",
		Recommendation:   "Virtual Network Gateway should have diagnostic settings enabled",
		LearnMoreUrl:     "https://learn.microsoft.com/en-us/azure/vpn-gateway/monitor-vpn-gateway",
	},
	"microsoft.network/virtualnetworks": {
		RecommendationID: "vnet-001",
		Recommendation:   "Virtual Network should have diagnostic settings enabled",
		LearnMoreUrl:     "https://learn.microsoft.com/en-us/azure/virtual-network/monitor-virtual-network#collection-and-routing",
	},
	"microsoft.network/virtualwans": {
		RecommendationID: "vwa-001",
		Recommendation:   "Virtual WAN should have diagnostic settings enabled",
		LearnMoreUrl:     "https://learn.microsoft.com/en-us/azure/virtual-wan/monitor-virtual-wan",
	},
	"microsoft.signalrservice/webpubsub": {
		RecommendationID: "wps-001",
		Recommendation:   "Web Pub Sub should have diagnostic settings enabled",
		LearnMoreUrl:     "https://learn.microsoft.com/en-us/azure/azure-web-pubsub/howto-troubleshoot-resource-logs",
	},
	// Legacy/system resource types that support diagnostic settings but don't have dedicated scanners with rules
	"microsoft.network/networkinterfaces":                       nil,
	"microsoft.network/routetables":                             nil,
	"microsoft.recoveryservices/vaults":                         nil,
	"specialized.workload/avd":                                  nil,
	"microsoft.compute/virtualmachines":                         nil,
	"microsoft.network/virtualnetworks/subnets":                 nil,
	"specialized.workload/hpc":                                  nil,
	"microsoft.automation/automationaccounts":                   nil,
	"microsoft.dashboard/grafana":                               nil,
	"microsoft.virtualmachineimages/imagetemplates":             nil,
	"microsoft.devices/iothubs":                                 nil,
	"microsoft.compute/disks":                                   nil,
	"microsoft.network/connections":                             nil,
	"microsoft.app/containerapps":                               nil,
	"microsoft.network/frontdoorwebapplicationfirewallpolicies": nil,
	"microsoft.batch/batchaccounts":                             nil,
	"microsoft.network/publicipaddresses":                       nil,
	"microsoft.sql/servers":                                     nil,
	"microsoft.sql/servers/elasticpools":                        nil,
	"microsoft.operationalinsights/workspaces":                  nil,
	"microsoft.insights/components":                             nil,
	"microsoft.compute/virtualmachinescalesets":                 nil,
	"microsoft.network/privateendpoints":                        nil,
	"microsoft.containerinstance/containergroups":               nil,
	"microsoft.resources/resourcegroups":                        nil,
	"microsoft.network/ipgroups":                                nil,
	"microsoft.compute/galleries":                               nil,
	"microsoft.network/privatednszones":                         nil,
	"microsoft.network/networkwatchers":                         nil,
	"microsoft.compute/availabilitysets":                        nil,
	"microsoft.web/connections":                                 nil,
	"microsoft.web/certificates":                                nil,
	"specialized.workload/sap":                                  nil,
}

// DiagnosticSettingsScanner - scanner for diagnostic settings
type DiagnosticSettingsScanner struct {
	ctx         context.Context
	httpClient  *az.HttpClient
	scanContext *models.ScanParams
}

// GetRecommendations returns all diagnostic settings recommendations in the format expected by the scanner system
// grouped by resource type. This allows other scanners to use diagnostic settings checks dynamically.
func GetRecommendations() map[string]map[string]models.GraphRecommendation {
	recommendations := make(map[string]map[string]models.GraphRecommendation)

	for resourceType, rec := range typesWithDiagnosticSettingsSupport {
		if rec == nil {
			continue // Skip legacy resource types without recommendation metadata
		}

		if recommendations[resourceType] == nil {
			recommendations[resourceType] = make(map[string]models.GraphRecommendation)
		}

		recommendations[resourceType][rec.RecommendationID] = models.GraphRecommendation{
			RecommendationID: rec.RecommendationID,
			ResourceType:     resourceType,
			Category:         string(models.CategoryMonitoringAndAlerting),
			Recommendation:   rec.Recommendation,
			Impact:           string(models.ImpactLow),
			LearnMoreLink: []struct {
				Name string `yaml:"name"`
				Url  string `yaml:"url"`
			}{
				{
					Name: "Diagnostic Settings",
					Url:  rec.LearnMoreUrl,
				},
			},
			Source: "AZQR",
		}
	}

	return recommendations
}

// Init - Initializes the DiagnosticSettingsScanner
func (d *DiagnosticSettingsScanner) Init(ctx context.Context, cred azcore.TokenCredential, scanCtx *models.ScanParams) error {
	d.ctx = ctx
	d.scanContext = scanCtx
	// Create HTTP client with built-in retry logic, authentication, and throttling
	d.httpClient = az.NewHttpClient(cred, az.DefaultHttpClientOptions(60*time.Second))
	return nil
}

// ListResourcesWithDiagnosticSettings returns a map of resource IDs that have diagnostic settings enabled
func (d *DiagnosticSettingsScanner) ListResourcesWithDiagnosticSettings(resources []*models.Resource) (map[string]bool, error) {
	res := map[string]bool{}

	// Filter resources to only include those that support diagnostic settings
	if len(resources) == 0 {
		log.Debug().Msg("No resources found to scan for diagnostic settings")
		return res, nil
	}
	filteredResources := []*string{}
	for _, resource := range resources {
		// Check if the resource type (in lowercase) supports diagnostic settings
		if _, ok := typesWithDiagnosticSettingsSupport[strings.ToLower(resource.Type)]; ok {
			filteredResources = append(filteredResources, &resource.ID)
		}
	}

	// if len(filteredResources) > 0 { // Uncomment this block to test with a large number of filteredResources
	// 	firstResource := filteredResources[0]
	// 	for len(filteredResources) < 100000 {
	// 		filteredResources = append(filteredResources, firstResource)
	// 	}
	// }

	if len(filteredResources) > 5000 {
		log.Warn().Msgf("%d resources detected. Scan will take longer than usual", len(filteredResources))
	}

	batches := int(math.Ceil(float64(len(filteredResources)) / 20))

	models.LogResourceTypeScan("Diagnostic Settings")

	if batches == 0 {
		return res, nil
	}

	log.Debug().Msgf("Number of diagnostic setting batches: %d", batches)
	jobs := make(chan []*string, batches)
	ch := make(chan map[string]bool, batches)
	var wg sync.WaitGroup

	// Use 30 workers to balance throughput with ARM API rate limits
	// ARM limits: 3 req/sec sustained, 100 burst capacity
	// 30 workers allows efficient use of burst capacity while maintaining sustainable rate
	numWorkers := 30
	if batches < numWorkers {
		numWorkers = batches
	}
	for w := 0; w < numWorkers; w++ {
		go d.worker(jobs, ch, &wg)
	}
	wg.Add(batches)

	// Split filteredResources into batches of 20 items.
	batchSize := 20
	for i := 0; i < len(filteredResources); i += batchSize {
		j := i + batchSize
		if j > len(filteredResources) {
			j = len(filteredResources)
		}
		jobs <- filteredResources[i:j]
	}

	// Wait for all workers to finish
	close(jobs)
	wg.Wait()

	for i := 0; i < batches; i++ {
		for k, v := range <-ch {
			res[k] = v
		}
	}

	return res, nil
}

func (d *DiagnosticSettingsScanner) worker(jobs <-chan []*string, results chan<- map[string]bool, wg *sync.WaitGroup) {
	for ids := range jobs {
		// doRequest now includes built-in retry logic via HttpClient
		resp, err := d.doRequest(d.ctx, ids)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to get diagnostic settings")
		}
		asyncRes := map[string]bool{}
		for _, response := range resp.Responses {
			if response.HttpStatusCode == http.StatusOK {
				// Decode the response content into a DiagnosticSettingsResourceCollection
				var diagnosticSettings armmonitor.DiagnosticSettingsResourceCollection
				contentBytes, err := json.Marshal(response.Content)
				if err != nil {
					log.Fatal().Err(err).Msg("Failed to marshal diagnostic settings content")
					continue
				}
				// Unmarshal the JSON bytes into the DiagnosticSettingsResourceCollection struct
				if err := json.Unmarshal(contentBytes, &diagnosticSettings); err != nil {
					log.Fatal().Err(err).Msg("Failed to unmarshal diagnostic settings response")
					continue
				}

				for _, diagnosticSetting := range diagnosticSettings.Value {
					id := parseResourceId(diagnosticSetting.ID)
					asyncRes[id] = true
				}
			}
		}
		results <- asyncRes
		wg.Done()
	}
}

// doRequest performs a batch request to retrieve diagnostic settings using the HTTP client with built-in retry logic.
func (d *DiagnosticSettingsScanner) doRequest(ctx context.Context, resourceIds []*string) (*ArmBatchResponse, error) {
	// Build the batch endpoint URL
	resourceManagerEndpoint := az.GetResourceManagerEndpoint()
	batchURL := fmt.Sprintf("%s/batch?api-version=2020-06-01", resourceManagerEndpoint)

	// Prepare the batch request payload
	batch := ArmBatchRequest{
		Requests: []ArmBatchRequestItem{},
	}
	for _, resourceId := range resourceIds {
		batch.Requests = append(batch.Requests, ArmBatchRequestItem{
			HttpMethod:  http.MethodGet,
			RelativeUrl: *resourceId + "/providers/microsoft.insights/diagnosticSettings?api-version=2021-05-01-preview",
		})
	}

	// Marshal the batch request to JSON
	bodyBytes, err := json.Marshal(batch)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal batch request: %w", err)
	}

	// Create ReadSeekCloser for the request body
	readSeekCloser := &bytesReadSeekCloser{
		Reader: bytes.NewReader(bodyBytes),
	}

	// Send POST request using HttpClient (handles auth, throttling, and retries)
	respBody, resp, err := d.httpClient.DoPost(ctx, batchURL, readSeekCloser)
	if err != nil {
		return nil, fmt.Errorf("batch request failed: %w", err)
	}

	// Log quota information if available
	if quotaStr := resp.Header.Get("x-ms-ratelimit-remaining-tenant-reads"); quotaStr != "" {
		if quota, err := strconv.Atoi(quotaStr); err == nil {
			log.Debug().Msgf("ARM batch remaining quota: %d", quota)
			if quota == 0 {
				log.Debug().Msg("ARM batch quota limit reached")
			}
		}
	}

	// Decode the response body
	var result ArmBatchResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal batch response: %w", err)
	}

	return &result, nil
}

func parseResourceId(diagnosticSettingID *string) string {
	id := *diagnosticSettingID
	i := strings.Index(id, "/providers/microsoft.insights/diagnosticSettings/")
	return strings.ToLower(id[:i])
}

// bytesReadSeekCloser wraps a bytes.Reader to implement io.ReadSeekCloser
// This allows the Azure SDK pipeline to seek back to the beginning for retries
type bytesReadSeekCloser struct {
	*bytes.Reader
}

// Close implements io.Closer (no-op for bytes.Reader)
func (b *bytesReadSeekCloser) Close() error {
	return nil
}

type (
	ArmBatchRequest struct {
		Requests []ArmBatchRequestItem `json:"requests"`
	}

	ArmBatchRequestItem struct {
		HttpMethod  string `json:"httpMethod"`
		RelativeUrl string `json:"relativeUrl"`
	}

	ArmBatchResponse struct {
		Responses []ArmBatchResponseItem `json:"responses"`
	}

	ArmBatchResponseItem struct {
		HttpStatusCode int         `json:"httpStatusCode"` // HTTP status code of the response
		Content        interface{} `json:"content"`        // armmonitor.DiagnosticSettingsResourceCollection
	}
)

func (d *DiagnosticSettingsScanner) Scan(resources []*models.Resource) []*models.GraphResult {
	// Get diagnostic settings status for all resources
	diagResults, err := d.ListResourcesWithDiagnosticSettings(resources)
	if err != nil {
		if models.ShouldSkipError(err) {
			diagResults = map[string]bool{}
		} else {
			log.Fatal().Err(err).Msg("Failed to list resources with Diagnostic Settings")
		}
	}

	// Get recommendations for all resource types
	recommendations := GetRecommendations()

	// Build results for resources WITHOUT diagnostic settings
	var results []*models.GraphResult
	for _, resource := range resources {
		resourceID := strings.ToLower(resource.ID)
		resourceType := strings.ToLower(resource.Type)

		// Skip resources that already have diagnostic settings enabled
		if diagResults[resourceID] {
			continue
		}

		if d.scanContext.Filters.Azqr.IsServiceExcluded(resource.ID) {
			continue
		}

		// Get the recommendation for this resource type
		recs, hasRecs := recommendations[resourceType]
		if !hasRecs {
			continue // Skip resource types without diagnostic settings recommendations
		}

		// Create a GraphResult for each recommendation (typically just one per resource type)
		for _, rec := range recs {
			// Extract learn more URL from LearnMoreLink
			learnURL := ""
			if len(rec.LearnMoreLink) > 0 {
				learnURL = rec.LearnMoreLink[0].Url
			}

			results = append(results, &models.GraphResult{
				RecommendationID:    rec.RecommendationID,
				ResourceType:        resource.Type,
				Recommendation:      rec.Recommendation,
				LongDescription:     rec.LongDescription,
				PotentialBenefits:   rec.PotentialBenefits,
				ResourceID:          resource.ID,
				SubscriptionID:      resource.SubscriptionID,
				ResourceGroup:       resource.ResourceGroup,
				Name:                resource.Name,
				Category:            models.CategoryMonitoringAndAlerting,
				Impact:              models.ImpactLow,
				Learn:               learnURL,
				AutomationAvailable: rec.AutomationAvailable,
				Source:              rec.Source,
			})
		}
	}

	return results
}
