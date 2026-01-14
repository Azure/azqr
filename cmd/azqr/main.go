// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package main

import (
	"github.com/Azure/azqr/cmd/azqr/commands"

	// Import internal plugins to register them
	_ "github.com/Azure/azqr/internal/scanners/plugins/carbon"
	_ "github.com/Azure/azqr/internal/scanners/plugins/openai"
	_ "github.com/Azure/azqr/internal/scanners/plugins/servicehealth"
	_ "github.com/Azure/azqr/internal/scanners/plugins/zone"
)

func main() {
	commands.Execute()
}
