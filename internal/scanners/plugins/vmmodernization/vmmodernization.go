// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

// Package vmmodernization is an internal plugin that reports Azure VM v6/v7
// series modernization readiness: Discover-phase workload pattern, Assess-phase
// readiness signals (Hyper-V generation, disk controller type, series generation,
// Azure Disk Encryption), and a Plan-phase execution-method recommendation with a
// suggested target v6/v7 SKU.
package vmmodernization

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/Azure/azqr/internal/az"
	"github.com/Azure/azqr/internal/graph"
	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/plugins"
	"github.com/Azure/azqr/internal/skus"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// argBool decodes a computed boolean column from an Azure Resource Graph KQL
// expression (isnotnull/isnotempty/case/iff). ARG serializes these as the JSON
// number 0 or 1, not a JSON true/false literal — confirmed by a live query
// against this plugin's own discover-assess.kql expressions via the Resource
// Graph REST API, not merely assumed. Plain JSON bool is still accepted for
// forward-compatibility if ARG's serialization ever changes.
type argBool bool

func (b *argBool) UnmarshalJSON(data []byte) error {
	switch string(data) {
	case "0", "null":
		*b = false
		return nil
	case "1":
		*b = true
		return nil
	default:
		var v bool
		if err := json.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("argBool: cannot unmarshal %s: %w", string(data), err)
		}
		*b = argBool(v)
		return nil
	}
}

//go:embed kql/discover-assess.kql
var discoverAssessQuery string

//go:embed kql/discover-assess-avd.kql
var discoverAssessAVDQuery string

//go:embed kql/discover-assess-aro.kql
var discoverAssessAROQuery string

//go:embed kql/discover-assess-vmss.kql
var discoverAssessVMSSQuery string

//go:embed kql/discover-assess-backup.kql
var discoverAssessBackupQuery string

// manualReviewRequired is the explicit flag value used whenever azqr cannot
// establish a hard-gate verdict (temp-disk dependency, SAP/ISV certification)
// from available data, instead of guessing Ready/Blocked (FR-004).
const manualReviewRequired = "Manual Review Required"

// alreadyOnTargetGeneration is the Suggested Target SKU value for a VM whose
// current size is already on a v6/v7 series.
const alreadyOnTargetGeneration = "Already on target generation"

// notPoolManaged is the fallback value for the three pool-level columns
// (P02) on a row with no vmssId (standalone VM, Availability Set member) —
// there is no parent scale set for these columns to describe.
const notPoolManaged = "N/A (not pool-managed)"

// Scanner is an internal plugin that scans VM v6/v7 modernization readiness.
type Scanner struct{}

// NewScanner creates a new VM modernization scanner.
func NewScanner() *Scanner {
	return &Scanner{}
}

// RegisterFlags registers plugin-specific flags (implements FlagProvider).
// target-regions is shared with the region-selection plugin so one set of
// regions drives both; when unset each VM is checked in its own region.
func (s *Scanner) RegisterFlags(cmd *cobra.Command) {
	cmd.Flags().StringSlice("target-regions", []string{}, "Target regions to check SKU availability and quota in (comma-separated, e.g., eastus,westeurope). Defaults to each VM's own region")
}

// GetMetadata returns plugin metadata.
func (s *Scanner) GetMetadata() plugins.PluginMetadata {
	return plugins.PluginMetadata{
		Name:        "vm-modernization",
		Version:     "0.2.0",
		Description: "Reports Azure VM v6/v7 series modernization readiness: Discover workload pattern, Assess readiness signals (Hyper-V generation, disk controller type, series generation, Azure Disk Encryption, heuristic Gen2/NVMe/MANA OS verdicts), a Plan-phase execution-method recommendation with a suggested target v6/v7 SKU, and a SKU availability and VM family quota sheet",
		Author:      "Azure Quick Review Team",
		License:     "MIT",
		Type:        plugins.PluginTypeInternal,
		// Column order is triage-oriented: an engineer reads left to right from
		// "what is this VM" to "what should I do" to "what blocks it" to raw data.
		// toRecord must emit values in exactly this order, and the pinned header
		// list in the tests must be updated with every change. Add each new column
		// to the group it belongs to, not at the end; the raw inventory group is
		// the only one meant to grow with unprocessed ARG fields.
		ColumnMetadata: []plugins.ColumnMetadata{
			// Identify: which VM and which workload pattern it was classified as.
			{Name: "Subscription"},
			{Name: "Resource Group"},
			{Name: "Name"},
			{Name: "Location"},
			{Name: "Current VM Size"},
			{Name: "Series Generation"},
			{Name: "Workload Pattern"},
			// Decide: the verdict, the route, and how hard the move is.
			{Name: "Platform Readiness Outcome"},
			{Name: "Recommended Execution Method"},
			{Name: "Modernization Complexity"},
			{Name: "Complexity Reason"},
			// Target: where to move to and what changes there.
			{Name: "Suggested Target SKU"},
			{Name: "Target SKU Compatibility Score"},
			{Name: "Documented Replacement Series"},
			{Name: "Series Specification Differences"},
			{Name: "Lifecycle/Pricing Notes"},
			// Blockers: signals that gate or complicate the move.
			{Name: "Generation 2 Ready"},
			{Name: "Remote NVMe OS Supported"},
			{Name: "MANA OS Supported"},
			{Name: "Hyper-V Generation"},
			{Name: "Disk Controller Type"},
			{Name: "Temp Disk Dependency"},
			{Name: "SAP/ISV Certification Gate"},
			{Name: "Azure Disk Encryption"},
			{Name: "Hibernation Enabled"},
			{Name: "Backup Policy Subtype"},
			{Name: "VM Power State"},
			{Name: "Instance View Captured"},
			// Cost: existing-SKU retail exposure.
			{Name: "Current Monthly Cost (USD)"},
			{Name: "Projected Monthly Cost After Increase (USD)"},
			{Name: "Monthly Cost Impact (USD)"},
			// Pool context: scale set, availability set and pool-model details.
			{Name: "Pool Membership"},
			{Name: "Membership Model"},
			{Name: "Orchestration Mode"},
			{Name: "VMSS Name"},
			{Name: "Pool Model VM Size"},
			{Name: "Pool Model Series Generation"},
			{Name: "Availability Set Name"},
			// Raw inventory: unprocessed ARG fields, kept last.
			{Name: "Zone Placement"},
			{Name: "Availability Zone"},
			{Name: "Platform Fault Domain"},
			{Name: "OS Type"},
			{Name: "OS Version"},
			{Name: "Computer Name"},
			{Name: "Is Marketplace VM"},
			{Name: "Image Publisher"},
			{Name: "Image Offer"},
			{Name: "Image SKU"},
			{Name: "Image Version"},
			{Name: "OS Disk Type"},
			{Name: "OS Disk Size (GB)"},
			{Name: "Ephemeral OS Disk"},
			{Name: "Data Disk Count"},
			{Name: "Ultra SSD Enabled"},
			{Name: "License Type"},
			{Name: "NIC Count"},
			{Name: "Accelerated Networking"},
			{Name: "IP Forwarding"},
			{Name: "Security Type"},
			{Name: "Secure Boot Enabled"},
			{Name: "vTPM Enabled"},
		},
	}
}

