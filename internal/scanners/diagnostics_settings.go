// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package scanners

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/monitor/armmonitor"
	"github.com/rs/zerolog/log"
)

// DiagnosticSettingsScanner - scanner for diagnostic settings
type DiagnosticSettingsScanner struct {
	ctx    context.Context
	client *arm.Client
}

// Init - Initializes the DiagnosticSettingsScanner
func (d *DiagnosticSettingsScanner) Init(ctx context.Context, cred azcore.TokenCredential, options *arm.ClientOptions) error {
	client, err := arm.NewClient(moduleName+".DiagnosticSettingsBatch", moduleVersion, cred, options)
	if err != nil {
		return err
	}
	d.client = client
	d.ctx = ctx
	return nil
}

// ListResourcesWithDiagnosticSettings - Lists all resources with diagnostic settings
func (d *DiagnosticSettingsScanner) ListResourcesWithDiagnosticSettings(resources []*string) (map[string]bool, error) {
	res := map[string]bool{}

	if len(resources) > 5000 {
		log.Warn().Msg(fmt.Sprintf("%d resources detected. Scan will take longer than usual", len(resources)))
	}

	batches := int(math.Ceil(float64(len(resources)) / 20))

	LogResourceTypeScan("Diagnostic Settings")

	if batches == 0 {
		return res, nil
	}

	log.Debug().Msgf("Number of diagnostic setting batches: %d", batches)
	jobs := make(chan []*string, batches)
	ch := make(chan map[string]bool, batches)
	var wg sync.WaitGroup

	// Start workers
	// Based on: https://medium.com/insiderengineering/concurrent-http-requests-in-golang-best-practices-and-techniques-f667e5a19dea
	numWorkers := 100 // Define the number of workers in the pool
	for w := 0; w < numWorkers; w++ {
		go d.worker(jobs, ch, &wg)
	}
	wg.Add(batches)

	// Split resources into batches of 20 items.
	batchSize := 20
	batchCount := 0
	for i := 0; i < len(resources); i += batchSize {
		j := i + batchSize
		if j > len(resources) {
			j = len(resources)
		}
		jobs <- resources[i:j]

		batchCount++
		if batchCount == numWorkers {
			log.Debug().Msgf("all %d workers are running. Sleeping for 4 seconds to avoid throttling", numWorkers)
			batchCount = 0
			// there are more batches to process
			// Staggering queries to avoid throttling. Max 15 queries each 5 seconds.
			// https://learn.microsoft.com/en-us/azure/governance/resource-graph/concepts/guidance-for-throttled-requests#staggering-queries
			time.Sleep(4 * time.Second)
		}
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
		resp, err := d.restCall(d.ctx, ids)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to get diagnostic settings")
		}
		asyncRes := map[string]bool{}
		for _, response := range resp.Responses {
			for _, diagnosticSetting := range response.Content.Value {
				id := parseResourceId(diagnosticSetting.ID)
				asyncRes[id] = true
			}
		}
		results <- asyncRes
		wg.Done()
	}
}

const (
	moduleName    = "armresources"
	moduleVersion = "v1.1.1"
)

func (d *DiagnosticSettingsScanner) restCall(ctx context.Context, resourceIds []*string) (*ArmBatchResponse, error) {
	req, err := runtime.NewRequest(ctx, http.MethodPost, runtime.JoinPaths(d.client.Endpoint(), "batch"))
	if err != nil {
		return nil, err
	}
	reqQP := req.Raw().URL.Query()
	reqQP.Set("api-version", "2020-06-01")
	req.Raw().URL.RawQuery = reqQP.Encode()
	req.Raw().Header["Accept"] = []string{"application/json"}

	batch := ArmBatchRequest{
		Requests: []ArmBatchRequestItem{},
	}

	for _, resourceId := range resourceIds {
		batch.Requests = append(batch.Requests, ArmBatchRequestItem{
			HttpMethod:  http.MethodGet,
			RelativeUrl: *resourceId + "/providers/microsoft.insights/diagnosticSettings?api-version=2021-05-01-preview",
		})
	}

	// set request body
	err = runtime.MarshalAsJSON(req, batch)
	if err != nil {
		return nil, err
	}

	resp, err := d.client.Pipeline().Do(req)
	if err != nil {
		return nil, err
	}
	if !runtime.HasStatusCode(resp, http.StatusOK, http.StatusAccepted) {
		return nil, runtime.NewResponseError(resp)
	}

	result := ArmBatchResponse{}
	if err := runtime.UnmarshalAsJSON(resp, &result); err != nil {
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
		Content armmonitor.DiagnosticSettingsResourceCollection `json:"content"`
	}
)

func (d *DiagnosticSettingsScanner) Scan(resources []*string) map[string]bool {
	diagResults, err := d.ListResourcesWithDiagnosticSettings(resources)
	if err != nil {
		if ShouldSkipError(err) {
			diagResults = map[string]bool{}
		} else {
			log.Fatal().Err(err).Msg("Failed to list resources with Diagnostic Settings")
		}
	}
	return diagResults
}
