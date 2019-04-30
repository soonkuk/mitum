package isaac

import (
	"encoding/json"
	"sort"
	"sync"

	"github.com/spikeekips/mitum/common"
)

type RoundVoting struct {
	sync.RWMutex
	current  *VotingProposal
	previous *VotingProposal
	unknown  *VotingUnknown
}

func NewRoundVoting() *RoundVoting {
	return &RoundVoting{
		unknown: NewVotingUnknown(),
	}
}

func (r *RoundVoting) Current() *VotingProposal {
	return r.current
}

func (r *RoundVoting) Previous() *VotingProposal {
	return r.previous
}

func (r *RoundVoting) Unknown() *VotingUnknown {
	return r.unknown
}

// Open starts new RoundVoting for node.
func (r *RoundVoting) Open(seal common.Seal) (*VotingProposal, error) {
	if seal.Type != ProposeSealType {
		return nil, InvalidSealTypeError
	}

	psHash, _, err := seal.Hash()
	if err != nil {
		return nil, err
	}

	if r.current != nil {
		if r.current.psHash == psHash {
			return r.current, nil
		}

		return nil, AnotherProposalIsOpenedError
	}

	var propose Propose
	if err := seal.UnmarshalBody(&propose); err != nil {
		return nil, err
	}

	r.Lock()
	defer r.Unlock()

	r.current = NewVotingProposal(psHash, propose.Block.Height, propose.Round)

	_, err = r.VoteManually(
		psHash,
		propose.Proposer,
		propose.Block.Height,
		propose.Round,
		VoteStageSIGN,
		VoteYES,
		psHash,
	)
	if err != nil {
		return nil, err
	}

	// import from others
	for _, u := range r.unknown.PSHash(psHash) {
		_, err := r.current.Vote(u.source, u.stage, u.vote, u.seal)
		if err != nil {
			return nil, err
		}
	}

	return r.current, nil
}

// Close finishes current running proposal; it's proposal reaches to ALLCONFIRM,
func (r *RoundVoting) Close() error {
	r.Lock()
	defer r.Unlock()

	if r.current == nil {
		return ProposalIsNotOpenedError
	}

	r.current.Close(VoteStageACCEPT)
	r.previous = r.current

	r.current = nil

	return nil
}

func (r *RoundVoting) VoteManually(
	psHash common.Hash,
	source common.Address,
	height common.Big,
	round Round,
	stage VoteStage,
	vote Vote,
	sHash common.Hash,
) (Majoritier, error) {
	if r.current == nil {
		return nil, ProposalIsNotOpenedError
	}

	if !vote.IsValid() || !vote.CanVote() {
		return nil, InvalidVoteError
	}

	if !stage.IsValid() || !stage.CanVote() {
		return nil, InvalidVoteStageError
	}

	if !r.current.psHash.Equal(psHash) && (r.previous == nil || !r.previous.psHash.Equal(psHash)) {
		_, err := r.unknown.Vote(
			psHash,
			source,
			height,
			round,
			stage,
			vote,
			sHash,
		)
		if err != nil {
			return nil, err
		}

		// NOTE remove last vote from source
		if r.current != nil {
			for _, vs := range r.current.Opened() {
				vs.Cancel(source)
			}
		}

		return r.unknown, nil
	}

	var p *VotingProposal
	if r.current.psHash.Equal(psHash) {
		p = r.current
	} else {
		p = r.previous
	}

	vs, err := p.Vote(source, stage, vote, sHash)
	if err != nil {
		return nil, err
	}

	// NOTE remove last vote from unknown
	r.unknown.Cancel(source)

	return vs, nil
}

func (r *RoundVoting) Vote(seal common.Seal) (Majoritier, error) {
	if seal.Type != BallotSealType {
		return nil, InvalidSealTypeError
	}

	sHash, _, err := seal.Hash()
	if err != nil {
		return nil, err
	}

	var ballot Ballot
	if err := seal.UnmarshalBody(&ballot); err != nil {
		return nil, err
	}

	m, err := r.VoteManually(
		ballot.ProposeSeal,
		ballot.Source,
		ballot.Height,
		ballot.Round,
		ballot.Stage,
		ballot.Vote,
		sHash,
	)
	if err != nil {
		return nil, err
	}

	return m, nil
}

