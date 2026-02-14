// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package pipeline

import (
	"fmt"
	"time"

	"github.com/Azure/azqr/internal/az"
	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/renderers"
	"github.com/Azure/azqr/internal/scanners/registry"
	"github.com/rs/zerolog/log"
)

// InitializationStage sets up the scan context (credentials, subscriptions, etc.)
type InitializationStage struct {
	*BaseStage
}

func NewInitializationStage() *InitializationStage {
	return &InitializationStage{
		BaseStage: NewBaseStage("Initialization", true),
	}
}

func (s *InitializationStage) Execute(ctx *ScanContext) error {
	// Step 1: Check scanner registry
	s.logScannerRegistryInfo()

	// Step 2: Generate output file name
	outputFile := s.generateOutputFileName(ctx.Params.OutputName)
	ctx.Params.OutputName = outputFile

	// Step 3: Validate and prepare filters
	s.validateAndPrepareFilters(ctx.Params)

	// Step 4: Create Azure credentials
	ctx.Cred = az.NewAzureCredential()

	// Step 5: Create client options
	ctx.ClientOptions = az.NewDefaultClientOptions()

	// Step 6: Initialize report data
	reportData := renderers.NewReportData(
		outputFile,
		ctx.Params.Mask,
		ctx.Params.Stages,
	)
	ctx.ReportData = &reportData

	log.Debug().Msg("Initialization stage completed")
	return nil
}

// logScannerRegistryInfo logs information about registered scanners (debug mode)
func (s *InitializationStage) logScannerRegistryInfo() {
	scannerInfo := registry.ListScannerInfo()

	log.Debug().
		Int("total_scanners", registry.GetScannerCount()).
		Int("unique_keys", len(registry.GetScannerKeys())).
		Int("scanner_types", len(scannerInfo)).
		Msg("Scanner registry initialized")

	// Log sample of registered scanners
	sampleSize := min(5, len(scannerInfo))

	for i := 0; i < sampleSize; i++ {
		info := scannerInfo[i]
		log.Debug().
			Str("key", info.Key).
			Strs("resource_types", info.ResourceTypes).
			Int("count", info.ScannerCount).
			Msg("Registered scanner")
	}

	if len(scannerInfo) > sampleSize {
		log.Debug().Msgf("... and %d more scanners", len(scannerInfo)-sampleSize)
	}
}

// generateOutputFileName generates output file name from params
func (s *InitializationStage) generateOutputFileName(outputName string) string {
	if outputName != "" {
		return outputName
	}

	current_time := time.Now()
	outputFileStamp := fmt.Sprintf("%d_%02d_%02d_T%02d%02d%02d",
		current_time.Year(), current_time.Month(), current_time.Day(),
		current_time.Hour(), current_time.Minute(), current_time.Second())

	return fmt.Sprintf("%s_%s", "azqr_action_plan", outputFileStamp)
}

// validateAndPrepareFilters validates input parameters and prepares filters
func (s *InitializationStage) validateAndPrepareFilters(params *models.ScanParams) {
	filters := params.Filters

	log.Debug().
		Int("scanners_before", len(filters.Azqr.Scanners)).
		Msg("Filters validation starting")

	// validate input
	if len(params.ManagementGroups) > 0 && (len(params.Subscriptions) > 0 || len(params.ResourceGroups) > 0) {
		log.Fatal().Msg("Management Group name cannot be used with a Subscription Id or Resource Group name")
	}

	if len(params.Subscriptions) < 1 && len(params.ResourceGroups) > 0 {
		log.Fatal().Msg("Resource Group name can only be used with a Subscription Id")
	}

	if len(params.Subscriptions) > 1 && len(params.ResourceGroups) > 0 {
		log.Fatal().Msg("Resource Group name can only be used with 1 Subscription Id")
	}

	if len(params.Subscriptions) > 0 {
		for _, sub := range params.Subscriptions {
			filters.Azqr.AddSubscription(sub)
		}
	}

	if len(params.ResourceGroups) > 0 {
		for _, rg := range params.ResourceGroups {
			filters.Azqr.AddResourceGroup(fmt.Sprintf("/subscriptions/%s/resourceGroups/%s", params.Subscriptions[0], rg))
		}
	}

	log.Debug().
		Int("scanners_after", len(filters.Azqr.Scanners)).
		Msg("Filters validation completed")
}
