package leveldbstorage

import (
	"bytes"
	"io/ioutil"
	"os"
	"sort"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/syndtr/goleveldb/leveldb"

	"github.com/spikeekips/mitum/common"
	"github.com/spikeekips/mitum/storage"
)

type testLevelDB struct {
	suite.Suite
	s *Storage
}

func (t *testLevelDB) SetupTest() {
	t.s, _ = NewStorage(Config{}) // open memory storage
}

func (t *testLevelDB) TearDownTest() {
	t.s.Close()
}

func (t *testLevelDB) insertData(s storage.Storage, prefix []byte, count int) [][][]byte {
	if s == nil {
		s = t.s
	}

	var data [][][]byte
	for i := 0; i < count; i++ {
		key := append(prefix, []byte(common.RandomUUID())...)
		value := []byte(common.RandomUUID())
		data = append(data, [][]byte{key, value})

		err := s.Insert(key, value)
		t.NoError(err)
	}

	return data
}

func (t *testLevelDB) TestOpen() {
	{ // memory storage
		s, err := NewStorage(Config{Path: ""})
		t.NoError(err)
		s.Close()
	}

	{ // create new storage
		path, _ := ioutil.TempDir("", "leveldbstorage")
		s, err := NewStorage(Config{Path: path})
		t.NoError(err)
		s.Close()
	}

	{ // open storage
		path, _ := ioutil.TempDir("", "leveldbstorage")
		s, err := OpenStorage(Config{Path: path})
		t.NotNil(err)
		t.Contains(err.Error(), "file does not exist")
		s.Close()
	}

	{ // create, but already exists
		path, _ := ioutil.TempDir("", "leveldbstorage")
		defer os.RemoveAll(path)
		s, err := NewStorage(Config{Path: path}) // create new one
		t.NoError(err)
		err = s.Close()
		t.NoError(err)

		// open again
		s, err = NewStorage(Config{Path: path})
		t.NotNil(err)
		t.Contains(err.Error(), "file already exists")
	}
}

func (t *testLevelDB) TestInsert() {
	key := []byte("showme")
	value := []byte("killme")

	err := t.s.Insert(key, value)
	t.NoError(err)

	returned, err := t.s.Get(key)
	t.NoError(err)
	t.Equal(value, returned)

	// same key
	err = t.s.Insert(key, []byte("findme"))
	t.True(storage.RecordAlreadyExistsError.Equal(err))
}

func (t *testLevelDB) TestUpdate() {
	key := []byte("showme")
	value := []byte("killme")

	err := t.s.Update(key, value)
	t.True(storage.RecordNotFoundError.Equal(err))

	// same key
	err = t.s.Insert(key, value)
	t.NoError(err)

	// update
	newValue := []byte("findme")
	err = t.s.Update(key, newValue)
	t.NoError(err)

	returned, err := t.s.Get(key)
	t.NoError(err)
	t.Equal(newValue, returned)
}

func (t *testLevelDB) TestExists() {
	key := []byte("showme")
	value := []byte("killme")

	err := t.s.Insert(key, value)
	t.NoError(err)

	exists, err := t.s.Exists(key)
	t.NoError(err)
	t.True(exists)

	exists, err = t.s.Exists([]byte("unknown"))
	t.NoError(err)
	t.False(exists)
}

func (t *testLevelDB) TestDelete() {
	key := []byte("showme")
	value := []byte("killme")

	err := t.s.Delete(key)
	t.True(storage.RecordNotFoundError.Equal(err))

	err = t.s.Insert(key, value)
	t.NoError(err)

	exists, err := t.s.Exists(key)
	t.NoError(err)
	t.True(exists)

	err = t.s.Delete(key)
	t.NoError(err)

	exists, err = t.s.Exists(key)
	t.NoError(err)
	t.False(exists)
}

func (t *testLevelDB) TestIterator() {
	data := t.insertData(nil, nil, 10)
	sort.SliceStable(data, func(i, j int) bool {
		return bytes.Compare(data[i][0], data[j][0]) < 0
	})

	var returned [][][]byte

	// iterate all
	err := t.s.Iterator(nil, false, func(key, value []byte) bool {
		returned = append(returned, [][]byte{key, value})
		return true
	})
	t.NoError(err)

	t.Equal(len(data), len(returned))

	t.Equal(data, returned)
}

