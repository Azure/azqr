// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package json

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/Azure/azqr/internal/renderers"
	"github.com/rs/zerolog/log"
)

func CreateJsonReport(data *renderers.ReportData) {
	results := []interface{}{}

	resources := renderers.ResourceResults{
		Resource: getResources(data),
	}
	results = append(results, resources)

	types := renderers.ResourceTypeCountResults{
		ResourceType: data.ResourceTypeCount,
	}
	results = append(results, types)

	writeData(results, data.OutputFileName, "json")
}

func writeData(data []interface{}, fileName, extension string) {
	filename := fmt.Sprintf("%s.%s", fileName, extension)
	log.Info().Msgf("Generating Report: %s", filename)

	f, err := os.Create(filename)
	if err != nil {
		log.Fatal().Err(err).Msg("error creating json:")
	}
	defer f.Close()

	js, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		log.Fatal().Err(err).Msg("error marshaling data:")
	}

	_, err = f.Write(js)
	if err != nil {
		log.Fatal().Err(err).Msg("error writing json:")
	}
}

func getResources(data *renderers.ReportData) []renderers.ResourceResult {
	rows := []renderers.ResourceResult{}

	for _, r := range data.AprlData {
		row := renderers.ResourceResult{
			ValidationAction: "Azure Resource Graph",
			RecommendationId: r.RecommendationID,
			Name:             r.Name,
			Id:               r.ResourceID,
			Param1:           r.Param1,
			Param2:           r.Param2,
			Param3:           r.Param3,
			Param4:           r.Param4,
			Param5:           r.Param5,
			CheckName:        "",
			Selector:         r.Source,
		}
		rows = append(rows, row)
	}

	for _, d := range data.AzqrData {
		for _, r := range d.Recommendations {
			if r.NotCompliant {
				row := renderers.ResourceResult{
					ValidationAction: "Azure Resource Manager",
					RecommendationId: r.RecommendationID,
					Name:             d.ServiceName,
					Id:               d.ResourceID(),
					Param1:           r.Result,
					Param2:           "",
					Param3:           "",
					Param4:           "",
					Param5:           "",
					CheckName:        "",
					Selector:         "AZQR",
				}
				rows = append(rows, row)
			}
		}
	}
	return rows
}
