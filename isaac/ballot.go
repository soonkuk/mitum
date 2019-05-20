package isaac

import (
	"github.com/spikeekips/mitum/common"
)

var CurrentBallotVersion common.Version = common.MustParseVersion("0.1.0-proto")

type Ballot interface {
	Version() common.Version
	Hash() common.Hash
	SerializeRLP() ([]interface{}, error)
	MarshalBinary() ([]byte, error)
	String() string
	Type() common.SealType
	Source() common.Address
	SignedAt() common.Time
	Wellformed() error
	SealVersion() common.Version
	SerializeMap() (map[string]interface{}, error)
	Signature() common.Signature
	Hint() string
	GenerateHash() (common.Hash, error)
	CheckSignature(common.NetworkID) error
	Stage() VoteStage
	Proposer() common.Address
	Block() common.Hash
	Proposal() common.Hash
	Validators() []common.Address
	Height() common.Big
	Round() Round
	Vote() Vote
}
