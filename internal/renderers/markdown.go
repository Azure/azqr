package renderers

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/cmendible/azqr/internal/scanners"
	"github.com/cmendible/azqr/internal/report/templates"
	"github.com/fbiville/markdown-table-formatter/pkg/markdown"
	mdparser "github.com/gomarkdown/markdown"
)

func CreateMarkdownReport(all []scanners.IAzureServiceResult, customer, outputFile string, detailedScan bool) {
	resultsTable := renderTable(all)

	var allFunctions []scanners.IAzureServiceResult
	for _, r := range all {
		v, ok := r.(scanners.AzureFunctionAppResult)
		if ok {
			allFunctions = append(allFunctions, v)
		}
	}

	reportTemplate := templates.GetTemplates("Report.md")
	reportTemplate = strings.Replace(reportTemplate, "{{results}}", resultsTable, 1)
	reportTemplate = strings.Replace(reportTemplate, "{{date}}", time.Now().Format("2006-01-02"), 1)
	reportTemplate = strings.Replace(reportTemplate, "{{customer}}", customer, -1)

	recommendations := ""
	dict := map[string]bool{}
	for _, r := range all {
		parsedType := strings.Replace(r.GetResourceType(), "/", ".", -1)
		if _, ok := dict[r.GetResourceType()]; !ok {
			dict[r.GetResourceType()] = true
			recommendations += "\n\n"
			recommendations += templates.GetTemplates(fmt.Sprintf("%s.md", parsedType))

			if r.GetResourceType() == "Microsoft.Web/serverfarms/sites" && len(allFunctions) > 0 && detailedScan {
				recommendations = strings.Replace(recommendations, "{{functions}}", renderDetailsTable(allFunctions), 1)
			} else {
				recommendations = strings.Replace(recommendations, "{{functions}}", "", 1)
			}
		}
	}

	reportTemplate = strings.Replace(reportTemplate, "{{recommendations}}", recommendations, 1)

	err := os.WriteFile(fmt.Sprintf("%s.md", outputFile), []byte(reportTemplate), 0644)
	if err != nil {
		log.Fatal(err)
	}

	md := []byte(reportTemplate)
	output := mdparser.ToHTML(md, nil, nil)
	err = os.WriteFile(fmt.Sprintf("%s.html", outputFile), output, 0644)
	if err != nil {
		log.Fatal(err)
	}
}

func renderTable(results []scanners.IAzureServiceResult) string {
	if len(results) == 0 {
		return "No results found."
	}

	heathers := results[0].GetProperties()

	rows := [][]string{}
	for _, r := range results {
		rows = append(mapToRow(heathers, r.ToMap()), rows...)
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

func renderDetailsTable(results []scanners.IAzureServiceResult) string {
	heathers := results[0].GetDetailProperties()

	rows := [][]string{}
	for _, r := range results {
		rows = append(mapToRow(heathers, r.ToDetail()), rows...)
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