// Scan executes the plugin and returns table data.
func (s *Scanner) Scan(ctx context.Context, cred azcore.TokenCredential, subscriptions map[string]string, params *models.ScanParams) ([]plugins.ExternalPluginOutput, error) {
	models.LogResourceTypeScan("VM v6/v7 Modernization Readiness")

	graphClient := graph.NewGraphQuery(cred)

	log.Debug().Msg("Executing VM modernization ARG query")

	result, err := graphClient.Query(ctx, discoverAssessQuery, subscriptions)
	if err != nil {
		return nil, fmt.Errorf("failed to query Azure Resource Graph for VM modernization resources: %w", err)
	}

	// AVD host-pool-type is fetched via a second, separate query rather than a
	// join inside discoverAssessQuery: Azure Resource Graph disallows
	// referencing more than one non-"resources" extension table in the same
	// query, and discoverAssessQuery already reads computeresources (for
	// Uniform-mode VMSS members) alongside desktopvirtualizationresources
	// (for AVD), which ARG rejects with "UnsupportedRemoteTableScenario".
	// This signal is supplementary to workload-pattern classification, so a
	// failure here is logged and does not fail the whole scan.
	log.Debug().Msg("Executing VM modernization AVD host-pool ARG query")
	avdHostPoolByVMId := map[string]string{}
	if avdResult, avdErr := graphClient.Query(ctx, discoverAssessAVDQuery, subscriptions); avdErr != nil {
		log.Warn().Err(avdErr).Msg("Failed to query Azure Resource Graph for AVD host-pool data; AVD workload-pattern classification will be unavailable for this scan")
	} else if avdResult.Data != nil {
		avdHostPoolByVMId = buildAVDHostPoolByVMId(graph.UnmarshalRows[avdHostPoolRow](avdResult.Data, "AVD Host Pool"))
	}

	aroClustersByResourceGroup := loadAROClusters(ctx, graphClient.Query, subscriptions)
	vmssByID := loadVMSSPools(ctx, graphClient.Query, subscriptions)
	backupSubtypeByVMId, backupLoaded := loadBackupPolicySubtypes(ctx, graphClient.Query, subscriptions)

	// Build header row from ColumnMetadata (single source of truth).
	meta := s.GetMetadata()
	table := [][]string{meta.HeaderRow()}
	var availabilitySheet *plugins.ExternalPluginOutput

	if result.Data != nil {
		rows := graph.UnmarshalRows[vmModernizationRow](result.Data, "VM Modernization")
		mergeAVDHostPoolType(rows, avdHostPoolByVMId)
		mergeAROClusters(rows, aroClustersByResourceGroup)
		mergeVMSSPools(rows, vmssByID)
		mergeBackupPolicySubtypes(rows, backupSubtypeByVMId, backupLoaded)

		// Pre-collect every unique (armSkuName, armRegionName) pair across the
		// whole scan and fetch them via one batched, concurrent pass (P02-T03/
		// NFR-002), before building any row, so every row's cost columns are
		// computed from an already-populated cache rather than triggering
		// lookups interleaved with record building.
		httpClient := az.NewHttpClient(cred, az.DefaultHttpClientOptions(30*time.Second))
		priceCache := collectVMPrices(ctx, httpClient, uniqueSKURegionPairs(rows))

		var activeRows []vmModernizationRow
		for _, r := range rows {
			if params.Filters.Azqr.IsSubscriptionExcluded(r.SubscriptionID) {
				continue
			}
			subscriptionName := subscriptions[r.SubscriptionID]
			if subscriptionName == "" {
				subscriptionName = r.SubscriptionID
			}
			table = append(table, r.toRecord(subscriptionName, priceCache))
			activeRows = append(activeRows, r)
		}

		groups := groupAvailabilityTargets(activeRows, parseTargetRegions(params))
		regionData := collectRegionSKUData(ctx, cred, httpClient, regionPairs(groups))
		availabilitySheet = buildSKUAvailabilitySheet(groups, regionData, subscriptions)
	}

	log.Info().Msgf("VM modernization scan completed with %d resources", len(table)-1)

	outputs := []plugins.ExternalPluginOutput{{
		Metadata:    meta,
		SheetName:   "VM Modernization",
		Description: "Azure VM v6/v7 series modernization readiness: workload pattern, Assess-phase signals, and Plan-phase recommendation",
		Table:       table,
	}}
	if availabilitySheet != nil {
		outputs = append(outputs, *availabilitySheet)
	}
	if sheet := buildGeneralNoticesSheet(); sheet != nil {
		outputs = append(outputs, *sheet)
	}

	return outputs, nil
}

