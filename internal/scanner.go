// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package internal

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Azure/azqr/internal/az"
	"github.com/Azure/azqr/internal/graph"
	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/plugins"
	"github.com/Azure/azqr/internal/renderers"
	"github.com/Azure/azqr/internal/renderers/csv"
	"github.com/Azure/azqr/internal/renderers/excel"
	"github.com/Azure/azqr/internal/renderers/json"
	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	_ "github.com/Azure/azqr/internal/scanners/aa"
	_ "github.com/Azure/azqr/internal/scanners/adf"
	_ "github.com/Azure/azqr/internal/scanners/afd"
	_ "github.com/Azure/azqr/internal/scanners/afw"
	_ "github.com/Azure/azqr/internal/scanners/agw"
	_ "github.com/Azure/azqr/internal/scanners/aif"
	_ "github.com/Azure/azqr/internal/scanners/aks"
	_ "github.com/Azure/azqr/internal/scanners/amg"
	_ "github.com/Azure/azqr/internal/scanners/apim"
	_ "github.com/Azure/azqr/internal/scanners/appcs"
	_ "github.com/Azure/azqr/internal/scanners/appi"
	_ "github.com/Azure/azqr/internal/scanners/arc"
	_ "github.com/Azure/azqr/internal/scanners/as"
	_ "github.com/Azure/azqr/internal/scanners/asp"
	_ "github.com/Azure/azqr/internal/scanners/avail"
	_ "github.com/Azure/azqr/internal/scanners/avd"
	_ "github.com/Azure/azqr/internal/scanners/avs"
	_ "github.com/Azure/azqr/internal/scanners/ba"
	_ "github.com/Azure/azqr/internal/scanners/ca"
	_ "github.com/Azure/azqr/internal/scanners/cae"
	_ "github.com/Azure/azqr/internal/scanners/ci"
	_ "github.com/Azure/azqr/internal/scanners/conn"
	_ "github.com/Azure/azqr/internal/scanners/cosmos"
	_ "github.com/Azure/azqr/internal/scanners/cr"
	_ "github.com/Azure/azqr/internal/scanners/dbw"
	_ "github.com/Azure/azqr/internal/scanners/dec"
	_ "github.com/Azure/azqr/internal/scanners/disk"
	_ "github.com/Azure/azqr/internal/scanners/erc"
	_ "github.com/Azure/azqr/internal/scanners/evgd"
	_ "github.com/Azure/azqr/internal/scanners/evh"
	_ "github.com/Azure/azqr/internal/scanners/fabric"
	_ "github.com/Azure/azqr/internal/scanners/fdfp"
	_ "github.com/Azure/azqr/internal/scanners/gal"
	_ "github.com/Azure/azqr/internal/scanners/hpc"
	_ "github.com/Azure/azqr/internal/scanners/hub"
	_ "github.com/Azure/azqr/internal/scanners/iot"
	_ "github.com/Azure/azqr/internal/scanners/it"
	_ "github.com/Azure/azqr/internal/scanners/kv"
	_ "github.com/Azure/azqr/internal/scanners/lb"
	_ "github.com/Azure/azqr/internal/scanners/log"
	_ "github.com/Azure/azqr/internal/scanners/logic"
	_ "github.com/Azure/azqr/internal/scanners/maria"
	_ "github.com/Azure/azqr/internal/scanners/mysql"
	_ "github.com/Azure/azqr/internal/scanners/netapp"
	_ "github.com/Azure/azqr/internal/scanners/ng"
	_ "github.com/Azure/azqr/internal/scanners/nic"
	_ "github.com/Azure/azqr/internal/scanners/nsg"
	_ "github.com/Azure/azqr/internal/scanners/nw"
	_ "github.com/Azure/azqr/internal/scanners/odb"
	_ "github.com/Azure/azqr/internal/scanners/pdnsz"
	_ "github.com/Azure/azqr/internal/scanners/pep"
	_ "github.com/Azure/azqr/internal/scanners/pip"
	_ "github.com/Azure/azqr/internal/scanners/psql"
	_ "github.com/Azure/azqr/internal/scanners/redis"
	_ "github.com/Azure/azqr/internal/scanners/rg"
	_ "github.com/Azure/azqr/internal/scanners/rsv"
	_ "github.com/Azure/azqr/internal/scanners/rt"
	_ "github.com/Azure/azqr/internal/scanners/sap"
	_ "github.com/Azure/azqr/internal/scanners/sb"
	_ "github.com/Azure/azqr/internal/scanners/sigr"
	_ "github.com/Azure/azqr/internal/scanners/sql"
	_ "github.com/Azure/azqr/internal/scanners/srch"
	_ "github.com/Azure/azqr/internal/scanners/st"
	_ "github.com/Azure/azqr/internal/scanners/synw"
	_ "github.com/Azure/azqr/internal/scanners/traf"
	_ "github.com/Azure/azqr/internal/scanners/vdpool"
	_ "github.com/Azure/azqr/internal/scanners/vgw"
	_ "github.com/Azure/azqr/internal/scanners/vm"
	_ "github.com/Azure/azqr/internal/scanners/vmss"
	_ "github.com/Azure/azqr/internal/scanners/vnet"
	_ "github.com/Azure/azqr/internal/scanners/vwan"
	_ "github.com/Azure/azqr/internal/scanners/wps"
)

