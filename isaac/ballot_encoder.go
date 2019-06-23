package isaac

import (
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/spikeekips/mitum/common"
	"github.com/spikeekips/mitum/seal"
	"golang.org/x/xerrors"
)

type ballotEncoder struct {
	t common.DataType
}

func newBallotEncoder(t common.DataType) ballotEncoder {
	return ballotEncoder{t: t}
}

func (be ballotEncoder) Type() common.DataType {
	return be.t
}

func (be ballotEncoder) Encode(v interface{}) ([]byte, error) {
	return rlp.EncodeToBytes(v)
}

func (be ballotEncoder) Decode(b []byte) (interface{}, error) {
	var raw seal.RLPDecodeSeal
	if err := rlp.DecodeBytes(b, &raw); err != nil {
		return nil, err
	}

	var sl seal.Seal
	switch be.t {
	case INITBallotType:
		var body INITBallotBody
		if err := rlp.DecodeBytes(raw.Body, &body); err != nil {
			return nil, err
		}
		bsl := &seal.BaseSeal{}
		bsl = bsl.
			SetHash(raw.Hash).
			SetHeader(raw.Header).
			SetBody(body)

		sl = INITBallot{BaseSeal: *bsl, body: body}
	default:
		return nil, xerrors.Errorf("invalid ballot type")
	}

	return sl, nil
}
