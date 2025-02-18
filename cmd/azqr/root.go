// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package azqr

import (
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	_ "github.com/Azure/azqr/internal/scanners/aa"
	_ "github.com/Azure/azqr/internal/scanners/adf"
	_ "github.com/Azure/azqr/internal/scanners/afd"
	_ "github.com/Azure/azqr/internal/scanners/afw"
	_ "github.com/Azure/azqr/internal/scanners/agw"
	_ "github.com/Azure/azqr/internal/scanners/aks"
	_ "github.com/Azure/azqr/internal/scanners/amg"
	_ "github.com/Azure/azqr/internal/scanners/apim"
	_ "github.com/Azure/azqr/internal/scanners/appcs"
	_ "github.com/Azure/azqr/internal/scanners/appi"
	_ "github.com/Azure/azqr/internal/scanners/as"
	_ "github.com/Azure/azqr/internal/scanners/asp"
	_ "github.com/Azure/azqr/internal/scanners/avd"
	_ "github.com/Azure/azqr/internal/scanners/avs"
	_ "github.com/Azure/azqr/internal/scanners/ba"
	_ "github.com/Azure/azqr/internal/scanners/ca"
	_ "github.com/Azure/azqr/internal/scanners/cae"
	_ "github.com/Azure/azqr/internal/scanners/ci"
	_ "github.com/Azure/azqr/internal/scanners/cog"
	_ "github.com/Azure/azqr/internal/scanners/conn"
	_ "github.com/Azure/azqr/internal/scanners/cosmos"
	_ "github.com/Azure/azqr/internal/scanners/cr"
	_ "github.com/Azure/azqr/internal/scanners/dbw"
	_ "github.com/Azure/azqr/internal/scanners/dec"
	_ "github.com/Azure/azqr/internal/scanners/disk"
	_ "github.com/Azure/azqr/internal/scanners/erc"
	_ "github.com/Azure/azqr/internal/scanners/evgd"
	_ "github.com/Azure/azqr/internal/scanners/evh"
	_ "github.com/Azure/azqr/internal/scanners/fdfp"
	_ "github.com/Azure/azqr/internal/scanners/gal"
	_ "github.com/Azure/azqr/internal/scanners/hpc"
	_ "github.com/Azure/azqr/internal/scanners/iot"
	_ "github.com/Azure/azqr/internal/scanners/it"
	_ "github.com/Azure/azqr/internal/scanners/kv"
	_ "github.com/Azure/azqr/internal/scanners/lb"
	_ "github.com/Azure/azqr/internal/scanners/log"
	_ "github.com/Azure/azqr/internal/scanners/logic"
	_ "github.com/Azure/azqr/internal/scanners/maria"
	_ "github.com/Azure/azqr/internal/scanners/mysql"
	_ "github.com/Azure/azqr/internal/scanners/netapp"
	_ "github.com/Azure/azqr/internal/scanners/ng"
	_ "github.com/Azure/azqr/internal/scanners/nsg"
	_ "github.com/Azure/azqr/internal/scanners/nw"
	_ "github.com/Azure/azqr/internal/scanners/pdnsz"
	_ "github.com/Azure/azqr/internal/scanners/pep"
	_ "github.com/Azure/azqr/internal/scanners/pip"
	_ "github.com/Azure/azqr/internal/scanners/psql"
	_ "github.com/Azure/azqr/internal/scanners/redis"
	_ "github.com/Azure/azqr/internal/scanners/rsv"
	_ "github.com/Azure/azqr/internal/scanners/rt"
	_ "github.com/Azure/azqr/internal/scanners/sap"
	_ "github.com/Azure/azqr/internal/scanners/sb"
	_ "github.com/Azure/azqr/internal/scanners/sigr"
	_ "github.com/Azure/azqr/internal/scanners/sql"
	_ "github.com/Azure/azqr/internal/scanners/st"
	_ "github.com/Azure/azqr/internal/scanners/synw"
	_ "github.com/Azure/azqr/internal/scanners/traf"
	_ "github.com/Azure/azqr/internal/scanners/vdpool"
	_ "github.com/Azure/azqr/internal/scanners/vgw"
	_ "github.com/Azure/azqr/internal/scanners/vm"
	_ "github.com/Azure/azqr/internal/scanners/vmss"
	_ "github.com/Azure/azqr/internal/scanners/vnet"
	_ "github.com/Azure/azqr/internal/scanners/vwan"
	_ "github.com/Azure/azqr/internal/scanners/wps"
	_ "github.com/Azure/azqr/internal/scanners/rg"
	_ "github.com/Azure/azqr/internal/scanners/avail"
	_ "github.com/Azure/azqr/internal/scanners/nic"
)

var (
	version = "dev"
)

var rootCmd = &cobra.Command{
	Use:     "azqr",
	Short:   "Azure Quick Review (azqr) goal is to produce a high level assessment of an Azure Subscription or Resource Group",
	Long:    `Azure Quick Review (azqr) goal is to produce a high level assessment of an Azure Subscription or Resource Group`,
	Args:    cobra.NoArgs,
	Version: version,
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Usage()
	},
}

func Execute() {
	output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}

	log.Logger = zerolog.New(output).With().Timestamp().Logger()

	cobra.CheckErr(rootCmd.Execute())
}
