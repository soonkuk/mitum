package hash

import (
	"sync"

	"github.com/spikeekips/mitum/common"
)

type Hashes struct {
	sync.RWMutex
	algorithms  map[ /*common.DataType*/ uint]HashAlgorithm
	defaultType common.DataType
}

func NewHashes() *Hashes {
	return &Hashes{
		algorithms: map[uint]HashAlgorithm{},
	}
}

func (h *Hashes) Register(algorithm HashAlgorithm) error {
	h.Lock()
	defer h.Unlock()

	// check duplication
	for t := range h.algorithms {
		if t == algorithm.Type().ID() {
			return HashAlgorithmAlreadyRegisteredError.Newf("type=%q", algorithm.Type().String())
		}
	}

	h.algorithms[algorithm.Type().ID()] = algorithm

	if h.defaultType.Empty() {
		h.defaultType = algorithm.Type()
	}

	return nil
}

func (h *Hashes) SetDefault(algorithmType common.DataType) error {
	algorithm, err := h.Algorithm(algorithmType)
	if err != nil {
		return err
	}

	h.Lock()
	defer h.Unlock()

	h.defaultType = algorithm.Type()

	return nil
}

func (h *Hashes) Algorithm(algorithmType common.DataType) (HashAlgorithm, error) {
	h.RLock()
	defer h.RUnlock()

	algorithm, found := h.algorithms[algorithmType.ID()]
	if !found {
		return nil, HashAlgorithmNotRegisteredError.Newf("type=%q", algorithmType.String())
	}

	return algorithm, nil
}

func (h *Hashes) NewHash(hint string, b []byte) (Hash, error) {
	algorithm, err := h.Algorithm(h.defaultType)
	if err != nil {
		panic(HashAlgorithmNotRegisteredError.Newf("type=%q", h.defaultType.String()))
	}

	return NewHash(algorithm.Type(), hint, algorithm.GenerateHash(b))
}

func (h *Hashes) NewHashByType(algorithmType common.DataType, hint string, b []byte) (Hash, error) {
	algorithm, err := h.Algorithm(algorithmType)
	if err != nil {
		panic(HashAlgorithmNotRegisteredError.Newf("type=%q", algorithmType.String()))
	}

	return NewHash(algorithm.Type(), hint, algorithm.GenerateHash(b))
}

func (h *Hashes) UnmarshalHash(b []byte) (Hash, error) {
	var hash Hash
	if err := hash.UnmarshalBinary(b); err != nil {
		return Hash{}, err
	}

	algorithm, err := h.Algorithm(hash.Algorithm())
	if err != nil {
		return Hash{}, err
	}

	hash.algorithm = algorithm.Type()

	if err := algorithm.IsValid(hash.Body()); err != nil {
		return Hash{}, err
	}

	return hash, nil
}

func (h *Hashes) Merge(base Hash, hashes ...Hash) (Hash, error) {
	if err := base.IsValid(); err != nil {
		return Hash{}, err
	}

	if len(hashes) < 1 {
		return base, nil
	}

	algorithm, err := h.Algorithm(base.Algorithm())
	if err != nil {
		return Hash{}, err
	}

	for _, i := range hashes {
		if err := i.IsValid(); err != nil {
			return Hash{}, err
		}
	}

	var body []byte
	for _, i := range append([]Hash{base}, hashes...) {
		a, err := i.MarshalBinary()
		if err != nil {
			return Hash{}, err
		}
		body = append(body, a...)
	}

	return NewHash(base.Algorithm(), base.Hint(), algorithm.GenerateHash(body))
}
