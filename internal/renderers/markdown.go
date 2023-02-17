package renderers

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/cmendible/azqr/internal/report/templates"
	"github.com/cmendible/azqr/internal/scanners"
	"github.com/fbiville/markdown-table-formatter/pkg/markdown"
	mdparser "github.com/gomarkdown/markdown"
)

func CreateMarkdownReport(data ReportData) {
	resultsTable := renderTable(data.MainData, data.Mask)

	var allFunctions []scanners.IAzureServiceResult
	for _, r := range data.MainData {
		v, ok := r.(scanners.AzureFunctionAppResult)
		if ok {
			allFunctions = append(allFunctions, v)
		}
	}

	reportTemplate := templates.GetTemplates("Report.md")
	reportTemplate = strings.Replace(reportTemplate, "{{results}}", resultsTable, 1)
	reportTemplate = strings.Replace(reportTemplate, "{{date}}", time.Now().Format("2006-01-02"), 1)

	bestPractices := ""
	dict := map[string]bool{}
	for _, r := range data.MainData {
		parsedType := strings.Replace(r.GetResourceType(), "/", ".", -1)
		if _, ok := dict[r.GetResourceType()]; !ok {
			dict[r.GetResourceType()] = true
			bestPractices += "\n\n"
			bestPractices += templates.GetTemplates(fmt.Sprintf("%s.md", parsedType))

			if r.GetResourceType() == "Microsoft.Web/serverfarms/sites" && len(allFunctions) > 0 && data.EnableDetailedScan {
				bestPractices = strings.Replace(bestPractices, "{{functions}}", renderDetailsTable(allFunctions, data.Mask), 1)
			} else {
				bestPractices = strings.Replace(bestPractices, "{{functions}}", "", 1)
			}
		}
	}

	if len(data.DefenderData) > 0 {
		bestPractices += "\n\n"
		bestPractices += templates.GetTemplates("Microsoft.Security.pricings.md")
		bestPractices = strings.Replace(bestPractices, "{{defender}}", renderDefenderTable(data.DefenderData, data.Mask), 1)
	}

	reportTemplate = strings.Replace(reportTemplate, "{{best_practices}}", bestPractices, 1)

	err := os.WriteFile(fmt.Sprintf("%s.md", data.OutputFileName), []byte(reportTemplate), 0644)
	if err != nil {
		log.Fatal(err)
	}

	md := []byte(reportTemplate)
	output := mdparser.ToHTML(md, nil, nil)
	err = os.WriteFile(fmt.Sprintf("%s.html", data.OutputFileName), output, 0644)
	if err != nil {
		log.Fatal(err)
	}
}

func renderTable(results []scanners.IAzureServiceResult, mask bool) string {
	if len(results) == 0 {
		return "No results found."
	}

	heathers := results[0].GetHeathers()

	rows := [][]string{}
	for _, r := range results {
		rows = append(mapToRow(heathers, r.ToMap(mask)), rows...)
	}

	prettyPrintedTable, err := markdown.NewTableFormatterBuilder().
		WithPrettyPrint().
		Build(heathers...).
		Format(rows)

	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("")
	fmt.Println(prettyPrintedTable)
	return prettyPrintedTable
}

func renderDetailsTable(results []scanners.IAzureServiceResult, mask bool) string {
	heathers := results[0].GetDetailHeathers()

	rows := [][]string{}
	for _, r := range results {
		rows = append(mapToRow(heathers, r.ToDetailMap(mask)), rows...)
	}

	prettyPrintedTable, err := markdown.NewTableFormatterBuilder().
		WithPrettyPrint().
		Build(heathers...).
		Format(rows)

	if err != nil {
		log.Fatal(err)
	}

	return prettyPrintedTable
}

func renderDefenderTable(results []scanners.DefenderResult, mask bool) string {
	if len(results) == 0 {
		return "No results found."
	}

	heathers := results[0].GetProperties()

	rows := [][]string{}
	for _, r := range results {
		rows = append(mapToRow(heathers, r.ToMap(mask)), rows...)
	}

	prettyPrintedTable, err := markdown.NewTableFormatterBuilder().
		WithPrettyPrint().
		Build(heathers...).
		Format(rows)

	if err != nil {
		log.Fatal(err)
	}

	return prettyPrintedTable
}

func mapToRow(heathers []string, m map[string]string) [][]string {
	v := make([]string, 0, len(m))

	for _, k := range heathers {
		v = append(v, m[k])
	}

	return [][]string{v}
}
