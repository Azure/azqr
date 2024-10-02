// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package pbi

import (
	"os"
	"path/filepath"

	"github.com/Azure/azqr/internal/embeded"
	"github.com/rs/zerolog/log"
)

func CreatePBIReport(path string) {
	if path == "" {
		log.Fatal().Msg("Please specify the path were the PowerBI template will be created using --template-path option")
	}

	log.Info().Msgf("Generating Power BI dashboard template: %sazqr.pbit", path)

	pbit := embeded.GetTemplates("azqr.pbit")
	err := os.WriteFile(filepath.Join(path, "azqr.pbit"), []byte(pbit), 0644)
	if err != nil {
		panic(err)
	}
}