func (r *RoundVoting) Voted(source common.Address) (
	/* current */ map[VoteStage]*VotingStage,
	/* unknown */ VotingUnknownVote,
) {
	current := r.current.Voted(source)
	unknown, _ := r.unknown.Voted(source)

	return current, unknown
}

type VotingProposal struct {
	psHash      common.Hash
	height      common.Big
	round       Round
	stage       VoteStage
	stageINIT   *VotingStage
	stageSIGN   *VotingStage
	stageACCEPT *VotingStage
}

func NewVotingProposal(psHash common.Hash, height common.Big, round Round) *VotingProposal {
	return &VotingProposal{
		psHash:      psHash,
		height:      height,
		round:       round,
		stage:       VoteStageINIT,
		stageINIT:   NewVotingStage(psHash, height, round, VoteStageINIT),
		stageSIGN:   NewVotingStage(psHash, height, round, VoteStageSIGN),
		stageACCEPT: NewVotingStage(psHash, height, round, VoteStageACCEPT),
	}
}

func (v *VotingProposal) Stage(stage VoteStage) *VotingStage {
	switch stage {
	case VoteStageINIT:
		return v.stageINIT
	case VoteStageSIGN:
		return v.stageSIGN
	case VoteStageACCEPT:
		return v.stageACCEPT
	}

	return nil
}

func (v *VotingProposal) Opened() []*VotingStage {
	allStages := []*VotingStage{
		v.stageACCEPT,
		v.stageSIGN,
		v.stageINIT,
	}

	var opened []*VotingStage
	for _, stage := range allStages {
		if stage.Closed() {
			break
		}
		opened = append(opened, stage)
	}

	return opened
}

func (v *VotingProposal) Closed() bool {
	return v.stageACCEPT.Closed()
}

func (v *VotingProposal) Close(stage VoteStage) error {
	if !stage.IsValid() || !stage.CanVote() {
		return InvalidVoteStageError
	}

	var vs []*VotingStage
	switch stage {
	case VoteStageINIT:
		vs = append(vs, v.stageINIT)
	case VoteStageSIGN:
		vs = append(vs, v.stageINIT, v.stageSIGN)
	case VoteStageACCEPT:
		vs = append(vs, v.stageINIT, v.stageSIGN, v.stageACCEPT)
	default:
		return InvalidVoteStageError
	}

	for _, v := range vs {
		v.Close()
	}

	return nil
}

func (v *VotingProposal) Voted(source common.Address) map[VoteStage]*VotingStage {
	allStages := []*VotingStage{
		v.stageACCEPT,
		v.stageSIGN,
		v.stageINIT,
	}

	voted := map[VoteStage]*VotingStage{}
	for _, s := range allStages {
		if _, found := s.Voted(source); !found {
			continue
		}
		voted[s.stage] = s
	}

	return voted
}

func (v *VotingProposal) Vote(source common.Address, stage VoteStage, vote Vote, sHash common.Hash) (
	*VotingStage,
	error,
) {
	var vs *VotingStage
	switch stage {
	case VoteStageINIT:
		vs = v.stageINIT
	case VoteStageSIGN:
		vs = v.stageSIGN
	case VoteStageACCEPT:
		vs = v.stageACCEPT
	default:
		return nil, InvalidVoteStageError
	}

	return vs.Vote(source, vote, sHash)
}

type VotingStage struct {
	sync.RWMutex
	psHash common.Hash
	height common.Big
	round  Round
	stage  VoteStage
	closed bool
	voted  map[ /* source */ common.Address]VotingStageNode
}

func NewVotingStage(psHash common.Hash, height common.Big, round Round, stage VoteStage) *VotingStage {
	return &VotingStage{
		psHash: psHash,
		height: height,
		round:  round,
		stage:  stage,
		voted:  map[common.Address]VotingStageNode{},
	}
}

func (v *VotingStage) Voted(source common.Address) (VotingStageNode, bool) {
	v.RLock()
	defer v.RUnlock()

	sn, found := v.voted[source]
	return sn, found
}

func (v *VotingStage) Closed() bool {
	v.RLock()
	defer v.RUnlock()

	return v.closed
}

func (v *VotingStage) Close() {
	if v.Closed() {
		return
	}

	v.Lock()
	defer v.Unlock()

	v.closed = true
}

