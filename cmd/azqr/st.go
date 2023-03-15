package azqr

import (
	"github.com/cmendible/azqr/internal/scanners"
	"github.com/cmendible/azqr/internal/scanners/st"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(stCmd)
}

var stCmd = &cobra.Command{
	Use:   "st",
	Short: "Scan Azure Storage",
	Long:  "Scan Azure Storage",
	Run: func(cmd *cobra.Command, args []string) {
		serviceScanners := []scanners.IAzureScanner{
			&st.StorageScanner{},
		}

		scan(cmd, serviceScanners)
	},
}
