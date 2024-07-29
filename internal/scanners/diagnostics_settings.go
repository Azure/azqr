// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package scanners

import (
	"context"
	"math"
	"net/http"
	"strings"
	"sync"

	"github.com/Azure/azqr/internal/azqr"
	"github.com/Azure/azqr/internal/graph"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/monitor/armmonitor"
	"github.com/rs/zerolog/log"
)

// DiagnosticSettingsScanner - scanner for diagnostic settings
type DiagnosticSettingsScanner struct {
	config     *azqr.ScannerConfig
	client     *arm.Client
	graphQuery *graph.GraphQuery
}

// Init - Initializes the DiagnosticSettingsScanner
func (d *DiagnosticSettingsScanner) Init(config *azqr.ScannerConfig) error {
	d.config = config
	client, err := arm.NewClient(moduleName+".DiagnosticSettingsBatch", moduleVersion, d.config.Cred, d.config.ClientOptions)
	if err != nil {
		return err
	}
	d.client = client
	d.graphQuery = graph.NewGraphQuery(d.config.Cred)
	return nil
}

// ListResourcesWithDiagnosticSettings - Lists all resources with diagnostic settings
func (d *DiagnosticSettingsScanner) ListResourcesWithDiagnosticSettings() (map[string]bool, error) {
	resources := []string{}
	res := map[string]bool{}

	azqr.LogSubscriptionScan(d.config.SubscriptionID, "Resource Ids")

	result := d.graphQuery.Query(d.config.Ctx, "resources | project id | order by id asc", []*string{&d.config.SubscriptionID})

	if result == nil || result.Data == nil {
		log.Info().Msg("Preflight: No resources found")
		return res, nil
	}

	for _, row := range result.Data {
		m := row.(map[string]interface{})
		resources = append(resources, strings.ToLower(m["id"].(string)))
	}

	batches := int(math.Ceil(float64(len(resources)) / 20))

	var wg sync.WaitGroup
	ch := make(chan map[string]bool, 5)
	wg.Add(batches)

	go func() {
		wg.Wait()
		close(ch)
	}()

	azqr.LogSubscriptionScan(d.config.SubscriptionID, "Diagnostic Settings")

	// Split resources into batches of 20 items.
	batchSize := 20
	for i := 0; i < len(resources); i += batchSize {
		j := i + batchSize
		if j > len(resources) {
			j = len(resources)
		}
		go func(r []string) {
			defer wg.Done()
			resp, err := d.restCall(d.config.Ctx, r)
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
			ch <- asyncRes
		}(resources[i:j])
	}

	for i := 0; i < batches; i++ {
		for k, v := range <-ch {
			res[k] = v
		}
	}

	return res, nil
}

const (
	moduleName    = "armresources"
	moduleVersion = "v1.1.1"
)

func (d *DiagnosticSettingsScanner) restCall(ctx context.Context, resourceIds []string) (*ArmBatchResponse, error) {
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
			RelativeUrl: resourceId + "/providers/microsoft.insights/diagnosticSettings?api-version=2021-05-01-preview",
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

func (d *DiagnosticSettingsScanner) Scan(config *azqr.ScannerConfig) map[string]bool {
	err := d.Init(config)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize Diagnostic Settings Scanner")
	}
	diagResults, err := d.ListResourcesWithDiagnosticSettings()
	if err != nil {
		if azqr.ShouldSkipError(err) {
			diagResults = map[string]bool{}
		} else {
			log.Fatal().Err(err).Msg("Failed to list resources with Diagnostic Settings")
		}
	}
	return diagResults
}
