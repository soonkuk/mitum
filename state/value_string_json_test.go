package state

import (
	"testing"

	"github.com/spikeekips/mitum/base/valuehash"
	"github.com/spikeekips/mitum/util"
	"github.com/spikeekips/mitum/util/encoder"
	"github.com/stretchr/testify/suite"
)

type testStateStringValueJSON struct {
	suite.Suite

	encs *encoder.Encoders
	enc  encoder.Encoder
}

func (t *testStateStringValueJSON) SetupSuite() {
	t.encs = encoder.NewEncoders()
	t.enc = encoder.NewJSONEncoder()
	_ = t.encs.AddEncoder(t.enc)

	_ = t.encs.AddHinter(valuehash.SHA256{})
	_ = t.encs.AddHinter(StringValue{})
}

func (t *testStateStringValueJSON) TestEncode() {
	sv, err := NewStringValue("showme")
	t.NoError(err)

	b, err := util.JSONMarshal(sv)
	t.NoError(err)

	decoded, err := t.enc.DecodeByHint(b)
	t.NoError(err)
	t.Implements((*Value)(nil), decoded)

	u := decoded.(Value)

	t.True(sv.Hint().Equal(u.Hint()))
	t.True(sv.Equal(u))
	t.Equal(sv.v, u.(StringValue).v)
}

func TestStateStringValueJSON(t *testing.T) {
	suite.Run(t, new(testStateStringValueJSON))
}
