// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package scanners

import (
	"sort"

	"github.com/Azure/azqr/internal/azqr"
	"github.com/Azure/azqr/internal/scanners/aa"
	"github.com/Azure/azqr/internal/scanners/adf"
	"github.com/Azure/azqr/internal/scanners/afd"
	"github.com/Azure/azqr/internal/scanners/afw"
	"github.com/Azure/azqr/internal/scanners/agw"
	"github.com/Azure/azqr/internal/scanners/aks"
	"github.com/Azure/azqr/internal/scanners/amg"
	"github.com/Azure/azqr/internal/scanners/apim"
	"github.com/Azure/azqr/internal/scanners/appcs"
	"github.com/Azure/azqr/internal/scanners/appi"
	"github.com/Azure/azqr/internal/scanners/as"
	"github.com/Azure/azqr/internal/scanners/asp"
	"github.com/Azure/azqr/internal/scanners/avd"
	"github.com/Azure/azqr/internal/scanners/avs"
	"github.com/Azure/azqr/internal/scanners/ba"
	"github.com/Azure/azqr/internal/scanners/ca"
	"github.com/Azure/azqr/internal/scanners/cae"
	"github.com/Azure/azqr/internal/scanners/ci"
	"github.com/Azure/azqr/internal/scanners/cog"
	"github.com/Azure/azqr/internal/scanners/conn"
	"github.com/Azure/azqr/internal/scanners/cosmos"
	"github.com/Azure/azqr/internal/scanners/cr"
	"github.com/Azure/azqr/internal/scanners/dbw"
	"github.com/Azure/azqr/internal/scanners/dec"
	"github.com/Azure/azqr/internal/scanners/disk"
	"github.com/Azure/azqr/internal/scanners/erc"
	"github.com/Azure/azqr/internal/scanners/evgd"
	"github.com/Azure/azqr/internal/scanners/evh"
	"github.com/Azure/azqr/internal/scanners/fdfp"
	"github.com/Azure/azqr/internal/scanners/gal"
	"github.com/Azure/azqr/internal/scanners/hpc"
	"github.com/Azure/azqr/internal/scanners/iot"
	"github.com/Azure/azqr/internal/scanners/it"
	"github.com/Azure/azqr/internal/scanners/kv"
	"github.com/Azure/azqr/internal/scanners/lb"
	"github.com/Azure/azqr/internal/scanners/log"
	"github.com/Azure/azqr/internal/scanners/logic"
	"github.com/Azure/azqr/internal/scanners/maria"
	"github.com/Azure/azqr/internal/scanners/mysql"
	"github.com/Azure/azqr/internal/scanners/netapp"
	"github.com/Azure/azqr/internal/scanners/ng"
	"github.com/Azure/azqr/internal/scanners/nsg"
	"github.com/Azure/azqr/internal/scanners/nw"
	"github.com/Azure/azqr/internal/scanners/pdnsz"
	"github.com/Azure/azqr/internal/scanners/pep"
	"github.com/Azure/azqr/internal/scanners/pip"
	"github.com/Azure/azqr/internal/scanners/psql"
	"github.com/Azure/azqr/internal/scanners/redis"
	"github.com/Azure/azqr/internal/scanners/rsv"
	"github.com/Azure/azqr/internal/scanners/rt"
	"github.com/Azure/azqr/internal/scanners/sap"
	"github.com/Azure/azqr/internal/scanners/sb"
	"github.com/Azure/azqr/internal/scanners/sigr"
	"github.com/Azure/azqr/internal/scanners/sql"
	"github.com/Azure/azqr/internal/scanners/st"
	"github.com/Azure/azqr/internal/scanners/synw"
	"github.com/Azure/azqr/internal/scanners/traf"
	"github.com/Azure/azqr/internal/scanners/vdpool"
	"github.com/Azure/azqr/internal/scanners/vgw"
	"github.com/Azure/azqr/internal/scanners/vm"
	"github.com/Azure/azqr/internal/scanners/vmss"
	"github.com/Azure/azqr/internal/scanners/vnet"
	"github.com/Azure/azqr/internal/scanners/wps"
)

