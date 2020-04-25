package leveldbstorage

import (
	"github.com/syndtr/goleveldb/leveldb"
	"golang.org/x/xerrors"

	"github.com/spikeekips/mitum/base/block"
	"github.com/spikeekips/mitum/base/operation"
	"github.com/spikeekips/mitum/base/state"
	"github.com/spikeekips/mitum/base/tree"
	"github.com/spikeekips/mitum/base/valuehash"
)

type LeveldbBlockStorage struct {
	st         *LeveldbStorage
	block      block.Block
	operations *tree.AVLTree
	states     *tree.AVLTree
	batch      *leveldb.Batch
}

func NewLeveldbBlockStorage(st *LeveldbStorage, blk block.Block) (*LeveldbBlockStorage, error) {
	bst := &LeveldbBlockStorage{
		st:    st,
		block: blk,
		batch: &leveldb.Batch{},
	}

	return bst, nil
}

func (bst *LeveldbBlockStorage) Block() block.Block {
	return bst.block
}

func (bst *LeveldbBlockStorage) SetBlock(blk block.Block) error {
	if bst.block.Height() != blk.Height() {
		return xerrors.Errorf(
			"block has different height from initial block; initial=%d != block=%d",
			bst.block.Height(),
			blk.Height(),
		)
	}

	if bst.block.Round() != blk.Round() {
		return xerrors.Errorf(
			"block has different round from initial block; initial=%d != block=%d",
			bst.block.Round(),
			blk.Round(),
		)
	}

	if b, err := LeveldbMarshal(bst.st.enc, blk); err != nil {
		return err
	} else {
		bst.batch.Put(leveldbBlockHashKey(blk.Hash()), b)
	}

	if b, err := LeveldbMarshal(bst.st.enc, blk.Manifest()); err != nil {
		return err
	} else {
		key := leveldbManifestKey(blk.Hash())
		bst.batch.Put(key, b)
	}

	if b, err := LeveldbMarshal(bst.st.enc, blk.Hash()); err != nil {
		return err
	} else {
		bst.batch.Put(leveldbBlockHeightKey(blk.Height()), b)
	}

	if err := bst.setOperations(blk.Operations()); err != nil {
		return err
	}

	if err := bst.setStates(blk.States()); err != nil {
		return err
	}

	bst.block = blk

	return nil
}

func (bst *LeveldbBlockStorage) setOperations(tr *tree.AVLTree) error {
	if tr == nil || tr.Empty() {
		return nil
	}

	if b, err := LeveldbMarshal(bst.st.enc, tr); err != nil { // block 1st
		return err
	} else {
		bst.batch.Put(leveldbBlockOperationsKey(bst.block), b)
	}

	// store operation hashes
	if err := tr.Traverse(func(node tree.Node) (bool, error) {
		op := node.Immutable().(operation.OperationAVLNode).Operation()

		raw, err := bst.st.enc.Encode(op.Hash())
		if err != nil {
			return false, err
		}

		bst.batch.Put(
			leveldbOperationHashKey(op.Hash()),
			LeveldbDataWithEncoder(bst.st.enc, raw),
		)

		return true, nil
	}); err != nil {
		return err
	}

	bst.operations = tr

	return nil
}

func (bst *LeveldbBlockStorage) setStates(tr *tree.AVLTree) error {
	if tr == nil || tr.Empty() {
		return nil
	}

	if b, err := LeveldbMarshal(bst.st.enc, tr); err != nil { // block 1st
		return err
	} else {
		bst.batch.Put(leveldbBlockStatesKey(bst.block), b)
	}

	if err := tr.Traverse(func(node tree.Node) (bool, error) {
		var st state.State
		if s, ok := node.Immutable().(state.StateV0AVLNode); !ok {
			return false, xerrors.Errorf("not state.StateV0AVLNode: %T", node)
		} else {
			st = s.State()
		}

		if b, err := LeveldbMarshal(bst.st.enc, st); err != nil {
			return false, err
		} else {
			bst.batch.Put(leveldbStateKey(st.Key()), b)
		}

		return true, nil
	}); err != nil {
		return err
	}

	bst.states = tr

	return nil
}

func (bst *LeveldbBlockStorage) UnstageOperationSeals(hs []valuehash.Hash) error {
	return leveldbUnstageOperationSeals(bst.st, bst.batch, hs)
}

func (bst *LeveldbBlockStorage) Commit() error {
	return LeveldbWrapError(bst.st.db.Write(bst.batch, nil))
}