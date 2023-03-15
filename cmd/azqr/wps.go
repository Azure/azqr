package azqr

import (
	"github.com/cmendible/azqr/internal/scanners"
	"github.com/cmendible/azqr/internal/scanners/wps"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(wpsCmd)
}

var wpsCmd = &cobra.Command{
	Use:   "wps",
	Short: "Scan Azure Web PubSub",
	Long:  "Scan Azure Web PubSub",
	Run: func(cmd *cobra.Command, args []string) {
		serviceScanners := []scanners.IAzureScanner{
			&wps.WebPubSubScanner{},
		}

		scan(cmd, serviceScanners)
	},
}
