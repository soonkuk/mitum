package main

import (
	"os"
	"syscall"
	"time"

	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "run contest",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if cmd.Flags().Changed("number-of-nodes") {
			if flagNumberOfNodes < 1 {
				cmd.Println("Error: `--number-of-nodes` should be greater than zero")
				os.Exit(1)
			}
		}

		log.Info("contest started")
		defer func() {
			log.Info("contest stopped")
		}()

		config, err := LoadConfig(args[0])
		if err != nil {
			cmd.Println("Error:", err.Error())
			os.Exit(1)
		}

		if flagNumberOfNodes > 0 {
			config.NumberOfNodes_ = &flagNumberOfNodes
		}

		log.Debug("config loaded", "config", config)

		go func() { // exit-after
			if flagExitAfter < time.Nanosecond {
				return
			}

			<-time.After(flagExitAfter)
			sigc <- syscall.SIGINT // interrupt process by force after timeout
		}()

		if err := run(cmd, config); err != nil {
			printError(cmd, err)
			os.Exit(1)
		}
	},
}

func init() {
	runCmd.Flags().DurationVar(&flagExitAfter, "exit-after", 0, "exit after; 0 forever")
	runCmd.Flags().UintVar(&flagNumberOfNodes, "number-of-nodes", 0, "number of nodes")

	rootCmd.AddCommand(runCmd)
}