func (v *VotingStage) Cancel(source common.Address) bool {
	if v.Closed() {
		return false
	}

	v.Lock()
	defer v.Unlock()

	if _, found := v.voted[source]; !found {
		return false
	}

	delete(v.voted, source)

	return true
}

func (v *VotingStage) Vote(source common.Address, vote Vote, sHash common.Hash) (*VotingStage, error) {
	v.Lock()
	defer v.Unlock()

	v.voted[source] = VotingStageNode{vote: vote, seal: sHash}
	return v, nil
}

func (v *VotingStage) Count() int {
	return len(v.voted)
}

func (v *VotingStage) VoteCount() (int, int, int) {
	var yes, nop, exp int
	for _, t := range v.voted {
		switch t.vote {
		case VoteYES:
			yes += 1
		case VoteNOP:
			nop += 1
		case VoteEXPIRE:
			exp += 1
		}
	}

	return yes, nop, exp
}

func (v *VotingStage) CanCount(total, threshold uint) bool {
	yes, nop, exp := v.VoteCount()

	return canCountVoting(total, threshold, yes, nop, exp)
}

func (v *VotingStage) Majority(total, threshold uint) VoteResultInfo {
	if v.closed {
		return NewVoteResultInfo()
	}

	yes, nop, exp := v.VoteCount()

	result := majority(total, threshold, yes, nop, exp)
	r := NewVoteResultInfo()
	r.Result = result

	if result == VoteResultNotYet {
		return r
	}

	r.Proposal = v.psHash
	r.Height = v.height
	r.Round = v.round
	r.Stage = v.stage

	return r
}

type VotingStageNode struct {
	vote Vote
	seal common.Hash
}

type VotingUnknown struct {
	sync.RWMutex
	voted map[ /* source */ common.Address]VotingUnknownVote
}

func NewVotingUnknown() *VotingUnknown {
	return &VotingUnknown{
		voted: map[common.Address]VotingUnknownVote{},
	}
}

// ClearBefore can clear voted by VoteResultInfo.LastVotedAt.
func (v *VotingUnknown) ClearBefore(t common.Time) {
	v.Lock()
	defer v.Unlock()

	for source, u := range v.voted {
		if u.votedAt.Before(t) {
			delete(v.voted, source)
		}
	}
}

func (v *VotingUnknown) Len() int {
	return len(v.voted)
}

func (v *VotingUnknown) Vote(
	psHash common.Hash,
	source common.Address,
	height common.Big,
	round Round,
	stage VoteStage,
	vote Vote,
	sHash common.Hash,
) (*VotingUnknown, error) {
	u, err := NewVotingUnknownVote(psHash, source, height, round, stage, vote, sHash)
	if err != nil {
		return v, err
	}

	v.Lock()
	defer v.Unlock()

	v.voted[source] = u

	return v, nil
}

func (v *VotingUnknown) Cancel(source common.Address) bool {
	_, found := v.voted[source]
	if !found {
		return false
	}

	v.Lock()
	defer v.Unlock()

	delete(v.voted, source)

	return true
}

func (v *VotingUnknown) Voted(source common.Address) (VotingUnknownVote, bool) {
	vu, found := v.voted[source]
	return vu, found
}

func (v *VotingUnknown) PSHash(psHash common.Hash) []VotingUnknownVote {
	var us []VotingUnknownVote
	for _, u := range v.voted {
		if !u.psHash.Equal(psHash) {
			continue
		}
		us = append(us, u)
	}

	return us
}

func (v *VotingUnknown) Height(height common.Big) []VotingUnknownVote {
	var us []VotingUnknownVote
	for _, u := range v.voted {
		if !u.height.Equal(height) {
			continue
		}
		us = append(us, u)
	}

	return us
}

func (v *VotingUnknown) CanCount(total, threshold uint) bool {
	if total < threshold {
		return false
	}

	v.RLock()
	defer v.RUnlock()

	if len(v.voted) < int(threshold) {
		return false
	}

	return true
}

func (v *VotingUnknown) Majority(total, threshold uint) VoteResultInfo {
	if !v.CanCount(total, threshold) {
		return NewVoteResultInfo()
	}

	r := v.MajorityPSHash(total, threshold)
	if !r.NotYet() {
		return r
	}

	return v.MajorityINIT(total, threshold)
}

