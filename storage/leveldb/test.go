// +build test

package leveldbstorage

func NewMemStorage() *Storage {
	s, _ := NewStorage(Config{}) // open memory storage

	return s
}