type (
	ScanParams struct {
		ManagementGroups       []string
		Subscriptions          []string
		ResourceGroups         []string
		OutputName             string
		Defender               bool
		Advisor                bool
		Arc                    bool
		Xlsx                   bool
		Cost                   bool
		Mask                   bool
		Csv                    bool
		Json                   bool
		Stdout                 bool
		Debug                  bool
		Policy                 bool
		ScannerKeys            []string
		Filters                *models.Filters
		UseAzqrRecommendations bool
		UseAprlRecommendations bool
		EnabledInternalPlugins map[string]bool
	}

	Scanner struct{}
)

func NewScanParams() *ScanParams {
	return &ScanParams{
		ManagementGroups:       []string{},
		Subscriptions:          []string{},
		ResourceGroups:         []string{},
		OutputName:             "",
		Defender:               true,
		Advisor:                true,
		Cost:                   true,
		Mask:                   true,
		Csv:                    false,
		Json:                   false,
		Debug:                  false,
		Policy:                 false,
		ScannerKeys:            []string{},
		Filters:                models.NewFilters(),
		UseAzqrRecommendations: true,
		UseAprlRecommendations: true,
	}
}

// initializeLogLevel sets the global log level based on debug flag
func (sc Scanner) initializeLogLevel(debug bool) {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		log.Debug().Msg("Debug logging enabled")
	}
}