// ScannerList is a map of service abbreviation to scanner
var ScannerList = map[string][]azqr.IAzureScanner{
	"aa":     {&aa.AutomationAccountScanner{}},
	"adf":    {&adf.DataFactoryScanner{}},
	"afd":    {&afd.FrontDoorScanner{}},
	"afw":    {&afw.FirewallScanner{}},
	"agw":    {&agw.ApplicationGatewayScanner{}},
	"aks":    {&aks.AKSScanner{}},
	"amg":    {&amg.ManagedGrafanaScanner{}},
	"apim":   {&apim.APIManagementScanner{}},
	"appcs":  {&appcs.AppConfigurationScanner{}},
	"appi":   {&appi.AppInsightsScanner{}},
	"as":     {&as.AnalysisServicesScanner{}},
	"asp":    {&asp.AppServiceScanner{}},
	"avd":    {&avd.AzureVirtualDesktopScanner{}},
	"avs":    {&avs.AVSScanner{}},
	"ba":     {&ba.BatchAccountScanner{}},
	"ca":     {&ca.ContainerAppsScanner{}},
	"cae":    {&cae.ContainerAppsEnvironmentScanner{}},
	"ci":     {&ci.ContainerInstanceScanner{}},
	"cog":    {&cog.CognitiveScanner{}},
	"con":    {&conn.ConnectionScanner{}},
	"cosmos": {&cosmos.CosmosDBScanner{}},
	"cr":     {&cr.ContainerRegistryScanner{}},
	"dbw":    {&dbw.DatabricksScanner{}},
	"dec":    {&dec.DataExplorerScanner{}},
	"disk":   {&disk.DiskScanner{}},
	"erc":    {&erc.ExpressRouteScanner{}},
	"evgd":   {&evgd.EventGridScanner{}},
	"evh":    {&evh.EventHubScanner{}},
	"fdfp":   {&fdfp.FrontDoorWAFPolicyScanner{}},
	"gal":    {&gal.GalleryScanner{}},
	"hpc":    {&hpc.HighPerformanceComputingScanner{}},
	"iot":    {&iot.IoTHubScanner{}},
	"it":     {&it.ImageTemplateScanner{}},
	"kv":     {&kv.KeyVaultScanner{}},
	"lb":     {&lb.LoadBalancerScanner{}},
	"log":    {&log.LogAnalyticsScanner{}},
	"logic":  {&logic.LogicAppScanner{}},
	"maria":  {&maria.MariaScanner{}},
	"mysql":  {&mysql.MySQLFlexibleScanner{}, &mysql.MySQLScanner{}},
	"netapp": {&netapp.NetAppScanner{}},
	"ng":     {&ng.NatGatewayScanner{}},
	"nsg":    {&nsg.NSGScanner{}},
	"nw":     {&nw.NetworkWatcherScanner{}},
	"pdnsz":  {&pdnsz.PrivateDNSZoneScanner{}},
	"pep":    {&pep.PrivateEndpointScanner{}},
	"pip":    {&pip.PublicIPScanner{}},
	"psql":   {&psql.PostgreFlexibleScanner{}, &psql.PostgreScanner{}},
	"redis":  {&redis.RedisScanner{}},
	"rsv":    {&rsv.RecoveryServiceScanner{}},
	"rt":     {&rt.RouteTableScanner{}},
	"sap":    {&sap.SAPScanner{}},
	"sb":     {&sb.ServiceBusScanner{}},
	"sigr":   {&sigr.SignalRScanner{}},
	"sql":    {&sql.SQLScanner{}},
	"st":     {&st.StorageScanner{}},
	"synw":   {&synw.SynapseWorkspaceScanner{}},
	"traf":   {&traf.TrafficManagerScanner{}},
	"vdpool": {&vdpool.VirtualDesktopScanner{}},
	"vgw":    {&vgw.VirtualNetworkGatewayScanner{}},
	"vm":     {&vm.VirtualMachineScanner{}},
	"vmss":   {&vmss.VirtualMachineScaleSetScanner{}},
	"vnet":   {&vnet.VirtualNetworkScanner{}},
	"wps":    {&wps.WebPubSubScanner{}},
}

// GetScanners returns a list of all scanners in ScannerList
func GetScanners() []azqr.IAzureScanner {
	var scanners []azqr.IAzureScanner
	keys := make([]string, 0, len(ScannerList))
	for key := range ScannerList {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		scanners = append(scanners, ScannerList[key]...)
	}
	return scanners
}

// GetScannerByKeys returns a list of scanners for the given keys
func GetScannerByKeys(keys []string) []azqr.IAzureScanner {
	var scanners []azqr.IAzureScanner
	for _, key := range keys {
		if scannerList, exists := ScannerList[key]; exists {
			scanners = append(scanners, scannerList...)
		}
	}
	return scanners
}
