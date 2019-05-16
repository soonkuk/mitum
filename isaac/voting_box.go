package isaac

import (
	"encoding/json"
	"sort"
	"sync"
	"time"

	"github.com/inconshreveable/log15"
	"github.com/spikeekips/mitum/common"
)

type VotingBox interface {
	Open(Proposal) (VoteResultInfo, error)
	Vote(Ballot) (VoteResultInfo, error)
	Close() error
	Clear() error
}

type DefaultVotingBox struct {
	sync.RWMutex
	*common.Logger
	home     *common.HomeNode
	policy   ConsensusPolicy
	current  *VotingBoxProposal
	previous *VotingBoxProposal
	unknown  *VotingBoxUnknown
}

func NewDefaultVotingBox(home *common.HomeNode, policy ConsensusPolicy) *DefaultVotingBox {
	return &DefaultVotingBox{
		home:    home,
		Logger:  common.NewLogger(log, "node", home.Name()),
		policy:  policy,
		unknown: NewVotingBoxUnknown(policy),
	}
}

func (r *DefaultVotingBox) Current() *VotingBoxProposal {
	return r.current
}

func (r *DefaultVotingBox) Previous() *VotingBoxProposal {
	return r.previous
}

func (r *DefaultVotingBox) Unknown() *VotingBoxUnknown {
	return r.unknown
}

// Open starts new DefaultVotingBox for node.
func (r *DefaultVotingBox) Open(proposal Proposal) (VoteResultInfo, error) {
	if r.current != nil {
		if r.current.proposal.Equal(proposal.Hash()) {
			return VoteResultInfo{}, VotingBoxProposalAlreadyStartedError.SetMessage(
				"close running VotingBox first",
			)
		}

		return VoteResultInfo{}, AnotherProposalIsOpenedError
	}

	r.Lock()
	defer r.Unlock()

	r.current = NewVotingBoxProposal(r.policy, proposal.Hash(), proposal.Block.Height, proposal.Round)

	// import from others
	for _, u := range r.unknown.Proposal(proposal.Hash()) {
		_, err := r.current.Vote(u.source, u.stage, u.vote, u.seal)
		if err != nil {
			return VoteResultInfo{}, err
		}
	}

	// NOTE result will be used to broadcast sign ballot
	result := VoteResultInfo{
		Result:      VoteResultYES,
		Proposal:    proposal.Hash(),
		Height:      proposal.Block.Height,
		Round:       proposal.Round,
		Stage:       VoteStageINIT, // NOTE will broadcast sign ballot
		Proposed:    true,
		LastVotedAt: common.Now(),
		Voted: map[common.Address]VotingBoxStageNode{
			proposal.Source(): NewVotingBoxStageNode(VoteYES, proposal.Hash()),
		},
	}

	return result, nil
}

// Close finishes current running proposal; it's proposal reaches to ALLCONFIRM,
func (r *DefaultVotingBox) Close() error {
	r.Lock()
	defer r.Unlock()

	if r.current == nil {
		return ProposalIsNotOpenedError
	}

	if err := r.current.Close(VoteStageACCEPT); err != nil {
		return err
	}

	var previous common.Hash
	if r.previous != nil {
		previous = r.previous.proposal
	}

	r.Log().Debug(
		"current votingbox closed",
		"current", r.current.proposal,
		"previous", previous,
	)

	r.previous = r.current
	r.current = nil

	return nil
}

// Clear clears votes; only leaves previous
func (r *DefaultVotingBox) Clear() error {
	r.Lock()
	defer r.Unlock()

	r.current = nil
	r.unknown = NewVotingBoxUnknown(r.policy)

	return nil
}

func (r *DefaultVotingBox) afterMajority(result VoteResultInfo) error {
	if err := r.current.Close(result.Stage); err != nil {
		return err
	}

	r.unknown.ClearBefore(result.LastVotedAt)

	if result.Stage == VoteStageACCEPT { // automaticall finish consensus
		if err := r.Close(); err != nil {
			return err
		}
	}

	return nil
}

