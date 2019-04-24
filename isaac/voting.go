package isaac

import (
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

// NewRound will only get the ProposeBallotSealType
func (r *RoundVoting) Open(seal common.Seal) (*VotingProposal, *VotingStage, error) {
	if seal.Type != ProposeBallotSealType {
		return nil, nil, InvalidSealTypeError
	}

	proposeBallotSealHash, _, err := seal.Hash()
	if err != nil {
		return nil, nil, err
	}

	if r.IsRunning(proposeBallotSealHash) {
		return nil, nil, VotingProposalAlreadyStartedError
	}

	var proposeBallot ProposeBallot
	if err := seal.UnmarshalBody(&proposeBallot); err != nil {
		return nil, nil, err
	}

	vp := NewVotingProposal(proposeBallot.Block.Height)

	r.Lock()
	defer r.Unlock()

	r.proposals[proposeBallotSealHash] = vp

	vs := vp.Stage(VoteStageSIGN)
	err = vs.Vote(
		proposeBallotSealHash,
		seal.Source,
		VoteYES,
	)
	if err != nil {
		return nil, nil, err
	}

	return vp, vs, nil
}

func (r *RoundVoting) IsRunning(proposeBallotSealHash common.Hash) bool {
	r.RLock()
	defer r.RUnlock()

	_, found := r.proposals[proposeBallotSealHash]
	return found
}

func (r *RoundVoting) Finish(proposeBallotSealHash common.Hash) error {
	if !r.IsRunning(proposeBallotSealHash) {
		return nil
	}

	r.Lock()
	defer r.Unlock()

	currentHeight := r.proposals[proposeBallotSealHash].height

	removeHashes := []common.Hash{proposeBallotSealHash}
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

func (r *RoundVoting) Proposal(proposeBallotSealHash common.Hash) *VotingProposal {
	r.RLock()
	defer r.RUnlock()

	vp, found := r.proposals[proposeBallotSealHash]
	if !found {
		return nil
	}

	return vp
}

func (r *RoundVoting) Vote(voteBallot VoteBallot) (*VotingProposal, *VotingStage, error) {
	if !r.IsRunning(voteBallot.ProposeBallotSeal) {
		return nil, nil, VotingProposalNotFoundError
	}

	vp := r.Proposal(voteBallot.ProposeBallotSeal)
	if vp == nil {
		return nil, nil, VotingProposalNotFoundError
	}

	stage := vp.Stage(voteBallot.Stage)
	if stage == nil {
		return nil, nil, InvalidVoteStageError
	}

	err := stage.Vote(
		voteBallot.ProposeBallotSeal,
		voteBallot.Source,
		voteBallot.Vote,
	)
	if err != nil {
		return nil, nil, err
	}

	return vp, stage, nil
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
	b := map[string]interface{}{
		"height":                 r.height,
		VoteStageINIT.String():   r.StageINIT,
		VoteStageSIGN.String():   r.StageSIGN,
		VoteStageACCEPT.String(): r.StageACCEPT,
	}

	return common.PrettyMap(b)
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

	b := map[string]interface{}{
		"yes": []interface{}{len(yes), yes},
		"nop": []interface{}{len(nop), nop},
		"exp": []interface{}{len(exp), exp},
	}

	return common.PrettyMap(b)
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

func (r *VotingStage) Vote(sealHash common.Hash, source common.Address, vote Vote) error {
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

	m[source] = sealHash

	return nil
}

func (r *VotingStage) Count() int {
	return len(r.yes) + len(r.nop) + len(r.exp)
}

func (r *VotingStage) CanCount(total, threshold uint) bool {
	r.RLock()
	defer r.RUnlock()

	to := int(total)
	th := int(threshold)

	margin := to - th

	// check majority
	yes := len(r.yes)
	if yes >= th || yes > margin {
		return true
	}

	nop := len(r.nop)
	if nop >= th || nop > margin {
		return true
	}

	exp := len(r.exp)
	if exp >= th || exp > margin {
		return true
	}

	// draw
	count := r.Count()
	if count == to {
		return true
	}

	var voted = []int{yes, nop, exp}
	sort.Ints(voted)

	if voted[len(voted)-1]+to-count < th {
		return true
	}

	return false
}

func (r *VotingStage) Majority(total, threshold uint) VoteResult {
	if !r.CanCount(total, threshold) {
		return VoteResultNotYet // not yet
	}

	th := int(threshold)

	if len(r.yes) >= th {
		return VoteResultYES
	}

	if len(r.nop) >= th {
		return VoteResultNOP
	}

	if len(r.exp) >= th {
		return VoteResultEXPIRE
	}

	return VoteResultDRAW
}
