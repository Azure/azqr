// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package vmmodernization

import (
	"regexp"
	"strings"
)

// Tri-state verdicts shared by the platform-readiness classifiers.
const (
	verdictTrue    = "TRUE"
	verdictFalse   = "FALSE"
	verdictUnknown = "UNKNOWN"
)

// classifyGeneration2Ready reports whether a VM is Generation 2 ready. The
// Hyper-V generation from the instance view is authoritative; when it is
// missing, an image SKU containing "gen2" or "g2" is used as a hint.
func classifyGeneration2Ready(hyperVGeneration, imageSKU string) string {
	switch strings.ToLower(strings.TrimSpace(hyperVGeneration)) {
	case "v2", "generation 2":
		return verdictTrue
	case "v1", "generation 1":
		return verdictFalse
	}
	sku := strings.ToLower(imageSKU)
	if strings.Contains(sku, "gen2") || strings.Contains(sku, "g2") {
		return verdictTrue
	}
	return verdictUnknown
}

const (
	complexityHigh   = "HIGH"
	complexityMedium = "MEDIUM"
	complexityLow    = "LOW"
	complexityReview = "REVIEW"

	complexityNotApplicable = "N/A (already on target generation)"
)

var sourceV4V5Re = regexp.MustCompile(`(?i)_v[45]$`)

// classifyModernizationComplexity rates the effort to move a VM to a v6/v7
// series from its Hyper-V generation, disk controller, current series and
// security type. A VM that fits none of the HIGH, MEDIUM or LOW rules is
// rated REVIEW. It returns the rating and a one-line reason.
func classifyModernizationComplexity(vmSize, hyperVGeneration, diskController, securityType string, isTargetGeneration bool) (rating, reason string) {
	if isTargetGeneration {
		return complexityNotApplicable, complexityNotApplicable
	}
	generation := strings.ToLower(strings.TrimSpace(hyperVGeneration))
	controller := strings.TrimSpace(diskController)
	security := strings.TrimSpace(securityType)
	if security == "" {
		security = "Standard"
	}
	switch generation {
	case "v1", "generation 1":
		return complexityHigh, "Generation 1 VM; a Gen2 migration is required before moving to v6/v7"
	case "v2", "generation 2":
		switch {
		case strings.EqualFold(controller, "SCSI"):
			return complexityMedium, "Generation 2 VM on a SCSI disk controller; the move to NVMe needs OS and disk preparation"
		case strings.EqualFold(controller, "NVMe") && sourceV4V5Re.MatchString(vmSize) && strings.EqualFold(security, "Standard"):
			return complexityLow, "Generation 2 VM already on NVMe, v4/v5 series and Standard security type"
		}
		return complexityReview, "Generation 2 VM that does not match a defined rule (check disk controller, series and security type)"
	}
	return complexityReview, "Hyper-V generation is unknown; confirm it before planning the move"
}

// remoteNVMeOSRules lists the OS images documented as supporting remote NVMe
// disks, matched against the lowercased image publisher, offer, SKU and guest
// OS name/version. Version patterns avoid digits adjacent to other digits so
// that, for example, "x86_64" is not read as version 8.
var remoteNVMeOSRules = []*regexp.Regexp{
	// Azure Linux 3.0
	regexp.MustCompile(`(azure.?linux|cbl.?mariner)[^0-9]{0,12}3([._ -]0)?([^0-9]|$)`),
	// AlmaLinux 8.x and 9.x
	regexp.MustCompile(`alma.{0,40}(^|[^0-9.])(8|9)([^0-9]|$)`),
	// Debian 11 and 12
	regexp.MustCompile(`debian.{0,40}(^|[^0-9.])(11|12)([^0-9]|$)`),
	// Oracle Linux 7.9, 8.5+, 9.x
	regexp.MustCompile(`oracle.?linux.{0,40}((^|[^0-9.])7[._ -]?9([^0-9]|$)|(^|[^0-9.])8[._ -]?([5-9]|1[0-9])([^0-9]|$)|(^|[^0-9.])8([- ]|$)|(^|[^0-9.])9([._ -]?[0-9]{1,2})?([^0-9]|$))`),
	regexp.MustCompile(`(^|[^a-z])ol((79|8[5-9]|8[1-9][0-9]|9[0-9]*))`),
	// RHEL 7.9, 8.6+, 9.x
	regexp.MustCompile(`(rhel|red.?hat).{0,40}((^|[^0-9.])7[._ -]?9([^0-9]|$)|(^|[^0-9.])8[._ -]?([6-9]|1[0-9])([^0-9]|$)|(^|[^0-9.])8([- ]|$)|(^|[^0-9.])9([._ -]?[0-9]{1,2})?([^0-9]|$))`),
	// SLES 15 SP4+ and SLES for SAP 15 SP3+
	regexp.MustCompile(`(sles|suse)(.{0,40}sap)?.{0,40}15.{0,10}sp[._ -]?([4-9]|1[0-9])`),
	regexp.MustCompile(`(sles|suse).{0,10}sap.{0,40}15.{0,10}sp[._ -]?([3-9]|1[0-9])`),
	// Ubuntu 18.04, 20.04, 22.04, 24.04
	regexp.MustCompile(`ubuntu.{0,60}((^|[^0-9])(18|20|22|24)[._ -]?04([^0-9]|$)|bionic|focal|jammy|noble)`),
	// Rocky Linux 8.10 and 9.6
	regexp.MustCompile(`rocky.{0,40}((^|[^0-9])8[._ -]10([^0-9]|$)|(^|[^0-9])9[._ -]6([^0-9]|$))`),
	// Windows Server 2019, 2022, 2025
	regexp.MustCompile(`windows.?server.{0,30}(^|[^0-9])(2019|2022|2025)([^0-9]|$)`),
	// Windows 10 and 11
	regexp.MustCompile(`(windows[._ -]?(10|11)|(^|[^a-z])win(10|11))([^0-9]|$)`),
}