func (r *DefaultVotingBox) Vote(ballot Ballot) (VoteResultInfo, error) {
	result, err := r.vote(
		ballot.Proposal,
		ballot.Source(),
		ballot.Height,
		ballot.Round,
		ballot.Stage,
		ballot.Vote,
		ballot.Hash(),
	)
	if err != nil {
		return result, err
	}

	if result.NotYet() {
		result.Height = ballot.Height
		result.Round = ballot.Round
		result.Stage = ballot.Stage
		result.Proposal = ballot.Proposal

		return result, nil
	}

	return result, nil
}

func (r *DefaultVotingBox) vote(
	proposal common.Hash,
	source common.Address,
	height common.Big,
	round Round,
	stage VoteStage,
	vote Vote,
	seal common.Hash,
) (VoteResultInfo, error) {
	// vote for unknown
	log_ := r.Log().New(log15.Ctx{
		"proposal": proposal,
		"source":   source,
		"height":   height,
		"round":    round,
		"stage":    stage,
		"vote":     vote,
		"seal":     seal,
	})

	if proposal.Empty() || (r.current == nil || !r.current.proposal.Equal(proposal)) &&
		(r.previous == nil || !r.previous.proposal.Equal(proposal)) {
		log_.Debug("trying to vote to unknown")
		return r.voteUnknown(
			proposal,
			source,
			height,
			round,
			stage,
			vote,
			seal,
		)
	}

	log_.Debug("trying to vote to known")
	return r.voteKnown(
		proposal,
		source,
		stage,
		vote,
		seal,
	)
}

func (r *DefaultVotingBox) voteKnown(
	proposal common.Hash,
	source common.Address,
	stage VoteStage,
	vote Vote,
	seal common.Hash,
) (VoteResultInfo, error) {
	if !vote.IsValid() || !vote.CanVote() {
		return VoteResultInfo{}, InvalidVoteError
	}

	if !stage.IsValid() || !stage.CanVote() {
		return VoteResultInfo{}, InvalidVoteStageError
	}

	// vote for current or previous
	var p *VotingBoxProposal
	var isCurrent bool
	if r.current != nil && r.current.proposal.Equal(proposal) {
		p = r.current
		isCurrent = true
	} else if r.previous != nil && r.previous.proposal.Equal(proposal) {
		p = r.previous
	} else {
		return VoteResultInfo{}, InvalidVoteError.SetMessage("unknown proposal in current and previous")
	}

	if p.SealVoted(seal) {
		return VoteResultInfo{}, SealAlreadyVotedError
	}

	vs, err := p.Vote(source, stage, vote, seal)
	if err != nil {
		return VoteResultInfo{}, err
	}

	// NOTE remove last vote from unknown
	r.unknown.Cancel(source)

	var result VoteResultInfo
	if !isCurrent {
		return result, nil
	}

	result = vs.Majority(r.policy.Total, r.policy.Threshold)
	if result.NotYet() {
		return result, nil
	}

	if err := r.afterMajority(result); err != nil {
		return VoteResultInfo{}, err
	}

	return result, nil
}

func (r *DefaultVotingBox) voteUnknown(
	proposal common.Hash,
	source common.Address,
	height common.Big,
	round Round,
	stage VoteStage,
	vote Vote,
	seal common.Hash,
) (VoteResultInfo, error) {
	if !vote.IsValid() || !vote.CanVote() {
		return VoteResultInfo{}, InvalidVoteError
	}

	if !stage.IsValid() || !stage.CanVote() {
		return VoteResultInfo{}, InvalidVoteStageError
	}

	if r.unknown.SealVoted(seal) {
		return VoteResultInfo{}, SealAlreadyVotedError
	}

	_, err := r.unknown.Vote(
		proposal,
		source,
		height,
		round,
		stage,
		vote,
		seal,
	)
	if err != nil {
		return VoteResultInfo{}, err
	}

	// NOTE remove last vote from source
	if r.current != nil {
		for _, vs := range r.current.Opened() {
			vs.Cancel(source)
		}
	}

	result := r.unknown.Majority(r.policy.Total, r.policy.Threshold)

	if result.NotYet() {
		return result, nil
	}

	// NOTE clear unknown
	r.unknown.ClearBefore(result.LastVotedAt)

	// close current
	if r.current != nil {
		if err := r.Close(); err != nil {
			return result, err
		}
	}

	return result, nil
}

