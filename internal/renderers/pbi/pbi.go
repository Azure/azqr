// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package pbi

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/Azure/azqr/internal/embeded"
	"github.com/rs/zerolog/log"
)

func CreatePBIReport(path string) {
	if runtime.GOOS != "windows" {
		log.Info().Msg("Skipping PowerBI report generation. Since it's only supported on Windows")
		return
	}

	if path == "" {
		log.Fatal().Msg("Please specify the path were the PowerBI template will be created using --template-path option")
	}

	log.Info().Msgf("Generating Power BI dashboard template: %s.pbit", path)

	pbit := embeded.GetTemplates("azqr.pbit")
	err := os.WriteFile(filepath.Join(path, "azqr.pbit"), []byte(pbit), 0644)
	if err != nil {
		panic(err)
	}
}
