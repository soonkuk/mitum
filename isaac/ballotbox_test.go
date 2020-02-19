package isaac

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/spikeekips/mitum/hint"
	"github.com/spikeekips/mitum/key"
	"github.com/spikeekips/mitum/valuehash"
)

type testBallotbox struct {
	suite.Suite

	pk key.Privatekey
}

func (t *testBallotbox) SetupSuite() {
	_ = hint.RegisterType(INITBallotType, "init-ballot")
	_ = hint.RegisterType(ProposalBallotType, "proposal")
	_ = hint.RegisterType(SIGNBallotType, "sign-ballot")
	_ = hint.RegisterType(ACCEPTBallotType, "accept-ballot")
	_ = hint.RegisterType((valuehash.SHA256{}).Hint().Type(), "sha256")

	t.pk, _ = key.NewBTCPrivatekey()
}

func (t *testBallotbox) newLocalState(total uint, percent float64) *LocalState {
	ls, err := NewLocalState(nil, nil)
	t.NoError(err)

	threshold, _ := NewThreshold(total, percent)
	_ = ls.Policy().SetThreshold(threshold)

	return ls
}

func (t *testBallotbox) TestNew() {
	bb := NewBallotbox(t.newLocalState(2, 67))
	ba := t.newINITBallot(Height(10), Round(0), NewShortAddress("test-for-init-ballot"))

	vp, err := bb.Vote(ba)
	t.NoError(err)
	t.NotEmpty(vp)
}

func (t *testBallotbox) newINITBallot(height Height, round Round, node Address) INITBallotV0 {
	vp := NewDummyVoteProof(
		height-1,
		Round(0),
		StageACCEPT,
		VoteProofMajority,
	)

	ib := INITBallotV0{
		BaseBallotV0: BaseBallotV0{
			node: node,
		},
		INITBallotFactV0: INITBallotFactV0{
			BaseBallotFactV0: BaseBallotFactV0{
				height: height,
				round:  round,
			},
			previousBlock: valuehash.RandomSHA256(),
			previousRound: vp.Round(),
		},
		voteProof: vp,
	}
	t.NoError(ib.Sign(t.pk, nil))

	return ib
}

func (t *testBallotbox) TestVoteRace() {
	bb := NewBallotbox(t.newLocalState(50, 100))

	checkDone := make(chan bool)
	vrChan := make(chan interface{}, 49)

	go func() {
		for i := range vrChan {
			switch c := i.(type) {
			case error:
				t.NoError(c)
			case VoteProof:
				t.Equal(VoteProofNotYet, c.Result())
			}
		}
		checkDone <- true
	}()

	var wg sync.WaitGroup
	wg.Add(49)
	for i := 0; i < 49; i++ {
		go func() {
			defer wg.Done()
			ba := t.newINITBallot(Height(10), Round(0), RandomShortAddress())

			vp, err := bb.Vote(ba)
			if err != nil {
				vrChan <- err
			} else {
				vrChan <- vp
			}
		}()
	}
	wg.Wait()
	close(vrChan)

	<-checkDone
}

func (t *testBallotbox) TestINITVoteProofNotYet() {
	bb := NewBallotbox(t.newLocalState(2, 67))
	ba := t.newINITBallot(Height(10), Round(0), NewShortAddress("test-for-init-ballot"))

	vp, err := bb.Vote(ba)
	t.NoError(err)
	t.Equal(VoteProofNotYet, vp.Result())

	t.Equal(ba.Height(), vp.Height())
	t.Equal(ba.Round(), vp.Round())
	t.Equal(ba.Stage(), vp.Stage())

	vrs := bb.loadVoteRecords(ba, false)
	t.NotNil(vrs)

	ib, found := vrs.ballots[ba.Node()]
	t.True(found)

	iba := ib.(INITBallotV0)
	t.True(ba.PreviousBlock().Equal(iba.PreviousBlock()))
	t.Equal(ba.Node(), iba.Node())
	t.Equal(ba.PreviousRound(), iba.PreviousRound())
}