func (r *DefaultVotingBox) Voted(source common.Address) (
	/* current */ map[VoteStage]*VotingBoxStage,
	/* unknown */ VotingBoxUnknownVote,
) {
	current := r.current.Voted(source)
	unknown, _ := r.unknown.Voted(source)

	return current, unknown
}

func (r *DefaultVotingBox) SealVoted(seal common.Hash) bool {
	r.RLock()
	defer r.RUnlock()

	if r.current != nil {
		if voted := r.current.SealVoted(seal); voted {
			return true
		}
	}

	if r.previous != nil {
		if voted := r.previous.SealVoted(seal); voted {
			return true
		}
	}

	return r.unknown.SealVoted(seal)
}

func (r *DefaultVotingBox) ProposalVoted(proposal common.Hash) bool {
	r.RLock()
	defer r.RUnlock()

	if r.current != nil {
		if r.current.proposal.Equal(proposal) {
			return true
		}
	}

	if r.previous != nil {
		if r.current.proposal.Equal(proposal) {
			return true
		}
	}

	return r.unknown.ProposalVoted(proposal)
}

type VotingBoxProposal struct {
	policy      ConsensusPolicy
	proposal    common.Hash
	height      common.Big
	round       Round
	stage       VoteStage
	stageSIGN   *VotingBoxStage
	stageACCEPT *VotingBoxStage
}

func NewVotingBoxProposal(
	policy ConsensusPolicy,
	proposal common.Hash,
	height common.Big,
	round Round,
) *VotingBoxProposal {
	return &VotingBoxProposal{
		policy:      policy,
		proposal:    proposal,
		height:      height,
		round:       round,
		stage:       VoteStageINIT,
		stageSIGN:   NewVotingBoxStage(policy, proposal, height, round, VoteStageSIGN),
		stageACCEPT: NewVotingBoxStage(policy, proposal, height, round, VoteStageACCEPT),
	}
}

func (v *VotingBoxProposal) Stage(stage VoteStage) *VotingBoxStage {
	switch stage {
	case VoteStageSIGN:
		return v.stageSIGN
	case VoteStageACCEPT:
		return v.stageACCEPT
	}

	return nil
}

func (v *VotingBoxProposal) Opened() []*VotingBoxStage {
	allStages := []*VotingBoxStage{
		v.stageACCEPT,
		v.stageSIGN,
	}

	var opened []*VotingBoxStage
	for _, stage := range allStages {
		if stage.Closed() {
			break
		}
		opened = append(opened, stage)
	}

	return opened
}

func (v *VotingBoxProposal) Closed() bool {
	return v.stageACCEPT.Closed()
}

func (v *VotingBoxProposal) Close(stage VoteStage) error {
	if !stage.IsValid() || !stage.CanVote() {
		return InvalidVoteStageError
	}

	var vs []*VotingBoxStage
	switch stage {
	case VoteStageSIGN:
		vs = append(vs, v.stageSIGN)
	case VoteStageACCEPT:
		vs = append(vs, v.stageSIGN, v.stageACCEPT)
	default:
		return InvalidVoteStageError
	}

	for _, v := range vs {
		v.Close()
	}

	return nil
}

func (v *VotingBoxProposal) Voted(source common.Address) map[VoteStage]*VotingBoxStage {
	allStages := []*VotingBoxStage{
		v.stageACCEPT,
		v.stageSIGN,
	}

	voted := map[VoteStage]*VotingBoxStage{}
	for _, s := range allStages {
		if _, found := s.Voted(source); !found {
			continue
		}
		voted[s.stage] = s
	}

	return voted
}

func (v *VotingBoxProposal) SealVoted(seal common.Hash) bool {
	if voted := v.stageSIGN.SealVoted(seal); voted {
		return true
	}

	if voted := v.stageACCEPT.SealVoted(seal); voted {
		return true
	}

	return false
}

func (v *VotingBoxProposal) Vote(source common.Address, stage VoteStage, vote Vote, seal common.Hash) (
	*VotingBoxStage,
	error,
) {
	var vs *VotingBoxStage
	switch stage {
	case VoteStageSIGN:
		vs = v.stageSIGN
	case VoteStageACCEPT:
		vs = v.stageACCEPT
	default:
		return nil, InvalidVoteStageError
	}

	return vs.Vote(source, vote, seal)
}

