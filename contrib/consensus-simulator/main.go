package main

import (
	"fmt"
	"os"

	"github.com/inconshreveable/log15"
	"github.com/spf13/cobra"

	"github.com/spikeekips/mitum/account"
	"github.com/spikeekips/mitum/big"
	"github.com/spikeekips/mitum/common"
	"github.com/spikeekips/mitum/encode"
	"github.com/spikeekips/mitum/hash"
	"github.com/spikeekips/mitum/isaac"
	"github.com/spikeekips/mitum/keypair"
	"github.com/spikeekips/mitum/network"
	"github.com/spikeekips/mitum/node"
	"github.com/spikeekips/mitum/seal"
	"github.com/spikeekips/mitum/transaction"
)

var rootCmd = &cobra.Command{
	Use:   "cs",
	Short: "cs is the consensus simulator of ISAAC+",
	Args:  cobra.NoArgs,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// set logging
		handler, _ := common.LogHandler(common.LogFormatter(flagLogFormat.f), FlagLogOut)
		handler = log15.CallerFileHandler(handler)
		handler = log15.LvlFilterHandler(flagLogLevel.lvl, handler)

		logs := []log15.Logger{
			log,
			account.Log(),
			big.Log(),
			common.Log(),
			encode.Log(),
			hash.Log(),
			isaac.Log(),
			keypair.Log(),
			network.Log(),
			node.Log(),
			seal.Log(),
			transaction.Log(),
		}
		for _, l := range logs {
			common.SetLogger(l, flagLogLevel.lvl, handler)
		}

		log.Debug("parsed flags", "flags", printFlags(cmd, flagLogFormat.f))
	},
}

func main() {
	rootCmd.PersistentFlags().Var(&flagLogLevel, "log-level", "log level: {debug error warn info crit}")
	rootCmd.PersistentFlags().Var(&flagLogFormat, "log-format", "log format: {json terminal}")
	rootCmd.PersistentFlags().StringVar(&FlagLogOut, "log", FlagLogOut, "log output file")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	os.Exit(0)
}
