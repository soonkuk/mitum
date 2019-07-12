package main

import (
	"context"
	"fmt"
	"time"

	"github.com/spikeekips/mitum/common"
	"github.com/spikeekips/mitum/contrib/consensus-simulator/modules"
	"github.com/spikeekips/mitum/isaac"
	"github.com/spikeekips/mitum/network"
	"github.com/spikeekips/mitum/node"
	"golang.org/x/xerrors"
)

type Node struct {
	*common.Logger
	homeState    *isaac.HomeState
	voteCompiler *isaac.VoteCompiler
	nt           *network.NodesTest
	client       isaac.ClientTest
	st           *isaac.StateTransition
}

func NewNode(name string, homeState *isaac.HomeState, homes []node.Node) (Node, error) {
	config := globalConfig.Node(name)

	alias := homeState.Home().Alias()

	policy, err := newPolicy(config.Policy)
	if err != nil {
		return Node{}, err
	}

	nt := network.NewNodesTest(homeState.Home())
	client := isaac.NewClientTest(nt)
	client.SetLogContext(nil, "node", alias)

	ballotbox := isaac.NewBallotbox(policy.Threshold)

	suffrage, err := newSuffrage(homes, config.Modules.Suffrage)
	if err != nil {
		return Node{}, err
	}

	log.Debug("suffrage created", "suffrage", suffrage)

	voteCompiler := isaac.NewVoteCompiler(homeState, suffrage, ballotbox)

	proposalValidator, err := newProposalValidator(config.Modules.ProposalValidator)
	if err != nil {
		return Node{}, err
	}

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
	proposalValidator.(common.Loggerable).SetLogContext(nil, "node", alias)

	n.Log().Info(
		"node created",
		"homeState", homeState,
		"policy", policy,
		"suffrage", suffrage,
		"config", config,
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

func newProposalValidator(c map[string]interface{}) (isaac.ProposalValidator, error) {
	switch name := c["name"].(string); name {
	case "DurationProposalValidator":
		d, ok := c["duration"]
		if !ok {
			return nil, xerrors.Errorf("duration should be set")
		}

		duration, ok := d.(time.Duration)
		if !ok {
			return nil, xerrors.Errorf("invalid duration; duration=%q", d)
		}

		log.Debug("DurationProposalValidator is loaded", "c", c)
		return modules.NewDurationProposalValidator(duration), nil
	case "WrongBlockProposalValidator":
		d, ok := c["duration"]
		if !ok {
			return nil, xerrors.Errorf("duration should be set")
		}

		duration, ok := d.(time.Duration)
		if !ok {
			return nil, xerrors.Errorf("invalid duration; duration=%q", d)
		}

		hs, ok := c["heights"]
		if !ok {
			return nil, xerrors.Errorf("heights should be set")
		}

		heights, ok := hs.([]isaac.Height)
		if !ok {
			return nil, xerrors.Errorf("invalid heights; heights=%q", hs)
		}

		log.Debug("WrongBlockProposalValidator is loaded", "c", c)
		return modules.NewWrongBlockProposalValidator(heights, duration), nil
	default:
		return nil, xerrors.Errorf("unknown ProposalValidator; name=%q", name)
	}
}

func newSuffrage(nodes []node.Node, c map[string]interface{}) (isaac.Suffrage, error) {
	switch name := c["name"].(string); name {
	case "FixedProposerSuffrage":
		var index int
		_, _ = fmt.Sscanf(c["proposer"].(string), "n%d", &index)

		var proposer node.Node
		for i, n := range nodes {
			if i == index {
				proposer = n
				break
			}
		}

		log.Debug("FixedProposerSuffrage is loaded", "c", c)
		return modules.NewFixedProposerSuffrage(proposer, nodes), nil
	case "RandomSuffrage":
		log.Debug("RandomSuffrage is loaded", "c", c)
		return modules.NewRandomSuffrage(nodes), nil
	default:
		return nil, xerrors.Errorf("unknown Suffrage; name=%q", name)
	}
}
