package azqr

import (
	"github.com/cmendible/azqr/internal/scanners"
	"github.com/cmendible/azqr/internal/scanners/afd"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(afdCmd)
}

var afdCmd = &cobra.Command{
	Use:   "afd",
	Short: "Scan Azure Front Door",
	Long:  "Scan Azure Front Door",
	Run: func(cmd *cobra.Command, args []string) {
		serviceScanners := []scanners.IAzureScanner{
			&afd.FrontDoorScanner{},
		}

		scan(cmd, serviceScanners)
	},
}