func (sc Scanner) Scan(params *ScanParams) string {
	startTime := time.Now()
	sc.initializeLogLevel(params.Debug)

	// generate output file name
	outputFile := sc.generateOutputFileName(params.OutputName)

	// validate and prepare filters
	sc.validateAndPrepareFilters(params)

	serviceScanners := params.Filters.Azqr.Scanners

	// create Azure credentials and context
	cred := az.NewAzureCredential()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	clientOptions := az.NewDefaultClientOptions()

	// list subscriptions
	subscriptions := sc.listSubscriptions(ctx, cred, params, clientOptions)

	// initialize scanners
	defenderScanner := scanners.DefenderScanner{}
	pipScanner := scanners.PublicIPScanner{}
	peScanner := scanners.PrivateEndpointScanner{}
	diagnosticsScanner := scanners.DiagnosticSettingsScanner{}
	advisorScanner := scanners.AdvisorScanner{}
	azurePolicyScanner := scanners.AzurePolicyScanner{}
	arcSQLScanner := scanners.ArcSQLScanner{}
	costScanner := scanners.CostScanner{}
	diagResults := map[string]bool{}

	// initialize report data
	reportData := renderers.NewReportData(outputFile, params.Mask, params.Policy, params.Arc, params.Defender, params.Advisor, params.Cost, true)

	resourceScanner := scanners.ResourceScanner{}
	reportData.Resources, reportData.ExludedResources = resourceScanner.GetAllResources(ctx, cred, subscriptions, params.Filters)

	// Check if the number of resources exceeds Excel's row limit (1,048,576 rows) - 10 rows reserved for headers
	const excelMaxRows = 1048566
	if len(reportData.Resources) > excelMaxRows {
		log.Fatal().Msgf("Number of resources (%d) exceeds Excel's maximum row limit (%d). Aborting scan.", len(reportData.Resources), excelMaxRows)
	}

	aprlScanner := graph.NewAprlScanner(serviceScanners, params.Filters, subscriptions)

	// Register YAML plugin recommendations with the APRL scanner
	yamlPluginRegistry := plugins.GetRegistry()
	for _, plugin := range yamlPluginRegistry.List() {
		if len(plugin.YamlRecommendations) > 0 {
			log.Info().
				Str("plugin", plugin.Metadata.Name).
				Int("queries", len(plugin.YamlRecommendations)).
				Msg("Registering YAML plugin queries with APRL scanner")
			for _, rec := range plugin.YamlRecommendations {
				aprlScanner.RegisterExternalQuery(rec.ResourceType, rec)
			}
		}
	}

	reportData.Recommendations, _ = aprlScanner.ListRecommendations()

	resourceTypes := resourceScanner.GetCountPerResourceType(ctx, cred, subscriptions, params.Filters)

	// Filter service scanners to include only those with resource types present across all subscriptions
	filteredServiceScanners := sc.buildFilteredServiceScanners(serviceScanners, resourceTypes)

	// Get plugin scanners - they will be executed separately
	pluginRegistry := plugins.GetRegistry()
	registeredPlugins := pluginRegistry.List()
	log.Debug().Msgf("Found %d registered plugins", len(registeredPlugins))
	for _, plugin := range registeredPlugins {
		if plugin.InternalScanner != nil {
			log.Debug().
				Str("plugin", plugin.Metadata.Name).
				Str("type", "internal").
				Str("version", plugin.Metadata.Version).
				Msg("Internal plugin ready for execution")
		}
	}

	// get the APRL scan results (built-in APRL recommendations only)
	aprlScanner = graph.NewAprlScanner(filteredServiceScanners, params.Filters, subscriptions)

	// Register YAML plugin recommendations with the APRL scanner
	for _, plugin := range yamlPluginRegistry.List() {
		if len(plugin.YamlRecommendations) > 0 {
			for _, rec := range plugin.YamlRecommendations {
				aprlScanner.RegisterExternalQuery(rec.ResourceType, rec)
			}
		}
	}

	reportData.Aprl = aprlScanner.Scan(ctx, cred)

	// Execute internal plugin scanners
	reportData.PluginResults = sc.executePlugins(ctx, cred, subscriptions, params)

	// get the count of resources per resource type
	reportData.ResourceTypeCount = resourceScanner.GetCountPerResourceTypeAndSubscription(ctx, cred, subscriptions, reportData.Recommendations, params.Filters)

	// Build a map of subscription ID -> resource type -> count for efficient lookup
	subscriptionResourceTypeMap := sc.buildSubscriptionResourceTypeMap(reportData.ResourceTypeCount, subscriptions)

	// For each service scanner, get the recommendations list
	if params.UseAzqrRecommendations {
		for _, s := range serviceScanners {
			for i, r := range s.GetRecommendations() {
				if params.Filters.Azqr.IsRecommendationExcluded(r.RecommendationID) {
					continue
				}

				if r.RecommendationType != models.TypeRecommendation {
					continue
				}

				if reportData.Recommendations[strings.ToLower(r.ResourceType)] == nil {
					reportData.Recommendations[strings.ToLower(r.ResourceType)] = map[string]models.AprlRecommendation{}
				}

				reportData.Recommendations[strings.ToLower(r.ResourceType)][i] = r.ToAzureAprlRecommendation()
			}
		}

		// scan diagnostic settings
		err := diagnosticsScanner.Init(ctx, cred, clientOptions)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to initialize diagnostic settings scanner")
		}

		diagResults = diagnosticsScanner.Scan(reportData.Resources)
	}

	// scan each subscription with AZQR scanners
	for sid, sn := range subscriptions {
		config := &models.ScannerConfig{
			Ctx:              ctx,
			SubscriptionID:   sid,
			SubscriptionName: sn,
			Cred:             cred,
			ClientOptions:    clientOptions,
		}

		if params.UseAzqrRecommendations {
			// Filter service scanners for this subscription only
			subscriptionFilteredScanners := sc.buildSubscriptionFilteredScanners(
				serviceScanners,
				subscriptionResourceTypeMap[sid],
				sid,
				params.Mask,
			)

			// Skip scanning if no scanners are needed for this subscription
			if len(subscriptionFilteredScanners) == 0 {
				log.Info().Msgf("No scanners needed for subscription %s, skipping AZQR scan", renderers.MaskSubscriptionID(sid, params.Mask))
				continue
			}

			// scan private endpoints
			peResults := peScanner.Scan(config)

			// scan public IPs
			pips := pipScanner.Scan(config)

			// initialize scan context
			scanContext := models.ScanContext{
				Filters:             params.Filters,
				PrivateEndpoints:    peResults,
				DiagnosticsSettings: diagResults,
				PublicIPs:           pips,
			}

			// Worker pool to limit concurrent scanner goroutines
			const numScannerWorkers = 10
			jobs := make(chan models.IAzureScanner, len(subscriptionFilteredScanners))
			results := make(chan []*models.AzqrServiceResult, len(subscriptionFilteredScanners))

			// Start worker pool
			var workerWg sync.WaitGroup
			for w := 0; w < numScannerWorkers; w++ {
				workerWg.Add(1)
				go func() {
					defer workerWg.Done()
					for s := range jobs {
						// Initialize scanner with this subscription's config
						err := s.Init(config)
						if err != nil {
							log.Fatal().Err(err).Msg("Failed to initialize scanner")
						}

						res, err := sc.retry(3, 10*time.Millisecond, s, &scanContext)
						if err != nil {
							cancel()
							log.Fatal().Err(err).Msg("Failed to scan")
						}
						results <- res
					}
				}()
			}

			// Send scanner jobs to workers
			go func() {
				for _, s := range subscriptionFilteredScanners {
					jobs <- s
				}
				close(jobs)
			}()

			// Wait for workers to finish and close results channel
			go func() {
				workerWg.Wait()
				close(results)
			}()

			// Collect results from all scanners
			for res := range results {
				for _, r := range res {
					// check if the resource is excluded
					if params.Filters.Azqr.IsServiceExcluded(r.ResourceID()) {
						continue
					}
					reportData.Azqr = append(reportData.Azqr, r)
				}
			}
		}

		// scan costs
		costs := costScanner.Scan(params.Cost, config)
		reportData.Cost.From = costs.From
		reportData.Cost.To = costs.To
		reportData.Cost.Items = append(reportData.Cost.Items, costs.Items...)
	}

	// scan advisor
	reportData.Advisor = append(reportData.Advisor, advisorScanner.Scan(ctx, params.Defender, cred, subscriptions, params.Filters)...)

	// scan Azure Policy
	if params.Policy {
		reportData.AzurePolicy = append(reportData.AzurePolicy, azurePolicyScanner.Scan(ctx, cred, subscriptions, params.Filters)...)
	}

	// scan Arc-enabled SQL Server
	if params.Arc {
		reportData.ArcSQL = append(reportData.ArcSQL, arcSQLScanner.Scan(ctx, cred, subscriptions, params.Filters)...)
	}

	// scan defender
	reportData.Defender = append(reportData.Defender, defenderScanner.Scan(ctx, params.Defender, cred, subscriptions, params.Filters)...)

	// get the defender recommendations
	reportData.DefenderRecommendations = append(reportData.DefenderRecommendations, defenderScanner.GetRecommendations(ctx, params.Defender, cred, subscriptions, params.Filters)...)

	// Render reports
	outputJson := sc.renderReports(&reportData, params)

	elapsedTime := time.Since(startTime)
	// Format the elapsed time as HH:MM:SS and log the scan completion time
	hours := int(elapsedTime.Hours())
	minutes := int(elapsedTime.Minutes()) % 60
	seconds := int(elapsedTime.Seconds()) % 60
	log.Info().Msgf("Scan completed in %02d:%02d:%02d", hours, minutes, seconds)

	return outputJson
}