// uniqueSKURegionPairs collects every unique (armSkuName, armRegionName) pair
// across all VM rows returned by the scan's ARG query (P02-T03/NFR-002), so
// the Retail Prices API lookup can be pre-collected and batched once rather
// than discovered lazily one VM at a time. Region values are normalized the
// same way normalizeRegion already does in lifecycle_dataset.go, since ARG's
// location values and the Retail Prices API's armRegionName values follow the
// same normalized form.
func uniqueSKURegionPairs(rows []vmModernizationRow) []skuRegionPair {
	seen := make(map[skuRegionPair]bool)
	var pairs []skuRegionPair
	for _, r := range rows {
		pair := skuRegionPair{ArmSkuName: r.VMSize, ArmRegionName: normalizeRegion(r.Location)}
		if !seen[pair] {
			seen[pair] = true
			pairs = append(pairs, pair)
		}
	}
	return pairs
}

// generalNotices are account-wide, not-VM-specific facts from the field alert
// (plan decision D4/Finding 6): each is true for every VM/account in scope
// rather than depending on a particular VM's family or region, so they are
// reported once per run on their own sheet instead of being repeated on every
// per-VM row of the "VM Modernization" sheet.
var generalNotices = []string{
	"Reserved Instance (RI) protection: customers on Reserved Instances are protected from the price changes until the RI itself expires, not until the 1 February 2027 effective date.",
	"PaaS compute is not affected by the price changes (Storage services are affected; see the Lifecycle/Pricing Notes column for Storage-specific entries).",
	"The price changes apply to all agreement types (pay-as-you-go, Enterprise Agreement, and Microsoft Customer Agreement).",
}

// buildGeneralNoticesSheet returns a second, broadcast-only Excel sheet
// listing generalNotices, following the proven multi-sheet-per-Scan-call
// pattern in internal/scanners/plugins/region/selection.go (accumulating
// sheets into one returned slice rather than a second Scan-like call). Like
// that package's output.BuildSvcAvailSheets/BuildCostComparisonSheet, this
// sheet sets only SheetName, Description, and Table, leaving Metadata as its
// zero value: it has no per-VM ColumnMetadata, so no change to
// GetMetadata().ColumnMetadata or the header-consistency tests is required.
// Returns nil when the active cloud is exempt (plan decision D3), matching
// matchLifecycleNotes' suppression of per-VM lifecycle/pricing notes.
func buildGeneralNoticesSheet() *plugins.ExternalPluginOutput {
	if isExemptCloud() {
		return nil
	}

	table := [][]string{{"General Notice"}}
	for _, notice := range generalNotices {
		table = append(table, []string{notice})
	}

	return &plugins.ExternalPluginOutput{
		SheetName:   "VM Modernization - Notices",
		Description: "Account-wide VM lifecycle/pricing notices that apply regardless of VM family or region",
		Table:       table,
	}
}

// avdHostPoolRow is the shape of a single row returned by the separate AVD
// host-pool-type lookup query (see kql/discover-assess-avd.kql). It is
// merged onto vmModernizationRow.AVDHostPoolType by vmId in Scan, since Azure
// Resource Graph disallows combining the desktopvirtualizationresources and
// computeresources tables in one query.
type avdHostPoolRow struct {
	HostPoolArmID   string `json:"hostPoolArmId"`
	VMId            string `json:"vmId"`
	AVDHostPoolType string `json:"avdHostPoolType"`
}

// buildAVDHostPoolByVMId joins the query's session-host rows (vmId set) to its
// host-pool rows (avdHostPoolType set) by host pool ARM ID, returning a
// lowercase vmId -> host pool type map covering every session host.
func buildAVDHostPoolByVMId(rows []avdHostPoolRow) map[string]string {
	typeByPool := map[string]string{}
	for _, r := range rows {
		if r.AVDHostPoolType != "" {
			typeByPool[strings.ToLower(r.HostPoolArmID)] = r.AVDHostPoolType
		}
	}
	byVM := map[string]string{}
	for _, r := range rows {
		if r.VMId == "" {
			continue
		}
		if t, ok := typeByPool[strings.ToLower(r.HostPoolArmID)]; ok {
			byVM[strings.ToLower(r.VMId)] = t
		}
	}
	return byVM
}

