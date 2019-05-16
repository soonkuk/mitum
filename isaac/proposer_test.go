package isaac

import (
	"fmt"
	"sort"
	"testing"

	"github.com/spikeekips/mitum/common"
	"github.com/stretchr/testify/suite"
)

type testProposer struct {
	suite.Suite
}

func (t *testProposer) TestElect() {
	cases := []struct {
		name             string
		height           common.Big
		round            Round
		candidates       []string // candidates
		block            string
		expectedProposer string // selected proposer
		expectedIndex    int    // proposer index in the candidates
		err              error
	}{
		{
			height:           common.NewBig(0),
			round:            Round(0),
			candidates:       []string{"GBZIQWDUSKHO3Z3HCEIQX65KVIFYQAH7GN3QW3O3D5VNIG3OXE5WGO3B"},
			block:            "bk-3BGMn2GASCCPf12UeqLbAVsZxWE9Neoi36a1EMS6XjbW",
			expectedProposer: "GBZIQWDUSKHO3Z3HCEIQX65KVIFYQAH7GN3QW3O3D5VNIG3OXE5WGO3B",
			expectedIndex:    0,
		},
		{
			height: common.NewBig(1),
			round:  Round(0),
			candidates: []string{
				"GBLB5YXAGABBJ2CAA5MZE4OS6UQX5LVWPUCLY57Z5YV6ZYLIWDFQZ7FH",
				"GATQBQ3N6IHHM35TRLRBMZN5ZSADX26BR5SDKPMPNMCSEXEXHJ7DLFAB",
				"GCPVWPESHWGV7B3AHL4VY74JJ5D5OIA3EX2HMATDX62SET56H77VLY52",
				"GDFA6ZH23ERWEJDDNGUOOKHI4SVTT56JF5RHAUGTB3LTGBVJ3OZGOKYO",
				"GBUIQ7HEWMYQDNB6PCSO7MYJESUFJP7UGAA2Q2DQATK37LE67QEE4IDZ",
			},
			block:            "bk-3MVFdQ2qZoRXAv2PtgyCUMfusTuegTfiHitBY2Ct3Lni",
			expectedProposer: "GCPVWPESHWGV7B3AHL4VY74JJ5D5OIA3EX2HMATDX62SET56H77VLY52",
			expectedIndex:    3,
		},
		{
			height: common.NewBig(5),
			round:  Round(0),
			candidates: []string{
				"GBQUR4YYGPOXAGTALFC7TUPY2654STGUIHC6MZTKDDJAFN6AFC55VBLY",
				"GBQW2AJ7FOPMCANAHGNBJ5ZFRTDWDQCDM4RADM6JIIEWUQX4RDOMAJQV",
				"GALIGLVHCZAAS44FHHG5FYGNIMF3DYP2XY7CGLRQH5H6EE4GZ6YUJWID",
				"GD2GNOFBHZLLUUWRRFFVBRQFYGOOQXVNDVDAVREQ7234VXLD4CKIN7WG",
				"GAZM5BNXBFXVFSDD5L3JOBB4WAIGQAX4C56TNO5FTYCEMVWVGLQJUA5S",
			},
			block:            "bk-43KGf1y5rR6dq2SsDgaBsPvwXJNoAcaaEcnnMyrrU2tR",
			expectedProposer: "GAZM5BNXBFXVFSDD5L3JOBB4WAIGQAX4C56TNO5FTYCEMVWVGLQJUA5S",
			expectedIndex:    1,
		},
		{
			height: common.NewBig(33),
			round:  Round(5),
			candidates: []string{
				"GAQNARCKRAIIY7F7ECYPFFIF7MCQ2CWSKPKTMRQ22FLBO5IO6VL3KHVL",
				"GBMB2KPATKLN4XRMULZ4YOT6QTF636E2UYLB4KFXDPIJHNTGPOIN4VO7",
				"GDOK2PAYLLZ3J76P6BBXYNEKATJVZY3HDLWXXILW6U4QF6N2BUNFH5O7",
				"GCP75GV5JC47X3ERCNY5I3GWYU4T7SEUCWSBPC4AR2WQQNI22HQ3ITLQ",
			},
			block:            "bk-D2Rx4dpFRe8LT2WtqgCTybJGqHgQpuPJC7i1hjaBdteP",
			expectedProposer: "GAQNARCKRAIIY7F7ECYPFFIF7MCQ2CWSKPKTMRQ22FLBO5IO6VL3KHVL",
			expectedIndex:    0,
		},
		{
			height: common.NewBig(36),
			round:  Round(0),
			candidates: []string{
				"GCUP3CQT4KK3BNSWBAGOWVHRIPQK5CC2TABRFDHTFYCYZONRH7HVUMFM",
				"GDRBPB3KVBROWPDJNQEBOLYKTFMM5GMYW6SYW5H4YK4FEIWZWBCYPAE3",
				"GCGSC2WHMXWKYXZGCCLDITDEFTTBIHSXQALMZMCKUKUB7Y32R34MIUHP",
				"GDCNRTNKAJSK4WK5TIFAG7SD2QWMEPGLO7PF6RNLAFIXQNMUTO23Z6HG",
				"GAX5CCEY7R32WII437G27AY72EGR2LSHTDX3RDXF2WOFJNPEKOBFDOEG",
				"GC266JQQ7OQ42DRACWXNU5LITH4BNDXXT52SVSAYYIHJ3C7CQ6A4OUE4",
			},
			block:            "bk-2U1sr32jbu3Zg2SzFRz5rEEHLRZcEA1BNnQ7jRfN7Knw",
			expectedProposer: "GCUP3CQT4KK3BNSWBAGOWVHRIPQK5CC2TABRFDHTFYCYZONRH7HVUMFM",
			expectedIndex:    3,
		},
		{
			height: common.NewBig(36),
			round:  Round(1),
			candidates: []string{
				"GA5TTCIGOX7YYD34ICSOJKGFOPNRJVPXHRU3VINLN46C7IQBV5V3BCCF",
				"GAT6XDX4OGPD662FSCXB5VQ45TJHVWOJ7757XWIVMVCPPIUUL5ANNAMS",
				"GD4BPMATI4PVEG3UIZUMWK72FZCZGIB3GEUZS5RX6FVB4NQRGSIHUSVH",
				"GDXE533ZEP2SFWVZE46FYNYDMJERYP23JGYCUFZDHB2NBF7GBWWRPYVF",
				"GBEARXZKUMSCYDASJHZKH4YIY2BCN7WBM53ZKDOE47YQL7YBLWG7CUMC",
				"GB53EQECFL7PAK53RBJARLCMLLBZFGBQWWXRUCQ56PAFTDLWOD2GNVMT",
			},
			block:            "bk-HJMZb2eoHfwkGsPwof7Z2oJKJbdpTDMpvBMPZF722Juo",
			expectedProposer: "GD4BPMATI4PVEG3UIZUMWK72FZCZGIB3GEUZS5RX6FVB4NQRGSIHUSVH",
			expectedIndex:    4,
		},
		{
			height: common.NewBig(36),
			round:  Round(2),
			candidates: []string{
				"GAO37AYQ65UCL2OOGN33474TQTXCBBIUPNFTXXGOLWKWQH2FZPKI23SL",
				"GA2S3P5PLKVJC3R37UJXBRPE5TCCKZDZEHVLR7C7T4LPZLKM4Y7VFAYL",
				"GAUSFTZGVDLI4WDZEKIYP6ELK43MVAITTS3OA2GDOF7ZVHIGCBDHAFBN",
				"GC2HST6RTH5AT6CYJNKS63BJIEEZYSRARKNUT7FBKI3V7XX6ZKREGOIO",
				"GBVPK66ZRDBWTLHUTMNY2CVOXBYUZ7Z4TU5ZXSETAH7USBDOYGYTBRL5",
				"GD35TLOZEDZC3BY57SPJ24R4LUA2AH5FFUEXLDPCEA2SHKIMDPPVTMJF",
			},
			block:            "bk-92wktrdD4CcXSS24vKJBhHBZCGgNBAU7aQgMm2Mfe3wa",
			expectedProposer: "GA2S3P5PLKVJC3R37UJXBRPE5TCCKZDZEHVLR7C7T4LPZLKM4Y7VFAYL",
			expectedIndex:    0,
		},
		{
			height: common.NewBig(41),
			round:  Round(0),
			candidates: []string{
				"GAEGYKW52IXO4JI6GK5P5MJEIA6MQVFHJSRZBZ2E35V7U2V7NXXIJ3WW",
				"GDUVEOBZBW5LRIKQCDWFOQ25GXKWUT3SFCHLK5SLX3CYUKVJONKJS3IU",
				"GA7AUMHDF4PABPEVERZOGWH4MAMEDE5ZUUKO3IQSM446D3TNH3ZKOOT4",
				"GD7O6J7R6TSN32RNNZSFYZN3JWABEGBBQXI36EOV2KC2UKCBM4ZXGHEM",
				"GAEC6ERG6ORZONTLVEWPCLFY7RMCDWTWMTQ7ZYLB6MQJRS5CPF2J32UI",
				"GCPHMJIMJZJFKJVVN6WUJN6I423AKWMEXPLAGTLANWNVHEGY4ODFJB5U",
			},
			block:            "bk-DYejqpU4FsRD4nkAvva3qa8nrd8Qj2eWoi5Pti3mRnvH",
			expectedProposer: "GD7O6J7R6TSN32RNNZSFYZN3JWABEGBBQXI36EOV2KC2UKCBM4ZXGHEM",
			expectedIndex:    4,
		},
		{
			height: common.NewBig(41),
			round:  Round(1),
			candidates: []string{
				"GCO745ODDL7T63MNJWFT3OTXLLYO6VZPPDW2KMQGQOGHWV2AYO2UUBC4",
				"GA4REJWS3P3OZGNNBPV7XP4MTYWP6YDAVQCYLP7DGB7AEHQAXVWESVPW",
				"GDYUJGB26HVUNLEIYRJRT4EOAJTYDVSZQRKILEYIK3A7FMJCECLOFWQU",
				"GB6RAMONVXNKTZQQYDMEHJANARD7FIA2FPUDHT7JGFB5ZJZ73UFAVAIX",
				"GCRQKAPXJ2FQG4CCJISUDPTKF25WMT4NPZPE6WMVCF2KPTOCEQZMGQ73",
				"GBIUVD7IYQFUVGG56TTPVYKIF5JBNILIS2RO6HORUNGRWOKH467M4WFJ",
			},
			block:            "bk-74RfuAJ88FBanExrCuMcwTgbKt79hEM2kiPELLj4WWma",
			expectedProposer: "GCO745ODDL7T63MNJWFT3OTXLLYO6VZPPDW2KMQGQOGHWV2AYO2UUBC4",
			expectedIndex:    3,
		},
		{
			height: common.NewBig(41),
			round:  Round(2),
			candidates: []string{
				"GD6CXGIV2RY5GEYFR7C4L7JNFKYIJCVOMXDYNVVWC2RRLCTIKJ5YQJEZ",
				"GANOCKTM3M6WKH5O5JHU2XSAUGKQZUMWSI5Z7OJ2FTWX5YD5J6IFQWLL",
				"GC4XTPQWJ2Q3BNW2ISCIXYVLWMNXUY4RKNTITIXO5VCZILC7N7KLCCZK",
				"GBWF6QWBPJGFQ4MDWTGND22S4EIC34ZZ3J4CORPYAE56Q2E5VBOHVLWL",
				"GCAOC7FT5N36G7VRGXXZHBG6VFMDSD6PCYW4F55WIGUPFEDCKDHMX4P4",
				"GBO2JBVWUZCLOMCBKYRWNLLU3T77MQBN7Q6TSMGJ3YLDPFS3OOFHDVFG",
			},
			block:            "bk-49FTVG7MSUSBeyBRYBLknwvSykL7TLBNGbqauoudn6N4",
			expectedProposer: "GANOCKTM3M6WKH5O5JHU2XSAUGKQZUMWSI5Z7OJ2FTWX5YD5J6IFQWLL",
			expectedIndex:    0,
		},
	}

	for i, c := range cases {
		i := i
		c := c
		c.name = fmt.Sprintf(
			"c=%v block=%s height=%v round=%v proposer=%v index=%v",
			len(c.candidates),
			c.block[:5],
			c.height,
			c.round,
			common.Address(c.expectedProposer).Alias(),
			c.expectedIndex,
		)
		t.T().Run(
			c.name,
			func(*testing.T) {
				var addresses []common.Address
				for _, a := range c.candidates {
					addresses = append(addresses, common.Address(a))
				}
				sort.Sort(common.SortAddress(addresses))

				var nodes []common.Node
				for _, a := range addresses {
					nodes = append(nodes, common.NewValidator(a, common.NetAddr{}))
				}

				block, err := common.ParseHash(c.block)
				t.NoError(err)

				selector, err := NewDefaultProposerSelector(nodes)
				t.NoError(err)
				selected, err := selector.Select(block, c.height, c.round)
				if c.err != nil {
					t.Error(c.err, err)
					return
				}

				// get index
				var index int = -1
				for j, candidate := range addresses {
					if selected.Address() != candidate {
						continue
					}
					index = j
					break
				}

				t.True(index >= 0)
				t.Equal(c.expectedProposer, selected.Address().String(), "%d: %v", i, c.name)
				t.Equal(c.expectedIndex, index, "%d: %v", i, c.name)
			},
		)
	}
}

func TestProposer(t *testing.T) {
	suite.Run(t, new(testProposer))
}