func (v *VotingBoxProposal) String() string {
	b, _ := json.Marshal(map[string]interface{}{
		"proposal":    v.proposal,
		"height":      v.height,
		"round":       v.round,
		"stage":       v.stage,
		"stageSIGN":   v.stageSIGN,
		"stageACCEPT": v.stageACCEPT,
	})

	return common.TerminalLogString(string(b))
}

type VotingBoxStage struct {
	sync.RWMutex
	policy   ConsensusPolicy
	proposal common.Hash
	height   common.Big
	round    Round
	stage    VoteStage
	closed   bool
	voted    map[ /* source */ common.Address]VotingBoxStageNode
}

func NewVotingBoxStage(
	policy ConsensusPolicy,
	proposal common.Hash,
	height common.Big,
	round Round,
	stage VoteStage,
) *VotingBoxStage {
	return &VotingBoxStage{
		policy:   policy,
		proposal: proposal,
		height:   height,
		round:    round,
		stage:    stage,
		voted:    map[common.Address]VotingBoxStageNode{},
	}
}

func (v *VotingBoxStage) Voted(source common.Address) (VotingBoxStageNode, bool) {
	v.RLock()
	defer v.RUnlock()

	sn, found := v.voted[source]
	return sn, found
}

func (v *VotingBoxStage) SealVoted(seal common.Hash) bool {
	v.RLock()
	defer v.RUnlock()

	for _, n := range v.voted {
		if seal.Equal(n.seal) {
			return !n.Expired(v.policy.ExpireDurationVote)
		}
	}

	return false
}

func (v *VotingBoxStage) Closed() bool {
	v.RLock()
	defer v.RUnlock()

	return v.closed
}

func (v *VotingBoxStage) Close() {
	if v.Closed() {
		return
	}

	v.Lock()
	defer v.Unlock()

	v.closed = true
}

func (v *VotingBoxStage) Cancel(source common.Address) bool {
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

func (v *VotingBoxStage) Vote(source common.Address, vote Vote, seal common.Hash) (*VotingBoxStage, error) {
	v.Lock()
	defer v.Unlock()

	v.voted[source] = NewVotingBoxStageNode(vote, seal)
	return v, nil
}

func (v *VotingBoxStage) Count() int {
	return len(v.voted)
}

func (v *VotingBoxStage) VoteCount() (int, int) {
	var yes, nop int
	for _, t := range v.voted {
		if t.Expired(v.policy.ExpireDurationVote) {
			continue
		}

		switch t.vote {
		case VoteYES:
			yes += 1
		case VoteNOP:
			nop += 1
		}
	}

	return yes, nop
}

func (v *VotingBoxStage) CanCount(total, threshold uint) bool {
	yes, nop := v.VoteCount()

	return canCountVoting(total, threshold, yes, nop)
}

func (v *VotingBoxStage) Majority(total, threshold uint) VoteResultInfo {
	if v.closed {
		return NewVoteResultInfo()
	}

	yes, nop := v.VoteCount()

	result := majority(total, threshold, yes, nop)

	r := NewVoteResultInfo()
	r.Result = result

	if result == VoteResultNotYet {
		return r
	}

	r.Proposal = v.proposal
	r.Height = v.height
	r.Round = v.round
	r.Stage = v.stage

	votedMap := map[common.Address]VotingBoxStageNode{}
	for k, v := range v.voted {
		votedMap[k] = v
	}

	r.Voted = votedMap

	return r
}

func (v *VotingBoxStage) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"proposal": v.proposal,
		"height":   v.height,
		"round":    v.round,
		"stage":    v.stage,
		"closed":   v.closed,
		"voted":    v.voted,
	})
}

func (v *VotingBoxStage) String() string {
	b, _ := json.Marshal(v)
	return common.TerminalLogString(string(b))
}

type VotingBoxStageNode struct {
	vote    Vote
	seal    common.Hash
	votedAt common.Time
}

