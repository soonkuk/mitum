package main

import (
	"fmt"
	"os"
	"os/signal"
	"runtime/pprof"
	"syscall"

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

var sigc chan os.Signal

var rootCmd = &cobra.Command{
	Use:   "cs",
	Short: "cs is the consensus simulator of ISAAC+",
	Args:  cobra.NoArgs,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// set logging
		var handler log15.Handler
		if len(FlagLogOut) > 0 {
			handler = LogFileByNodeHandler(FlagLogOut, common.LogFormatter(flagLogFormat.f), flagQuiet)
		} else {
			handler, _ = common.LogHandler(common.LogFormatter(flagLogFormat.f), FlagLogOut)
		}
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

		if len(flagCPUProfile) > 0 {
			f, err := os.Create(flagCPUProfile)
			if err != nil {
				panic(err)
			}
			pprof.StartCPUProfile(f)

			sigc := make(chan os.Signal, 1)
			signal.Notify(
				sigc,
				syscall.SIGTERM,
				syscall.SIGQUIT,
				syscall.SIGKILL,
			)

			go func() {
				_ = <-sigc
				pprof.StopCPUProfile()
				log.Debug("cpuprofile closed")
				os.Exit(0)
			}()
			log.Debug("cpuprofile enabled")
		}
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		if len(flagCPUProfile) > 0 {
			pprof.StopCPUProfile()
			log.Debug("cpuprofile closed")
		}
		log.Debug("stopped")
	},
}

func main() {
	rootCmd.PersistentFlags().Var(&flagLogLevel, "log-level", "log level: {debug error warn info crit}")
	rootCmd.PersistentFlags().Var(&flagLogFormat, "log-format", "log format: {json terminal}")
	rootCmd.PersistentFlags().StringVar(&FlagLogOut, "log", FlagLogOut, "log output directory")
	rootCmd.PersistentFlags().StringVar(&flagCPUProfile, "cpuprofile", flagCPUProfile, "write cpu profile to file")
	rootCmd.PersistentFlags().BoolVar(&flagQuiet, "quiet", flagQuiet, "quiet")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	os.Exit(0)
}