// mergeAVDHostPoolType sets rows[i].AVDHostPoolType from avdHostPoolByVMId by
// matching rows[i].ID case-insensitively (the AVD query's own vmId values are
// already lowercased, see kql/discover-assess-avd.kql). A row with no
// matching entry keeps its zero-value (empty) AVDHostPoolType.
func mergeAVDHostPoolType(rows []vmModernizationRow, avdHostPoolByVMId map[string]string) {
	for i := range rows {
		rows[i].AVDHostPoolType = avdHostPoolByVMId[strings.ToLower(rows[i].ID)]
	}
}

type aroClusterRow struct {
	AROClusterID       string `json:"AROClusterId"`
	AROClusterName     string `json:"AROClusterName"`
	AROResourceGroupID string `json:"AROResourceGroupId"`
}

func loadAROClusters(ctx context.Context, query func(context.Context, string, map[string]string, ...graph.QueryOptions) (*graph.GraphResult, error), subscriptions map[string]string) map[string]aroClusterRow {
	clusters := make(map[string]aroClusterRow)
	log.Debug().Msg("Executing VM modernization ARO cluster ARG query")
	result, err := query(ctx, discoverAssessAROQuery, subscriptions)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to query Azure Resource Graph for ARO cluster data; ARO workload-pattern classification will be unavailable for this scan")
		return clusters
	}
	if result == nil {
		log.Warn().Msg("Empty Azure Resource Graph response for ARO cluster data; ARO workload-pattern classification will be unavailable for this scan")
		return clusters
	}

	for _, cluster := range graph.UnmarshalRows[aroClusterRow](result.Data, "ARO Cluster") {
		cluster.AROClusterID = strings.ToLower(cluster.AROClusterID)
		cluster.AROResourceGroupID = strings.ToLower(cluster.AROResourceGroupID)
		clusters[cluster.AROResourceGroupID] = cluster
	}
	return clusters
}

func mergeAROClusters(rows []vmModernizationRow, clustersByResourceGroup map[string]aroClusterRow) {
	for i := range rows {
		rows[i].IsAROManaged = false
		if rows[i].SubscriptionID == "" || rows[i].ResourceGroup == "" {
			continue
		}
		key := strings.ToLower(fmt.Sprintf("/subscriptions/%s/resourceGroups/%s", rows[i].SubscriptionID, rows[i].ResourceGroup))
		_, rows[i].IsAROManaged = clustersByResourceGroup[key]
	}
}

type backupPolicyRow struct {
	VMID                string `json:"VMId"`
	BackupPolicySubtype string `json:"BackupPolicySubtype"`
}

const (
	backupNotConfigured = "Not configured"
	backupUnknown       = "Unknown"
)

// loadBackupPolicySubtypes returns a lowercase VM ID to policy subtype map. The
// bool reports whether the query succeeded, so a failed query can be told
// apart from a VM that simply has no backup.
func loadBackupPolicySubtypes(ctx context.Context, query func(context.Context, string, map[string]string, ...graph.QueryOptions) (*graph.GraphResult, error), subscriptions map[string]string) (map[string]string, bool) {
	subtypes := make(map[string]string)
	log.Debug().Msg("Executing VM modernization backup policy ARG query")
	result, err := query(ctx, discoverAssessBackupQuery, subscriptions)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to query Azure Resource Graph for backup policy data; backup policy subtype will be Unknown for this scan")
		return subtypes, false
	}
	if result == nil {
		log.Warn().Msg("Empty Azure Resource Graph response for backup policy data; backup policy subtype will be Unknown for this scan")
		return subtypes, false
	}
	if result.Data == nil {
		return subtypes, true
	}
	for _, row := range graph.UnmarshalRows[backupPolicyRow](result.Data, "Backup Policy") {
		subtypes[strings.ToLower(row.VMID)] = row.BackupPolicySubtype
	}
	return subtypes, true
}

// mergeBackupPolicySubtypes sets each row's backup policy subtype. When the
// lookup failed every row is Unknown; otherwise a VM without a protected item
// is Not configured.
func mergeBackupPolicySubtypes(rows []vmModernizationRow, subtypeByVMId map[string]string, loaded bool) {
	for i := range rows {
		switch subtype, ok := subtypeByVMId[strings.ToLower(rows[i].ID)]; {
		case !loaded:
			rows[i].BackupPolicySubtype = backupUnknown
		case ok:
			rows[i].BackupPolicySubtype = subtype
		default:
			rows[i].BackupPolicySubtype = backupNotConfigured
		}
	}
}

type vmssPoolRow struct {
	VMSSID            string `json:"VMSSId"`
	OrchestrationMode string `json:"OrchestrationMode"`
	PoolModelVMSize   string `json:"PoolModelVMSize"`
}

func loadVMSSPools(ctx context.Context, query func(context.Context, string, map[string]string, ...graph.QueryOptions) (*graph.GraphResult, error), subscriptions map[string]string) map[string]vmssPoolRow {
	pools := make(map[string]vmssPoolRow)
	log.Debug().Msg("Executing VM modernization VMSS pool ARG query")
	result, err := query(ctx, discoverAssessVMSSQuery, subscriptions)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to query Azure Resource Graph for scale set data; pool-level columns will be unavailable for this scan")
		return pools
	}
	if result == nil || result.Data == nil {
		return pools
	}
	for _, pool := range graph.UnmarshalRows[vmssPoolRow](result.Data, "VMSS Pool") {
		pools[strings.ToLower(pool.VMSSID)] = pool
	}
	return pools
}