func NewVotingBoxStageNode(vote Vote, hash common.Hash) VotingBoxStageNode {
	return VotingBoxStageNode{
		vote:    vote,
		seal:    hash,
		votedAt: common.Now(),
	}
}

func (v VotingBoxStageNode) Expired(d time.Duration) bool {
	if d == 0 {
		return false
	}

	if d > 0 {
		d = d * -1
	}

	return v.votedAt.Before(common.Now().Add(d))
}

func (v VotingBoxStageNode) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"vote": v.vote,
		"seal": v.seal,
	})
}

func (v VotingBoxStageNode) String() string {
	b, _ := json.Marshal(v)
	return common.TerminalLogString(string(b))
}

type VotingBoxUnknown struct {
	sync.RWMutex
	policy ConsensusPolicy
	voted  map[ /* source */ common.Address]VotingBoxUnknownVote
}

func NewVotingBoxUnknown(policy ConsensusPolicy) *VotingBoxUnknown {
	return &VotingBoxUnknown{
		policy: policy,
		voted:  map[common.Address]VotingBoxUnknownVote{},
	}
}

// ClearBefore can clear voted by VoteResultInfo.LastVotedAt.
func (v *VotingBoxUnknown) ClearBefore(t common.Time) {
	v.Lock()
	defer v.Unlock()

	for source, u := range v.voted {
		if u.votedAt.Before(t) {
			delete(v.voted, source)
		}
	}
}

func (v *VotingBoxUnknown) Len() int {
	return len(v.voted)
}

func (v *VotingBoxUnknown) Vote(
	proposal common.Hash,
	source common.Address,
	height common.Big,
	round Round,
	stage VoteStage,
	vote Vote,
	seal common.Hash,
) (*VotingBoxUnknown, error) {
	u, err := NewVotingBoxUnknownVote(proposal, source, height, round, stage, vote, seal)
	if err != nil {
		return v, err
	}

	v.Lock()
	defer v.Unlock()

	v.voted[source] = u

	return v, nil
}

func (v *VotingBoxUnknown) Cancel(source common.Address) bool {
	_, found := v.voted[source]
	if !found {
		return false
	}

	v.Lock()
	defer v.Unlock()

	delete(v.voted, source)

	return true
}

func (v *VotingBoxUnknown) SealVoted(seal common.Hash) bool {
	for _, n := range v.voted {
		if seal.Equal(n.seal) {
			return !n.Expired(v.policy.ExpireDurationVote)
		}
	}

	return false
}

func (v *VotingBoxUnknown) ProposalVoted(proposal common.Hash) bool {
	for _, n := range v.voted {
		if proposal.Equal(n.proposal) {
			return true
		}
	}

	return false
}

func (v *VotingBoxUnknown) Voted(source common.Address) (VotingBoxUnknownVote, bool) {
	vu, found := v.voted[source]
	return vu, found
}

func (v *VotingBoxUnknown) Proposal(proposal common.Hash) []VotingBoxUnknownVote {
	var us []VotingBoxUnknownVote
	for _, u := range v.voted {
		if !u.proposal.Equal(proposal) {
			continue
		}
		us = append(us, u)
	}

	return us
}

func (v *VotingBoxUnknown) Height(height common.Big) []VotingBoxUnknownVote {
	var us []VotingBoxUnknownVote
	for _, u := range v.voted {
		if !u.height.Equal(height) {
			continue
		}
		us = append(us, u)
	}

	return us
}

func (v *VotingBoxUnknown) CanCount(total, threshold uint) bool {
	if total < threshold {
		return false
	}

	v.RLock()
	defer v.RUnlock()

	return len(v.voted) >= int(threshold)
}

func (v *VotingBoxUnknown) Majority(total, threshold uint) VoteResultInfo {
	if !v.CanCount(total, threshold) {
		return NewVoteResultInfo()
	}

	r := v.MajorityProposal(total, threshold)
	if !r.NotYet() {
		return r
	}

	return v.MajorityINIT(total, threshold)
}

