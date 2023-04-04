package azqr

import (
	"github.com/cmendible/azqr/internal/scanners"
	"github.com/cmendible/azqr/internal/scanners/agw"
	"github.com/spf13/cobra"
)

func init() {
	scanCmd.AddCommand(agwCmd)
}

var agwCmd = &cobra.Command{
	Use:   "agw",
	Short: "Scan Azure Application Gateway",
	Long:  "Scan Azure Application Gateway",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		serviceScanners := []scanners.IAzureScanner{
			&agw.ApplicationGatewayScanner{},
		}

		scan(cmd, serviceScanners)
	},
}
