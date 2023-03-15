package azqr

import (
	"github.com/cmendible/azqr/internal/scanners"
	"github.com/cmendible/azqr/internal/scanners/evgd"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(evgdCmd)
}

var evgdCmd = &cobra.Command{
	Use:   "evgd",
	Short: "Scan Azure Event Grid Domains",
	Long:  "Scan Azure Event Grid Domains",
	Run: func(cmd *cobra.Command, args []string) {
		serviceScanners := []scanners.IAzureScanner{
			&evgd.EventGridScanner{},
		}

		scan(cmd, serviceScanners)
	},
}
