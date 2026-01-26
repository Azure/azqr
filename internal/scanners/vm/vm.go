// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package vm

import (
	"context"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v4"
)

func init() {
	models.ScannerList["vm"] = []models.IAzureScanner{NewVirtualMachineScanner()}
}

// NewVirtualMachineScanner creates a new Virtual Machine scanner using the generic framework
func NewVirtualMachineScanner() models.IAzureScanner {
	return models.NewGenericScanner(
		models.GenericScannerConfig[armcompute.VirtualMachine, *armcompute.VirtualMachinesClient]{
			ResourceTypes: []string{"Microsoft.Compute/virtualMachines"},

			ClientFactory: func(config *models.ScannerConfig) (*armcompute.VirtualMachinesClient, error) {
				return armcompute.NewVirtualMachinesClient(
					config.SubscriptionID,
					config.Cred,
					config.ClientOptions,
				)
			},

			ListResources: func(client *armcompute.VirtualMachinesClient, ctx context.Context) ([]*armcompute.VirtualMachine, error) {
				pager := client.NewListAllPager(nil)
				vms := make([]*armcompute.VirtualMachine, 0)

				for pager.More() {
					resp, err := pager.NextPage(ctx)
					if err != nil {
						return nil, err
					}
					vms = append(vms, resp.Value...)
				}

				return vms, nil
			},

			GetRecommendations: getRecommendations,

			ExtractResourceInfo: func(vm *armcompute.VirtualMachine) models.ResourceInfo {
				return models.ExtractStandardARMResourceInfo(
					vm.ID,
					vm.Name,
					vm.Location,
					vm.Type,
				)
			},
		},
	)
}
