package main

import (
	"context"
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
	proposalValidator := isaac.NewTestProposalValidator(policy, time.Second*2)

	alias := homeState.Home().Alias()

	go func() {
		for message := range nt.Reader() {
			voteCompiler.Write(message)
		}
	}()

	st := isaac.NewStateTransition(homeState, voteCompiler)

	{
		booting := isaac.NewBootingStateHandler(homeState, st.ChanState())
		sync := isaac.NewSyncStateHandler(homeState, suffrage, policy, client, st.ChanState())
		join := isaac.NewJoinStateHandler(homeState, policy, client, st.ChanState())
		con := isaac.NewConsensusStateHandler(homeState, suffrage, policy, client, proposalValidator, st.ChanState())
		stopped := isaac.NewStoppedStateHandler(homeState)

		stateHandlers := []isaac.StateHandler{booting, sync, join, con, stopped}

		for _, handler := range stateHandlers {
			handler.SetLogContext(nil, "node", alias)
			if err := st.SetStateHandler(handler); err != nil {
				return Node{}, err
			}
		}
	}

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
	no.st.ChanState() <- common.SetContext(context.TODO(), "state", node.StateBooting)

	return nil
}

func (no Node) Stop() error {
	no.Log().Debug("stop Node")
	_ = no.homeState.SetState(node.StateStopped)
	return no.st.Stop()
}
