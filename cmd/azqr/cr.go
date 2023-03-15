package azqr

import (
	"github.com/cmendible/azqr/internal/scanners"
	"github.com/cmendible/azqr/internal/scanners/cr"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(crCmd)
}

var crCmd = &cobra.Command{
	Use:   "cr",
	Short: "Scan Azure Container Registries",
	Long:  "Scan Azure Container Registries",
	Run: func(cmd *cobra.Command, args []string) {
		serviceScanners := []scanners.IAzureScanner{
			&cr.ContainerRegistryScanner{},
		}

		scan(cmd, serviceScanners)
	},
}
