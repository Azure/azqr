// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package scanners

import (
	"context"
	"math"
	"net/http"
	"strings"

	"github.com/Azure/azqr/internal/azqr"
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

	batches := int(math.Ceil(float64(len(resources)) / 20))

	ch := make(chan map[string]bool, 5)

	maxConcurrency := 200 // max number of concurrent requests
	limiter := make(chan struct{}, maxConcurrency)

	azqr.LogResourceTypeScan("Diagnostic Settings")

	// Split resources into batches of 20 items.
	batchSize := 20
	for i := 0; i < len(resources); i += batchSize {
		j := i + batchSize
		if j > len(resources) {
			j = len(resources)
		}
		limiter <- struct{}{} // Acquire a token. Waits here for token releases from the limiter. 
		go func(r []*string) {
			defer func() { <-limiter }() // Release the token
			resp, err := d.restCall(d.ctx, r)
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
		if azqr.ShouldSkipError(err) {
			diagResults = map[string]bool{}
		} else {
			log.Fatal().Err(err).Msg("Failed to list resources with Diagnostic Settings")
		}
	}
	return diagResults
}
