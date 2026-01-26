// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package it

import (
	"context"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/virtualmachineimagebuilder/armvirtualmachineimagebuilder/v2"
)

func init() {
	models.ScannerList["it"] = []models.IAzureScanner{NewImageTemplateScanner()}
}

// NewImageTemplateScanner creates a new Image Template Scanner
func NewImageTemplateScanner() models.IAzureScanner {
	return models.NewGenericScanner(
		models.GenericScannerConfig[armvirtualmachineimagebuilder.ImageTemplate, *armvirtualmachineimagebuilder.VirtualMachineImageTemplatesClient]{
			ResourceTypes: []string{"Microsoft.VirtualMachineImages/imageTemplates"},

			ClientFactory: func(config *models.ScannerConfig) (*armvirtualmachineimagebuilder.VirtualMachineImageTemplatesClient, error) {
				return armvirtualmachineimagebuilder.NewVirtualMachineImageTemplatesClient(config.SubscriptionID, config.Cred, config.ClientOptions)
			},

			ListResources: func(client *armvirtualmachineimagebuilder.VirtualMachineImageTemplatesClient, ctx context.Context) ([]*armvirtualmachineimagebuilder.ImageTemplate, error) {
				pager := client.NewListPager(nil)
				svcs := make([]*armvirtualmachineimagebuilder.ImageTemplate, 0)
				for pager.More() {
					resp, err := pager.NextPage(ctx)
					if err != nil {
						return nil, err
					}
					svcs = append(svcs, resp.Value...)
				}
				return svcs, nil
			},

			GetRecommendations: getRecommendations,

			ExtractResourceInfo: func(resource *armvirtualmachineimagebuilder.ImageTemplate) models.ResourceInfo {
				return models.ExtractStandardARMResourceInfo(
					resource.ID,
					resource.Name,
					resource.Location,
					resource.Type,
				)
			},
		},
	)
}
