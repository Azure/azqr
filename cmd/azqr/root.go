package azqr

import (
	"fmt"
	"os"

	"github.com/cmendible/azqr/internal/scanners"
	"github.com/cmendible/azqr/internal/scanners/afd"
	"github.com/cmendible/azqr/internal/scanners/agw"
	"github.com/cmendible/azqr/internal/scanners/aks"
	"github.com/cmendible/azqr/internal/scanners/apim"
	"github.com/cmendible/azqr/internal/scanners/appcs"
	"github.com/cmendible/azqr/internal/scanners/cae"
	"github.com/cmendible/azqr/internal/scanners/ci"
	"github.com/cmendible/azqr/internal/scanners/cosmos"
	"github.com/cmendible/azqr/internal/scanners/cr"
	"github.com/cmendible/azqr/internal/scanners/evgd"
	"github.com/cmendible/azqr/internal/scanners/evh"
	"github.com/cmendible/azqr/internal/scanners/kv"
	"github.com/cmendible/azqr/internal/scanners/plan"
	"github.com/cmendible/azqr/internal/scanners/psql"
	"github.com/cmendible/azqr/internal/scanners/redis"
	"github.com/cmendible/azqr/internal/scanners/sb"
	"github.com/cmendible/azqr/internal/scanners/sigr"
	"github.com/cmendible/azqr/internal/scanners/sql"
	"github.com/cmendible/azqr/internal/scanners/st"
	"github.com/cmendible/azqr/internal/scanners/wps"
	"github.com/spf13/cobra"
)

var (
	version = "dev"
)

const (
	defaultConcurrency = 4
)

func init() {
	rootCmd.PersistentFlags().StringP("subscription-id", "s", "", "Azure Subscription Id (Required)")
	rootCmd.PersistentFlags().StringP("resource-group", "r", "", "Azure Resource Group")
	rootCmd.PersistentFlags().StringP("output-prefix", "o", "azqr_report", "Output file prefix")
	rootCmd.PersistentFlags().BoolP("mask", "m", false, "Mask the subscription id in the report")
	rootCmd.PersistentFlags().IntP("concurrency", "p", defaultConcurrency, fmt.Sprintf("Parallel processes. Default to %d. A < 0 value will use the maxmimum concurrency.", defaultConcurrency))
	// err := rootCmd.MarkFlagRequired("subscription-id")
	// if err != nil {
	// 	panic(err)
	// }
}

var rootCmd = &cobra.Command{
	Use:     "azqr",
	Short:   "Azure Quick Review (azqr) goal is to produce a high level assessment of an Azure Subscription or Resource Group",
	Long:    `Azure Quick Review (azqr) goal is to produce a high level assessment of an Azure Subscription or Resource Group`,
	Version: version,
	Run: func(cmd *cobra.Command, args []string) {
		serviceScanners := []scanners.IAzureScanner{
			&aks.AKSScanner{},
			&apim.APIManagementScanner{},
			&agw.ApplicationGatewayScanner{},
			&cae.ContainerAppsScanner{},
			&ci.ContainerInstanceScanner{},
			&cosmos.CosmosDBScanner{},
			&cr.ContainerRegistryScanner{},
			&evh.EventHubScanner{},
			&evgd.EventGridScanner{},
			&kv.KeyVaultScanner{},
			&appcs.AppConfigurationScanner{},
			&plan.AppServiceScanner{},
			&redis.RedisScanner{},
			&sb.ServiceBusScanner{},
			&sigr.SignalRScanner{},
			&wps.WebPubSubScanner{},
			&st.StorageScanner{},
			&psql.PostgreScanner{},
			&psql.PostgreFlexibleScanner{},
			&sql.SQLScanner{},
			&afd.FrontDoorScanner{},
		}

		scan(cmd, serviceScanners)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
