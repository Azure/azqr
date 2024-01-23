// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package azqr

import (
	"github.com/Azure/azqr/internal"
	"github.com/Azure/azqr/internal/scanners"

	"github.com/spf13/cobra"
)

func init() {
	scanCmd.PersistentFlags().StringP("subscription-id", "s", "", "Azure Subscription Id")
	scanCmd.PersistentFlags().StringP("resource-group", "g", "", "Azure Resource Group (Use with --subscription-id)")
	scanCmd.PersistentFlags().BoolP("defender", "d", true, "Scan Defender Status")
	scanCmd.PersistentFlags().BoolP("advisor", "a", true, "Scan Azure Advisor Recommendations")
	scanCmd.PersistentFlags().BoolP("costs", "c", false, "Scan Azure Costs")
	scanCmd.PersistentFlags().StringP("output-name", "o", "", "Output file name")
	scanCmd.PersistentFlags().BoolP("mask", "m", true, "Mask the subscription id in the report")
	scanCmd.PersistentFlags().BoolP("azure-cli-credential", "f", false, "Force the use of Azure CLI Credential")
	scanCmd.PersistentFlags().BoolP("debug", "", false, "Set log level to debug")

	rootCmd.AddCommand(scanCmd)
}

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan Azure Resources",
	Long:  "Scan Azure Resources",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		serviceScanners := internal.GetScanners()
		scan(cmd, serviceScanners)
	},
}

func scan(cmd *cobra.Command, serviceScanners []scanners.IAzureScanner) {
	subscriptionID, _ := cmd.Flags().GetString("subscription-id")
	resourceGroupName, _ := cmd.Flags().GetString("resource-group")
	outputFileName, _ := cmd.Flags().GetString("output-name")
	defender, _ := cmd.Flags().GetBool("defender")
	advisor, _ := cmd.Flags().GetBool("advisor")
	cost, _ := cmd.Flags().GetBool("costs")
	mask, _ := cmd.Flags().GetBool("mask")
	debug, _ := cmd.Flags().GetBool("debug")
	forceAzureCliCredential, _ := cmd.Flags().GetBool("azure-credential-type")

	params := internal.ScanParams{
		SubscriptionID:          subscriptionID,
		ResourceGroup:           resourceGroupName,
		OutputName:              outputFileName,
		Defender:                defender,
		Advisor:                 advisor,
		Cost:                    cost,
		Mask:                    mask,
		Debug:                   debug,
		ServiceScanners:         serviceScanners,
		ForceAzureCliCredential: forceAzureCliCredential,
	}

	internal.Scan(&params)
}
