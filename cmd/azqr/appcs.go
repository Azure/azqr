package azqr

import (
	"github.com/cmendible/azqr/internal/scanners"
	"github.com/cmendible/azqr/internal/scanners/appcs"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(appcsCmd)
}

var appcsCmd = &cobra.Command{
	Use:   "appcs",
	Short: "Scan Azure App Configuration",
	Long:  "Scan Azure App Configuration",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		serviceScanners := []scanners.IAzureScanner{
			&appcs.AppConfigurationScanner{},
		}

		scan(cmd, serviceScanners)
	},
}
