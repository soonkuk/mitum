package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/spf13/cobra"

	"github.com/spikeekips/mitum/hash"
	"github.com/spikeekips/mitum/isaac"
	"github.com/spikeekips/mitum/keypair"
	"github.com/spikeekips/mitum/node"
)

var nodesCmd = &cobra.Command{
	Use:   "nodes",
	Short: "run multiple nodes",
	Args:  cobra.MinimumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		if err := run(); err != nil {
			printError(cmd, err)
		}
	},
}

func init() {
	nodesCmd.Flags().UintVar(&FlagNumberOfNodes, "nodes", FlagNumberOfNodes, "number of nodes")

	rootCmd.AddCommand(nodesCmd)
}

func run() error {
	// create homes
	var homes []node.Node
	for i := uint(0); i < FlagNumberOfNodes; i++ {
		//n := node.NewRandomHome()
		n := newHome(i)
		homes = append(homes, n)
		log.Debug("home created", "home", n, "node", n.Alias())
	}

	a := float64(FlagNumberOfNodes) * 0.62
	fmt.Println(a, int64(a), a > float64(int64(a)))

	policy := isaac.NewTestPolicy()
	policy.TimeoutINITBallot = time.Second * 3
	policy.IntervalINITBallotOfJoin = time.Second * 3
	policy.BasePercent = 66

	threshold, err := isaac.NewThreshold(FlagNumberOfNodes, policy.BasePercent)
	if err != nil {
		return err
	}
	policy.Threshold = threshold

	log.Debug("policy created", "policy", policy)

	suffrage := isaac.NewSuffrageTest(
		homes,
		func(height isaac.Height, round isaac.Round, nodes []node.Node) []node.Node {
			return homes
		},
	)
	log.Debug("suffrage created", "suffrage", suffrage)

	previousBlock := isaac.NewRandomBlock()
	currentBlock := isaac.NewRandomNextBlock(previousBlock)

	var nodes []Node
	for _, home := range homes {
		homeState := isaac.NewHomeState(home.(node.Home), previousBlock)
		homeState.SetBlock(currentBlock)

		n, err := NewNode(homeState, policy, suffrage)
		if err != nil {
			return err
		}

		nodes = append(nodes, n)
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
		for {
			select {
			case err := <-errChan:
				if err != nil {
					log.Error("failed to start Node", "error", err)
				}
				wg.Done()
			}
		}
	}()

	wg.Wait()

	select {}

	return nil
}

func newHome(n uint) node.Home {
	pk, _ := keypair.NewStellarPrivateKey()

	prefix := []byte(fmt.Sprintf("%d", n))

	var b [32]byte
	copy(b[:], prefix)

	h, _ := hash.NewHash(node.AddressHashHint, b[:])
	address := node.Address{Hash: h}

	return node.NewHome(address, pk)
}
