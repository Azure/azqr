// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package scanners

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/throttling"
	"github.com/Azure/azqr/internal/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/monitor/armmonitor"
	"github.com/rs/zerolog/log"
)

var typesWithDiagnosticSettingsSupport = map[string]*bool{
	"microsoft.network/virtualwans":                 to.Ptr(true),
	"microsoft.apimanagement/service":               to.Ptr(true),
	"microsoft.compute/virtualmachines":             to.Ptr(true),
	"microsoft.containerservice/managedclusters":    to.Ptr(true),
	"microsoft.machinelearningservices/workspaces":  to.Ptr(true),
	"microsoft.dbforpostgresql/servers":             to.Ptr(true),
	"microsoft.dbforpostgresql/flexibleservers":     to.Ptr(true),
	"microsoft.network/virtualnetworks":             to.Ptr(true),
	"microsoft.network/loadbalancers":               to.Ptr(true),
	"microsoft.synapse/workspaces":                  to.Ptr(true),
	"microsoft.synapse workspaces/bigdatapools":     to.Ptr(true),
	"microsoft.synapse/workspaces/sqlpools":         to.Ptr(true),
	"microsoft.containerregistry/registries":        to.Ptr(true),
	"microsoft.eventgrid/domains":                   to.Ptr(true),
	"microsoft.dashboard/managedgrafana":            to.Ptr(true),
	"microsoft.virtualmachineimages/imagetemplates": to.Ptr(true),
	"microsoft.dbformysql/servers":                  to.Ptr(true),
	"microsoft.dbformysql/flexibleservers":          to.Ptr(true),
	"microsoft.network/virtualnetworkgateways":      to.Ptr(true),
	"microsoft.network/routetables":                 to.Ptr(true),
	"microsoft.app/containerapps":                   to.Ptr(true),
	"microsoft.network/trafficmanagerprofiles":      to.Ptr(true),
	"microsoft.search/searchservices":               to.Ptr(true),
	"microsoft.signalrservice/webpubsub":            to.Ptr(true),
	"microsoft.network/publicipaddresses":           to.Ptr(true),
	"microsoft.sql/servers":                         to.Ptr(true),
	"microsoft.sql/servers/databases":               to.Ptr(true),
	"microsoft.sql/servers/elasticpools":            to.Ptr(true),
	"microsoft.keyvault/vaults":                     to.Ptr(true),
	"microsoft.network/natgateways":                 to.Ptr(true),
	"microsoft.operationalinsights/workspaces":      to.Ptr(true),
	"microsoft.analysisservices/servers":            to.Ptr(true),
	"microsoft.insights/components":                 to.Ptr(true),
	"microsoft.datafactory/factories":               to.Ptr(true),
	"microsoft.cognitiveservices/accounts":          to.Ptr(true),
	"microsoft.kusto/clusters":                      to.Ptr(true),
	"microsoft.app/managedenvironments":             to.Ptr(true),
	"microsoft.compute/virtualmachinescalesets":     to.Ptr(true),
	"microsoft.storage/storageaccounts":             to.Ptr(true),
	"microsoft.network/privateendpoints":            to.Ptr(true),
	"microsoft.containerinstance/containergroups":   to.Ptr(true),
}

// DiagnosticSettingsScanner - scanner for diagnostic settings
type DiagnosticSettingsScanner struct {
	ctx         context.Context
	httpClient  *http.Client
	accessToken string
}

const (
	bucketCapacity = 250
	refillRate     = 25
)

