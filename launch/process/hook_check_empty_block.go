package process

import (
	"context"

	"github.com/spikeekips/mitum/isaac"
	"github.com/spikeekips/mitum/launch/config"
	"github.com/spikeekips/mitum/storage"
	"github.com/spikeekips/mitum/util/logging"
	"golang.org/x/xerrors"
)

const HookNameCheckEmptyBlock = "check_empty_block"

// HookCheckEmptyBlock checks whether local has empty block. If empty block and
// there are no other nodes, stop process with error.
func HookCheckEmptyBlock(ctx context.Context) (context.Context, error) {
	var local *isaac.Local
	if err := LoadLocalContextValue(ctx, &local); err != nil {
		return ctx, err
	}

	var st storage.Storage
	if err := LoadStorageContextValue(ctx, &st); err != nil {
		return ctx, err
	}

	var blockFS *storage.BlockFS
	if err := LoadBlockFSContextValue(ctx, &blockFS); err != nil {
		return ctx, err
	}

	var log logging.Logger
	if err := config.LoadLogContextValue(ctx, &log); err != nil {
		return ctx, err
	}

	if blk, err := storage.CheckBlockEmpty(st, blockFS); err != nil {
		return ctx, err
	} else if blk == nil {
		log.Debug().Msg("empty block found; storage will be empty")

		if err := storage.Clean(st, blockFS, false); err != nil {
			return nil, err
		}

		if local.Nodes().Len() < 1 {
			return ctx, xerrors.Errorf("empty block, but no other nodes; can not sync")
		}
	} else if err := blk.IsValid(local.Policy().NetworkID()); err != nil {
		return ctx, xerrors.Errorf("invalid block found, clean up block: %w", err)
	} else {
		log.Debug().Hinted("block", blk.Manifest()).Msg("valid initial block found")
	}

	return ctx, nil
}
