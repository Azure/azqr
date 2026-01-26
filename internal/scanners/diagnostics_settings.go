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
	"github.com/Azure/azqr/internal/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/monitor/armmonitor"
	"github.com/rs/zerolog/log"
)

var typesWithDiagnosticSettingsSupport = map[string]*bool{
	// Resource types from all scanners' ResourceTypes()
	"microsoft.network/networkinterfaces":                       to.Ptr(true),
	"microsoft.keyvault/vaults":                                 to.Ptr(true),
	"microsoft.network/trafficmanagerprofiles":                  to.Ptr(true),
	"microsoft.network/applicationgateways":                     to.Ptr(true),
	"microsoft.network/routetables":                             to.Ptr(true),
	"microsoft.recoveryservices/vaults":                         to.Ptr(true),
	"specialized.workload/avd":                                  to.Ptr(true),
	"microsoft.compute/virtualmachines":                         to.Ptr(true),
	"microsoft.network/virtualwans":                             to.Ptr(true),
	"microsoft.network/virtualnetworks":                         to.Ptr(true),
	"microsoft.network/virtualnetworks/subnets":                 to.Ptr(true),
	"specialized.workload/hpc":                                  to.Ptr(true),
	"microsoft.automation/automationaccounts":                   to.Ptr(true),
	"microsoft.machinelearningservices/workspaces":              to.Ptr(true),
	"microsoft.containerservice/managedclusters":                to.Ptr(true),
	"microsoft.dbforpostgresql/flexibleservers":                 to.Ptr(true),
	"microsoft.dbforpostgresql/servers":                         to.Ptr(true),
	"microsoft.network/loadbalancers":                           to.Ptr(true),
	"microsoft.signalrservice/signalr":                          to.Ptr(true),
	"specialized.workload/sap":                                  to.Ptr(true),
	"microsoft.dashboard/grafana":                               to.Ptr(true),
	"microsoft.containerregistry/registries":                    to.Ptr(true),
	"microsoft.virtualmachineimages/imagetemplates":             to.Ptr(true),
	"microsoft.devices/iothubs":                                 to.Ptr(true),
	"microsoft.dbformysql/servers":                              to.Ptr(true),
	"microsoft.dbformysql/flexibleservers":                      to.Ptr(true),
	"microsoft.eventgrid/domains":                               to.Ptr(true),
	"microsoft.compute/disks":                                   to.Ptr(true),
	"microsoft.network/connections":                             to.Ptr(true),
	"microsoft.app/containerapps":                               to.Ptr(true),
	"microsoft.network/virtualnetworkgateways":                  to.Ptr(true),
	"microsoft.network/frontdoorwebapplicationfirewallpolicies": to.Ptr(true),
	"microsoft.batch/batchaccounts":                             to.Ptr(true),
	"microsoft.search/searchservices":                           to.Ptr(true),
	"microsoft.network/publicipaddresses":                       to.Ptr(true),
	"microsoft.signalrservice/webpubsub":                        to.Ptr(true),
	"microsoft.sql/servers":                                     to.Ptr(true),
	"microsoft.sql/servers/databases":                           to.Ptr(true),
	"microsoft.sql/servers/elasticpools":                        to.Ptr(true),
	"microsoft.network/natgateways":                             to.Ptr(true),
	"microsoft.operationalinsights/workspaces":                  to.Ptr(true),
	"microsoft.analysisservices/servers":                        to.Ptr(true),
	"microsoft.insights/components":                             to.Ptr(true),
	"microsoft.datafactory/factories":                           to.Ptr(true),
	"microsoft.cognitiveservices/accounts":                      to.Ptr(true),
	"microsoft.kusto/clusters":                                  to.Ptr(true),
	"microsoft.app/managedenvironments":                         to.Ptr(true),
	"microsoft.compute/virtualmachinescalesets":                 to.Ptr(true),
	"microsoft.storage/storageaccounts":                         to.Ptr(true),
	"microsoft.network/privateendpoints":                        to.Ptr(true),
	"microsoft.containerinstance/containergroups":               to.Ptr(true),
	"microsoft.resources/resourcegroups":                        to.Ptr(true),
	"microsoft.servicebus/namespaces":                           to.Ptr(true),
	"microsoft.network/azurefirewalls":                          to.Ptr(true),
	"microsoft.network/ipgroups":                                to.Ptr(true),
	"microsoft.cache/redis":                                     to.Ptr(true),
	"microsoft.network/networksecuritygroups":                   to.Ptr(true),
	"microsoft.eventhub/namespaces":                             to.Ptr(true),
	"microsoft.documentdb/databaseaccounts":                     to.Ptr(true),
	"microsoft.compute/galleries":                               to.Ptr(true),
	"microsoft.appconfiguration/configurationstores":            to.Ptr(true),
	"microsoft.cdn/profiles":                                    to.Ptr(true),
	"microsoft.logic/workflows":                                 to.Ptr(true),
	"microsoft.databricks/workspaces":                           to.Ptr(true),
	"microsoft.network/privatednszones":                         to.Ptr(true),
	"microsoft.network/networkwatchers":                         to.Ptr(true),
	"microsoft.compute/availabilitysets":                        to.Ptr(true),
	"microsoft.web/serverfarms":                                 to.Ptr(true),
	"microsoft.web/sites":                                       to.Ptr(true),
	"microsoft.web/connections":                                 to.Ptr(true),
	"microsoft.web/certificates":                                to.Ptr(true),
}

// DiagnosticSettingsScanner - scanner for diagnostic settings
type DiagnosticSettingsScanner struct {
	ctx        context.Context
	httpClient *az.HttpClient
}

// Init - Initializes the DiagnosticSettingsScanner
func (d *DiagnosticSettingsScanner) Init(ctx context.Context, cred azcore.TokenCredential, options *arm.ClientOptions) error {
	d.ctx = ctx
	// Create HTTP client with built-in retry logic, authentication, and throttling
	d.httpClient = az.NewHttpClient(cred, az.DefaultHttpClientOptions(60*time.Second))
	return nil
}

// ListResourcesWithDiagnosticSettings - Lists all resources with diagnostic settings
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

	// Use default scope for Azure Management API
	scope := ""

	// Send POST request using HttpClient (handles auth, throttling, and retries)
	respBody, resp, err := d.httpClient.DoPost(ctx, batchURL, readSeekCloser, &scope)
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

func (d *DiagnosticSettingsScanner) Scan(resources []*models.Resource) map[string]bool {
	diagResults, err := d.ListResourcesWithDiagnosticSettings(resources)
	if err != nil {
		if models.ShouldSkipError(err) {
			diagResults = map[string]bool{}
		} else {
			log.Fatal().Err(err).Msg("Failed to list resources with Diagnostic Settings")
		}
	}
	return diagResults
}