func (t *testBallotbox) TestINITVoteProofDraw() {
	bb := NewBallotbox(t.newLocalState(2, 67))

	// 2 ballot have the differnt previousBlock hash
	{
		ba := t.newINITBallot(Height(10), Round(0), NewShortAddress("node0"))
		vp, err := bb.Vote(ba)
		t.NoError(err)
		t.Equal(VoteProofNotYet, vp.Result())
	}
	{
		ba := t.newINITBallot(Height(10), Round(0), NewShortAddress("node1"))
		vp, err := bb.Vote(ba)
		t.NoError(err)
		t.Equal(VoteProofDraw, vp.Result())
		t.True(vp.IsFinished())
		t.NotNil(vp.FinishedAt())
		t.True(time.Now().Sub(vp.FinishedAt()) < time.Second)
	}

	{ // already finished
		ba := t.newINITBallot(Height(10), Round(0), NewShortAddress("node2"))
		vp, err := bb.Vote(ba)
		t.NoError(err)
		t.Equal(VoteProofDraw, vp.Result())
		t.True(vp.IsFinished())
		t.True(vp.IsClosed())
	}
}

func (t *testBallotbox) TestINITVoteProofMajority() {
	bb := NewBallotbox(t.newLocalState(3, 66))

	// 2 ballot have the differnt previousBlock hash
	ba0 := t.newINITBallot(Height(10), Round(0), NewShortAddress("node0"))
	ba1 := t.newINITBallot(Height(10), Round(0), NewShortAddress("node1"))

	{ // set same previousBlock and previousRound
		ba1.previousBlock = ba0.previousBlock
		ba1.previousRound = ba0.previousRound

		t.NoError(ba1.Sign(t.pk, nil))
	}

	{
		vp, err := bb.Vote(ba0)
		t.NoError(err)
		t.Equal(VoteProofNotYet, vp.Result())
	}
	{
		vp, err := bb.Vote(ba1)
		t.NoError(err)
		t.Equal(VoteProofMajority, vp.Result())
	}
}

func (t *testBallotbox) TestINITVoteProofClean() {
	bb := NewBallotbox(t.newLocalState(3, 66))

	// 2 ballot have the differnt previousBlock hash
	ba0 := t.newINITBallot(Height(10), Round(0), NewShortAddress("node0"))
	ba1 := t.newINITBallot(Height(10), Round(0), NewShortAddress("node1"))
	bar := t.newINITBallot(Height(9), Round(0), NewShortAddress("node0"))

	{ // set same previousBlock and previousRound
		ba1.previousBlock = ba0.previousBlock
		ba1.previousRound = ba0.previousRound

		t.NoError(ba1.Sign(t.pk, nil))
	}

	{
		vp, err := bb.Vote(ba0)
		t.NoError(err)
		t.Equal(VoteProofNotYet, vp.Result())
	}

	{
		_, err := bb.Vote(bar)
		t.NoError(err)
	}

	{
		vp, err := bb.Vote(ba1)
		t.NoError(err)
		t.Equal(VoteProofMajority, vp.Result())
	}

	var remains []string
	bb.vrs.Range(func(k, v interface{}) bool {
		remains = append(remains, k.(string))
		return true
	})
	t.Equal(1, len(remains))

	var barFound bool
	for _, r := range remains {
		if r == "9-0-1" {
			barFound = true
			break
		}
	}
	t.False(barFound)
}

func (t *testBallotbox) newACCEPTBallot(height Height, round Round, node Address) ACCEPTBallotV0 {
	vp := NewDummyVoteProof(
		height,
		round,
		StageINIT,
		VoteProofMajority,
	)

	ib := ACCEPTBallotV0{
		BaseBallotV0: BaseBallotV0{
			node: node,
		},
		ACCEPTBallotFactV0: ACCEPTBallotFactV0{
			BaseBallotFactV0: BaseBallotFactV0{
				height: height,
				round:  round,
			},
			proposal: valuehash.RandomSHA256(),
			newBlock: valuehash.RandomSHA256(),
		},
		voteProof: vp,
	}
	t.NoError(ib.Sign(t.pk, nil))

	return ib
}

