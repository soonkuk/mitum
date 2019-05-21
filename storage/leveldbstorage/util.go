package leveldbstorage

import (
	leveldbUtil "github.com/syndtr/goleveldb/leveldb/util"
)

func PrefixSlice(prefix []byte) *leveldbUtil.Range {
	return leveldbUtil.BytesPrefix(prefix)
}