func (v *VotingBoxUnknown) MajorityProposal(total, threshold uint) VoteResultInfo {
	th := int(threshold)

	// within same proposal and same stage
	v.RLock()
	byProposal := map[common.Hash][]VotingBoxUnknownVote{}
	for _, u := range v.voted {
		if u.Expired(v.policy.ExpireDurationVote) {
			continue
		}

		byProposal[u.proposal] = append(byProposal[u.proposal], u)
	}
	v.RUnlock()

	var found []VotingBoxUnknownVote
	for _, l := range byProposal {
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
	byStage := map[VoteStage][]VotingBoxUnknownVote{}
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
	var yes, nop int
	for _, u := range found {
		switch u.vote {
		case VoteYES:
			yes += 1
		case VoteNOP:
			nop += 1
		}
	}

	sort.Slice(found, func(i, j int) bool {
		return found[i].votedAt.After(found[j].votedAt)
	})

	result := majority(total, threshold, yes, nop)

	r := NewVoteResultInfo()
	r.Result = result

	if result == VoteResultNotYet {
		return r
	}

	var voted []VotingBoxUnknownVote
	for _, l := range byProposal {
		voted = append(voted, l...)
	}

	sort.Slice(voted, func(i, j int) bool {
		return voted[i].votedAt.After(voted[j].votedAt)
	})

	r.LastVotedAt = voted[0].votedAt

	r.Proposal = found[0].proposal
	r.Height = found[0].height
	r.Round = found[0].round
	r.Stage = found[0].stage

	votedMap := map[common.Address]VotingBoxStageNode{}
	for k, v := range v.voted {
		votedMap[k] = v.VotingBoxStageNode
	}

	r.Voted = votedMap

	return r
}

// MajorityINIT checks INIT stage votes which have same height and same round
func (v *VotingBoxUnknown) MajorityINIT(total, threshold uint) VoteResultInfo {
	v.RLock()
	byRound := map[Round][]VotingBoxUnknownVote{}
	for _, u := range v.voted {
		if u.stage != VoteStageINIT {
			continue
		}

		if u.Expired(v.policy.ExpireDurationVote) {
			continue
		}

		byRound[u.round] = append(byRound[u.round], u)
	}
	v.RUnlock()

	if len(byRound) < 1 {
		return NewVoteResultInfo()
	}

	th := int(threshold)

	var found []VotingBoxUnknownVote
	for _, l := range byRound {
		if len(l) < th {
			continue
		}

		found = l
		break
	}

	result := majority(total, threshold, len(found), 0)
	if result == VoteResultNotYet {
		return NewVoteResultInfo()
	}

	r := NewVoteResultInfo()
	r.Result = VoteResultYES
	r.Height = found[0].height
	r.Round = found[0].round
	r.Stage = VoteStageINIT

	votedMap := map[common.Address]VotingBoxStageNode{}
	for k, v := range v.voted {
		votedMap[k] = v.VotingBoxStageNode
	}

	r.Voted = votedMap

	return r
}

func (v *VotingBoxUnknown) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.voted)
}

func (v *VotingBoxUnknown) String() string {
	b, _ := json.Marshal(v)
	return common.TerminalLogString(string(b))
}

type VotingBoxUnknownVote struct {
	VotingBoxStageNode
	source   common.Address
	proposal common.Hash
	height   common.Big
	round    Round
	stage    VoteStage
}

func NewVotingBoxUnknownVote(
	proposal common.Hash,
	source common.Address,
	height common.Big,
	round Round,
	stage VoteStage,
	vote Vote,
	seal common.Hash,
) (VotingBoxUnknownVote, error) {
	return VotingBoxUnknownVote{
		VotingBoxStageNode: NewVotingBoxStageNode(vote, seal),
		proposal:           proposal,
		source:             source,
		height:             height,
		round:              round,
		stage:              stage,
	}, nil
}

func (v VotingBoxUnknownVote) Empty() bool {
	return !v.proposal.IsValid()
}

func (v VotingBoxUnknownVote) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"source":   v.source,
		"proposal": v.proposal,
		"height":   v.height,
		"round":    v.round,
		"stage":    v.stage,
		"vote":     v.vote,
		"seal":     v.seal,
		"voted_at": v.votedAt,
	})
}

func (v VotingBoxUnknownVote) String() string {
	b, _ := json.Marshal(v)
	return common.TerminalLogString(string(b))
}