func (t *testBallotbox) TestACCEPTVoteProofNotYet() {
	bb := NewBallotbox(t.newLocalState(2, 67))
	ba := t.newACCEPTBallot(Height(10), Round(0), NewShortAddress("test-for-accept-ballot"))

	vp, err := bb.Vote(ba)
	t.NoError(err)
	t.Equal(VoteProofNotYet, vp.Result())

	t.Equal(ba.Height(), vp.Height())
	t.Equal(ba.Round(), vp.Round())
	t.Equal(ba.Stage(), vp.Stage())

	vrs := bb.loadVoteRecords(ba, false)
	t.NotNil(vrs)

	ib, found := vrs.ballots[ba.Node()]
	t.True(found)

	iba := ib.(ACCEPTBallotV0)
	t.True(ba.Proposal().Equal(iba.Proposal()))
	t.Equal(ba.Node(), iba.Node())
	t.Equal(ba.NewBlock(), iba.NewBlock())
}

func (t *testBallotbox) TestACCEPTVoteProofDraw() {
	bb := NewBallotbox(t.newLocalState(2, 67))

	// 2 ballot have the differnt previousBlock hash
	ba0 := t.newACCEPTBallot(Height(10), Round(0), NewShortAddress("node0"))
	ba1 := t.newACCEPTBallot(Height(10), Round(0), NewShortAddress("node1"))

	{
		vp, err := bb.Vote(ba0)
		t.NoError(err)
		t.Equal(VoteProofNotYet, vp.Result())
	}
	{
		vp, err := bb.Vote(ba1)
		t.NoError(err)
		t.Equal(VoteProofDraw, vp.Result())
	}
}

func (t *testBallotbox) TestACCEPTVoteProofMajority() {
	bb := NewBallotbox(t.newLocalState(3, 66))

	// 2 ballot have the differnt previousBlock hash
	ba0 := t.newACCEPTBallot(Height(10), Round(0), NewShortAddress("node0"))
	ba1 := t.newACCEPTBallot(Height(10), Round(0), NewShortAddress("node1"))

	{ // set same previousBlock and previousRound
		ba1.proposal = ba0.proposal
		ba1.newBlock = ba0.newBlock

		t.NoError(ba1.Sign(t.pk, nil))
	}

	{
		vp, err := bb.Vote(ba0)
		t.NoError(err)
		t.Equal(VoteProofNotYet, vp.Result())
	}
	{
		vp, err := bb.Vote(ba1)
		t.NoError(err)
		t.Equal(VoteProofMajority, vp.Result())
	}
}

func (t *testBallotbox) TestINITVoteProofMajorityClosed() {
	bb := NewBallotbox(t.newLocalState(3, 66))

	// 2 ballot have the differnt previousBlock hash
	ba0 := t.newINITBallot(Height(10), Round(0), NewShortAddress("n0"))
	ba1 := t.newINITBallot(Height(10), Round(0), NewShortAddress("n1"))
	ba2 := t.newINITBallot(Height(10), Round(0), NewShortAddress("n2"))

	{ // set same previousBlock and previousRound
		ba1.previousBlock = ba0.previousBlock
		ba1.previousRound = ba0.previousRound

		t.NoError(ba1.Sign(t.pk, nil))
	}

	{
		vp, err := bb.Vote(ba0)
		t.NoError(err)
		t.Equal(VoteProofNotYet, vp.Result())
		t.False(vp.IsClosed())
	}

	{
		vp, err := bb.Vote(ba1)
		t.NoError(err)
		t.Equal(VoteProofMajority, vp.Result())
		t.False(vp.IsClosed())
	}

	{
		vp, err := bb.Vote(ba2)
		t.NoError(err)
		t.Equal(VoteProofMajority, vp.Result())
		t.True(vp.IsClosed())
	}
}

func TestBallotbox(t *testing.T) {
	suite.Run(t, new(testBallotbox))
}