// mergeVMSSPools sets the pool-level fields from the scale set matching each
// row's VMSSId; rows without a match keep empty values.
func mergeVMSSPools(rows []vmModernizationRow, poolsByID map[string]vmssPoolRow) {
	for i := range rows {
		pool, ok := poolsByID[strings.ToLower(rows[i].VMSSID)]
		if rows[i].VMSSID == "" || !ok {
			continue
		}
		rows[i].OrchestrationMode = pool.OrchestrationMode
		rows[i].PoolModelVMSize = pool.PoolModelVMSize
	}
}

// vmModernizationRow is the shape of a single row returned by the Discover/Assess
// ARG query. Field names match the KQL `project` output exactly (see
// kql/discover-assess.kql).
type vmModernizationRow struct {
	Name                string  `json:"Name"`
	ID                  string  `json:"Id"`
	ResourceGroup       string  `json:"ResourceGroup"`
	SubscriptionID      string  `json:"SubscriptionId"`
	Location            string  `json:"Location"`
	VMSize              string  `json:"VMSize"`
	HyperVGeneration    string  `json:"HyperVGeneration"`
	DiskControllerType  string  `json:"DiskControllerType"`
	HasScaleSet         argBool `json:"HasScaleSet"`
	IsAKSManaged        argBool `json:"IsAKSManaged"`
	IsDatabricksManaged argBool `json:"IsDatabricksManaged"`
	AVDHostPoolType     string  `json:"AVDHostPoolType"`
	IsAROManaged        bool    `json:"-"`
	BackupPolicySubtype string  `json:"-"`
	AzureDiskEncryption argBool `json:"AzureDiskEncryption"`

	// Full raw inventory signals (Assess article's starter query), surfaced
	// as-is alongside the derived columns above.
	MembershipModel       string  `json:"MembershipModel"`
	ZonePlacement         string  `json:"ZonePlacement"`
	AvailabilityZone      string  `json:"AvailabilityZone"`
	VMSSName              string  `json:"VMSSName"`
	VMSSID                string  `json:"VMSSId"`
	AvailabilitySetName   string  `json:"AvailabilitySetName"`
	PlatformFaultDomain   string  `json:"PlatformFaultDomain"`
	VMPowerState          string  `json:"VMPowerState"`
	InstanceViewCaptured  string  `json:"InstanceViewCaptured"`
	OSType                string  `json:"OSType"`
	OSVersion             string  `json:"OSVersion"`
	OSName                string  `json:"OSName"`
	ComputerName          string  `json:"ComputerName"`
	IsMarketplaceVM       argBool `json:"IsMarketplaceVM"`
	ImagePublisher        string  `json:"ImagePublisher"`
	ImageOffer            string  `json:"ImageOffer"`
	ImageSKU              string  `json:"ImageSKU"`
	ImageVersion          string  `json:"ImageVersion"`
	OSDiskType            string  `json:"OSDiskType"`
	OSDiskSizeGB          int64   `json:"OSDiskSizeGB"`
	EphemeralOSDisk       string  `json:"EphemeralOSDisk"`
	DataDiskCount         int64   `json:"DataDiskCount"`
	HibernationEnabled    string  `json:"HibernationEnabled"`
	UltraSSDEnabled       string  `json:"UltraSSDEnabled"`
	LicenseType           string  `json:"LicenseType"`
	NICCount              int64   `json:"NICCount"`
	AcceleratedNetworking string  `json:"AcceleratedNetworking"`
	IPForwarding          string  `json:"IPForwarding"`
	SecurityType          string  `json:"SecurityType"`
	SecureBootEnabled     string  `json:"SecureBootEnabled"`
	VTpmEnabled           string  `json:"VTpmEnabled"`
	OrchestrationMode     string  `json:"OrchestrationMode"`
	PoolModelVMSize       string  `json:"PoolModelVMSize"`
}

