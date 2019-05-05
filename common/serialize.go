package common

import "github.com/ethereum/go-ethereum/rlp"

type RLPSerializer interface {
	SerializeRLP() ([]interface{}, error)
}

type RLPUnserializer interface {
	UnserializeRLP([]rlp.RawValue) error
}

type JSONMapSerializer interface {
	SerializeMap() (map[string]interface{}, error)
}
