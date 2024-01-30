// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package renderers

import (
	"fmt"
	"strconv"

	"github.com/Azure/azqr/internal/scanners"
)

type ReportData struct {
	OutputFileName     string
	Mask               bool
	MainData           []scanners.AzureServiceResult
	DefenderData       []scanners.DefenderResult
	AdvisorData        []scanners.AdvisorResult
	CostData           *scanners.CostResult
}

func (rd *ReportData) ServicesTable() [][]string {
	headers := []string{"Subscription", "Resource Group", "Location", "Type", "Service Name", "Broken", "Category", "Subcategory", "Severity", "Description", "Result", "Learn"}

	rbroken := [][]string{}
	rok := [][]string{}
	for _, d := range rd.MainData {
		for _, r := range d.Rules {
			row := []string{
				scanners.MaskSubscriptionID(d.SubscriptionID, rd.Mask),
				d.ResourceGroup,
				scanners.ParseLocation(d.Location),
				d.Type,
				d.ServiceName,
				fmt.Sprintf("%t", r.IsBroken),
				string(r.Category),
				string(r.Subcategory),
				string(r.Severity),
				r.Description,
				r.Result,
				r.Learn,
			}
			if r.IsBroken {
				rbroken = append([][]string{row}, rbroken...)
			} else {
				rok = append([][]string{row}, rok...)
			}
		}
	}

	rows := append(rbroken, rok...)
	rows = append([][]string{headers}, rows...)
	return rows
}

func (rd *ReportData) CostTable() [][]string {
	headers := []string{"Subscription", "ServiceName", "Value", "Currency"}

	rows := [][]string{}
	for _, r := range rd.CostData.Items {
		row := []string{
			scanners.MaskSubscriptionID(r.SubscriptionID, rd.Mask),
			r.ServiceName,
			r.Value,
			r.Currency,
		}
		rows = append(rows, row)
	}

	rows = append([][]string{headers}, rows...)
	return rows
}

func (rd *ReportData) DefenderTable() [][]string {
	headers := []string{"Subscription", "Name", "Tier", "Deprecated"}
	rows := [][]string{}
	for _, d := range rd.DefenderData {
		row := []string{
			scanners.MaskSubscriptionID(d.SubscriptionID, rd.Mask),
			d.Name,
			d.Tier,
			fmt.Sprintf("%t", d.Deprecated),
		}
		rows = append(rows, row)
	}

	rows = append([][]string{headers}, rows...)
	return rows
}

func (rd *ReportData) AdvisorTable() [][]string {
	headers := []string{"Subscription", "Name", "Type", "Category", "Description", "PotentialBenefits", "Risk", "LearnMoreLink"}
	rows := [][]string{}
	for _, d := range rd.AdvisorData {
		row := []string{
			scanners.MaskSubscriptionID(d.SubscriptionID, rd.Mask),
			d.Name,
			d.Type,
			d.Category,
			d.Description,
			d.PotentialBenefits,
			d.Risk,
			d.LearnMoreLink,
		}
		rows = append(rows, row)
	}

	rows = append([][]string{headers}, rows...)
	return rows
}

func (rd *ReportData) OverviewTable() [][]string {
	headers := []string{"Subscription", "ResourceGroup", "Location", "Type", "Name", "SKU", "SLA", "AZ", "PVT", "DS", "CAF"}
	rows := [][]string{}
	for _, d := range rd.MainData {
		sku := ""
		sla := ""
		az := ""
		pvt := ""
		ds := ""
		caf := ""
		for _, v := range d.Rules {
			switch v.Subcategory {
			case scanners.RulesSubcategoryReliabilitySKU:
				sku = v.Result
			case scanners.RulesSubcategoryReliabilitySLA:
				sla = v.Result
			case scanners.RulesSubcategoryReliabilityAvailabilityZones:
				az = strconv.FormatBool(!v.IsBroken)
			case scanners.RulesSubcategorySecurityPrivateEndpoint:
				pvt = strconv.FormatBool(!v.IsBroken)
			case scanners.RulesSubcategoryReliabilityDiagnosticLogs:
				ds = strconv.FormatBool(!v.IsBroken)
			case scanners.RulesSubcategoryOperationalExcellenceCAF:
				caf = strconv.FormatBool(!v.IsBroken)
			}
		}
		row := []string{
			scanners.MaskSubscriptionID(d.SubscriptionID, rd.Mask),
			d.ResourceGroup,
			scanners.ParseLocation(d.Location),
			d.Type,
			d.ServiceName,
			sku,
			sla,
			az,
			pvt,
			ds,
			caf,
		}
		rows = append(rows, row)
	}

	rows = append([][]string{headers}, rows...)
	return rows
}