func (t *testLevelDB) TestIteratorReverse() {
	data := t.insertData(nil, nil, 10)
	sort.SliceStable(data, func(i, j int) bool {
		return bytes.Compare(data[i][0], data[j][0]) > 0 // reverse
	})

	var returned [][][]byte

	// iterate all
	err := t.s.Iterator(nil, true /* reverse */, func(key, value []byte) bool {
		returned = append(returned, [][]byte{key, value})
		return true
	})
	t.NoError(err)

	t.Equal(len(data), len(returned))

	t.Equal(data, returned)
}

func (t *testLevelDB) TestIteratorLimit() {
	data := t.insertData(nil, nil, 10)
	sort.SliceStable(data, func(i, j int) bool {
		return bytes.Compare(data[i][0], data[j][0]) < 0
	})

	var returned [][][]byte

	// iterate all
	limit := 3
	err := t.s.Iterator(nil, false, func(key, value []byte) bool {
		returned = append(returned, [][]byte{key, value})

		if len(returned) == limit {
			return false
		}

		return true
	})
	t.NoError(err)

	t.Equal(limit, len(returned))

	t.Equal(data[:limit], returned)
}

func (t *testLevelDB) TestIteratorReverseLimit() {
	data := t.insertData(nil, nil, 10)
	sort.SliceStable(data, func(i, j int) bool {
		return bytes.Compare(data[i][0], data[j][0]) > 0
	})

	var returned [][][]byte

	// iterate all
	limit := 3
	err := t.s.Iterator(nil, true, func(key, value []byte) bool {
		returned = append(returned, [][]byte{key, value})
		if len(returned) == limit {
			return false
		}

		return true
	})
	t.NoError(err)

	t.Equal(limit, len(returned))

	t.Equal(data[:limit], returned)
}

func (t *testLevelDB) TestIteratorSlice() {
	prefix0 := []byte("ahowme")
	prefix1 := []byte("billme")
	data0 := t.insertData(nil, prefix0, 5)
	data1 := t.insertData(nil, prefix1, 5)

	{
		// iterate only prefix0
		sort.SliceStable(data0, func(i, j int) bool {
			return bytes.Compare(data0[i][0], data0[j][0]) < 0
		})
		var returned [][][]byte
		err := t.s.Iterator(prefix0, false, func(key, value []byte) bool {
			returned = append(returned, [][]byte{key, value})
			return true
		})
		t.NoError(err)

		t.Equal(len(data0), len(returned))

		t.Equal(data0, returned)
	}

	{
		// iterate only prefix0 with reverse
		sort.SliceStable(data0, func(i, j int) bool {
			return bytes.Compare(data0[i][0], data0[j][0]) > 0
		})
		var returned [][][]byte
		err := t.s.Iterator(prefix0, true, func(key, value []byte) bool {
			returned = append(returned, [][]byte{key, value})
			return true
		})
		t.NoError(err)

		t.Equal(len(data0), len(returned))

		t.Equal(data0, returned)
	}

	{
		// iterate only prefix1
		sort.SliceStable(data1, func(i, j int) bool {
			return bytes.Compare(data1[i][0], data1[j][0]) < 0
		})
		var returned [][][]byte
		err := t.s.Iterator(prefix1, false, func(key, value []byte) bool {
			returned = append(returned, [][]byte{key, value})
			return true
		})
		t.NoError(err)

		t.Equal(len(data1), len(returned))
		t.Equal(data1, returned)
	}

	{
		// iterate only prefix1 with reverse
		sort.SliceStable(data1, func(i, j int) bool {
			return bytes.Compare(data1[i][0], data1[j][0]) > 0
		})
		var returned [][][]byte
		err := t.s.Iterator(prefix1, true, func(key, value []byte) bool {
			returned = append(returned, [][]byte{key, value})
			return true
		})
		t.NoError(err)

		t.Equal(len(data1), len(returned))
		t.Equal(data1, returned)
	}
}

func (t *testLevelDB) TestNewTransaction() {
	{ // new
		ts, err := t.s.NewTransaction()
		t.NoError(err)
		t.NotNil(ts)
	}

	{ // open again
		_, err := t.s.NewTransaction()
		t.Contains(err.Error(), "transaction already opened")
	}

	{ // open by ts
		ts := t.s.Transaction()
		t.NotNil(ts)

		_, err := ts.NewTransaction()
		t.Contains(err.Error(), "transaction already opened")
	}
}

