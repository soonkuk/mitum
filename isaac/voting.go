package isaac

import (
	"encoding/json"
	"sort"
	"sync"

	"github.com/spikeekips/mitum/common"
)

type RoundVoting struct {
	sync.RWMutex
	proposals map[common.Hash]*VotingProposal
}

func NewRoundVoting() *RoundVoting {
	return &RoundVoting{
		proposals: map[common.Hash]*VotingProposal{},
	}
}

// NewRound will only get the ProposeSealType
func (r *RoundVoting) Open(seal common.Seal) (*VotingProposal, *VotingStage, error) {
	if seal.Type != ProposeSealType {
		return nil, nil, InvalidSealTypeError
	}

	psHash, _, err := seal.Hash()
	if err != nil {
		return nil, nil, err
	}

	if r.IsRunning(psHash) {
		return nil, nil, VotingProposalAlreadyStartedError
	}

	var Propose Propose
	if err := seal.UnmarshalBody(&Propose); err != nil {
		return nil, nil, err
	}

	vp := NewVotingProposal(Propose.Block.Height)

	r.Lock()
	defer r.Unlock()

	r.proposals[psHash] = vp

	vs := vp.Stage(VoteStageSIGN)
	err = vs.Vote(
		psHash,
		seal.Source,
		VoteYES,
	)
	if err != nil {
		return nil, nil, err
	}

	return vp, vs, nil
}

func (r *RoundVoting) IsRunning(psHash common.Hash) bool {
	r.RLock()
	defer r.RUnlock()

	_, found := r.proposals[psHash]
	return found
}

func (r *RoundVoting) Finish(psHash common.Hash) error {
	if !r.IsRunning(psHash) {
		return nil
	}

	r.Lock()
	defer r.Unlock()

	currentHeight := r.proposals[psHash].height

	removeHashes := []common.Hash{psHash}
	for ph, vp := range r.proposals {
		if vp.height.Cmp(currentHeight) < 1 { // same or lower
			removeHashes = append(removeHashes, ph)
		}
	}

	for _, h := range removeHashes {
		delete(r.proposals, h)
	}

	return nil
}

func (r *RoundVoting) Proposal(psHash common.Hash) *VotingProposal {
	r.RLock()
	defer r.RUnlock()

	vp, found := r.proposals[psHash]
	if !found {
		return nil
	}

	return vp
}

func (r *RoundVoting) Vote(ballot Ballot) (*VotingProposal, *VotingStage, error) {
	if !r.IsRunning(ballot.ProposeSeal) {
		return nil, nil, VotingProposalNotFoundError
	}

	vp := r.Proposal(ballot.ProposeSeal)
	if vp == nil {
		return nil, nil, VotingProposalNotFoundError
	}

	stage := vp.Stage(ballot.Stage)
	if stage == nil {
		return nil, nil, InvalidVoteStageError
	}

	err := stage.Vote(
		ballot.ProposeSeal,
		ballot.Source,
		ballot.Vote,
	)
	if err != nil {
		return nil, nil, err
	}

	return vp, stage, nil
}

// Agreed clean up the other proposals except agreed proposal
func (r *RoundVoting) Agreed(psHash common.Hash) error {
	if !r.IsRunning(psHash) {
		return VotingProposalNotFoundError
	}

	r.Lock()
	defer r.Unlock()

	for k, _ := range r.proposals {
		if k.Equal(psHash) {
			continue
		}
		log.Debug(
			"voting agreed; the other proposals will be removed",
			"agreed-seal", psHash,
			"proposal", k,
		)
		delete(r.proposals, k)
	}

	return nil
}

type VotingProposal struct {
	height      common.Big
	StageINIT   *VotingStage
	StageSIGN   *VotingStage
	StageACCEPT *VotingStage
}

func NewVotingProposal(height common.Big) *VotingProposal {
	return &VotingProposal{
		height:      height,
		StageINIT:   NewVotingStage(),
		StageSIGN:   NewVotingStage(),
		StageACCEPT: NewVotingStage(),
	}
}

func (r *VotingProposal) Stage(stage VoteStage) *VotingStage {
	var vs *VotingStage
	switch stage {
	case VoteStageINIT:
		vs = r.StageINIT
	case VoteStageSIGN:
		vs = r.StageSIGN
	case VoteStageACCEPT:
		vs = r.StageACCEPT
	default:
		return nil
	}

	return vs
}

