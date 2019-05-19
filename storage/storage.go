package storage

type Storage interface {
	Close() error
	Get(key []byte) ([]byte, error)
	Exists(key []byte) (bool, error)
	Insert(key []byte, value []byte) error
	Update(key []byte, value []byte) error
	Delete(key []byte) error
	Iterator(
		prefix []byte,
		reverse bool,
		callback func([]byte, []byte) bool,
	) error
	NewTransaction() error // NOTE deprecated; NewTransaction() available only one at a time
	Batch() Batch
	WriteBatch(Batch) error
}

type Batch interface {
	Len() int
	Put(key []byte, value []byte)
	Delete(key []byte)
}
