// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package renderers

import (
	"fmt"

	"github.com/Azure/azqr/internal/scanners"
)

type ReportData struct {
	OutputFileName string
	Mask           bool
	MainData       []scanners.AzureServiceResult
	DefenderData   []scanners.DefenderResult
	AdvisorData    []scanners.AdvisorResult
	CostData       *scanners.CostResult
}

func (rd *ReportData) ServicesTable() [][]string {
	headers := []string{"Subscription", "Resource Group", "Location", "Type", "Service Name", "Compliant", "Impact", "Category", "Recommendation", "Result", "Learn"}

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
				fmt.Sprintf("%t", !r.NotCompliant),
				string(r.Impact),
				string(r.Category),
				r.Recommendation,
				r.Result,
				r.Learn,
			}
			if r.NotCompliant {
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
	headers := []string{"From", "To", "Subscription", "ServiceName", "Value", "Currency"}

	rows := [][]string{}
	for _, r := range rd.CostData.Items {
		row := []string{
			rd.CostData.From.Format("2006-01-02"),
			rd.CostData.To.Format("2006-01-02"),
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
