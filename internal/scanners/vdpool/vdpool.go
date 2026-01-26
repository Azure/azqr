// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package vdpool

import (
	"github.com/Azure/azqr/internal/models"
)

func init() {
	models.ScannerList["vdpool"] = []models.IAzureScanner{
		models.NewBaseScanner(
			"Microsoft.DesktopVirtualization/hostPools",
			"Microsoft.DesktopVirtualization/scalingPlans",
			"Microsoft.DesktopVirtualization/workspaces",
		),
	}
}
