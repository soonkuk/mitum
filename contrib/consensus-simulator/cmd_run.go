package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/spikeekips/mitum/isaac"
	"github.com/spikeekips/mitum/node"
)

var (
	globalConfig NodesGlobalConfig
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "run consensus simulator",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		{
			log.Debug("trying to load config", "file", args[0])
			b, err := ioutil.ReadFile(args[0])
			if err != nil {
				cmd.Println("Error:", err.Error())
				os.Exit(1)
			}

			gc, err := newNodesConfigFromBytes(b)
			if err != nil {
				cmd.Println("Error:", err.Error())
				os.Exit(1)
			}

			globalConfig = gc

			if flagExitAfter > 0 {
				globalConfig.ExitAfter = flagExitAfter
			}

			if flagNumberOfNodes > 0 {
				globalConfig.NumberOfNodes = flagNumberOfNodes
			}

			log.Debug("config loaded", "config", globalConfig)
		}

		if err := run(); err != nil {
			printError(cmd, err)
		}
	},
}

func init() {
	runCmd.Flags().DurationVar(&flagExitAfter, "exit-after", 0, "exit after; 0 forever")
	runCmd.Flags().UintVar(&flagNumberOfNodes, "number-of-nodes", 0, "number of nodes")

	rootCmd.AddCommand(runCmd)
}

func runNodes() error {
	// create homes
	var homes []node.Node
	for i := uint(0); i < globalConfig.NumberOfNodes; i++ {
		n := newHome(i)
		homes = append(homes, n)
		log.Info("home created", "home", n, "node", n.Alias())
	}

	previousBlock := newRandomBlock(globalConfig.Global.Block.StartHeight, globalConfig.Global.Block.StartRound)
	currentBlock := newRandomBlock(globalConfig.Global.Block.StartHeight+1, globalConfig.Global.Block.StartRound+1)

	var nodes []Node
	for i, home := range homes {
		homeState := isaac.NewHomeState(home.(node.Home), previousBlock)
		homeState.SetBlock(currentBlock)

		n, err := NewNode(fmt.Sprintf("n%d", i), homeState, homes)
		if err != nil {
			return err
		}

		nodes = append(nodes, n)
	}

	// connect networks
	for _, a := range nodes {
		for _, b := range nodes {
			if a.nt.Home().Equal(b.nt.Home()) {
				continue
			}

			a.nt.AddReceiver(b.nt.Home().Address(), b.nt.ReceiveFunc)
		}
	}

	// start nodes
	errChan := make(chan error, 100)
	for _, n := range nodes {
		go func(n Node) {
			err := n.Start()
			errChan <- err
		}(n)
		/*
			defer func(n Node) {
				if err := n.Stop(); err != nil {
					n.Log().Error("failed to stop Node", "error", err)
				}
			}(n)
		*/
	}

	go func() {
		for err := range errChan {
			if err != nil {
				log.Error("failed to start Node", "error", err)
				os.Exit(1)
			}
		}
	}()

	return nil
}

func run() error {
	go func() {
		if err := runNodes(); err != nil {
			log.Error("failed to start nodes", "error", err)
			os.Exit(1)
		}
	}()

	if globalConfig.ExitAfter == 0 {
		select {}
	} else {
		select {
		case <-time.After(globalConfig.ExitAfter):
			break
		}
	}

	return nil
}
