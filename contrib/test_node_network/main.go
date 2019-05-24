package main

import (
	"time"

	"github.com/inconshreveable/log15"

	"github.com/spikeekips/mitum/common"
	"github.com/spikeekips/mitum/isaac"
	"github.com/spikeekips/mitum/network"
	"github.com/spikeekips/mitum/storage"

	"github.com/spikeekips/mitum/contrib/test_node_network/lib"
)

var log log15.Logger = log15.New("module", "test-node-network")

var (
	seeds []string = []string{
		"SACCEWPT5O5RW67F53M6TMVELMJ34DC5EDWQPSG4RRTQ72BJBSASJTWY", // GAEU.6BSA
		"SCVEZIL32YTRJBFQYJSZZDNAE2CZBRABMKG524LZPBQXHFLBP3OPYFLK", // GCHI.OEDQ
		"SB4NCVQZQDX6NYVVQPIAJJGES3GOK5PUJI7DP6ZTJV62N3K7CS6JN5XN", // GDJ7.M4FG
		"SAPIKULLJIWWSR3O3RKW3ZGIRAMJWFPDIR7B6H54ADSNWKMGHGU4HBQO", // GDZH.X5S3
		//"SAUR7Z7SPYCMAILQYMFWMGNFLPFQKI3DRPLNNEUBCHT23N2GSSHWW7QS",
		//"SDHMHAPZ3HCHK54MDTPYGTBN5YISVZNQP6U2ZYIXQNQINH7FLWU7IK3S",
		//"SAAMSDCC5WX6H3EMKN5FK2FRCMN5GA2HULXVBKJDPQZN4ZKHUZWSPQQM",
		//"SDGBJJQAWEE3JRNNXPFZPF4ESE3JAJVV5EM3YGVZVX7Q24VBIVGGDNLE",
		//"SBP3FD3YDKSJD36CWJF4KWYGMYVYS5ADRYKKNZ5YS2BKVOT5NTWPMZD7",
		//"SBJNPBE4RPLKV5C2PP47NPEWUNJ6TKKBYGDAWMXHSRBBZER6SEB255BW",
		//"SDKHEIABEUHJFDX5LXTQHSLR3DUB2TKLQIQPUN3NZGWQLWNORMTINHPI",
		//"SDTWLWDSIBYZUYF7VROKJ4GUQ4ABPUY5EKFJGOPMIX72P4EOEG2PTP7M",
		//"SDEPMOUYNUNWIEIGRFHR6JMUHRA35TFWRM4LDFCALW6H7XCWTXAVZRC7",
	}
	height     common.Big            = common.NewBig(33)
	block      common.Hash           = common.NewRandomHash("bk")
	blockState []byte                = []byte("initial state")
	policy     isaac.ConsensusPolicy = isaac.ConsensusPolicy{
		NetworkID:                 common.TestNetworkID,
		Total:                     uint(len(seeds)),
		Threshold:                 uint(len(seeds) - 1),
		TimeoutWaitSeal:           time.Second * 3,
		AvgBlockRoundInterval:     time.Second * 5,
		SealSignedAtAllowDuration: time.Second * 3,
	}
)

func init() {
	common.InTest = false

	handler, _ := common.LogHandler(common.LogFormatter("json"), "")
	handler = log15.LvlFilterHandler(log15.LvlDebug, handler)
	handler = log15.CallerFileHandler(handler)

	loggers := []log15.Logger{
		log,
		common.Log(),
		isaac.Log(),
		network.Log(),
		storage.Log(),
		lib.Log(),
	}
	for _, l := range loggers {
		l.SetHandler(handler)
	}

	// NOTE TimeSyncer
	//{
	//	syncer, err := common.NewTimeSyncer("zero.bora.net", time.Second*10)
	//	if err != nil {
	//		panic(err)
	//	}
	//	_ = syncer.Start()
	//	common.SetTimeSyncer(syncer)
	//}
}

func createNodes() []*lib.Node {
	log.Debug("starting", "policy", policy)

	var nodes []*lib.Node
	for _, seedString := range seeds {
		node, err := lib.CreateNode(seedString, height, block, blockState, policy)
		if err != nil {
			panic(err)
		}
		node.Log().Debug("node created")
		nodes = append(nodes, node)
	}

	var validators []common.Node
	for _, node := range nodes {
		validators = append(validators, node.Home.AsValidator())
	}

	for _, node := range nodes {
		node.ProposerSelector, _ = isaac.NewDefaultProposerSelector(validators)
	}

	return nodes
}

func stopProposal(nodes []*lib.Node) {
	ps := isaac.NewFixedProposerSelector()
	ps.SetProposer(nodes[0].Home)

	for _, node := range nodes {
		node.ProposerSelector = ps
	}

	// NOTE proposer does not send proposal
	nodes[0].SealBroadcaster.SetManipulateFuncs(lib.StopProposal, nil)
}

func highHeightACCEPTBallot(nodes []*lib.Node) {
	for _, node := range nodes {
		node.SealBroadcaster.SetManipulateFuncs(lib.HighHeightACCEPTBallot, nil)
	}
}

func higherHeight(nodes []*lib.Node, higherNodes []*lib.Node) {
	// TODO needs to consider the proposer selection
	ps := isaac.NewFixedProposerSelector()
	ps.SetProposer(nodes[0].Home)

	for _, node := range nodes {
		node.ProposerSelector = ps
	}

	for _, node := range higherNodes {
		node.State.SetHeight(node.State.Height().Inc().Inc())
	}
}

func main() {
	nodes := createNodes()

	//stopProposal(nodes)
	higherHeight(nodes, nodes[:3])

	if err := lib.PrepareNodePool(nodes); err != nil {
		panic(err)
	}

	if err := lib.StartNodes(nodes); err != nil {
		panic(err)
	}

	defer lib.StopNodes(nodes)

	select {}
}
