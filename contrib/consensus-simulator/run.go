package main

import (
	"io/ioutil"
	"os"
	"sync"
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

func run() error {
	// create homes
	var homes []node.Node
	for i := uint(0); i < globalConfig.NumberOfNodes; i++ {
		n := newHome(i)
		homes = append(homes, n)
		log.Info("home created", "home", n, "node", n.Alias())
	}

	previousBlock := newRandomBlock(globalConfig.Global.Block.StartHeight-1, 0)
	currentBlock := newRandomBlock(globalConfig.Global.Block.StartHeight, globalConfig.Global.Block.StartRound)

	var nodes []Node
	for _, home := range homes {
		homeState := isaac.NewHomeState(home.(node.Home), previousBlock)
		homeState.SetBlock(currentBlock)

		n, err := NewNode(homeState, homes)
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
		defer func(n Node) {
			if err := n.Stop(); err != nil {
				n.Log().Error("failed to stop Node", "error", err)
			}
		}(n)
	}

	var wg sync.WaitGroup
	wg.Add(len(nodes))
	go func() {
		for err := range errChan {
			if err != nil {
				log.Error("failed to start Node", "error", err)
			}
			wg.Done()
		}
	}()

	wg.Wait()

	if globalConfig.ExitAfter == 0 {
		select {}
	} else {
		select {
		case <-time.After(globalConfig.ExitAfter):
			break
		}
	}
	defer log.Debug("exited")

	return nil
}