// toRecord flattens a vmModernizationRow into a table row in the same column
// order as the plugin's ColumnMetadata, computing every derived (non-ARG-sourced)
// column: workload pattern, series generation, hard-gate flags, the Plan-phase
// execution-method recommendation, the suggested target v6/v7 SKU, and matching
// lifecycle/pricing notes.
func (r vmModernizationRow) toRecord(subscriptionName string, priceCache *vmPriceCache) []string {
	patternCode, poolMembership := classifyWorkloadPattern(bool(r.HasScaleSet), bool(r.IsAKSManaged), bool(r.IsDatabricksManaged), r.AVDHostPoolType, r.ImagePublisher, r.ImageOffer, r.IsAROManaged)
	seriesGeneration, isTargetGeneration := classifySeriesGeneration(r.VMSize)
	executionMethod := recommendExecutionMethod(patternCode, r.HyperVGeneration, isTargetGeneration)
	tempDiskDependency, sapISVCertificationGate := recommendHardGates(patternCode)

	var suggestedSKU string
	targetScore := "N/A"
	if isActionableWorkloadPattern(patternCode) {
		if isTargetGeneration {
			suggestedSKU = alreadyOnTargetGeneration
		} else if name, score, ok := suggestTargetSKU(r.VMSize); ok {
			suggestedSKU = name
			targetScore = fmt.Sprintf("%.3f", score)
		} else {
			suggestedSKU = manualReviewRequired
		}
	} else {
		suggestedSKU = "N/A (pool-managed)"
	}

	if patternCode == patternComputePool && poolMembership == "VMSS" {
		executionMethod = manualReviewRequired
		tempDiskDependency = manualReviewRequired
		sapISVCertificationGate = manualReviewRequired
	}

	currentMonthlyCost, projectedMonthlyCost, monthlyCostImpact := r.computeMonetaryImpact(priceCache)
	complexityRating, complexityReason := classifyModernizationComplexity(r.VMSize, r.HyperVGeneration, r.DiskControllerType, r.SecurityType, isTargetGeneration)
	generation2Ready := classifyGeneration2Ready(r.HyperVGeneration, r.ImageSKU)
	nvmeOSSupported := classifyRemoteNVMeOSSupport(r.ImagePublisher, r.ImageOffer, r.ImageSKU, r.OSName, r.OSVersion)
	manaOSSupported := classifyMANAOSSupport(r.ImagePublisher, r.ImageOffer, r.ImageSKU, r.OSName, r.OSVersion)
	readinessOutcome := classifyPlatformReadinessOutcome(generation2Ready, nvmeOSSupported, manaOSSupported, isTargetGeneration)
	documentedTargetSeries, documentedDifferences := lookupDocumentedReplacement(r.VMSize, isTargetGeneration)

	orchestrationMode, poolModelVMSize, poolModelSeriesGeneration := r.poolLevelColumns()

	// Keep the groups below in lockstep with GetMetadata: Identify, Decide,
	// Target, Blockers, Cost, Pool context, Raw inventory.
	return []string{
		// Identify
		subscriptionName,
		r.ResourceGroup,
		r.Name,
		r.Location,
		r.VMSize,
		seriesGeneration,
		patternCode,
		// Decide
		readinessOutcome,
		executionMethod,
		complexityRating,
		complexityReason,
		// Target
		suggestedSKU,
		targetScore,
		documentedTargetSeries,
		documentedDifferences,
		matchLifecycleNotes(r.VMSize, r.Location),
		// Blockers
		generation2Ready,
		nvmeOSSupported,
		manaOSSupported,
		r.HyperVGeneration,
		r.DiskControllerType,
		tempDiskDependency,
		sapISVCertificationGate,
		formatBool(bool(r.AzureDiskEncryption)),
		orNA(r.HibernationEnabled),
		orNA(r.BackupPolicySubtype),
		orNA(r.VMPowerState),
		orNA(r.InstanceViewCaptured),
		// Cost
		currentMonthlyCost,
		projectedMonthlyCost,
		monthlyCostImpact,
		// Pool context
		poolMembership,
		orNA(r.MembershipModel),
		orchestrationMode,
		orNA(r.VMSSName),
		poolModelVMSize,
		poolModelSeriesGeneration,
		orNA(r.AvailabilitySetName),
		// Raw inventory
		orNA(r.ZonePlacement),
		orNA(r.AvailabilityZone),
		orNA(r.PlatformFaultDomain),
		orNA(r.OSType),
		orNA(r.OSVersion),
		orNA(r.ComputerName),
		formatBool(bool(r.IsMarketplaceVM)),
		orNA(r.ImagePublisher),
		orNA(r.ImageOffer),
		orNA(r.ImageSKU),
		orNA(r.ImageVersion),
		orNA(r.OSDiskType),
		fmt.Sprintf("%d", r.OSDiskSizeGB),
		orNA(r.EphemeralOSDisk),
		fmt.Sprintf("%d", r.DataDiskCount),
		orNA(r.UltraSSDEnabled),
		orNA(r.LicenseType),
		fmt.Sprintf("%d", r.NICCount),
		orNA(r.AcceleratedNetworking),
		orNA(r.IPForwarding),
		orNA(r.SecurityType),
		orNA(r.SecureBootEnabled),
		orNA(r.VTpmEnabled),
	}
}

// poolLevelColumns renders the three pool-level columns (P02) joined onto
// this row from its parent scale set via vmssId: orchestration mode, the
// scale set's model/default VM profile size, and whether that model size is
// already on the v6/v7 target generation (reusing classifySeriesGeneration
// unchanged). Rows with HasScaleSet=false (standalone VMs, Availability Set
// members) read notPoolManaged for all three, since there is no parent scale
// set. A VMSS-backed row whose join did not resolve a model VM size (for
// example a transient ARG replication gap) reads "N/A" rather than
// notPoolManaged, since it is still pool-managed, just missing this one
// joined value.
func (r vmModernizationRow) poolLevelColumns() (orchestrationMode, poolModelVMSize, poolModelSeriesGeneration string) {
	if !bool(r.HasScaleSet) {
		return notPoolManaged, notPoolManaged, notPoolManaged
	}
	orchestrationMode = orNA(r.OrchestrationMode)
	if r.PoolModelVMSize == "" {
		return orchestrationMode, "N/A", "N/A"
	}
	label, _ := classifySeriesGeneration(r.PoolModelVMSize)
	return orchestrationMode, r.PoolModelVMSize, label
}

func formatBool(b bool) string {
	if b {
		return "Enabled"
	}
	return "Disabled"
}

