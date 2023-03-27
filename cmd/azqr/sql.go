package azqr

import (
	"github.com/cmendible/azqr/internal/scanners"
	"github.com/cmendible/azqr/internal/scanners/sql"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(sqlCmd)
}

var sqlCmd = &cobra.Command{
	Use:   "sql",
	Short: "Scan Azure SQL Database",
	Long:  "Scan Azure SQL Database",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		serviceScanners := []scanners.IAzureScanner{
			&sql.SQLScanner{},
		}

		scan(cmd, serviceScanners)
	},
}
