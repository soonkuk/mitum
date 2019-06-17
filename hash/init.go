package hash

var (
	DefaultHashes *Hashes
)

func init() {
	// register hashes
	DefaultHashes = NewHashes()
	DefaultHashes.Register(NewArgon2Hash())
	DefaultHashes.Register(NewDoubleSHA256Hash())
	DefaultHashes.SetDefault(Argon2HashType)
}
