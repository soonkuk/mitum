package main

import (
	"time"

	"golang.org/x/xerrors"

	"github.com/inconshreveable/log15"
	contest_module "github.com/spikeekips/mitum/contrib/contest/module"
	"github.com/spikeekips/mitum/isaac"
	"github.com/spikeekips/mitum/keypair"
	"github.com/spikeekips/mitum/node"
	"github.com/spikeekips/mitum/seal"
)

type Node struct {
	homeState *isaac.HomeState
	nt        *contest_module.ChannelNetwork
	sc        *isaac.StateController
}

func NewNode(
	home node.Home,
	nodes []node.Node,
	globalConfig *Config,
	config *NodeConfig,
) (*Node, error) {
	log_ := log.New(log15.Ctx{"node": home.Alias()})

	lastBlock := config.LastBlock()
	previousBlock := config.Block(lastBlock.Height().Sub(1))
	homeState := isaac.NewHomeState(home, previousBlock).
		SetBlock(lastBlock)

	suffrage := newSuffrage(config, home, nodes)
	ballotChecker := isaac.NewCompilerBallotChecker(homeState, suffrage)

	numberOfActing := uint((*config.Modules.Suffrage)["number_of_acting"].(int))
	thr, _ := isaac.NewThreshold(numberOfActing, *config.Policy.Threshold)
	if err := thr.Set(isaac.StageINIT, globalConfig.NumberOfNodes(), *config.Policy.Threshold); err != nil {
		return nil, err
	}

	cm := isaac.NewCompiler(homeState, isaac.NewBallotbox(thr), ballotChecker)
	cm.SetLogContext(nil, "node", home.Alias())
	nt := contest_module.NewChannelNetwork(
		home,
		func(sl seal.Seal) (seal.Seal, error) {
			return sl, xerrors.Errorf("echo back")
		},
	)
	nt.SetLogContext(nil, "node", home.Alias())

	pv := contest_module.NewDummyProposalValidator()

	var sc *isaac.StateController
	{ // state handlers
		bs := isaac.NewBootingStateHandler(homeState)
		bs.SetLogContext(nil, "node", home.Alias())
		js, err := isaac.NewJoinStateHandler(
			homeState,
			cm,
			nt,
			pv,
			*config.Policy.IntervalBroadcastINITBallotInJoin,
			*config.Policy.TimeoutWaitVoteResultInJoin,
		)
		if err != nil {
			return nil, err
		}
		js.SetLogContext(nil, "node", home.Alias())

		dp := newProposalMaker(config, home)

		cs, err := isaac.NewConsensusStateHandler(
			homeState,
			cm,
			nt,
			suffrage,
			pv,
			dp,
			*config.Policy.TimeoutWaitBallot,
		)
		if err != nil {
			return nil, err
		}
		cs.SetLogContext(nil, "node", home.Alias())

		ss := isaac.NewStoppedStateHandler()
		ss.SetLogContext(nil, "node", home.Alias())

		ssr := newSealStorage(config, home)

		sc = isaac.NewStateController(homeState, cm, ssr, bs, js, cs, ss)
		sc.SetLogContext(nil, "node", home.Alias())

		go func() {
			for m := range nt.Reader() {
				_ = sc.Write(m)
			}
		}()
	}

	log_.Debug(
		"node created",
		"config", config,
		"home", home,
		"homeState", homeState,
		"threshold", thr,
		"suffrage", suffrage,
	)

	n := &Node{
		homeState: homeState,
		nt:        nt,
		sc:        sc,
	}

	return n, nil
}

func (no *Node) Home() node.Home {
	return no.homeState.Home()
}

func (no *Node) Start() error {
	if err := no.nt.Start(); err != nil {
		return err
	}

	if err := no.sc.Start(); err != nil {
		return err
	}

	return nil
}

func (no *Node) Stop() error {
	if err := no.sc.Stop(); err != nil {
		return err
	}

	if err := no.nt.Stop(); err != nil {
		return err
	}

	return nil
}

func NewHome(i uint) node.Home {
	pk, _ := keypair.NewStellarPrivateKey()

	h, _ := node.NewAddress([]byte{uint8(i)})
	return node.NewHome(h, pk)
}

func newSuffrage(config *NodeConfig, home node.Home, nodes []node.Node) isaac.Suffrage {
	sc := *config.Modules.Suffrage

	numberOfActing := uint(sc["number_of_acting"].(int))

	switch sc["name"] {
	case "FixedProposerSuffrage":
		// find proposer
		proposer := interface{}(home).(node.Node)
		for _, n := range nodes {
			if n.Alias() == sc["proposer"].(string) {
				proposer = n
				break
			}
		}

		return contest_module.NewFixedProposerSuffrage(proposer, numberOfActing, nodes...)
	case "RoundrobinSuffrage":
		return contest_module.NewRoundrobinSuffrage(numberOfActing, nodes...)
	default:
		panic(xerrors.Errorf("unknown suffrage config: %v", config))
	}
}

func newProposalMaker(config *NodeConfig, home node.Home) isaac.ProposalMaker {
	pc := *config.Modules.ProposalMaker
	switch pc["name"] {
	case "DefaultProposalMaker":
		delay, err := time.ParseDuration(pc["delay"].(string))
		if err != nil {
			panic(err)
		}

		dp := isaac.NewDefaultProposalMaker(home, delay)
		dp.SetLogContext(nil, "node", home.Alias())
		return dp
	default:
		panic(xerrors.Errorf("unknown proposal maker config: %v", config))
	}
}

func newSealStorage(config *NodeConfig, home node.Home) isaac.SealStorage {
	ss := contest_module.NewMemorySealStorage()
	ss.SetLogContext(nil, "node", home.Alias())

	return ss
}
