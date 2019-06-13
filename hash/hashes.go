package hash

import (
	"sync"
)

type Hashes struct {
	sync.RWMutex
	algorithms  map[ /*HashAlgorithmType*/ uint]HashAlgorithm
	defaultType HashAlgorithmType
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

func (h *Hashes) SetDefault(algorithmType HashAlgorithmType) error {
	algorithm, err := h.Algorithm(algorithmType)
	if err != nil {
		return err
	}

	h.Lock()
	defer h.Unlock()

	h.defaultType = algorithm.Type()

	return nil
}

func (h *Hashes) Algorithm(algorithmType HashAlgorithmType) (HashAlgorithm, error) {
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
		return Hash{}, err
	}

	body, err := algorithm.GenerateHash(b)
	if err != nil {
		return Hash{}, HashFailedError.New(err)
	}

	return NewHash(algorithm.Type(), hint, body)
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

	if err := algorithm.IsValid(hash.body); err != nil {
		return Hash{}, err
	}

	return hash, nil
}
