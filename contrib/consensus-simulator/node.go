package main

import (
	"time"

	"github.com/spikeekips/mitum/common"
	"github.com/spikeekips/mitum/isaac"
	"github.com/spikeekips/mitum/network"
	"github.com/spikeekips/mitum/node"
)

type Node struct {
	*common.Logger
	homeState    *isaac.HomeState
	voteCompiler *isaac.VoteCompiler
	nt           *network.NodesTest
	client       isaac.ClientTest
	st           *isaac.StateTransition
}

func NewNode(homeState *isaac.HomeState, policy isaac.Policy, suffrage isaac.Suffrage) (Node, error) {
	nt := network.NewNodesTest(homeState.Home())
	client := isaac.NewClientTest(nt)
	ballotbox := isaac.NewBallotbox(policy.Threshold)
	voteCompiler := isaac.NewVoteCompiler(homeState, suffrage, ballotbox)

	go func() {
		for message := range nt.Reader() {
			voteCompiler.Write(message)
		}
	}()

	st := isaac.NewStateTransition(homeState, voteCompiler)

	stateHandlers := []isaac.StateHandler{
		isaac.NewBootingStateHandler(homeState, st.ChanState()),
		isaac.NewSyncStateHandler(homeState, suffrage, policy, client, st.ChanState()),
		isaac.NewJoinStateHandler(homeState, policy, client, st.ChanState()),
		isaac.NewConsensusStateHandler(homeState, suffrage, policy, client, st.ChanState()),
	}

	for _, handler := range stateHandlers {
		if err := st.SetStateHandler(handler); err != nil {
			return Node{}, err
		}
	}

	alias := homeState.Home().Alias()
	n := Node{
		Logger:       common.NewLogger(log, "node", alias),
		homeState:    homeState,
		voteCompiler: voteCompiler,
		nt:           nt,
		client:       client,
		st:           st,
	}

	nt.SetLogContext(nil, "node", alias)
	client.SetLogContext(nil, "node", alias)
	ballotbox.SetLogContext(nil, "node", alias)
	voteCompiler.SetLogContext(nil, "node", alias)
	st.SetLogContext(nil, "node", alias)

	n.Log().Debug(
		"node created",
		"homeState", homeState,
		"policy", policy,
		"suffrage", suffrage,
	)

	return n, nil
}

func (no Node) Start() error {
	no.Log().Debug("start Node")
	if err := no.voteCompiler.Start(); err != nil {
		return err
	}
	if err := no.nt.Start(); err != nil {
		return err
	}

	if err := no.st.Start(); err != nil {
		return err
	}
	<-time.After(time.Millisecond * 50)
	no.st.ChanState() <- node.StateBooting

	return nil
}

func (no Node) Stop() error {
	no.Log().Debug("stop Node")
	_ = no.homeState.SetState(node.StateStopped)
	return no.st.Stop()
}
