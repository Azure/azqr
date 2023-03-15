package azqr

import (
	"github.com/cmendible/azqr/internal/scanners"
	"github.com/cmendible/azqr/internal/scanners/psql"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(psqlCmd)
}

var psqlCmd = &cobra.Command{
	Use:   "psql",
	Short: "Scan Azure Database for psql",
	Long:  "Scan Azure Database for psql",
	Run: func(cmd *cobra.Command, args []string) {
		serviceScanners := []scanners.IAzureScanner{
			&psql.PostgreScanner{},
			&psql.PostgreFlexibleScanner{},
		}

		scan(cmd, serviceScanners)
	},
}
