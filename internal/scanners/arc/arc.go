// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package arc

import (
	"github.com/Azure/azqr/internal/models"
)

func init() {
	models.ScannerList["arc"] = []models.IAzureScanner{
		models.NewBaseScanner("Microsoft.AzureArcData/sqlServerInstances"),
	}
}

// This scanner doesn't perform actual scans - it's here to register the resource type
// Actual Arc SQL scanning is done by the graph-based ArcSQLScanner
