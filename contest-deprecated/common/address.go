package common

import (
	"encoding/json"
	"fmt"

	"golang.org/x/xerrors"

	"github.com/spikeekips/mitum/base"
	jsonencoder "github.com/spikeekips/mitum/util/encoder/json"
	"github.com/spikeekips/mitum/util/hint"
	"github.com/spikeekips/mitum/util/logging"
)

var (
	ContestAddressType = hint.MustNewType(0xd0, 0x00, "contest-address")
	ContestAddressHint = hint.MustHint(ContestAddressType, "0.0.1")
)

type ContestAddress string

func NewContestAddress(id int) ContestAddress {
	return ContestAddress(fmt.Sprintf("node:%02d", id))
}

func (sa ContestAddress) String() string {
	return string(sa)
}

func (sa ContestAddress) Hint() hint.Hint {
	return ContestAddressHint
}

func (sa ContestAddress) IsValid([]byte) error {
	if len(sa) < 1 {
		return xerrors.Errorf("empty address")
	}

	return nil
}

func (sa ContestAddress) Equal(a base.Address) bool {
	if sa.Hint().Type() != a.Hint().Type() {
		return false
	}

	return sa == a.(ContestAddress)
}

func (sa ContestAddress) Bytes() []byte {
	return []byte(sa)
}

func (sa ContestAddress) MarshalJSON() ([]byte, error) {
	return jsonencoder.Marshal(struct {
		jsonencoder.HintedHead
		A string `json:"address"`
	}{
		HintedHead: jsonencoder.NewHintedHead(sa.Hint()),
		A:          sa.String(),
	})
}

func (sa *ContestAddress) UnpackJSON(b []byte, _ *jsonencoder.Encoder) error {
	var s struct {
		jsonencoder.HintedHead
		A string `json:"address"`
	}
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	} else if err := sa.Hint().IsCompatible(s.H); err != nil {
		return err
	} else if len(s.A) < 5 {
		return xerrors.Errorf("not enough address")
	}

	*sa = ContestAddress(s.A)

	return nil
}

func (sa ContestAddress) MarshalLog(key string, e logging.Emitter, verbose bool) logging.Emitter {
	if !verbose {
		return e.Str(key, sa.String())
	}

	return e.Dict(key, logging.Dict().
		Str("address", sa.String()).
		HintedVerbose("hint", sa.Hint(), true),
	)
}