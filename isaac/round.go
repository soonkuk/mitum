package isaac

import (
	"encoding/json"
	"sort"
	"sync"

	"github.com/spikeekips/mitum/common"
)

type Round uint64

type RoundVotingManager struct {
	sync.RWMutex
	proposals map[common.Hash]*VotingProposal
}

func NewRoundVotingManager() *RoundVotingManager {
	return &RoundVotingManager{
		proposals: map[common.Hash]*VotingProposal{},
	}
}

func (r *RoundVotingManager) NewRound(seal common.Seal) (*VotingProposal, error) {
	if seal.Type != ProposeBallotSealType {
		return nil, InvalidSealTypeError
	}

	proposeBallotSealHash, _, err := seal.Hash()
	if err != nil {
		return nil, err
	}

	var proposeBallot ProposeBallot
	if err := seal.UnmarshalBody(&proposeBallot); err != nil {
		return nil, err
	}

	if r.IsRunning(proposeBallotSealHash) {
		return nil, VotingProposalAlreadyStartedError
	}

	r.Lock()
	defer r.Unlock()

	vp := NewVotingProposal(proposeBallot.Block.Height)
	r.proposals[proposeBallotSealHash] = vp

	// seal.Source automatically votes for SIGN
	vr, err := vp.NewRound(0)
	if err != nil {
		return nil, err
	}

	vr.Stage(VoteStageSIGN).Vote(
		proposeBallotSealHash,
		seal.Source,
		VoteYES,
	)

	return vp, nil
}

func (r *RoundVotingManager) IsRunning(proposeBallotSealHash common.Hash) bool {
	r.RLock()
	defer r.RUnlock()

	_, found := r.proposals[proposeBallotSealHash]
	return found
}

func (r *RoundVotingManager) FinishRound(proposeBallotSealHash common.Hash) error {
	if !r.IsRunning(proposeBallotSealHash) {
		return nil
	}

	r.Lock()
	defer r.Unlock()

	current := r.proposals[proposeBallotSealHash]

	removeHashes := []common.Hash{proposeBallotSealHash}
	for ph, vrs := range r.proposals {
		if vrs.height.Cmp(current.height) < 1 { // same or lower
			removeHashes = append(removeHashes, ph)
		}
	}

	for _, h := range removeHashes {
		delete(r.proposals, h)
	}

	return nil
}

func (r *RoundVotingManager) Round(proposeBallotSealHash common.Hash, round Round) *VotingRound {
	r.RLock()
	defer r.RUnlock()

	vrs, found := r.proposals[proposeBallotSealHash]
	if !found {
		return nil
	}

	return vrs.Round(round)
}

type VotingProposal struct {
	sync.RWMutex
	height common.Big // block height
	rounds map[Round]*VotingRound
}

func NewVotingProposal(height common.Big) *VotingProposal {
	return &VotingProposal{
		height: height,
		rounds: map[Round]*VotingRound{},
	}
}

func (vrs *VotingProposal) NewRound(round Round) (*VotingRound, error) {
	if vrs.IsRunning(round) {
		return nil, VotingRoundAlreadyStartedError
	}

	vrs.Lock()
	defer vrs.Unlock()

	vr := NewVotingRound()
	vrs.rounds[round] = vr

	return vr, nil
}

func (vrs *VotingProposal) IsRunning(round Round) bool {
	vrs.RLock()
	defer vrs.RUnlock()

	_, found := vrs.rounds[round]
	return found
}

func (vrs *VotingProposal) Round(round Round) *VotingRound {
	vrs.RLock()
	defer vrs.RUnlock()

	vr, found := vrs.rounds[round]
	if !found {
		return nil
	}

	return vr
}

type VotingRound struct {
	StageINIT   *VotingStage
	StageSIGN   *VotingStage
	StageACCEPT *VotingStage
}

func NewVotingRound() *VotingRound {
	return &VotingRound{
		StageINIT:   NewVotingStage(),
		StageSIGN:   NewVotingStage(),
		StageACCEPT: NewVotingStage(),
	}
}

func (r *VotingRound) Stage(stage VoteStage) *VotingStage {
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

	b, _ := json.Marshal(map[string][]interface{}{
		"yes": []interface{}{len(yes), yes},
		"nop": []interface{}{len(nop), nop},
		"exp": []interface{}{len(exp), exp},
	})

	return string(b)
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

func (r *VotingStage) CanCount(total, threshold int) bool {
	// TODO implement
	r.RLock()
	defer r.RUnlock()

	margin := total - threshold

	// check majority
	yes := len(r.yes)
	if yes >= threshold || yes > margin {
		return true
	}

	nop := len(r.nop)
	if nop >= threshold || nop > margin {
		return true
	}

	exp := len(r.exp)
	if exp >= threshold || exp > margin {
		return true
	}

	// draw
	count := r.Count()
	if count == total {
		return true
	}

	var voted = []int{yes, nop, exp}
	sort.Ints(voted)

	if voted[len(voted)-1]+total-count < threshold {
		return true
	}

	return false
}

func (r *VotingStage) Majority(total, threshold int) VoteResult {
	if !r.CanCount(total, threshold) {
		return VoteResultNotYet // not yet
	}

	if len(r.yes) >= threshold {
		return VoteResultYES
	}

	if len(r.nop) >= threshold {
		return VoteResultNOP
	}

	if len(r.exp) >= threshold {
		return VoteResultEXPIRE
	}

	return VoteResultDRAW
}