func (t *testLevelDB) TestTransactionInsert() {
	inserted := t.insertData(nil, nil, 10)
	sort.SliceStable(inserted, func(i, j int) bool {
		return bytes.Compare(inserted[i][0], inserted[j][0]) < 0
	})

	ts, err := t.s.NewTransaction()
	t.NoError(err)

	insertedTransaction := t.insertData(ts, nil, 10) // insert 10 thru transaction

	{ // before Commit(), :)
		var returned [][][]byte
		err = t.s.Iterator(nil, false, func(key, value []byte) bool {
			returned = append(returned, [][]byte{key, value})
			return true
		})
		t.NoError(err)
		t.Equal(inserted, returned)
	}

	{ // after Commit(), :)
		err = ts.(*Storage).Commit()
		t.NoError(err)

		var returned [][][]byte
		err = t.s.Iterator(nil, false, func(key, value []byte) bool {
			returned = append(returned, [][]byte{key, value})
			return true
		})

		data := append(inserted, insertedTransaction...)
		sort.SliceStable(data, func(i, j int) bool {
			return bytes.Compare(data[i][0], data[j][0]) < 0
		})

		t.NoError(err)
		t.Equal(data, returned)
	}
}

func (t *testLevelDB) TestTransactionDiscard() {
	inserted := t.insertData(nil, nil, 10)
	sort.SliceStable(inserted, func(i, j int) bool {
		return bytes.Compare(inserted[i][0], inserted[j][0]) < 0
	})

	ts, err := t.s.NewTransaction()
	t.NoError(err)

	_ = t.insertData(ts, nil, 10) // insert 10 thru transaction

	{ // before Commit(), :)
		var returned [][][]byte
		err = t.s.Iterator(nil, false, func(key, value []byte) bool {
			returned = append(returned, [][]byte{key, value})
			return true
		})
		t.NoError(err)
		t.Equal(inserted, returned)
	}

	{ // after Discard(), :)
		err = ts.(*Storage).Discard()
		t.NoError(err)

		var returned [][][]byte
		err = t.s.Iterator(nil, false, func(key, value []byte) bool {
			returned = append(returned, [][]byte{key, value})
			return true
		})

		t.NoError(err)
		t.Equal(inserted, returned)
	}
}

func (t *testLevelDB) TestTransactionInsertBeforeCommit() {
	_ = t.insertData(nil, nil, 10)

	ts, err := t.s.NewTransaction()
	t.NoError(err)

	_ = t.insertData(ts, nil, 10) // insert 10 thru transaction

	err = t.s.Insert([]byte("a"), []byte("b"))
	t.Contains(err.Error(), "transaction already opened")
}

func (t *testLevelDB) TestBatch() {
	b := t.s.Batch()
	t.IsType(&leveldb.Batch{}, b)
}

func (t *testLevelDB) TestWriteBatch() {
	b := t.s.Batch()

	var data [][][]byte
	for i := 0; i < 10; i++ {
		key := []byte(common.RandomUUID())
		value := []byte(common.RandomUUID())
		data = append(data, [][]byte{key, value})

		b.Put(key, value)
	}

	err := t.s.WriteBatch(b)
	t.NoError(err)

	{
		var returned [][][]byte
		err = t.s.Iterator(nil, false, func(key, value []byte) bool {
			returned = append(returned, [][]byte{key, value})
			return true
		})

		sort.SliceStable(data, func(i, j int) bool {
			return bytes.Compare(data[i][0], data[j][0]) < 0
		})

		t.NoError(err)
		t.Equal(data, returned)
	}
}

func (t *testLevelDB) TestWriteBatchDeleteUnknown() {
	b := t.s.Batch()

	var data [][][]byte
	for i := 0; i < 10; i++ {
		key := []byte(common.RandomUUID())
		value := []byte(common.RandomUUID())
		data = append(data, [][]byte{key, value})

		b.Put(key, value)
	}
	b.Delete([]byte("unknown")) // will be ignored

	err := t.s.WriteBatch(b)
	t.NoError(err)

	{
		var returned [][][]byte
		err = t.s.Iterator(nil, false, func(key, value []byte) bool {
			returned = append(returned, [][]byte{key, value})
			return true
		})

		sort.SliceStable(data, func(i, j int) bool {
			return bytes.Compare(data[i][0], data[j][0]) < 0
		})

		t.NoError(err)
		t.Equal(data, returned)
	}
}

func TestLevelDB(t *testing.T) {
	suite.Run(t, new(testLevelDB))
}
