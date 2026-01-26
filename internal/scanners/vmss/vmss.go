// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package vmss

import (
	"context"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v4"
)

func init() {
	models.ScannerList["vmss"] = []models.IAzureScanner{NewVirtualMachineScaleSetScanner()}
}

// NewVirtualMachineScaleSetScanner creates a new VMSS scanner using the generic framework
func NewVirtualMachineScaleSetScanner() models.IAzureScanner {
	return models.NewGenericScanner(
		models.GenericScannerConfig[armcompute.VirtualMachineScaleSet, *armcompute.VirtualMachineScaleSetsClient]{
			ResourceTypes: []string{"Microsoft.Compute/virtualMachineScaleSets"},

			ClientFactory: func(config *models.ScannerConfig) (*armcompute.VirtualMachineScaleSetsClient, error) {
				return armcompute.NewVirtualMachineScaleSetsClient(
					config.SubscriptionID,
					config.Cred,
					config.ClientOptions,
				)
			},

			ListResources: func(client *armcompute.VirtualMachineScaleSetsClient, ctx context.Context) ([]*armcompute.VirtualMachineScaleSet, error) {
				pager := client.NewListAllPager(nil)
				vmss := make([]*armcompute.VirtualMachineScaleSet, 0)

				for pager.More() {
					resp, err := pager.NextPage(ctx)
					if err != nil {
						return nil, err
					}
					vmss = append(vmss, resp.Value...)
				}

				return vmss, nil
			},

			GetRecommendations: getRecommendations,

			ExtractResourceInfo: func(vmss *armcompute.VirtualMachineScaleSet) models.ResourceInfo {
				return models.ExtractStandardARMResourceInfo(
					vmss.ID,
					vmss.Name,
					vmss.Location,
					vmss.Type,
				)
			},
		},
	)
}
