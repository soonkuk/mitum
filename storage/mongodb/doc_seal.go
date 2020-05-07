package mongodbstorage

import (
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/xerrors"

	"github.com/spikeekips/mitum/base/seal"
	"github.com/spikeekips/mitum/base/valuehash"
	"github.com/spikeekips/mitum/util/encoder"
	bsonencoder "github.com/spikeekips/mitum/util/encoder/bson"
	"github.com/spikeekips/mitum/util/localtime"
)

type SealDoc struct {
	BaseDoc
	seal seal.Seal
}

func NewSealDoc(sl seal.Seal, enc encoder.Encoder) (SealDoc, error) {
	b, err := NewBaseDoc(sl.Hash().String(), sl, enc)
	if err != nil {
		return SealDoc{}, err
	}

	return SealDoc{
		BaseDoc: b,
		seal:    sl,
	}, nil
}

func (sd SealDoc) MarshalBSON() ([]byte, error) {
	m, err := sd.BaseDoc.M()
	if err != nil {
		return nil, err
	}

	m["hash_string"] = sd.seal.Hash().String()
	m["hash"] = sd.seal.Hash()
	m["inserted_at"] = localtime.Now()

	return bsonencoder.Marshal(m)
}

func loadSealFromDecoder(decoder func(interface{}) error, encs *encoder.Encoders) (seal.Seal, error) {
	var b bson.Raw
	if err := decoder(&b); err != nil {
		return nil, err
	}

	var sl seal.Seal
	_, hinter, err := loadWithEncoder(b, encs)
	if err != nil {
		return nil, err
	} else if i, ok := hinter.(seal.Seal); !ok {
		return nil, xerrors.Errorf("not Seal: %T", hinter)
	} else {
		sl = i
	}

	return sl, nil
}

func loadSealHashFromDecoder(decoder func(interface{}) error, encs *encoder.Encoders) (valuehash.Hash, error) {
	var hd HashIDDoc
	if err := decoder(&hd); err != nil {
		return nil, err
	}

	enc, err := encs.Encoder(hd.E.Type(), hd.E.Version())
	if err != nil {
		return nil, err
	}

	var h valuehash.Hash
	if hinter, err := enc.DecodeByHint(hd.H); err != nil {
		return nil, err
	} else if i, ok := hinter.(valuehash.Hash); !ok {
		return nil, xerrors.Errorf("not valuehash.Hash: %T", hinter)
	} else {
		h = i
	}

	return h, nil
}
