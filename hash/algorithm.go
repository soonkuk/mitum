package hash

type HashAlgorithm interface {
	Type() HashAlgorithmType
	GenerateHash([]byte) ([]byte, error)
	IsValid([]byte) error
}
