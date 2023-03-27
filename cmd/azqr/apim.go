package azqr

import (
	"github.com/cmendible/azqr/internal/scanners"
	"github.com/cmendible/azqr/internal/scanners/apim"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(apimCmd)
}

var apimCmd = &cobra.Command{
	Use:   "apim",
	Short: "Scan Azure API Management",
	Long:  "Scan Azure API Management",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		serviceScanners := []scanners.IAzureScanner{
			&apim.APIManagementScanner{},
		}

		scan(cmd, serviceScanners)
	},
}
