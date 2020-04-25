package mongodbstorage

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/xerrors"

	"github.com/spikeekips/mitum/base/block"
	"github.com/spikeekips/mitum/base/operation"
	"github.com/spikeekips/mitum/base/state"
	"github.com/spikeekips/mitum/base/tree"
	"github.com/spikeekips/mitum/base/valuehash"
	"github.com/spikeekips/mitum/storage"
)

type MongodbBlockStorage struct {
	st                  *MongodbStorage
	block               block.Block
	operations          *tree.AVLTree
	states              *tree.AVLTree
	blockModels         []mongo.WriteModel
	manifestModels      []mongo.WriteModel
	operationSealModels []mongo.WriteModel
	operationModels     []mongo.WriteModel
	stateModels         []mongo.WriteModel
}

func NewMongodbBlockStorage(st *MongodbStorage, blk block.Block) (*MongodbBlockStorage, error) {
	bst := &MongodbBlockStorage{
		st:    st,
		block: blk,
	}

	return bst, nil
}

func (bst *MongodbBlockStorage) Block() block.Block {
	return bst.block
}

func (bst *MongodbBlockStorage) SetBlock(blk block.Block) error {
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

	enc := bst.st.enc

	if doc, err := NewBlockDoc(blk, enc); err != nil {
		return err
	} else {
		bst.blockModels = append(bst.blockModels, mongo.NewInsertOneModel().SetDocument(doc))
	}

	if doc, err := NewManifestDoc(blk.Manifest(), enc); err != nil {
		return err
	} else {
		bst.manifestModels = append(bst.manifestModels, mongo.NewInsertOneModel().SetDocument(doc))
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

func (bst *MongodbBlockStorage) UnstageOperationSeals(seals []valuehash.Hash) error {
	for _, h := range seals {
		bst.operationSealModels = append(bst.operationSealModels,
			mongo.NewDeleteOneModel().SetFilter(NewFilter("_id", h.String()).D()),
		)
	}

	return nil
}

func (bst *MongodbBlockStorage) Commit() error {
	if res, err := bst.writeModels("block", bst.blockModels); err != nil {
		return storage.WrapError(err)
	} else if res != nil && res.InsertedCount < 1 {
		return xerrors.Errorf("block not inserted")
	}

	if res, err := bst.writeModels("manifest", bst.manifestModels); err != nil {
		return storage.WrapError(err)
	} else if res != nil && res.InsertedCount < 1 {
		return xerrors.Errorf("manifest not inserted")
	}

	if res, err := bst.writeModels("operation", bst.operationModels); err != nil {
		return storage.WrapError(err)
	} else if res != nil && res.InsertedCount < 1 {
		return xerrors.Errorf("operation not inserted")
	}

	if res, err := bst.writeModels("state", bst.stateModels); err != nil {
		return storage.WrapError(err)
	} else if res != nil && res.InsertedCount < 1 {
		return xerrors.Errorf("state not inserted")
	}

	if _, err := bst.writeModels("operation_seal", bst.operationSealModels); err != nil {
		return storage.WrapError(err)
	}

	return nil
}

func (bst *MongodbBlockStorage) setOperations(tr *tree.AVLTree) error {
	if tr == nil || tr.Empty() {
		return nil
	}

	var models []mongo.WriteModel
	if err := tr.Traverse(func(node tree.Node) (bool, error) {
		op := node.Immutable().(operation.OperationAVLNode).Operation()

		doc, err := NewOperationDoc(op, bst.st.enc)
		if err != nil {
			return false, err
		}
		models = append(models, mongo.NewInsertOneModel().SetDocument(doc))

		return true, nil
	}); err != nil {
		return err
	}

	bst.operationModels = models
	bst.operations = tr

	return nil
}

func (bst *MongodbBlockStorage) setStates(tr *tree.AVLTree) error {
	if tr == nil || tr.Empty() {
		return nil
	}

	var models []mongo.WriteModel
	if err := tr.Traverse(func(node tree.Node) (bool, error) {
		st := node.Immutable().(state.StateV0AVLNode).State()

		doc, err := NewStateDoc(st, bst.st.enc)
		if err != nil {
			return false, err
		}
		models = append(models,
			// NOTE state is managed by it's Key()
			mongo.NewDeleteOneModel().SetFilter(NewFilter("_id", doc.ID()).D()),
			mongo.NewInsertOneModel().SetDocument(doc),
		)

		return true, nil
	}); err != nil {
		return err
	}

	bst.stateModels = models
	bst.states = tr

	return nil
}

func (bst *MongodbBlockStorage) writeModels(col string, models []mongo.WriteModel) (*mongo.BulkWriteResult, error) {
	if len(models) < 1 {
		return nil, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	opts := options.BulkWrite().SetOrdered(true)
	res, err := bst.st.client.Collection(col).BulkWrite(ctx, models, opts)
	if err != nil {
		return nil, storage.WrapError(err)
	}

	return res, nil
}