// computeMonetaryImpact derives the three monetary-impact columns (current
// monthly cost, projected monthly cost after the increase, and the monthly
// impact) following the plan's confirmed fallback precedence: exempt cloud ->
// deallocated VM -> retail price lookup failure -> no applicable
// impactPercent -> full calculation. This always computes a per-instance
// figure from the row's own current SKU, independent of whether the row's
// workload pattern is actionable at the per-VM level (that gate only applies
// to the target-SKU suggestion; see isActionableWorkloadPattern), since a
// pool-managed VM (AKS node, VMSS member) still accrues real billing on its
// current SKU today.
func (r vmModernizationRow) computeMonetaryImpact(priceCache *vmPriceCache) (current, projected, monthlyImpact string) {
	if isExemptCloud() {
		return "", "", ""
	}
	if isDeallocated(r.VMPowerState) {
		const na = "N/A (VM deallocated)"
		return na, na, na
	}

	hourlyPrice, ok := priceCache.hourlyPrice(r.VMSize, normalizeRegion(r.Location), r.OSType)
	if !ok {
		return manualReviewRequired, manualReviewRequired, manualReviewRequired
	}

	currentMonthly := hourlyPrice * 730
	current = formatUSD(currentMonthly)

	percent, ok := matchImpactPercent(r.VMSize, r.Location)
	if !ok {
		return current, current, formatUSD(0)
	}

	projectedMonthly := currentMonthly * (1 + percent/100)
	monthlyDiff := projectedMonthly - currentMonthly

	return current, formatUSD(projectedMonthly), formatUSD(monthlyDiff)
}

// isDeallocated reports whether a VM's power state indicates compute is
// released and no longer billed ("deallocating" or "deallocated"),
// case-insensitively. A stopped but still allocated VM keeps accruing compute
// charges, so it is deliberately not matched.
func isDeallocated(powerState string) bool {
	return strings.Contains(strings.ToLower(powerState), "deallocat")
}

// formatUSD renders a monetary value using the plan's confirmed "$%.2f"
// formatting convention.
func formatUSD(value float64) string {
	return fmt.Sprintf("$%.2f", value)
}

// orNA normalizes an empty raw ARG string signal (for example a deallocated
// VM's missing instance-view fields) to "N/A" instead of an empty cell.
func orNA(s string) string {
	if s == "" {
		return "N/A"
	}
	return s
}

// Workload pattern codes (Discover article categories A-G). Patterns are
// layered in incrementally as reliable signals become available from the ARG
// query; see classifyWorkloadPattern for the current precedence order.
// patternSAPWorkload (F) is never emitted by classifyWorkloadPattern today —
// no SAP-detection signal exists yet — but is declared so recommendHardGates'
// pattern-F handling stays correct and testable ahead of a future detector.
const (
	patternPlatformManaged    = "C: Service-managed compute"
	patternComputePool        = "A: Compute pool"
	patternAVDPooled          = "B: AVD pooled host pool"
	patternCertifiedAppliance = "G: Certified ISV appliance"
	patternCustomerManaged    = "E: Customer-managed VM (default)"
	patternSAPWorkload        = "F: SAP workload"
)

// classifyWorkloadPattern derives the Discover-phase workload pattern using
// exactly one canonical precedence order: C (Databricks-managed) -> A
// (ARO/AKS membership) -> B (AVD pooled host pool) -> G (certified
// appliance) -> A for a generic VMSS member -> E (default). VMSS membership
// alone does not prove a stateless pool, so toRecord recommends manual
// validation for those rows. Each step's rationale:
//   - C first: the Vendor=Databricks tag is authoritative regardless of
//     whether the underlying compute happens to be VMSS-backed, so a
//     Databricks-tagged VM is never misreported as a generic compute pool.
//   - A next: pool/cluster-managed compute (including a pool that happens to
//     also carry an AVD host-pool type or an appliance image) is evaluated at
//     the pool level per the Discover article, so pool membership outranks
//     both B and G.
//   - B next: once pool/platform management is ruled out, the AVD host-pool
//     type is the authoritative signal for distinguishing a pooled session
//     host from a personal one or a plain customer-managed VM.
//   - G next: the certified-appliance vendor match is the most specific
//     remaining signal once C/A/B have not matched.
//   - E is the default for every VM none of the above signals identify.
func classifyWorkloadPattern(hasScaleSet, isAKSManaged, isDatabricksManaged bool, avdHostPoolType, imagePublisher, imageOffer string, isAROManaged bool) (patternCode, poolMembership string) {
	switch {
	case isDatabricksManaged:
		return patternPlatformManaged, "Azure Databricks"
	case isAROManaged:
		return patternComputePool, "ARO Cluster"
	case isAKSManaged:
		return patternComputePool, "AKS Node Pool"
	case strings.EqualFold(avdHostPoolType, "Pooled"):
		return patternAVDPooled, "AVD Pooled Host Pool"
	}
	if vendor, ok := matchApplianceVendor(imagePublisher, imageOffer); ok {
		return patternCertifiedAppliance, vendor
	}
	if hasScaleSet {
		return patternComputePool, "VMSS"
	}
	return patternCustomerManaged, "Standalone"
}

// isActionableWorkloadPattern reports whether a workload pattern represents a
// VM whose own SKU/Hyper-V generation a user can act on directly. Pool-,
// AVD-pool-, and platform-managed patterns (A, B, and C) are evaluated at the
// pool/cluster level instead, so toRecord and recommendExecutionMethod skip
// the per-VM SKU
// suggestion and execution-method recommendation for them.
func isActionableWorkloadPattern(patternCode string) bool {
	return patternCode == patternCustomerManaged || patternCode == patternCertifiedAppliance
}

