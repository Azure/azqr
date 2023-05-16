// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package azqr

import (
	"fmt"
	"sort"

	"github.com/cmendible/azqr/internal/scanners"
	"github.com/cmendible/azqr/internal/scanners/afd"
	"github.com/cmendible/azqr/internal/scanners/afw"
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
	"github.com/cmendible/azqr/internal/scanners/mysql"
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

func init() {
	rootCmd.AddCommand(rulesCmd)
}

var rulesCmd = &cobra.Command{
	Use:   "rules",
	Short: "Print all azqr rules",
	Long:  "Print all azqr rules as markdown table",
	Args:  cobra.NoArgs,
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
			&afw.FirewallScanner{},
			&mysql.MySQLScanner{},
			&mysql.MySQLFlexibleScanner{},
		}

		fmt.Println("#  | Id | Category | Subcategory | Name | Severity | More Info")
		fmt.Println("---|---|---|---|---|---|---")

		i := 0
		for _, scanner := range serviceScanners {
			rulesMap := scanner.GetRules()

			rules := map[string]scanners.AzureRule{}
			for _, r := range rulesMap {
				rules[r.Id] = r
			}

			keys := make([]string, 0, len(rules))
			for k := range rules {
				keys = append(keys, k)
			}
			sort.Strings(keys)

			for _, k := range keys {
				rule := rules[k]
				i++
				fmt.Printf("%s | %s | %s | %s | %s | %s | %s", fmt.Sprint(i), rule.Id, rule.Category, rule.Subcategory, rule.Description, rule.Severity, rule.Url)
				fmt.Println()
			}
		}
	},
}
