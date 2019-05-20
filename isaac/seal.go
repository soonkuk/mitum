package isaac

import "github.com/spikeekips/mitum/common"

var (
	ProposalSealType     common.SealType = "proposal"
	TransactionSealType  common.SealType = "transaction"
	INITBallotSealType   common.SealType = "init-ballot"
	SIGNBallotSealType   common.SealType = "sign-ballot"
	ACCEPTBallotSealType common.SealType = "accept-ballot"
)

var sealCodec *common.SealCodec = common.NewSealCodec()

func init() {
	if err := sealCodec.Register(Proposal{}); err != nil {
		panic(err)
	}

	if err := sealCodec.Register(INITBallot{}); err != nil {
		panic(err)
	}

	if err := sealCodec.Register(SIGNBallot{}); err != nil {
		panic(err)
	}

	if err := sealCodec.Register(ACCEPTBallot{}); err != nil {
		panic(err)
	}

	// TODO register Transaction
}