// classifyRemoteNVMeOSSupport reports whether the VM's image or guest OS is
// documented as supporting remote NVMe disks. It returns UNKNOWN when no OS
// metadata is available and FALSE when metadata exists but matches no
// supported OS.
func classifyRemoteNVMeOSSupport(publisher, offer, sku, osName, osVersion string) string {
	return classifyOSSupport(remoteNVMeOSRules, publisher, offer, sku, osName, osVersion)
}

const (
	outcomePlanGeneration2   = "Gen2 migration needed"
	outcomeImageRefresh      = "OS image refresh needed"
	outcomeNeedsValidation   = "OS support unconfirmed"
	outcomeChecksPassed      = "Platform checks passed"
	outcomeNotApplicableText = "N/A (already on target generation)"
)

// classifyPlatformReadinessOutcome combines the Generation 2, remote NVMe OS
// and MANA OS verdicts into one outcome. Precedence: a non-Gen2 VM needs a
// Generation 2 path first; an unsupported NVMe or MANA OS needs an image
// refresh; any remaining unknown needs validation.
func classifyPlatformReadinessOutcome(generation2Ready, nvmeOSSupported, manaOSSupported string, isTargetGeneration bool) string {
	switch {
	case isTargetGeneration:
		return outcomeNotApplicableText
	case generation2Ready == verdictFalse:
		return outcomePlanGeneration2
	case nvmeOSSupported == verdictFalse || manaOSSupported == verdictFalse:
		return outcomeImageRefresh
	case generation2Ready == verdictUnknown || nvmeOSSupported == verdictUnknown || manaOSSupported == verdictUnknown:
		return outcomeNeedsValidation
	}
	return outcomeChecksPassed
}

// manaOSRules lists the OS images documented as supporting the Microsoft Azure
// Network Adapter (MANA), matched the same way as remoteNVMeOSRules.
var manaOSRules = []*regexp.Regexp{
	// Windows Server 2016, 2019, 2022, 2025 and Windows 11
	regexp.MustCompile(`windows.?server.{0,30}(^|[^0-9])(2016|2019|2022|2025)([^0-9]|$)`),
	regexp.MustCompile(`(windows[._ -]?11|(^|[^a-z])win11)([^0-9]|$)`),
	// Azure Linux 3 and 4
	regexp.MustCompile(`(azure.?linux|cbl.?mariner)[^0-9]{0,12}(3|4)([._ -]0)?([^0-9]|$)`),
	// AlmaLinux, Rocky Linux and RHEL 9.6 and 10.0
	regexp.MustCompile(`(alma|rocky|rhel|red.?hat).{0,40}((^|[^0-9.])9[._ -]?6([^0-9]|$)|(^|[^0-9.])10[._ -]?0([^0-9]|$))`),
	// Debian 12 and 13
	regexp.MustCompile(`debian.{0,40}(^|[^0-9.])(12|13)([^0-9]|$)`),
	// Oracle Linux 9.x, 10.x and UEK R7/R8
	regexp.MustCompile(`oracle.?linux.{0,40}((^|[^0-9.])(9|10)([._ -]?[0-9]{1,2})?([^0-9]|$)|uek.{0,5}(7|8)([^0-9]|$))`),
	regexp.MustCompile(`(^|[^a-z])ol(9|10)[0-9]*([^0-9]|$)`),
	// SLES 15 SP5-SP7 and SLES 16
	regexp.MustCompile(`(sles|suse).{0,40}(15.{0,10}sp[._ -]?[5-7]([^0-9]|$)|(^|[^0-9.])16([^0-9]|$))`),
	// Ubuntu 22.04, 24.04, 26.04
	regexp.MustCompile(`ubuntu.{0,60}(^|[^0-9])(22|24|26)[._ -]?04([^0-9]|$)`),
}

// classifyMANAOSSupport reports whether the VM's image or guest OS is
// documented as supporting MANA, with the same UNKNOWN/FALSE semantics as
// classifyRemoteNVMeOSSupport.
func classifyMANAOSSupport(publisher, offer, sku, osName, osVersion string) string {
	return classifyOSSupport(manaOSRules, publisher, offer, sku, osName, osVersion)
}

// classifyOSSupport matches the lowercased image and guest OS text against
// rules. It returns UNKNOWN when no OS metadata is available and FALSE when
// metadata exists but matches no rule.
func classifyOSSupport(rules []*regexp.Regexp, publisher, offer, sku, osName, osVersion string) string {
	if strings.TrimSpace(publisher+offer+sku+osName+osVersion) == "" {
		return verdictUnknown
	}
	text := strings.ToLower(strings.Join([]string{publisher, offer, sku, osName, osVersion}, " "))
	for _, rule := range rules {
		if rule.MatchString(text) {
			return verdictTrue
		}
	}
	return verdictFalse
}
