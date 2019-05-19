package leveldbstorage

import (
	"sync"

	"github.com/syndtr/goleveldb/leveldb"
	leveldbIterator "github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/opt"
	leveldbStorage "github.com/syndtr/goleveldb/leveldb/storage"
	leveldbUtil "github.com/syndtr/goleveldb/leveldb/util"

	"github.com/spikeekips/mitum/storage"
)

type LevelDBCore interface {
	Get([]byte, *opt.ReadOptions) ([]byte, error)
	Has([]byte, *opt.ReadOptions) (bool, error)
	Put([]byte, []byte, *opt.WriteOptions) error
	NewIterator(*leveldbUtil.Range, *opt.ReadOptions) leveldbIterator.Iterator
	Delete([]byte, *opt.WriteOptions) error
	Write(*leveldb.Batch, *opt.WriteOptions) error
}

type Config struct {
	Path string
}

type Storage struct {
	sync.RWMutex
	config      Config
	core        *leveldb.DB
	db          LevelDBCore
	transaction *Storage
}

func NewStorage(config Config) (*Storage, error) {
	s := &Storage{config: config}

	return s, s.open(true)
}

func OpenStorage(config Config) (*Storage, error) {
	s := &Storage{config: config}

	return s, s.open(false)
}

func (s *Storage) open(create bool) error {
	s.Lock()
	defer s.Unlock()

	if s.db != nil {
		return DBNotClosedError
	}

	var db *leveldb.DB
	var err error

	if len(s.config.Path) < 1 {
		db, err = leveldb.Open(leveldbStorage.NewMemStorage(), nil)
	} else {
		opts := &opt.Options{}
		if create {
			opts.ErrorIfExist = true
		} else {
			opts.ErrorIfMissing = true
		}
		db, err = leveldb.OpenFile(s.config.Path, opts)
	}

	if err != nil {
		return LevelDBError.SetError(err)
	}

	s.core = db
	s.db = db

	return nil
}

func (s *Storage) Close() error {
	s.Lock()
	defer s.Unlock()

	if s.db == nil {
		return nil
	}

	if err := s.core.Close(); err != nil {
		return LevelDBError.SetError(err)
	}

	s.db = nil

	return nil
}

func (s *Storage) Get(key []byte) ([]byte, error) {
	value, err := s.db.Get(key, nil)
	if err != nil {
		return nil, LevelDBError.SetError(err)
	}

	return value, nil
}

func (s *Storage) Exists(key []byte) (bool, error) {
	exists, err := s.db.Has(key, nil)
	if err != nil {
		return false, LevelDBError.SetError(err)
	}

	return exists, nil
}

func (s *Storage) Insert(key, value []byte) error {
	s.RLock()
	defer s.RUnlock()

	if s.transaction != nil {
		return TransactionAlreadyOpenedError
	}

	if found, err := s.Exists(key); err != nil {
		return LevelDBError.SetError(err)
	} else if found {
		return storage.RecordAlreadyExistsError.AppendMessage("key=%v", key)
	}

	if err := s.db.Put(key, value, nil); err != nil {
		return LevelDBError.SetError(err)
	}

	return nil
}

func (s *Storage) Update(key, value []byte) error {
	s.RLock()
	defer s.RUnlock()

	if s.transaction != nil {
		return TransactionAlreadyOpenedError
	}

	if found, err := s.Exists(key); err != nil {
		return LevelDBError.SetError(err)
	} else if !found {
		return storage.RecordNotFoundError.AppendMessage("key=%v", key)
	}

	if err := s.db.Put(key, value, nil); err != nil {
		return LevelDBError.SetError(err)
	}

	return nil
}

func (s *Storage) Delete(key []byte) error {
	s.RLock()
	defer s.RUnlock()

	if s.transaction != nil {
		return TransactionAlreadyOpenedError
	}

	if found, err := s.Exists(key); err != nil {
		return LevelDBError.SetError(err)
	} else if !found {
		return storage.RecordNotFoundError.AppendMessage("key=%v", key)
	}

	if err := s.db.Delete(key, nil); err != nil {
		return LevelDBError.SetError(err)
	}

	return nil
}

func (s *Storage) Iterator(prefix []byte, reverse bool, callback func([]byte, []byte) bool) error {
	var slice *leveldbUtil.Range
	if prefix != nil {
		slice = leveldbUtil.BytesPrefix(prefix)
	}

	iter := s.db.NewIterator(slice, nil)
	defer iter.Release()

	var next func() bool
	if reverse {
		if !iter.Last() {
			return nil // NOTE empty
		}

		next = func() bool {
			return iter.Prev()
		}
	} else {
		if !iter.First() {
			return nil // NOTE empty
		}
		next = iter.Next
	}

	if !iteratorCallback(iter, callback) {
		return nil
	}

	for next() {
		if !iteratorCallback(iter, callback) {
			break
		}
	}

	if err := iter.Error(); err != nil {
		return LevelDBError.SetError(err)
	}

	return nil
}

func (s *Storage) NewTransaction() (*Storage, error) {
	s.Lock()
	defer s.Unlock()

	if _, ok := s.db.(*leveldb.Transaction); ok {
		return nil, TransactionAlreadyOpenedError
	}

	if s.transaction != nil {
		return nil, TransactionAlreadyOpenedError
	}

	t, err := s.core.OpenTransaction()
	if err != nil {
		return nil, LevelDBError.SetError(err)
	}

	ts := &Storage{
		config: s.config,
		core:   s.core,
		db:     t,
	}
	s.transaction = ts

	return ts, nil
}

func (s *Storage) Transaction() *Storage {
	return s.transaction
}

func (s *Storage) closeTransaction() {
	s.Lock()
	defer s.Unlock()

	if s.transaction == nil {
		return
	}

	s.transaction = nil
}

func (s *Storage) Commit() error {
	t, ok := s.db.(*leveldb.Transaction)
	if !ok {
		return TransactionNotOpenedError
	}

	if err := t.Commit(); err != nil {
		return LevelDBError.SetError(err)
	}

	s.closeTransaction()

	return nil
}

func (s *Storage) Discard() error {
	t, ok := s.db.(*leveldb.Transaction)
	if !ok {
		return TransactionNotOpenedError
	}

	t.Discard()

	s.closeTransaction()

	return nil
}

func (s *Storage) Batch() storage.Batch {
	return &leveldb.Batch{}
}

func (s *Storage) WriteBatch(batch storage.Batch) error {
	b, ok := batch.(*leveldb.Batch)
	if !ok {
		return WrongBatchError.AppendMessage("type=%T", batch)
	}

	return s.db.Write(b, nil)
}

func iteratorCallback(iter leveldbIterator.Iterator, callback func([]byte, []byte) bool) bool {
	var key, value []byte
	{
		b := iter.Key()
		key = make([]byte, len(b))
		copy(key, b)
	}

	{
		b := iter.Value()
		value = make([]byte, len(b))
		copy(value, b)
	}

	return callback(key, value)
}
