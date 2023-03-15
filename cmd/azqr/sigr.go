package azqr

import (
	"github.com/cmendible/azqr/internal/scanners"
	"github.com/cmendible/azqr/internal/scanners/sigr"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(sigrCmd)
}

var sigrCmd = &cobra.Command{
	Use:   "sigr",
	Short: "Scan Azure SignalR",
	Long:  "Scan Azure SignalR",
	Run: func(cmd *cobra.Command, args []string) {
		serviceScanners := []scanners.IAzureScanner{
			&sigr.SignalRScanner{},
		}

		scan(cmd, serviceScanners)
	},
}