// seriesVersionRe extracts the generation suffix _vN from an Azure VM SKU name,
// following the same `_vN` convention internal/skus/advisor.go uses for its own
// (unexported) version parsing.
var seriesVersionRe = regexp.MustCompile(`(?i)_v(\d+)`)

// classifySeriesGeneration reports the VM's series generation label and whether
// it is already on the v6/v7 target generation.
func classifySeriesGeneration(vmSize string) (label string, isTargetGeneration bool) {
	m := seriesVersionRe.FindStringSubmatch(vmSize)
	if m == nil {
		return "Unknown (no _vN suffix)", false
	}
	version := m[1]
	switch version {
	case "6", "7":
		return "v" + version, true
	default:
		return "v" + version, false
	}
}

// Recommended Execution Method values. The column holds only these four;
// the reasons behind a value live in the signal columns (Hyper-V Generation,
// Generation 2 Ready, Temp Disk Dependency, SAP/ISV Certification Gate,
// Workload Pattern and Pool Membership).
const (
	methodManagedByPool = "Managed by pool/cluster"
	methodNoAction      = "No action needed"
	methodRedeploy      = "Redeploy"
)

// recommendExecutionMethod applies the Plan-phase redeploy-vs-in-place decision
// rule. Only pattern-E (customer-managed) and pattern-G (certified appliance)
// VMs not already on the target generation get a Redeploy or Manual Review
// Required value; pool-managed VMs are evaluated at the pool level per the
// Discover article, not per VM.
//
// Today only the Hyper-V generation signal downgrades Redeploy to Manual Review
// Required automatically: azqr has no data to assert a hard block from the
// temp-disk-dependency or SAP/ISV-certification gates (see recommendHardGates),
// so a reader should consult those two columns before acting on Redeploy.
func recommendExecutionMethod(patternCode, hyperVGeneration string, isTargetGeneration bool) string {
	if !isActionableWorkloadPattern(patternCode) || patternCode == patternAVDPooled {
		return methodManagedByPool
	}
	if isTargetGeneration {
		return methodNoAction
	}
	if strings.EqualFold(hyperVGeneration, "V2") {
		return methodRedeploy
	}
	return manualReviewRequired
}

// recommendHardGates reports the Temp Disk Dependency and SAP/ISV Certification
// Gate columns for a classified workload pattern. Neither gate is a guessed
// Ready/Blocked verdict (FR-004): azqr has no Azure Resource Graph data source
// to assert either hard gate automatically, so a pattern the gate applies to
// always reads Manual Review Required, never a false "clear". The N/A values
// below are reserved strictly for patterns the gate does not apply to at all.
//
//   - Temp Disk Dependency is an Assess-phase, per-VM concern that applies to
//     patterns B, E, and G (F is never emitted by classifyWorkloadPattern
//     today, so it is included only for forward-compatibility). Pattern B (AVD
//     pooled host pool) is included even though its host pool itself is
//     managed/replaced as a pool: Microsoft Learn's Assess/Plan guidance still
//     applies to the golden image the pool is built from, matching the same
//     rationale already encoded in the deck skill's ACTIONABLE_PATTERNS
//     handling. Patterns A and C are pool- or platform-managed with no
//     golden-image rebuild concern and are evaluated at the pool level per
//     Learn's own exit criteria, so Temp Disk Dependency reads
//     "N/A (pool/platform-managed)" for them.
//   - SAP/ISV Certification Gate reads Manual Review Required for every
//     pattern except the service-managed A and C, because azqr cannot rule out
//     an SAP or ISV-certified workload on a VM it only infers about.
func recommendHardGates(patternCode string) (tempDiskDependency, sapISVCertificationGate string) {
	switch patternCode {
	case patternAVDPooled, patternCustomerManaged, patternCertifiedAppliance, patternSAPWorkload:
		tempDiskDependency = manualReviewRequired
	default:
		tempDiskDependency = "N/A (pool/platform-managed)"
	}

	switch patternCode {
	case patternComputePool, patternPlatformManaged:
		sapISVCertificationGate = "N/A (pool/platform-managed)"
	default:
		sapISVCertificationGate = manualReviewRequired
	}

	return tempDiskDependency, sapISVCertificationGate
}

// isV6V7Family reports whether a SKU family string belongs to the v6 or v7
// generation, following the `...v6Family`/`...v7Family` naming convention visible
// in known_skus.yaml.
func isV6V7Family(family string) bool {
	f := strings.ToLower(family)
	return strings.Contains(f, "v6family") || strings.Contains(f, "v7family")
}

// suggestTargetSKU looks up the current SKU and returns the highest-scoring
// v6/v7-family alternative. It calls skus.FindAlternatives with n<=0 (unbounded)
// before filtering to v6/v7 families: the function sorts and truncates across
// every family, so filtering after a pre-truncated top-N could silently miss a
// valid v6/v7 match ranked below the truncation point (critique PC-001).
func suggestTargetSKU(currentSKUName string) (name string, score float64, ok bool) {
	current, found := skus.Lookup(currentSKUName)
	if !found {
		return "", 0, false
	}
	for _, rec := range skus.FindAlternatives(current, 0) {
		if isV6V7Family(rec.SKU.Family) {
			return rec.SKU.Name, rec.CompatibilityScore, true
		}
	}
	return "", 0, false
}

// init registers the plugin automatically.
func init() {
	plugins.RegisterInternalPlugin("vm-modernization", NewScanner())
}
