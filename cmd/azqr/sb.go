package azqr

import (
	"github.com/cmendible/azqr/internal/scanners"
	"github.com/cmendible/azqr/internal/scanners/sb"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(sbCmd)
}

var sbCmd = &cobra.Command{
	Use:   "sb",
	Short: "Scan Azure Service Bus",
	Long:  "Scan Azure Service Bus",
	Run: func(cmd *cobra.Command, args []string) {
		serviceScanners := []scanners.IAzureScanner{
			&sb.ServiceBusScanner{},
		}

		scan(cmd, serviceScanners)
	},
}
