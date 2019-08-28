package main

import (
	"fmt"
	"strings"
	"time"

	"golang.org/x/xerrors"

	"github.com/rs/zerolog"
	"github.com/spikeekips/mitum/common"
	contest_module "github.com/spikeekips/mitum/contrib/contest/module"
	"github.com/spikeekips/mitum/isaac"
	"github.com/spikeekips/mitum/keypair"
	"github.com/spikeekips/mitum/node"
	"github.com/spikeekips/mitum/seal"
)

type Node struct {
	*common.Logger
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
	rootLog := log.With().Str("node", home.Alias()).Logger()
	log_ := rootLog.With().Str("module", "node").Logger()

	lastBlock := config.LastBlock()
	previousBlock := config.Block(lastBlock.Height().Sub(1))
	homeState := isaac.NewHomeState(home, previousBlock).SetBlock(lastBlock)

	suffrage := newSuffrage(config, home, nodes)
	ballotChecker := isaac.NewCompilerBallotChecker(homeState, suffrage)
	ballotChecker.SetLogger(rootLog)

	numberOfActing := uint((*config.Modules.Suffrage)["number_of_acting"].(int))
	thr, _ := isaac.NewThreshold(numberOfActing, *config.Policy.Threshold)
	if err := thr.Set(isaac.StageINIT, globalConfig.NumberOfNodes(), *config.Policy.Threshold); err != nil {
		return nil, err
	}

	cm := isaac.NewCompiler(homeState, isaac.NewBallotbox(thr), ballotChecker)
	cm.SetLogger(rootLog)

	nt := contest_module.NewChannelNetwork(
		home,
		func(sl seal.Seal) (seal.Seal, error) {
			return sl, xerrors.Errorf("echo back")
		},
	)
	nt.SetLogger(rootLog)

	pv := contest_module.NewDummyProposalValidator()

	ballotMaker := newBallotMaker(config, home, rootLog)

	var sc *isaac.StateController
	{ // state handlers
		bs := isaac.NewBootingStateHandler(homeState)
		bs.SetLogger(rootLog)

		js, err := isaac.NewJoinStateHandler(
			homeState,
			cm,
			nt,
			suffrage,
			ballotMaker,
			pv,
			*config.Policy.IntervalBroadcastINITBallotInJoin,
			*config.Policy.TimeoutWaitVoteResultInJoin,
		)
		if err != nil {
			return nil, err
		}
		js.SetLogger(rootLog)

		dp := newProposalMaker(config, home, rootLog)

		cs, err := isaac.NewConsensusStateHandler(
			homeState,
			cm,
			nt,
			suffrage,
			ballotMaker,
			pv,
			dp,
			*config.Policy.TimeoutWaitBallot,
		)
		if err != nil {
			return nil, err
		}
		cs.SetLogger(rootLog)

		ss := isaac.NewStoppedStateHandler()
		ss.SetLogger(rootLog)

		ssr := newSealStorage(config, home, rootLog)

		sc = isaac.NewStateController(homeState, cm, ssr, bs, js, cs, ss)
		sc.SetLogger(rootLog)
	}

	log_.Info().
		Object("config", config).
		Object("home", home).
		Object("homeState", homeState).
		Object("threshold", thr).
		Interface("suffrage", suffrage).
		Msg("node created")

	n := &Node{
		Logger: common.NewLogger(func(c zerolog.Context) zerolog.Context {
			return c.Str("node", home.Alias())
		}),
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
	started := time.Now()

	if err := no.nt.Start(); err != nil {
		return err
	}

	if err := no.sc.Start(); err != nil {
		return err
	}

	go func() {
		for m := range no.nt.Reader() {
			go func(m interface{}) {
				started := time.Now()
				err := no.sc.Receive(m)
				no.sc.Log().Debug().
					Err(err).
					Dur("elapsed", time.Now().Sub(started)).
					Msg("message received")
			}(m)
		}
	}()
	no.Log().Debug().Dur("elapsed", time.Now().Sub(started)).Msg("node started")

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
		var proposer node.Node
		for _, n := range nodes {
			if n.Alias() == sc["proposer"].(string) {
				proposer = n
				break
			}
		}
		if proposer == nil {
			panic(xerrors.Errorf("failed to find proposer: %v", config))
		}

		return contest_module.NewFixedProposerSuffrage(proposer, numberOfActing, nodes...)
	case "RoundrobinSuffrage":
		return contest_module.NewRoundrobinSuffrage(numberOfActing, nodes...)
	default:
		panic(xerrors.Errorf("unknown suffrage config: %v", config))
	}
}

func newProposalMaker(config *NodeConfig, home node.Home, l zerolog.Logger) isaac.ProposalMaker {
	pc := *config.Modules.ProposalMaker
	switch pc["name"] {
	case "DefaultProposalMaker":
		delay, err := time.ParseDuration(pc["delay"].(string))
		if err != nil {
			panic(err)
		}

		dp := isaac.NewDefaultProposalMaker(home, delay)
		dp.SetLogger(l)
		return dp
	default:
		panic(xerrors.Errorf("unknown proposal maker config: %v", config))
	}
}

func newSealStorage(config *NodeConfig, home node.Home, l zerolog.Logger) isaac.SealStorage {
	ss := contest_module.NewMemorySealStorage()
	ss.SetLogger(l)

	return ss
}

func newBallotMaker(config *NodeConfig, home node.Home, l zerolog.Logger) isaac.BallotMaker {
	pc := *config.Modules.BallotMaker
	switch pc["name"] {
	case "DefaultBallotMaker":
		return isaac.NewDefaultBallotMaker(home)
	case "DamangedBallotMaker":
		bmc := config.Modules.BallotMaker

		var height, round, stage string

		if s, found := (*bmc)["height"]; !found {
			height = "*"
		} else {
			height = fmt.Sprintf("%v", s)
		}

		if s, found := (*bmc)["round"]; !found {
			round = "*"
		} else {
			round = fmt.Sprintf("%v", s)
		}

		if s, found := (*bmc)["stage"]; !found {
			stage = "*"
		} else {
			stage = strings.ToUpper(s.(string))
		}

		db := contest_module.NewDamangedBallotMaker(home)
		db = db.AddPoint(height, round, stage)

		return db
	default:
		panic(xerrors.Errorf("unknown ballot_maker found: %v", pc["name"]))
	}
}
