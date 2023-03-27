package azqr

import (
	"github.com/cmendible/azqr/internal/scanners"
	"github.com/cmendible/azqr/internal/scanners/redis"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(redisCmd)
}

var redisCmd = &cobra.Command{
	Use:   "redis",
	Short: "Scan Azure Cache for Redis",
	Long:  "Scan Azure Cache for Redis",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		serviceScanners := []scanners.IAzureScanner{
			&redis.RedisScanner{},
		}

		scan(cmd, serviceScanners)
	},
}
