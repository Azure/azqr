package azqr

import (
	"github.com/cmendible/azqr/internal/scanners"
	"github.com/cmendible/azqr/internal/scanners/aks"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(aksCmd)
}

var aksCmd = &cobra.Command{
	Use:   "aks",
	Short: "Scan Azure Kubernetes Service",
	Long:  "Scan Azure Kubernetes Service",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		serviceScanners := []scanners.IAzureScanner{
			&aks.AKSScanner{},
		}

		scan(cmd, serviceScanners)
	},
}