// ScanPlugins performs a fast plugin-only scan without resource or APRL scanning
func (sc Scanner) ScanPlugins(params *ScanParams) string {
	startTime := time.Now()

	sc.initializeLogLevel(params.Debug)

	log.Info().Msg("Running in plugin-only mode: skipping resource and APRL scanning")

	// generate output file name
	outputFile := sc.generateOutputFileName(params.OutputName)

	// validate and prepare filters
	sc.validateAndPrepareFilters(params)

	// create Azure credentials and context
	cred := az.NewAzureCredential()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	clientOptions := az.NewDefaultClientOptions()

	// list subscriptions
	subscriptions := sc.listSubscriptions(ctx, cred, params, clientOptions)

	// initialize minimal report data for plugin-only mode
	reportData := renderers.NewReportData(outputFile, params.Mask, false, false, false, false, false, false)

	// Execute plugins and collect results
	reportData.PluginResults = sc.executePlugins(ctx, cred, subscriptions, params)

	// Render reports
	outputJson := sc.renderReports(&reportData, params)

	elapsedTime := time.Since(startTime)
	hours := int(elapsedTime.Hours())
	minutes := int(elapsedTime.Minutes()) % 60
	seconds := int(elapsedTime.Seconds()) % 60
	log.Info().Msgf("Plugin-only scan completed in %02d:%02d:%02d", hours, minutes, seconds)

	return outputJson
}

