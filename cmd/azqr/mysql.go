package azqr

import (
	"github.com/cmendible/azqr/internal/scanners"
	"github.com/cmendible/azqr/internal/scanners/mysql"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(mysqlCmd)
}

var mysqlCmd = &cobra.Command{
	Use:   "mysql",
	Short: "Scan Azure Database for MySQL",
	Long:  "Scan Azure Database for MySQL",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		serviceScanners := []scanners.IAzureScanner{
			&mysql.MySQLScanner{},
			&mysql.MySQLFlexibleScanner{},
		}

		scan(cmd, serviceScanners)
	},
}