func (r *VotingProposal) String() string {
	b, _ := json.Marshal(map[string]interface{}{
		"height": r.height,
		"stages": map[string]interface{}{
			VoteStageINIT.String():   r.StageINIT.String(),
			VoteStageSIGN.String():   r.StageSIGN.String(),
			VoteStageACCEPT.String(): r.StageACCEPT.String(),
		},
	})

	return common.TerminalLogString(string(b))
}

type VotingStage struct {
	sync.RWMutex
	yes map[ /* source */ common.Address]common.Hash
	nop map[ /* source */ common.Address]common.Hash
	exp map[ /* source */ common.Address]common.Hash
}

func NewVotingStage() *VotingStage {
	return &VotingStage{
		yes: map[common.Address]common.Hash{},
		nop: map[common.Address]common.Hash{},
		exp: map[common.Address]common.Hash{},
	}
}

func (r *VotingStage) String() string {
	var yes, nop, exp []string
	for k, _ := range r.yes {
		yes = append(yes, k.Alias())
	}
	for k, _ := range r.nop {
		nop = append(nop, k.Alias())
	}
	for k, _ := range r.exp {
		exp = append(exp, k.Alias())
	}

	b, _ := json.Marshal(map[string]interface{}{
		"yes": []interface{}{len(yes), yes},
		"nop": []interface{}{len(nop), nop},
		"exp": []interface{}{len(exp), exp},
	})

	return common.TerminalLogString(string(b))
}

func (r *VotingStage) YES() map[common.Address]common.Hash {
	r.RLock()
	defer r.RUnlock()

	return r.yes
}

func (r *VotingStage) NOP() map[common.Address]common.Hash {
	r.RLock()
	defer r.RUnlock()

	return r.nop
}

func (r *VotingStage) EXP() map[common.Address]common.Hash {
	r.RLock()
	defer r.RUnlock()

	return r.exp
}

func (r *VotingStage) Voted(source common.Address) (Vote, bool) {
	r.RLock()
	defer r.RUnlock()

	if _, found := r.yes[source]; found {
		return VoteYES, true
	}
	if _, found := r.nop[source]; found {
		return VoteNOP, true
	}
	if _, found := r.exp[source]; found {
		return VoteEXPIRE, true
	}

	return VoteNONE, false
}

func (r *VotingStage) Vote(sHash common.Hash, source common.Address, vote Vote) error {
	var m map[common.Address]common.Hash
	var others []map[common.Address]common.Hash
	switch vote {
	case VoteYES:
		m = r.yes
		others = append(others, r.nop, r.exp)
	case VoteNOP:
		m = r.nop
		others = append(others, r.yes, r.exp)
	case VoteEXPIRE:
		m = r.exp
		others = append(others, r.yes, r.nop)
	default:
		return InvalidVoteError
	}

	r.Lock()
	defer r.Unlock()

	for _, o := range others {
		if _, found := o[source]; !found {
			continue
		}
		delete(o, source)
	}

	m[source] = sHash

	return nil
}

func (r *VotingStage) Count() int {
	return len(r.yes) + len(r.nop) + len(r.exp)
}

func (r *VotingStage) CanCount(total, threshold uint) bool {
	r.RLock()
	defer r.RUnlock()

	return canCountVoting(total, threshold, len(r.yes), len(r.nop), len(r.exp))
}

func (r *VotingStage) Majority(total, threshold uint) VoteResult {
	r.RLock()
	defer r.RUnlock()

	return majority(total, threshold, len(r.yes), len(r.nop), len(r.exp))
}

func canCountVoting(total, threshold uint, yes, nop, exp int) bool {
	if threshold > total {
		return false
	}

	to := int(total)
	count := yes + nop + exp
	if count >= to {
		return true
	}

	th := int(threshold)

	margin := to - th

	// check majority
	if yes >= th || yes > margin {
		return true
	}

	if nop >= th || nop > margin {
		return true
	}

	if exp >= th || exp > margin {
		return true
	}

	// draw
	var voted = []int{yes, nop, exp}
	sort.Ints(voted)

	major := voted[len(voted)-1]

	if major+to-count < th {
		return true
	}

	return false
}

func majority(total, threshold uint, yes, nop, exp int) VoteResult {
	if !canCountVoting(total, threshold, yes, nop, exp) {
		return VoteResultNotYet // not yet
	}

	to := int(total)
	count := yes + nop + exp
	if count > to {
		return VoteResultDRAW
	}

	th := int(threshold)

	if nop >= th {
		return VoteResultNOP
	}

	if yes >= th {
		return VoteResultYES
	}

	if exp >= th {
		return VoteResultEXPIRE
	}

	return VoteResultDRAW
}