func (v *VotingUnknown) MajorityPSHash(total, threshold uint) VoteResultInfo {
	th := int(threshold)

	// within same psHash and same stage
	v.RLock()
	byPSHash := map[common.Hash][]VotingUnknownVote{}
	for _, u := range v.voted {
		byPSHash[u.psHash] = append(byPSHash[u.psHash], u)
	}
	v.RUnlock()

	var found []VotingUnknownVote
	for _, l := range byPSHash {
		if len(l) < th {
			continue
		}
		found = l
		break
	}

	if len(found) < th {
		return NewVoteResultInfo()
	}

	// check stage
	byStage := map[VoteStage][]VotingUnknownVote{}
	for _, u := range found {
		byStage[u.stage] = append(byStage[u.stage], u)
	}

	for _, sl := range byStage {
		if len(sl) < th {
			continue
		}

		found = sl
	}

	if len(found) < 1 {
		return NewVoteResultInfo()
	}

	// collect Vote
	var yes, nop, exp int
	for _, u := range found {
		switch u.vote {
		case VoteYES:
			yes += 1
		case VoteNOP:
			nop += 1
		case VoteEXPIRE:
			exp += 1
		}
	}

	sort.Slice(found, func(i, j int) bool {
		return found[i].votedAt.After(found[j].votedAt)
	})

	result := majority(total, threshold, yes, nop, exp)

	r := NewVoteResultInfo()
	r.Result = result

	if result == VoteResultNotYet {
		return r
	}

	var voted []VotingUnknownVote
	for _, l := range byPSHash {
		for _, u := range l {
			voted = append(voted, u)
		}
	}

	sort.Slice(voted, func(i, j int) bool {
		return voted[i].votedAt.After(voted[j].votedAt)
	})

	r.LastVotedAt = voted[0].votedAt

	r.Proposal = found[0].psHash
	r.Height = found[0].height
	r.Round = found[0].round
	r.Stage = found[0].stage

	return r
}

// MajorityINIT checks INIT stage votes which have same height and same round
func (v *VotingUnknown) MajorityINIT(total, threshold uint) VoteResultInfo {
	v.RLock()
	byRound := map[Round][]VotingUnknownVote{}
	for _, u := range v.voted {
		if u.stage != VoteStageINIT {
			continue
		}

		byRound[u.round] = append(byRound[u.round], u)
	}
	v.RUnlock()

	if len(byRound) < 1 {
		return NewVoteResultInfo()
	}

	th := int(threshold)

	var found []VotingUnknownVote
	for _, l := range byRound {
		if len(l) < th {
			continue
		}
		found = l
		break
	}

	result := majority(total, threshold, len(found), 0, 0)
	if result == VoteResultNotYet {
		return NewVoteResultInfo()
	}

	r := NewVoteResultInfo()
	r.Result = VoteResultYES
	r.Height = found[0].height
	r.Round = found[0].round
	r.Stage = VoteStageINIT

	return r
}

func (v *VotingUnknown) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.voted)
}

func (v *VotingUnknown) String() string {
	b, _ := json.Marshal(v)
	return common.TerminalLogString(string(b))
}

type VotingUnknownVote struct {
	source  common.Address
	psHash  common.Hash
	height  common.Big
	round   Round
	stage   VoteStage
	vote    Vote
	seal    common.Hash
	votedAt common.Time
}

func NewVotingUnknownVote(
	psHash common.Hash,
	source common.Address,
	height common.Big,
	round Round,
	stage VoteStage,
	vote Vote,
	sHash common.Hash,
) (VotingUnknownVote, error) {
	return VotingUnknownVote{
		psHash:  psHash,
		source:  source,
		height:  height,
		round:   round,
		stage:   stage,
		vote:    vote,
		seal:    sHash,
		votedAt: common.Now(),
	}, nil
}

func (v VotingUnknownVote) Empty() bool {
	return v.psHash.Empty()
}

func (v VotingUnknownVote) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"source":   v.source,
		"psHash":   v.psHash,
		"height":   v.height,
		"round":    v.round,
		"stage":    v.stage,
		"vote":     v.vote,
		"seal":     v.seal,
		"voted_at": v.votedAt,
	})
}

func (v VotingUnknownVote) String() string {
	b, _ := json.Marshal(v)
	return common.TerminalLogString(string(b))
}