// Init - Initializes the DiagnosticSettingsScanner
func (d *DiagnosticSettingsScanner) Init(ctx context.Context, cred azcore.TokenCredential, options *arm.ClientOptions) error {
	// Create a new HTTP client with a timeout
	httpClient := &http.Client{
		Timeout: 60 * time.Second,
	}
	d.httpClient = httpClient
	d.ctx = ctx

	// Acquire an access token using the provided credential
	token, err := cred.GetToken(context.Background(), policy.TokenRequestOptions{
		Scopes: []string{"https://management.azure.com/.default"},
	})
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to acquire Azure access token")
	}

	// Set the access token string
	d.accessToken = token.Token
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

	// Create a burst limiter to control the rate of requests
	limiter := throttling.NewLimiter(bucketCapacity, refillRate, 1*time.Second, 0*time.Millisecond)
	burstLimiter := limiter.Start()

	numWorkers := bucketCapacity
	for w := 0; w < numWorkers; w++ {
		go d.worker(jobs, ch, &wg, burstLimiter)
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

func (d *DiagnosticSettingsScanner) worker(jobs <-chan []*string, results chan<- map[string]bool, wg *sync.WaitGroup, burstLimiter <-chan struct{}) {
	// Wait for a token from the burstLimiter channel before starting the scan
	for ids := range jobs {
		<-burstLimiter
		resp, err := d.retry(d.ctx, 3, 10*time.Millisecond, ids)
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

// retry executes the Resource Graph query with retries and exponential backoff.
// Returns the QueryResponse or error.
func (d *DiagnosticSettingsScanner) retry(ctx context.Context, attempts int, sleep time.Duration, resourceIds []*string) (*ArmBatchResponse, error) {
	var err error
	for i := 0; ; i++ {
		resp, err := d.doRequest(ctx, resourceIds)
		if err == nil {
			return resp, nil
		}

		errAsString := err.Error()

		if i >= (attempts - 1) {
			log.Info().Msgf("Retry limit reached. Error: %s", errAsString)
			break
		}

		log.Debug().Msgf("Retrying after error: %s", errAsString)

		time.Sleep(sleep)
		sleep *= 2
	}
	return nil, err
}

// restCall performs a batch request to retrieve diagnostic settings using an HTTP client.
func (d *DiagnosticSettingsScanner) doRequest(ctx context.Context, resourceIds []*string) (*ArmBatchResponse, error) {
	// Build the batch endpoint URL.
	batchURL := "https://management.azure.com/batch?api-version=2020-06-01"

	// Prepare the batch request payload.
	batch := ArmBatchRequest{
		Requests: []ArmBatchRequestItem{},
	}
	for _, resourceId := range resourceIds {
		batch.Requests = append(batch.Requests, ArmBatchRequestItem{
			HttpMethod:  http.MethodGet,
			RelativeUrl: *resourceId + "/providers/microsoft.insights/diagnosticSettings?api-version=2021-05-01-preview",
		})
	}

	// Marshal the batch request to JSON.
	body, err := json.Marshal(batch)
	if err != nil {
		return nil, err
	}

	// Create the HTTP request.
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, batchURL, strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+d.accessToken)

	// Send the HTTP request.
	resp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Fatal().Err(err).Msg("Failed to close response body")
		}
	}()

	quotaStr := resp.Header.Get("x-ms-ratelimit-remaining-tenant-reads")
	if quotaStr != "" {
		quota, err := strconv.Atoi(quotaStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse quota header: %w", err)
		}
		log.Debug().Msgf("Graph query remaining quota: %d", quota)
		// Quota limit reached, sleep for the duration specified in the response header
		if quota == 0 {
			log.Debug().Msg("Graph query quota limit reached.")
		}
	}

	// Check for successful status code.
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		// Check if the response status code is 429 (Too Many Requests).
		if resp.StatusCode == http.StatusTooManyRequests {
			// Parse the Retry-After header from the response headers
			retryAfterStr := resp.Header.Get("Retry-After")
			retryAfter, _ := strconv.Atoi(retryAfterStr)
			log.Debug().Msgf("Received 429 Too Many Requests. Retry-After: %d seconds", retryAfter)
			time.Sleep(time.Duration(retryAfter) * time.Second)
		}

		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Decode the response body.
	var result ArmBatchResponse
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func parseResourceId(diagnosticSettingID *string) string {
	id := *diagnosticSettingID
	i := strings.Index(id, "/providers/microsoft.insights/diagnosticSettings/")
	return strings.ToLower(id[:i])
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
		Content        interface{} `json:"content"`        //armmonitor.DiagnosticSettingsResourceCollection
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