// retry retries the Azure scanner Scan, a number of times with an increasing delay between retries
func (sc Scanner) retry(attempts int, sleep time.Duration, a models.IAzureScanner, scanContext *models.ScanContext) ([]*models.AzqrServiceResult, error) {
	var err error
	for i := 0; ; i++ {
		res, err := a.Scan(scanContext)
		if err == nil {
			return res, nil
		}

		if models.ShouldSkipError(err) {
			return []*models.AzqrServiceResult{}, nil
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

func (sc Scanner) generateOutputFileName(outputName string) string {
	outputFile := outputName
	if outputFile == "" {
		current_time := time.Now()
		outputFileStamp := fmt.Sprintf("%d_%02d_%02d_T%02d%02d%02d",
			current_time.Year(), current_time.Month(), current_time.Day(),
			current_time.Hour(), current_time.Minute(), current_time.Second())

		outputFile = fmt.Sprintf("%s_%s", "azqr_action_plan", outputFileStamp)
	}
	return outputFile
}

// buildSubscriptionResourceTypeMap creates a map of subscription ID -> resource type -> count
// from the ResourceTypeCount data for efficient per-subscription lookups
func (sc Scanner) buildSubscriptionResourceTypeMap(resourceTypeCounts []models.ResourceTypeCount, subscriptions map[string]string) map[string]map[string]float64 {
	subscriptionResourceTypeMap := make(map[string]map[string]float64)
	for _, rtc := range resourceTypeCounts {
		// Find subscription ID by name
		for sid, sname := range subscriptions {
			if sname == rtc.Subscription {
				if subscriptionResourceTypeMap[sid] == nil {
					subscriptionResourceTypeMap[sid] = make(map[string]float64)
				}
				subscriptionResourceTypeMap[sid][strings.ToLower(rtc.ResourceType)] = rtc.Count
				break
			}
		}
	}
	return subscriptionResourceTypeMap
}

// buildFilteredServiceScanners filters service scanners based on resource types
// present across all subscriptions, returning only scanners that have resources to scan
func (sc Scanner) buildFilteredServiceScanners(serviceScanners []models.IAzureScanner, resourceTypes map[string]float64) []models.IAzureScanner {
	var filteredScanners []models.IAzureScanner
	for _, s := range serviceScanners {
		add := true
		for _, resourceType := range s.ResourceTypes() {
			resourceType = strings.ToLower(resourceType)

			// Check if the resource type exists across any subscription
			if count, exists := resourceTypes[resourceType]; !exists || count <= 0 {
				log.Debug().Msgf("Skipping scanner for resource type %s as it has no resources", resourceType)
				continue
			} else if add {
				filteredScanners = append(filteredScanners, s)
				add = false
				log.Info().Msgf("Scanner for resource type %s will be used", resourceType)
			}
		}
	}
	return filteredScanners
}

// validateAndPrepareFilters validates input parameters and prepares filters
func (sc Scanner) validateAndPrepareFilters(params *ScanParams) {
	filters := params.Filters

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
}

// listSubscriptions returns subscriptions based on management groups or subscription IDs
func (sc Scanner) listSubscriptions(ctx context.Context, cred azcore.TokenCredential, params *ScanParams, clientOptions *arm.ClientOptions) map[string]string {
	if len(params.ManagementGroups) > 0 {
		managementGroupScanner := scanners.ManagementGroupsScanner{}
		return managementGroupScanner.ListSubscriptions(ctx, cred, params.ManagementGroups, params.Filters, clientOptions)
	}
	subscriptionScanner := scanners.SubcriptionScanner{}
	return subscriptionScanner.ListSubscriptions(ctx, cred, params.Subscriptions, params.Filters, clientOptions)
}

// executePlugins executes enabled internal plugins and returns results
func (sc Scanner) executePlugins(ctx context.Context, cred azcore.TokenCredential, subscriptions map[string]string, params *ScanParams) []renderers.PluginResult {
	var pluginResults []renderers.PluginResult

	// Get plugin scanners
	pluginRegistry := plugins.GetRegistry()
	registeredPlugins := pluginRegistry.List()
	log.Debug().Msgf("Found %d registered plugins", len(registeredPlugins))

	var internalPluginScanners []plugins.InternalPluginScanner
	for _, plugin := range registeredPlugins {
		if plugin.InternalScanner != nil {
			log.Debug().
				Str("plugin", plugin.Metadata.Name).
				Str("type", "internal").
				Str("version", plugin.Metadata.Version).
				Msg("Internal plugin ready for execution")
			internalPluginScanners = append(internalPluginScanners, plugin.InternalScanner)
		}
	}

	// Execute internal plugin scanners (only if enabled)
	if len(internalPluginScanners) > 0 {
		for _, internalScanner := range internalPluginScanners {
			pluginName := internalScanner.GetMetadata().Name
			// Check if plugin is enabled (default: false)
			enabled := false
			if params.EnabledInternalPlugins != nil {
				enabled = params.EnabledInternalPlugins[pluginName]
			}

			if !enabled {
				log.Debug().
					Str("plugin", pluginName).
					Msg("Internal plugin skipped (not enabled)")
				continue
			}

			log.Info().
				Str("plugin", pluginName).
				Str("type", "internal").
				Str("version", internalScanner.GetMetadata().Version).
				Msg("Executing internal plugin")

			output, err := internalScanner.Scan(ctx, cred, subscriptions, params.Filters)
			if err != nil {
				log.Error().
					Err(err).
					Str("plugin", pluginName).
					Msg("Internal plugin execution failed")
				continue
			}

			// Add internal plugin results
			pluginResults = append(pluginResults, renderers.PluginResult{
				PluginName:  pluginName,
				SheetName:   output.SheetName,
				Description: output.Description,
				Table:       output.Table,
			})

			log.Debug().
				Str("plugin", pluginName).
				Int("rows", len(output.Table)-1).
				Msg("Internal plugin completed")
		}
	}

	return pluginResults
}

// renderReports generates all requested output formats
func (sc Scanner) renderReports(reportData *renderers.ReportData, params *ScanParams) string {
	if params.Xlsx {
		excel.CreateExcelReport(reportData)
	}

	if params.Json {
		json.CreateJsonReport(reportData)
	}

	if params.Csv {
		csv.CreateCsvReport(reportData)
	}

	outputJson := json.CreateJsonOutput(reportData)
	if params.Stdout {
		fmt.Println(outputJson)
	}

	return outputJson
}

// buildSubscriptionFilteredScanners filters service scanners based on resource types
// present in a specific subscription, returning only scanners needed for that subscription
func (sc Scanner) buildSubscriptionFilteredScanners(serviceScanners []models.IAzureScanner, subscriptionResourceTypes map[string]float64, subscriptionID string, mask bool) []models.IAzureScanner {
	if subscriptionResourceTypes == nil {
		subscriptionResourceTypes = make(map[string]float64)
	}

	var filteredScanners []models.IAzureScanner
	for _, s := range serviceScanners {
		add := true
		for _, resourceType := range s.ResourceTypes() {
			resourceType = strings.ToLower(resourceType)

			// Check if the resource type is in this subscription's resource types
			if count, exists := subscriptionResourceTypes[resourceType]; !exists || count <= 0 {
				log.Debug().Msgf("Skipping scanner for resource type %s in subscription %s as it has no resources", resourceType, renderers.MaskSubscriptionID(subscriptionID, mask))
				continue
			} else if add {
				filteredScanners = append(filteredScanners, s)
				add = false
				log.Info().Msgf("Scanner for resource type %s will be used in subscription %s", resourceType, renderers.MaskSubscriptionID(subscriptionID, mask))
			}
		}
	}
	return filteredScanners
}
